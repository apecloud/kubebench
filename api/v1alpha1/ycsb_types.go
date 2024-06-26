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

	// the distribution of requests accross the keyspace.
	// uniform: each key has an equal probability of being accessed.
	// sequential: keys are accessed in sequential order.
	// zipfian: some keys are accessed more frequently than others.
	// latest: the most recently inserted keys are accessed more frequently.
	// hotspot: a small number of keys are accessed more frequently.
	// exponential: keys are accessed in an exponential distribution.
	// +kubebuilder:validation:Enum={uniform,sequential,zipfian,latest,hotspot,exponential}
	// +kubebuilder:default=uniform
	// +optional
	RequestDistribution string `json:"requestDistribution,omitempty"`

	// the distribution of scan lengths
	// +kubebuilder:validation:Enum={uniform,zipfian}
	// +kubebuilder:default=uniform
	// +optional
	ScanLengthDistribution string `json:"scanLengthDistribution,omitempty"`

	// the distribution of field lengths
	// +kubebuilder:validation:Enum={constant,uniform,zipfian,histogram}
	// +kubebuilder:default=constant
	// +optional
	FieldLengthDistribution string `json:"fieldLengthDistribution,omitempty"`

	// the number of threads
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:default={1}
	// +optional
	Threads []int `json:"threads,omitempty"`

	// TODO: achieve the following fields in target
	RedisMode             string `json:"redisMode,omitempty"`
	MasterName            string `json:"masterName,omitempty"`
	RedisSentinelUsername string `json:"redisSentinelUsername,omitempty"`
	RedisSentinelPassword string `json:"redisSentinelPassword,omitempty"`
	RedisAddr             string `json:"redisAddr,omitempty"`

	BenchCommon `json:",inline"`
}

// YcsbStatus defines the observed state of Ycsb
type YcsbStatus struct {
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
