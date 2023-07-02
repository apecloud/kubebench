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
	// the parameters for the sysbench init command
	InitArgs SysbenchInitArgs `json:"initArgs,omitempty"`

	// the parameters for the sysbench run command
	RunArgs SysbenchRunArgs `json:"runArgs,omitempty"`

	// the target for the sysbench run command
	Target SysbenchTarget `json:"target,omitempty"`
}

type SysbenchInitArgs struct {
	// the number of tables to use for sysbench
	Tables int `json:"tables,omitempty"`

	// the data volume of tables to use for sysbench
	Size int `json:"size,omitempty"`

	// the other sysbench init command flags to use for sysbench
	// +optional
	OtherFlags string `json:"otherFlags"`
}

type SysbenchRunArgs struct {
	// the number of threads to use for sysbench
	// +kubebuilder:validation:MinItems=1
	Threads []int `json:"threads,omitempty"`

	// the sysbench test types to run
	// +kubebuilder:validation:MinItems=1
	Types []string `json:"types"`

	// the time to run the sysbench test
	// +optional
	Times int `json:"times,omitempty"`

	// the other sysbench run command flags to use for sysbench
	// +optional
	OtherFlags string `json:"others,omitempty"`
}

type SysbenchTarget struct {
	// the name of the sysbench target
	Name string `json:"name,omitempty"`

	// the driver of the sysbench target
	Driver string `json:"driver,omitempty"`

	// the host of the sysbench target
	Host string `json:"host,omitempty"`

	// the port of the sysbench target
	Port int `json:"port,omitempty"`

	// the user of the sysbench target
	User string `json:"user,omitempty"`

	// the password of the sysbench target
	Password string `json:"password,omitempty"`

	// the database of the sysbench target
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

	// Describes the current state of add-on API installation conditions.
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
