package exporter

import "github.com/prometheus/client_golang/prometheus"

func NewGauge(name string, help string, labels []string) *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: name,
			Help: help,
		},
		labels,
	)
}
