package constants

import (
	"fmt"
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
	KubebenchEnvEsrally    = "KUBEBENCH_ESRALLY_IMAGE"
	KubebenchExporter      = "KUBEBENCH_EXPORTER_IMAGE"
	KubebenchTools         = "KUBEBENCH_TOOLS_IMAGE"
)

const (
	CfgKeyCtrlrMgrTolerations = "CM_TOLERATIONS"
)

const (
	DefaultImageRegistry = "apecloud-registry.cn-zhangjiakou.cr.aliyuncs.com"
)

func init() {
	viper.SetDefault(KubebenchEnvPgbench, fmt.Sprintf("%s/apecloud/spilo:14.8.0", DefaultImageRegistry))
	viper.SetDefault(KubebenchEnvSysbench, fmt.Sprintf("%s/apecloud/customsuites:latest", DefaultImageRegistry))
	viper.SetDefault(KubebenchEnvTpcc, fmt.Sprintf("%s/apecloud/benchmarksql:1.0", DefaultImageRegistry))
	viper.SetDefault(KubebenchEnvTpcds, fmt.Sprintf("%s/apecloud/tpcds:latest", DefaultImageRegistry))
	viper.SetDefault(KubebenchEnvTpch, fmt.Sprintf("%s/apecloud/customsuites:latest", DefaultImageRegistry))
	viper.SetDefault(KubebenchEnvYcsb, fmt.Sprintf("%s/apecloud/go-ycsb:latest", DefaultImageRegistry))
	viper.SetDefault(KubebenchEnvFio, fmt.Sprintf("%s/apecloud/fio:latest", DefaultImageRegistry))
	viper.SetDefault(KubebenchEnvRedisBench, fmt.Sprintf("%s/apecloud/redis:7.0.5", DefaultImageRegistry))
	viper.SetDefault(KubebenchEnvEsrally, fmt.Sprintf("%s/apecloud/kubebench-esrally:2.12.0", DefaultImageRegistry))
	viper.SetDefault(KubebenchExporter, fmt.Sprintf("%s/apecloud/kubebench:0.0.14", DefaultImageRegistry))
	viper.SetDefault(KubebenchTools, fmt.Sprintf("%s/apecloud/kubebench:0.0.14", DefaultImageRegistry))
	viper.SetDefault(CfgKeyCtrlrMgrTolerations, os.Getenv(CfgKeyCtrlrMgrTolerations))
}

// GetBenchmarkImage get benchmark image
func GetBenchmarkImage(envName string) string {
	image := os.Getenv(envName)
	if image == "" {
		return viper.GetString(envName)
	}
	return image
}
