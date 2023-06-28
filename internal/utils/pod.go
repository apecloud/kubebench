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

	corev1 "k8s.io/api/core/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

func GetLogFromPod(rsc *rest.Config, reqCtx context.Context, podName string, namespace string) (string, error) {
	clientset, err := corev1client.NewForConfig(rsc)
	if err != nil {
		return "", err
	}
	req := clientset.Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{})
	data, err := req.DoRaw(reqCtx)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
