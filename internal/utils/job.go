/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"context"
	"fmt"
	"github.com/apecloud/kubebench/api/v1alpha1"
	"github.com/apecloud/kubebench/pkg/constants"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func IsJobExisted(cli client.Client, reqCtx context.Context, jobName string, namespace string) (bool, error) {
	var job batchv1.Job
	if err := cli.Get(reqCtx, client.ObjectKey{Name: jobName, Namespace: namespace}, &job); err != nil {
		return false, client.IgnoreNotFound(err)
	}

	return true, nil
}

func GetJobStatus(cli client.Client, reqCtx context.Context, jobName string, namespace string) (*batchv1.JobStatus, error) {
	var job batchv1.Job
	if err := cli.Get(reqCtx, client.ObjectKey{Name: jobName, Namespace: namespace}, &job); err != nil {
		return nil, err
	}

	return &job.Status, nil
}

func IsJobSuccessful(cli client.Client, reqCtx context.Context, jobName string, namespace string) (bool, error) {
	var job batchv1.Job
	if err := cli.Get(reqCtx, client.ObjectKey{Name: jobName, Namespace: namespace}, &job); err != nil {
		return false, err
	}

	if job.Status.Succeeded == 1 {
		return true, nil
	}
	return false, nil
}

func GetPodListFromJob(cli client.Client, reqCtx context.Context, jobName string, namespace string) (*corev1.PodList, error) {
	var pods corev1.PodList
	if err := cli.List(reqCtx, &pods,
		client.InNamespace(namespace),
		client.MatchingLabels{
			"job-name": jobName,
		}); err != nil {
		return nil, err
	}

	return &pods, nil
}

func DeleteJob(cli client.Client, reqCtx context.Context, jobName string, namespace string) error {
	var job batchv1.Job
	if err := cli.Get(reqCtx, client.ObjectKey{Name: jobName, Namespace: namespace}, &job); err != nil {
		return err
	}

	// delete the job
	deletetions := []client.DeleteOption{
		client.PropagationPolicy(metav1.DeletePropagationBackground),
	}
	if err := cli.Delete(reqCtx, &job, deletetions...); err != nil {
		return err
	}

	return nil
}

// LogJobPodToCond record the log of job's pods to the conditions
func LogJobPodToCond(cli client.Client, restConfig *rest.Config, reqCtx context.Context, jobName string, namespace string, conditions *[]metav1.Condition, call func(string) string) error {
	l := log.FromContext(reqCtx)

	podList, err := GetPodListFromJob(cli, reqCtx, jobName, namespace)
	if err != nil {
		l.Error(err, "failed to get pod list from job", "job", jobName, "namespace", namespace)
		return err
	}

	for _, pod := range podList.Items {
		// get the log of the pod
		msg, err := GetLogFromPod(restConfig, reqCtx, pod.Name, namespace)
		if err != nil {
			l.Error(err, "failed to get log from pod", "pod", pod.Name, "namespace", namespace)
		}

		if call != nil {
			msg = call(msg)
		}

		msg = trimTooLongLog(msg)

		switch pod.Status.Phase {
		case corev1.PodSucceeded:
			RecordSuccessfulLogToCond(conditions, msg)
		case corev1.PodFailed:
			RecordFailedLogToCond(conditions, msg)
		}
	}

	return nil
}

func RecordSuccessfulLogToCond(conditions *[]metav1.Condition, msg string) {
	meta.SetStatusCondition(conditions, metav1.Condition{
		Type:               "Successful",
		Status:             metav1.ConditionTrue,
		Reason:             "RecordSuccessfulLog",
		Message:            msg,
		LastTransitionTime: metav1.Now(),
	})
}

func RecordFailedLogToCond(conditions *[]metav1.Condition, msg string) {
	meta.SetStatusCondition(conditions, metav1.Condition{
		Type:               "Failed",
		Status:             metav1.ConditionFalse,
		Reason:             "RecordFailedLog",
		Message:            msg,
		LastTransitionTime: metav1.Now(),
	})
}

func trimTooLongLog(log string) string {
	// spit the log by line
	lines := strings.Split(log, "\n")
	reuslt := ""
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if len(reuslt)+len(line) > 32768 {
			// delete from the start
			reuslt = reuslt[len(line)+1:]
		}
		reuslt += line + "\n"
	}
	// remove the last "\n"
	return strings.TrimSpace(reuslt)
}

func JobTemplate(name, namespace string) *batchv1.Job {
	backoffLimit := int32(0) // no retry

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &backoffLimit,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/managed-by": "kubebench",
					},
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "log",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
		},
	}

	return job
}

func AddTolerationToJobs(jobs []*batchv1.Job, tolerations []corev1.Toleration) {
	for _, job := range jobs {
		job.Spec.Template.Spec.Tolerations = append(job.Spec.Template.Spec.Tolerations, tolerations...)
	}
}

func AddResourceLimitsToJobs(jobs []*batchv1.Job, limits *v1alpha1.ResourceList) {
	for _, job := range jobs {
		addResourceLimitsToJob(job, limits)
	}
}

func addResourceLimitsToJob(job *batchv1.Job, limits *v1alpha1.ResourceList) {
	if job == nil || limits == nil {
		return
	}

	// validate the limits parameters
	if _, err := resource.ParseQuantity(limits.Cpu); err != nil {
		return
	}
	if _, err := resource.ParseQuantity(limits.Memory); err != nil {
		return
	}

	for i, container := range job.Spec.Template.Spec.Containers {
		container.Resources.Limits = corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(limits.Cpu),
			corev1.ResourceMemory: resource.MustParse(limits.Memory),
		}
		job.Spec.Template.Spec.Containers[i] = container
	}
}

func AddResourceRequestsToJobs(jobs []*batchv1.Job, requests *v1alpha1.ResourceList) {
	for _, job := range jobs {
		addResourceRequestsToJob(job, requests)
	}
}

func addResourceRequestsToJob(job *batchv1.Job, requests *v1alpha1.ResourceList) {
	if job == nil || requests == nil {
		return
	}

	// validate the requests parameters
	if _, err := resource.ParseQuantity(requests.Cpu); err != nil {
		return
	}
	if _, err := resource.ParseQuantity(requests.Memory); err != nil {
		return
	}

	for i, container := range job.Spec.Template.Spec.Containers {
		container.Resources.Requests = corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(requests.Cpu),
			corev1.ResourceMemory: resource.MustParse(requests.Memory),
		}
		job.Spec.Template.Spec.Containers[i] = container
	}
}

func AddLabelsToJobs(jobs []*batchv1.Job, labels map[string]string) {
	for _, job := range jobs {
		// add label to the job and pod template
		if job.Labels == nil {
			job.Labels = make(map[string]string)
		}
		if job.Spec.Template.Labels == nil {
			job.Spec.Template.Labels = make(map[string]string)
		}

		for k, v := range labels {
			job.Labels[k] = v
			job.Spec.Template.Labels[k] = v
		}
	}
}

// NewMysqlPreCheckJob create a job to check the mysql connection
func NewMysqlPreCheckJob(name, namespace string, target v1alpha1.Target) *batchv1.Job {
	job := JobTemplate(fmt.Sprintf("%s-precheck", name), namespace)
	job.Spec.Template.Spec.Containers = append(
		job.Spec.Template.Spec.Containers,
		corev1.Container{
			Name:            constants.ContainerName,
			Image:           constants.GetBenchmarkImage(constants.KubebenchTools),
			ImagePullPolicy: corev1.PullIfNotPresent,
			Command:         []string{"/tools"},
			Args: []string{"mysql", "ping",
				"--user", target.User,
				"--password", target.Password,
				"--host", target.Host,
				"--port", fmt.Sprintf("%d", target.Port),
			},
		},
	)

	return job
}

// NewPgbenchPreCheckJob create a job to check the pgbench connection
func NewPgbenchPreCheckJob(name string, namespace string, target v1alpha1.Target) *batchv1.Job {
	job := JobTemplate(fmt.Sprintf("%s-precheck", name), namespace)
	job.Spec.Template.Spec.Containers = append(
		job.Spec.Template.Spec.Containers,
		corev1.Container{
			Name:            constants.ContainerName,
			Image:           constants.GetBenchmarkImage(constants.KubebenchTools),
			ImagePullPolicy: corev1.PullIfNotPresent,
			Command:         []string{"/tools"},
			Args: []string{"postgresql", "ping",
				"--user", target.User,
				"--password", target.Password,
				"--host", target.Host,
				"--port", fmt.Sprintf("%d", target.Port),
			},
		},
	)

	return job
}

func NewMongodbPreCheckJob(name string, namespace string, target v1alpha1.Target) *batchv1.Job {
	job := JobTemplate(fmt.Sprintf("%s-precheck", name), namespace)
	job.Spec.Template.Spec.Containers = append(
		job.Spec.Template.Spec.Containers,
		corev1.Container{
			Name:            constants.ContainerName,
			Image:           constants.GetBenchmarkImage(constants.KubebenchTools),
			ImagePullPolicy: corev1.PullIfNotPresent,
			Command:         []string{"/tools"},
			Args: []string{"mongodb", "ping",
				"--user", target.User,
				"--password", target.Password,
				"--host", target.Host,
				"--port", fmt.Sprintf("%d", target.Port),
			},
		},
	)

	return job
}

func NewRedisPreCheckJob(name, namespace string, target v1alpha1.Target) *batchv1.Job {
	job := JobTemplate(fmt.Sprintf("%s-precheck", name), namespace)
	job.Spec.Template.Spec.Containers = append(
		job.Spec.Template.Spec.Containers,
		corev1.Container{
			Name:            constants.ContainerName,
			Image:           constants.GetBenchmarkImage(constants.KubebenchTools),
			ImagePullPolicy: corev1.PullIfNotPresent,
			Command:         []string{"/tools"},
			Args: []string{"redis", "ping",
				"--user", target.User,
				"--password", target.Password,
				"--host", target.Host,
				"--port", fmt.Sprintf("%d", target.Port),
			},
		},
	)

	return job
}

// NewPreCheckJob create a job to check the connection
func NewPreCheckJob(name, namespace string, driver string, target *v1alpha1.Target) *batchv1.Job {
	switch driver {
	case constants.MySqlDriver:
		return NewMysqlPreCheckJob(name, namespace, *target)
	case constants.PostgreSqlDriver:
		return NewPgbenchPreCheckJob(name, namespace, *target)
	case constants.MongoDbDriver:
		return NewMongodbPreCheckJob(name, namespace, *target)
	case constants.RedisDriver:
		return NewRedisPreCheckJob(name, namespace, *target)
	default:
		return nil
	}
}
