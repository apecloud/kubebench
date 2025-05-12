package v1alpha1

import corev1 "k8s.io/api/core/v1"

// BenchmarkPhase is the current state of the test.
// +kubebuilder:validation:Enum={Pending,Running,Completed,Failed}
type BenchmarkPhase string

const (
	Pending   BenchmarkPhase = "Pending"
	Running   BenchmarkPhase = "Running"
	Completed BenchmarkPhase = "Completed"
	Failed    BenchmarkPhase = "Failed"
)

// BenchCommon defines common attributes for all benchmarks.
type BenchCommon struct {
	// step is all, will exec cleanup, prepare, run
	// step is cleanup, will exec cleanup
	// step is prepare, will exec prepare
	// step is run, will exec run
	// +kubebuilder:default=all
	// +kubebuilder:validation:Enum={all,cleanup,prepare,run}
	// +optional
	Step string `json:"step,omitempty"`

	// the database target to run benchmark
	// +required
	Target Target `json:"target"`

	// the other sysbench run command flags to use for benchmark
	// +optional
	ExtraArgs []string `json:"extraArgs,omitempty"`

	// If specified, the pod's tolerations.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty" protobuf:"bytes,22,opt,name=tolerations"`

	// the resource requirements for the benchmark
	ResourceLimits   *ResourceList `json:"resourceLimits,omitempty"`
	ResourceRequests *ResourceList `json:"resourceRequests,omitempty"`
}

type ResourceList struct {
	Cpu string `json:"cpu,omitempty" protobuf:"bytes,23,opt,name=cpu"`

	Memory string `json:"memory,omitempty" protobuf:"bytes,24,opt,name=memory"`
}

type Target struct {
	// the driver represents the database type
	// +optional
	// +kubebuilder:validation:Enum={mysql,postgresql,mongodb,redis,oceanbase-oracle,dameng,minio,tidb}
	Driver string `json:"driver,omitempty"`

	// The database server's host name
	// +required
	Host string `json:"host"`

	// The database server's port number
	// +required
	Port int `json:"port"`

	// The username to connect as
	// +optional
	User string `json:"user,omitempty"`

	// The database server's password
	// +optional
	Password string `json:"password,omitempty"`

	// The database name of the target
	// +optional
	// +kubebuilder:default=kubebench
	Database string `json:"database,omitempty"`
}
