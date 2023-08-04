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

// SysbenchSpec defines the desired state of Sysbench
type SysbenchSpec struct {
	// the number of tables to use for sysbench
	Tables int `json:"tables,omitempty"`

	// the data volume of tables to use for sysbench
	Size int `json:"size,omitempty"`

	// the number of threads to use for sysbench
	// +kubebuilder:validation:MinItems=1
	Threads []int `json:"threads,omitempty"`

	// the sysbench test types to run
	// +kubebuilder:validation:MinItems=1
	Types []string `json:"types"`

	// the number of seconds to run sysbench
	// +kubebuilder:validation:Minimum=1
	// +optional
	Duration int `json:"duration,omitempty"`

	// step is all, will exec cleanup, prepare, run
	// step is cleanup, will exec cleanup
	// step is prepare, will exec prepare
	// step is run, will exec run
	// +kubebuilder:default=all
	// +kubebuilder:validation:Enum={all,cleanup,prepare,run}
	// +optional
	Step string `json:"step,omitempty"`

	// the other sysbench run command flags to use for sysbench
	// +optional
	ExtraArgs []string `json:"extraArgs,omitempty"`

	// the target for the sysbench run command
	Target SysbenchTarget `json:"target,omitempty"`
}

type SysbenchTarget struct {
	// the driver of the sysbench target
	// +required
	Driver string `json:"driver,omitempty"`

	// The database server's host name
	// +kubebuilder:default=localhost
	// +required
	Host string `json:"host,omitempty"`

	// The database server's port number
	// +required
	Port int `json:"port,omitempty"`

	// the user of the sysbench target
	// +required
	User string `json:"user,omitempty"`

	// The database server's password
	// +optional
	Password string `json:"password,omitempty"`

	// the database of the sysbench target
	// +required
	Database string `json:"database,omitempty"`
}

// SysbenchStatus defines the observed state of Sysbench
type SysbenchStatus struct {
	// Phase is the current state of the test. Valid values are Disabled, Enabled, Failed, Enabling, Disabling.
	// +kubebuilder:validation:Enum={Pending,Running,Complete,Failed}
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
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.phase",description="status phase"
// +kubebuilder:printcolumn:name="COMPLETIONS",type="string",JSONPath=".status.completions",description="completions"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"

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
