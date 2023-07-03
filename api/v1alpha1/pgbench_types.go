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
	// the parameters for the pgbench init command
	InitArgs PgbenchInitArgs `json:"initArgs,omitempty"`

	// the parameters for the pgbench run command
	RunArgs PgbenchRunArgs `json:"runArgs,omitempty"`

	// the target for the pgbench run command
	Target PgbenchTargets `json:"target,omitempty"`
}

// PgbenchInitArgs defines the parameters for the pgbench init command
type PgbenchInitArgs struct {
	// the scale factor to use for pgbench
	// +kubebuilder:validation:Minimum=1
	Scale int `json:"scale,omitempty"`

	// the other pgbench init command flags to use for pgbench
	// +optional
	OtherArgs []string `json:"others,omitempty"`
}

// PgbenchRunArgs defines the parameters for the pgbench run command
type PgbenchRunArgs struct {
	// clients are provided as a list of client counts to run pgbench with multiple times
	// +kubebuilder:validation:MinItems=1
	Clients []int `json:"clients,omitempty"`

	// the number of threads to use for pgbench
	// +kubebuilder:validation:Minimum=1
	// +optional
	Threads int `json:"threads,omitempty"`

	// establish a connection for each transaction
	// +optional
	Connect bool `json:"connect,omitempty"`

	// only run the select-only part of the benchmark
	// +optional
	SelectOnly bool `json:"selectOnly,omitempty"`

	// Note: the transactions and time parameters are mutually exclusive
	// the number of transactions to run for pgbench
	// +kubebuilder:validation:Minimum=1
	// +optional
	Transactions int `json:"transactions,omitempty"`

	// the number of seconds to run pgbench
	// +kubebuilder:validation:Minimum=1
	// +optional
	Time int `json:"time,omitempty"`

	// the other pgbench run command flags to use for pgbench
	// +optional
	OtherArgs []string `json:"others,omitempty"`
}

// PgbenchTargets defines the parameters for the pgbench target
type PgbenchTargets struct {
	// the name of the pgbench target
	Name string `json:"name,omitempty"`

	// the host of the pgbench target
	Host string `json:"host,omitempty"`

	// the port of the pgbench target
	Port int `json:"port,omitempty"`

	// the user of the pgbench target
	User string `json:"user,omitempty"`

	// the password of the pgbench target
	Password string `json:"password,omitempty"`

	// the database of the pgbench target
	Database string `json:"database,omitempty"`
}

// PgbenchStatus defines the observed state of Pgbench
type PgbenchStatus struct {
	// Phase is the current state of the test. Valid values are Disabled, Enabled, Failed, Enabling, Disabling.
	// +kubebuilder:validation:Enum={Pending,Running,Complete,Failed}
	Phase BenchmarkPhase `json:"phase,omitempty"`

	// completions is the completed/total number of pgbench runs
	Completions string `json:"completions,omitempty"`

	// ready is true when the pgbench benchmark is ready
	Ready bool `json:"ready,omitempty"`

	// succeeded is the number of successful pgbench runs
	Succeeded int `json:"succeeded,omitempty"`

	// total is the number of pgbench runs
	Total int `json:"total,omitempty"`

	// Describes the current state of pgbench benchmark.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.phase",description="status phase"
// +kubebuilder:printcolumn:name="COMPLETIONS",type="string",JSONPath=".status.completions",description="completions"
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
