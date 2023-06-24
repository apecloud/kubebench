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

package v1alpha1

import (
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SysbenchSpec defines the desired state of Sysbench
type SysbenchSpec struct {
	// JobTemplate defines the job that will run the benchmark.
	JobTemplate batchv1.JobTemplateSpec `json:"jobTemplate"`
}

// SysbenchStatus defines the observed state of Sysbench
type SysbenchStatus struct {
	// Phase is the current state of the test.
	Phase BenchmarkPhase `json:"phase,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Sysbench is the Schema for the sysbenches API
type Sysbench struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SysbenchSpec   `json:"spec,omitempty"`
	Status SysbenchStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SysbenchList contains a list of Sysbench
type SysbenchList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Sysbench `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Sysbench{}, &SysbenchList{})
}
