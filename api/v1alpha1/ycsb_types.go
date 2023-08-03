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

// YcsbSpec defines the desired state of ycsb
// Reference https://github.com/pingcap/go-ycsb
type YcsbSpec struct {
	// the number of records in the table to be inserted in
	// the load phase or the number of records already in the
	// table before the run phase.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=10000
	// +optional
	RecordCount int `json:"recordCount,omitempty"`

	// The number of operations to use during the run phase.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=10000
	// +optional
	OperationCount int `json:"operationCount,omitempty"`

	// the proportion of reads in the run phase.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:default=0
	// +optional
	ReadProportion int `json:"readProportion,omitempty"`

	// the proportion of updates in the run phase.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:default=0
	// +optional
	UpdateProportion int `json:"updateProportion,omitempty"`

	// the proportion of inserts in the run phase.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:default=0
	// +optional
	InsertProportion int `json:"insertProportion,omitempty"`

	// the proportion of operations read then modify a record in the run phase.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:default=0
	// +optional
	ReadModifyWriteProportion int `json:"readModifyWriteProportion,omitempty"`

	// the proportion of scans in the run phase.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:default=0
	// +optional
	ScanProportion int `json:"scanProportion,omitempty"`

	// the number of threads
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:default={1}
	// +optional
	Threads []int `json:"threads,omitempty"`

	// mode is all, will run cleanup, prepare, run
	// mode is cleanup, will run cleanup
	// mode is prepare, will run prepare
	// mode is run, will run cleanup, prepare, run
	// +kubebuilder:default=all
	// +kubebuilder:validation:Enum={all,cleanup,prepare,run}
	// +optional
	Mode string `json:"mode,omitempty"`

	// the other ycsb run command options to use for ycsb
	// +optional
	ExtraArgs []string `json:"extraArgs,omitempty"`

	// the target of the ycsb benchmark
	// +required
	Target YcsbTarget `json:"target,omitempty"`
}

type YcsbTarget struct {
	// the driver of the ycsb target
	// +required
	Driver string `json:"driver"`

	// The database server's host name
	// +kubebuilder:default=localhost
	// +required
	Host string `json:"host"`

	// The database server's port
	// +required
	Port int `json:"port"`

	// The database server's username
	// +optional
	User string `json:"user,omitempty"`

	// The database server's password
	// +optional
	Password string `json:"password,omitempty"`

	// The database server's database name
	// +optional
	Database string `json:"database,omitempty"`
}

// YcsbStatus defines the observed state of Ycsb
type YcsbStatus struct {
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

// Ycsb is the Schema for the ycsbs API
type Ycsb struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   YcsbSpec   `json:"spec,omitempty"`
	Status YcsbStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// YcsbList contains a list of Ycsb
type YcsbList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Ycsb `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Ycsb{}, &YcsbList{})
}
