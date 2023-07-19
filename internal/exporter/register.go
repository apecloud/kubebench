package exporter

func InitMetrics() {
	InitPgbench()
	InitSysbench()
}

// Register registers all metrics.
func Register() {
	RegisterCommon()
	RegisterPgbenchMetrics()
	RegisterSysbenchMetrics()
}
