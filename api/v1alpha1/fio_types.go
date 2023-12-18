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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FioSpec defines the desired state of Fio
// Reference https://fio.readthedocs.io/en/latest/fio_doc.html
type FioSpec struct {
	// The total size of file I/O for each thread of the test.
	// Please refer to https://fio.readthedocs.io/en/latest/fio_doc.html#i-o-size
	// +kubebuilder:validation:Pattern=^[0-9]+[kKmMgG]?$
	// +kubebuilder:default="1G"
	// +optional
	Size string `json:"size,omitempty"`

	// The block size to use for the test.
	// Please refer to https://fio.readthedocs.io/en/latest/fio_doc.html#block-size
	// +kubebuilder:validation:Pattern=^[0-9]+[kKmMgG]?$
	// +kubebuilder:default="4k"
	// +optional
	Bs string `json:"bs,omitempty"`

	// The number of threads to use for the test.
	// Please refer to https://fio.readthedocs.io/en/latest/fio_doc.html#job-description
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:default={1}
	// +optional
	Numjobs []int `json:"numjobs,omitempty"`

	// Limit runtime. The test will run until it completes the configured I/O workload
	// or until it has run for this specified amount of time, whichever occurs first.
	// Please refer to https://fio.readthedocs.io/en/latest/fio_doc.html##time-related-parameters
	// +kubebuilder:validation:Minimum=0
	// +optional
	RunTime int `json:"runtime,omitempty"`

	// Number of I/O units to keep in flight against the file.
	// Please refer to https://fio.readthedocs.io/en/latest/fio_doc.html#i-o-depth
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=1
	// +optional
	Iodepth int `json:"iodepth,omitempty"`

	// The I/O engine to use.
	// Please refer to https://fio.readthedocs.io/en/latest/fio_doc.html#cmdoption-arg-ioengine
	// +kubebuilder:default=psync
	// +optional
	IoEngine string `json:"ioengine,omitempty"`

	// Whether to use non-buffered I/O.
	// Please refer to https://fio.readthedocs.io/en/latest/fio_doc.html#i-o-type
	// +kubebuilder:default=true
	// +optional
	Direct bool `json:"direct,omitempty"`

	// The type of I/O pattern to use.
	// Please refer to https://fio.readthedocs.io/en/latest/fio_doc.html#i-o-type
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:default={read}
	// +optional
	Rws []string `json:"rws,omitempty"`

	// The other fio run command flags to use for benchmark
	// +optional
	ExtraArgs []string `json:"extraArgs,omitempty"`

	// If specified, the pod's tolerations.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty" protobuf:"bytes,22,opt,name=tolerations"`

	// If specified, the pod's cpu.
	// +optional
	Cpu string `json:"cpu,omitempty" protobuf:"bytes,23,opt,name=cpu"`

	// If specified, the pod's memory.
	// +optional
	Memory string `json:"memory,omitempty" protobuf:"bytes,24,opt,name=memory"`
}

// FioStatus defines the observed state of Fio
type FioStatus struct {
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

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Fio is the Schema for the fios API
type Fio struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FioSpec   `json:"spec,omitempty"`
	Status FioStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FioList contains a list of Fio
type FioList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Fio `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Fio{}, &FioList{})
}
