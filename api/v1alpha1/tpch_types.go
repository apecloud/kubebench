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

// TpchSpec defines the desired state of Tpch
type TpchSpec struct {
	// overall scale of the tpch test
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=1
	// +required
	Size int `json:"size,omitempty"`

	BenchCommon `json:",inline"`
}

// TpchStatus defines the observed state of Tpch
type TpchStatus struct {
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

	// the completion timestamp of the test
	CompletionTimestamp *metav1.Time `json:"completionTimestamp,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.phase",description="status phase"
// +kubebuilder:printcolumn:name="COMPLETIONS",type="string",JSONPath=".status.completions",description="completions"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"

// Tpch is the Schema for the tpches API
type Tpch struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TpchSpec   `json:"spec,omitempty"`
	Status TpchStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TpchList contains a list of Tpch
type TpchList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Tpch `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Tpch{}, &TpchList{})
}
