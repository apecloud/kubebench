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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RedisbenchSpec defines the desired state of Redisbench
type RedisbenchSpec struct {
	// clients provides a list of client counts to run redis-benchmark with multiple times.
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:default={1}
	Clients []int `json:"clients,omitempty"`

	// total number of requests
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=100000
	Requests int `json:"requests,omitempty"`

	// data size of set/get value in bytes
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=3
	DataSize int `json:"dataSize,omitempty"`

	// use random keys for SET/GET/INCR, random values for SADD.
	// +optional
	KeySpace *int `json:"keySpace,omitempty"`

	// only run the comma-separated list of tests. The test names are the same as the ones produced as output.
	Tests string `json:"tests,omitempty"`

	// pipeline number of commands to pipeline. Default is 1 (no pipeline).
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=1
	// +optional
	Pipeline int `json:"pipeline,omitempty"`

	// Quiet. Just show query/sec values.
	// +kubebuilder:default=true
	// +optional
	Quiet bool `json:"quiet,omitempty"`

	BenchCommon `json:",inline"`
}

// RedisbenchStatus defines the observed state of Redisbench
type RedisbenchStatus struct {
	// Phase is the current state of the test. Valid values are Disabled, Enabled, Failed, Enabling, Disabling.
	// +kubebuilder:validation:Enum={Pending,Running,Complete,Failed}
	Phase BenchmarkPhase `json:"phase,omitempty"`

	// completions is the completed/total number of pgbench runs
	Completions string `json:"completions,omitempty"`

	// succeeded is the number of successful pgbench runs
	Succeeded int `json:"succeeded,omitempty"`

	// total is the number of pgbench runs
	Total int `json:"total,omitempty"`

	// Describes the current state of benchmark conditions.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.phase",description="status phase"
// +kubebuilder:printcolumn:name="COMPLETIONS",type="string",JSONPath=".status.completions",description="completions"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"

// Redisbench is the Schema for the redisbenches API
type Redisbench struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RedisbenchSpec   `json:"spec,omitempty"`
	Status RedisbenchStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RedisbenchList contains a list of Redisbench
type RedisbenchList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Redisbench `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Redisbench{}, &RedisbenchList{})
}
