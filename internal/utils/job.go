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
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/apecloud/kubebench/api/v1alpha1"
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

func DelteJob(cli client.Client, reqCtx context.Context, jobName string, namespace string) error {
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

func NewJob(jobName string, namespace string, objectMeta metav1.ObjectMeta, image v1alpha1.ImageSpec) *batchv1.Job {
	backoffLimit := int32(0) // no retry

	job := &batchv1.Job{
		ObjectMeta: objectMeta,
		Spec: batchv1.JobSpec{
			BackoffLimit: &backoffLimit,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            jobName,
							Image:           image.Image,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command:         image.Cmds,
							Args:            image.Args,
							Env:             image.Env,
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
		},
	}

	return job
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

		// record the log to the conditions
		meta.SetStatusCondition(conditions, metav1.Condition{
			Type:               fmt.Sprintf("%s-%s", pod.Name, pod.Status.Phase),
			Status:             metav1.ConditionTrue,
			ObservedGeneration: pod.Generation,
			Reason:             "RecordLog",
			Message:            msg,
			LastTransitionTime: metav1.Now(),
		})
	}

	return nil
}

func trimTooLongLog(log string) string {
	lines := strings.Split(log, "\n")
	reuslt := ""
	for _, line := range lines {
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
