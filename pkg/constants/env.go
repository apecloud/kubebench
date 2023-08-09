package constants

import "os"

const (
	KubebenchPgbench  = "KUBEBENCH_PGBENCH_IMAGE"
	KubebenchSysbench = "KUBEBENCH_SYSBENCH_IMAGE"
	KubebenchTpcc     = "KUBEBENCH_TPCC_IMAGE"
	KubebenchTpch     = "KUBEBENCH_TPCH_IMAGE"
	KubebenchYcsb     = "KUBEBENCH_YCSB_IMAGE"
)

// GetPgbenchImage get pgbench image from env
// if env is empty, return default image
func GetPgbenchImage() string {
	image := os.Getenv(KubebenchPgbench)
	if image == "" {
		return PgbenchImage
	}
	return image
}

// GetSysbenchImage get sysbench image from env
// if env is empty, return default image
func GetSysbenchImage() string {
	image := os.Getenv(KubebenchSysbench)
	if image == "" {
		return SysbenchImage
	}
	return image
}

// GetTpccImage get tpcc image from env
// if env is empty, return default image
func GetTpccImage() string {
	image := os.Getenv(KubebenchTpcc)
	if image == "" {
		return TpccImage
	}
	return image
}

// GetTpchImage get tpch image from env
// if env is empty, return default image
func GetTpchImage() string {
	image := os.Getenv(KubebenchTpch)
	if image == "" {
		return TpchImage
	}
	return image
}

// GetYcsbImage get ycsb image from env
// if env is empty, return default image
func GetYcsbImage() string {
	image := os.Getenv(KubebenchYcsb)
	if image == "" {
		return YcsbImage
	}
	return image
}
