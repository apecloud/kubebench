package constants

const (
	ContainerName           = "kubebench"
	PrometheusExporterImage = "registry.cn-hangzhou.aliyuncs.com/apecloud/kubebench:0.0.1"
	BenchToolsImage         = "guang/kubebench:latest"
)

const (
	KubeBenchNameLabel = "kubebench.apecloud.io/name"
	KubeBenchTypeLabel = "kubebench.apecloud.io/type"
)

const (
	PgbenchType  = "pgbench"
	SysbenchType = "sysbench"
	TpccType     = "tpcc"
	TpchType     = "tpch"
	YcsbType     = "ycsb"
	FioType      = "fio"
)
