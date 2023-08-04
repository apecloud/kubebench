package v1alpha1

import corev1 "k8s.io/api/core/v1"

// BenchmarkPhase is the current state of the test.
// +kubebuilder:validation:Enum={Pending,Running,Complete,Failed}
type BenchmarkPhase string

const (
	Pending  BenchmarkPhase = "Pending"
	Running  BenchmarkPhase = "Running"
	Complete BenchmarkPhase = "Complete"
	Failed   BenchmarkPhase = "Failed"
)

// PodSpec defines the desired state of Pod
type PodSpec struct {
	// If specified, the pod's tolerations.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty" protobuf:"bytes,22,opt,name=tolerations"`
}
