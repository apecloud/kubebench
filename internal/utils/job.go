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

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

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

	// delete the pods created by the job
	podList, err := GetPodListFromJob(cli, reqCtx, jobName, namespace)
	if err != nil {
		return err
	}
	for _, pod := range podList.Items {
		if err := cli.Delete(reqCtx, &pod); err != nil {
			return err
		}
	}

	// delete the job
	if err := cli.Delete(reqCtx, &job); err != nil {
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
