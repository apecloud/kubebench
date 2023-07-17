package exporter

import "github.com/prometheus/client_golang/prometheus"

func InitMetrics() {
	InitPgbench()
	InitSysbench()
}

// Register registers all metrics.
func Register() {
	for _, counter := range PgbenchGaugeMap {
		prometheus.MustRegister(counter)
	}

	for _, counter := range SysbenchGaugeMap {
		prometheus.MustRegister(counter)
	}
}
