package constants

import (
	"os"

	"github.com/spf13/viper"
)

const (
	KubebenchEnvPgbench  = "KUBEBENCH_PGBENCH_IMAGE"
	KubebenchEnvSysbench = "KUBEBENCH_SYSBENCH_IMAGE"
	KubebenchEnvTpcc     = "KUBEBENCH_TPCC_IMAGE"
	KubebenchEnvTpch     = "KUBEBENCH_TPCH_IMAGE"
	KubebenchEnvYcsb     = "KUBEBENCH_YCSB_IMAGE"
)

func init() {
	viper.SetDefault(KubebenchEnvPgbench, "registry.cn-hangzhou.aliyuncs.com/apecloud/spilo:14.8.0")
	viper.SetDefault(KubebenchEnvSysbench, "registry.cn-hangzhou.aliyuncs.com/apecloud/customsuites:latest")
	viper.SetDefault(KubebenchEnvTpcc, "registry.cn-hangzhou.aliyuncs.com/apecloud/benchmarksql:latest")
	viper.SetDefault(KubebenchEnvTpch, "registry.cn-hangzhou.aliyuncs.com/apecloud/customsuites:latest")
	viper.SetDefault(KubebenchEnvYcsb, "registry.cn-hangzhou.aliyuncs.com/apecloud/go-ycsb:latest")
}

// GetBenchmarkImage get benchmark image
func GetBenchmarkImage(envName string) string {
	image := os.Getenv(envName)
	if image == "" {
		return viper.GetString(envName)
	}
	return image
}
