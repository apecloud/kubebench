package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

// BenchmarkPhase is the current state of the test.
// +kubebuilder:validation:Enum={Pending,Running,Complete,Failed}
type BenchmarkPhase string

const (
	Pending  BenchmarkPhase = "Pending"
	Running  BenchmarkPhase = "Running"
	Complete BenchmarkPhase = "Complete"
	Failed   BenchmarkPhase = "Failed"
)

// EnvVar is an environment variable
type EnvVar struct {
	// Name is the name of the environment variable
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Value is the value of the environment variable
	// +kubebuilder:validation:Required
	Value string `json:"value"`
}

type ImageSpec struct {
	// Name is the Docker Image location including tag
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// +optional
	PullPolicy corev1.PullPolicy `json:"pullPolicy,omitempty"`

	// PullSecret is the secret to use to pull the image
	// +optional
	PullSecret string `json:"pullSecret,omitempty"`

	// Cmds is the commands to run in the container
	// +optional
	Cmds []string `json:"cmds,omitempty"`

	// Args is the arguments to pass to the command
	// +optional
	Args []string `json:"args,omitempty"`

	// Env is the environment variables to set in the container
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`
}

type PodConfigSpec struct {
	// Annotations is the annotations to add to the pod
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// Labels is the labels to add to the pod
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
}
