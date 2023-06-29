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

// PgbenchSpec defines the desired state of Pgbench
type PgbenchSpec struct {
	// Image defines the image to use for the benchmark.
	Image ImageSpec `json:"image,omitempty"`

	// Pod Contains the pod specification for the benchmark.
	PodConfig PodConfigSpec `json:"podConfig,omitempty"`
}

// PgbenchStatus defines the observed state of Pgbench
type PgbenchStatus struct {
	// Phase is the current state of the test. Valid values are Disabled, Enabled, Failed, Enabling, Disabling.
	// +kubebuilder:validation:Enum={Pending,Running,Complete,Failed}
	Phase BenchmarkPhase `json:"phase,omitempty"`

	// Describes the current state of add-on API installation conditions.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.phase",description="status phase"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"

// Pgbench is the Schema for the pgbenches API
type Pgbench struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PgbenchSpec   `json:"spec,omitempty"`
	Status PgbenchStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PgbenchList contains a list of Pgbench
type PgbenchList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Pgbench `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Pgbench{}, &PgbenchList{})
}
