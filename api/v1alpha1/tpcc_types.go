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

// TpccSpec defines the desired state of Tpcc
type TpccSpec struct {
	// overall scale of the tpcc test
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=1
	// +required
	WareHouses int `json:"wareHouses,omitempty"`

	// the number of threads to use for tpcc
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:default={1}
	// +required
	Threads []int `json:"threads,omitempty"`

	// the number of transactions to use for tpcc each thread
	// +kubebuilder:validation:Minimum=1
	// +optional
	Transactions int `json:"transactions,omitempty"`

	// the number of minutes to run tpcc
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=1
	// +optional
	Duration int `json:"duration,omitempty"`

	// number of transactions to use for each minute
	// 0 means no limit
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=0
	// +optional
	LimitTxPerMin int `json:"limitTxPerMin,omitempty"`

	// percentage of new order transactions
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:default=45
	// +optional
	NewOrder int `json:"newOrder,omitempty"`

	// percentage of payment transactions
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:default=43
	// +optional
	Payment int `json:"payment,omitempty"`

	// percentage of order status transactions
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:default=4
	// +optional
	OrderStatus int `json:"orderStatus,omitempty"`

	// percentage of delivery transactions
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:default=4
	// +optional
	Delivery int `json:"delivery,omitempty"`

	// percentage of stock level transactions
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:default=4
	// +optional
	StockLevel int `json:"stockLevel,omitempty"`

	BenchCommon `json:",inline"`
}

// TpccStatus defines the observed state of Tpcc
type TpccStatus struct {
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

// Tpcc is the Schema for the tpccs API
type Tpcc struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TpccSpec   `json:"spec,omitempty"`
	Status TpccStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TpccList contains a list of Tpcc
type TpccList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Tpcc `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Tpcc{}, &TpccList{})
}
