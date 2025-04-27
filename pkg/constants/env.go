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

const (
	CfgKeyCtrlrMgrTolerations = "CM_TOLERATIONS"
)

func init() {
	viper.SetDefault(KubebenchEnvPgbench, "apecloud-registry.cn-zhangjiakou.cr.aliyuncs.com/apecloud/spilo:14.8.0")
	viper.SetDefault(KubebenchEnvSysbench, "apecloud-registry.cn-zhangjiakou.cr.aliyuncs.com/apecloud/customsuites:latest")
	viper.SetDefault(KubebenchEnvTpcc, "apecloud-registry.cn-zhangjiakou.cr.aliyuncs.com/apecloud/benchmarksql:latest")
	viper.SetDefault(KubebenchEnvTpcds, "apecloud-registry.cn-zhangjiakou.cr.aliyuncs.com/apecloud/tpcds:latest")
	viper.SetDefault(KubebenchEnvTpch, "apecloud-registry.cn-zhangjiakou.cr.aliyuncs.com/apecloud/customsuites:latest")
	viper.SetDefault(KubebenchEnvYcsb, "apecloud-registry.cn-zhangjiakou.cr.aliyuncs.com/apecloud/go-ycsb:latest")
	viper.SetDefault(KubebenchEnvFio, "apecloud-registry.cn-zhangjiakou.cr.aliyuncs.com/apecloud/fio:latest")
	viper.SetDefault(KubebenchEnvRedisBench, "apecloud-registry.cn-zhangjiakou.cr.aliyuncs.com/apecloud/redis:7.0.5")
	viper.SetDefault(KubebenchExporter, "apecloud-registry.cn-zhangjiakou.cr.aliyuncs.com/apecloud/kubebench:0.0.10")
	viper.SetDefault(KubebenchTools, "apecloud-registry.cn-zhangjiakou.cr.aliyuncs.com/apecloud/kubebench:0.0.10")
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
