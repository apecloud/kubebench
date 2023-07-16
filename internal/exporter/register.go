package exporter

import "github.com/prometheus/client_golang/prometheus"

var (
	// SysbenchMetrics Sysbench Metrics
	SysbenchMetrics = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kubebench_sysbench_metrics",
			Help: "Sysbench metrics record sysbench test result.",
		},
		[]string{
			"query_read",
			"query_write",
			"query_other",
			"query_total",
			"transactions",
			"queries",
			"ignored_errors",
			"reconnects",
			"total_time",
			"total_events",
			"latency_min",
			"latency_avg",
			"latency_max",
			"latency_95th",
			"latency_sum",
			"threads_events_avg",
			"threads_events_stddev",
			"threads_exec_avg",
			"threads_exec_stddev",
		},
	)

	// PgbenchMetrics Pgbench Metrics
	PgbenchMetrics = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kubebench_pgbench_metrics",
			Help: "Pgbench metrics record pgbench test result.",
		},
		[]string{
			"scale",
			"query_mode",
			"clients",
			"threads",
			"maximum_try",
			"transactions_per_client",
			"transactions_processed",
			"transactions_failed",
			"avg_latency",
			"std_latency",
			"initial_connections_time",
			"tps",
		},
	)
)

const (
	Sysbench = "sysbench"
	Pgbench  = "pgbench"
)

// Register registers all metrics.
func Register(benchType string) {
	switch benchType {
	case Pgbench:
		prometheus.MustRegister(PgbenchMetrics)
	case Sysbench:
		prometheus.MustRegister(SysbenchMetrics)
	default:
		panic("not support benchmark type")
	}
}
