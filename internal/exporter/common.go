package exporter

import "github.com/prometheus/client_golang/prometheus"

const (
	KubebenchTotalName = "kubebench_total"
	KubebenchTotalHelp = "Total number of kubebench runs"
)

var (
	KubebenchTotalLabels = []string{"benchmark", "name", "type"}
	KubebenchCounter     = NewCounter(KubebenchTotalName, KubebenchTotalHelp, KubebenchTotalLabels)
)

func RegisterCommon() {
	prometheus.MustRegister(KubebenchCounter)
}

func CommonCouterInc(benchmark, name, typ string) {
	KubebenchCounter.WithLabelValues(benchmark, name, typ).Inc()
}
