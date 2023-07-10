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

// RedisSpec defines the desired state of Redis
// Reference https://redis.io/docs/management/optimization/benchmarks/
type RedisSpec struct {
	// clients provides a list of client counts to run redis-benchmark with multiple times.
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:default={1}
	// +optional
	Clients []int `json:"clients,omitempty"`

	// Number of requests to perform. Default is 100000.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=100000
	// +optional
	Requests int `json:"requests,omitempty"`

	// Data size of SET/GET value in bytes. Default is 3.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=3
	// +optional
	DataSize int `json:"dataSize,omitempty"`

	// keepAlive enables TCP keepalive.
	// +kubebuilder:default=true
	// +optional
	KeepAlive bool `json:"keepAlive,omitempty"`

	// use random keys for SET/GET/INCR, random values for SADD.
	// +optional
	KeySpace int `json:"keySpace,omitempty"`

	// pipeline number of commands to pipeline. Default is 1 (no pipeline).
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=1
	// +optional
	Pipeline int `json:"pipeline,omitempty"`

	// The other options for redis-benchmark
	// +optional
	ExtraArgs []string `json:"extraArgs,omitempty"`

	// The target database server
	// +required
	Target RedisTarget `json:"target"`
}

// RedisStatus defines the observed state of Redis
type RedisStatus struct {
	// Phase is the current state of the test. Valid values are Disabled, Enabled, Failed, Enabling, Disabling.
	// +kubebuilder:validation:Enum={Pending,Running,Complete,Failed}
	Phase BenchmarkPhase `json:"phase,omitempty"`

	// completions is the completed/total number of redis-benchmark runs
	Completions string `json:"completions,omitempty"`

	// ready is true when the redis-benchmark is ready to run
	Ready bool `json:"ready,omitempty"`

	// succeeded is the number of successful redis-benchmark runs
	Succeeded int `json:"succeeded,omitempty"`

	// total is the number of redis-benchmark runs
	Total int `json:"total,omitempty"`

	// Describes the current state of redis-benchmark.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

type RedisTarget struct {
	// The database server's host name
	// +kubebuilder:default=localhost
	// +required
	Host string `json:"host"`

	// The database server's port number
	// +kubebuilder:default=6379
	// +required
	Port int `json:"port"`

	// The database server's passwork
	// +optional
	Password string `json:"password,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.phase",description="status phase"
// +kubebuilder:printcolumn:name="COMPLETIONS",type="string",JSONPath=".status.completions",description="completions"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"

// Redis is the Schema for the redis API
type Redis struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RedisSpec   `json:"spec,omitempty"`
	Status RedisStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RedisList contains a list of Redis
type RedisList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Redis `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Redis{}, &RedisList{})
}
