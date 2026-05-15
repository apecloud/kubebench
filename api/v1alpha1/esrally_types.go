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

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// EsrallySpec defines the desired state of Esrally.
type EsrallySpec struct {
	// track is the Rally track name to run from the configured track repository.
	// It is ignored when trackPath is set.
	// +kubebuilder:default=geonames
	// +optional
	Track string `json:"track,omitempty"`

	// trackRepository is the Rally track repository name.
	// +optional
	TrackRepository string `json:"trackRepository,omitempty"`

	// trackPath is a local file or directory path inside the Rally container.
	// When set, Rally uses --track-path instead of --track.
	// +optional
	TrackPath string `json:"trackPath,omitempty"`

	// challenge is the Rally challenge to run.
	// +optional
	Challenge string `json:"challenge,omitempty"`

	// includeTasks limits the race to specific task names.
	// +optional
	IncludeTasks []string `json:"includeTasks,omitempty"`

	// trackParams are Rally track template parameters.
	// +optional
	TrackParams map[string]string `json:"trackParams,omitempty"`

	// targetHosts overrides spec.target.host:spec.target.port with one or more Rally target hosts.
	// +optional
	TargetHosts []string `json:"targetHosts,omitempty"`

	// clientOptions is passed to Rally --client-options. If empty, basic auth
	// is synthesized from spec.target.user and spec.target.password when present.
	// +optional
	ClientOptions string `json:"clientOptions,omitempty"`

	// onError controls how Rally handles request errors.
	// +kubebuilder:validation:Enum={abort,continue,continue-on-network}
	// +kubebuilder:default=abort
	// +optional
	OnError string `json:"onError,omitempty"`

	// offline runs Rally without updating remote repositories or downloading data.
	// +optional
	Offline bool `json:"offline,omitempty"`

	// testMode enables Rally test mode for smoke checks. Results are not valid benchmark numbers.
	// +optional
	TestMode bool `json:"testMode,omitempty"`

	// telemetry lists Rally telemetry devices.
	// +optional
	Telemetry []string `json:"telemetry,omitempty"`

	// telemetryParams is passed to Rally --telemetry-params.
	// +optional
	TelemetryParams string `json:"telemetryParams,omitempty"`

	BenchCommon `json:",inline"`
}

// EsrallyStatus defines the observed state of Esrally.
type EsrallyStatus struct {
	// Phase is the current state of the test.
	// +kubebuilder:validation:Enum={Pending,Running,Completed,Failed}
	Phase BenchmarkPhase `json:"phase,omitempty"`

	// completions is the completed/total number of Rally jobs.
	Completions string `json:"completions,omitempty"`

	// succeeded is the number of successful Rally jobs.
	Succeeded int `json:"succeeded,omitempty"`

	// total is the number of Rally jobs.
	Total int `json:"total,omitempty"`

	// Describes the current state of benchmark conditions.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// the completion timestamp of the test.
	CompletionTimestamp *metav1.Time `json:"completionTimestamp,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.phase",description="status phase"
// +kubebuilder:printcolumn:name="COMPLETIONS",type="string",JSONPath=".status.completions",description="completions"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"

// Esrally is the Schema for the esrallies API.
type Esrally struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EsrallySpec   `json:"spec,omitempty"`
	Status EsrallyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// EsrallyList contains a list of Esrally.
type EsrallyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Esrally `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Esrally{}, &EsrallyList{})
}
