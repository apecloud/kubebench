package exporter

func InitMetrics() {
	InitPgbench()
	InitSysbench()
	InitEsrally()
}

// Register registers all metrics.
func Register() {
	RegisterCommon()
	RegisterPgbenchMetrics()
	RegisterSysbenchMetrics()
	RegisterEsrallyMetrics()
}
