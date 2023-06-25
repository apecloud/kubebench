package v1alpha1

// BenchmarkPhase is the current state of the test.
// +enum
type BenchmarkPhase string

const (
	Pending  BenchmarkPhase = "Pending"
	Running  BenchmarkPhase = "Running"
	Complete BenchmarkPhase = "Complete"
	Failed   BenchmarkPhase = "Failed"
)
