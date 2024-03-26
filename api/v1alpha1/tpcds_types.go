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

// TpcdsSpec defines the desired state of Tpcds
type TpcdsSpec struct {
	// overall scale of the tpcds test
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=1
	// +required
	Size int `json:"size,omitempty"`

	// create pk and fk for tpcds test, it will cost extra time
	// +kubebuilder:default=false
	// +optional
	UseKey bool `json:"useKey,omitempty"`

	BenchCommon `json:",inline"`
}

// TpcdsStatus defines the observed state of Tpcds
type TpcdsStatus struct {
	// Phase is the current state of the test. Valid values are Disabled, Enabled, Failed, Enabling, Disabling.
	// +kubebuilder:validation:Enum={Pending,Running,Completed,Failed}
	Phase BenchmarkPhase `json:"phase,omitempty"`

	// completions is the completed/total number of sysbench runs
	Completions string `json:"completions,omitempty"`

	// succeeded is the number of successful sysbench runs
	Succeeded int `json:"succeeded,omitempty"`

	// failed is the number of failed sysbench runs
	Total int `json:"total,omitempty"`

	// Describes the current state of benchmark conditions.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Tpcds is the Schema for the tpcds API
type Tpcds struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TpcdsSpec   `json:"spec,omitempty"`
	Status TpcdsStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TpcdsList contains a list of Tpcds
type TpcdsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Tpcds `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Tpcds{}, &TpcdsList{})
}
