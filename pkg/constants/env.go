package constants

import (
	"os"

	"github.com/spf13/viper"
)

const (
	KubebenchEnvPgbench    = "KUBEBENCH_PGBENCH_IMAGE"
	KubebenchEnvSysbench   = "KUBEBENCH_SYSBENCH_IMAGE"
	KubebenchEnvTpcc       = "KUBEBENCH_TPCC_IMAGE"
	KubebenchEnvTpcds      = "KUBEBENCH_TPCDS_IMAGE"
	KubebenchEnvTpch       = "KUBEBENCH_TPCH_IMAGE"
	KubebenchEnvYcsb       = "KUBEBENCH_YCSB_IMAGE"
	KubebenchEnvFio        = "KUBEBENCH_FIO_IMAGE"
	KubebenchEnvRedisBench = "KUBEBENCH_REDISBENCH_IMAGE"
	KubebenchExporter      = "KUBEBENCH_EXPORTER_IMAGE"
	KubebenchTools         = "KUBEBENCH_TOOLS_IMAGE"
)

func init() {
	viper.SetDefault(KubebenchEnvPgbench, "registry.cn-hangzhou.aliyuncs.com/apecloud/spilo:14.8.0")
	viper.SetDefault(KubebenchEnvSysbench, "registry.cn-hangzhou.aliyuncs.com/apecloud/customsuites:latest")
	viper.SetDefault(KubebenchEnvTpcc, "registry.cn-hangzhou.aliyuncs.com/apecloud/benchmarksql:latest")
	viper.SetDefault(KubebenchEnvTpcds, "infracreate-registry.cn-zhangjiakou.cr.aliyuncs.com/apecloud/tpcds:latest")
	viper.SetDefault(KubebenchEnvTpch, "registry.cn-hangzhou.aliyuncs.com/apecloud/customsuites:latest")
	viper.SetDefault(KubebenchEnvYcsb, "registry.cn-hangzhou.aliyuncs.com/apecloud/go-ycsb:latest")
	viper.SetDefault(KubebenchEnvFio, "registry.cn-hangzhou.aliyuncs.com/apecloud/fio:latest")
	viper.SetDefault(KubebenchEnvRedisBench, "registry.cn-hangzhou.aliyuncs.com/apecloud/redis:7.0.5")
	viper.SetDefault(KubebenchExporter, "registry.cn-hangzhou.aliyuncs.com/apecloud/kubebench:0.0.6")
	viper.SetDefault(KubebenchTools, "registry.cn-hangzhou.aliyuncs.com/apecloud/kubebench:0.0.6")
}

// GetBenchmarkImage get benchmark image
func GetBenchmarkImage(envName string) string {
	image := os.Getenv(envName)
	if image == "" {
		return viper.GetString(envName)
	}
	return image
}
