package constants

const (
	ContainerName = "kubebench"
)

const (
	KubeBenchNameLabel = "kubebench.apecloud.io/name"
	KubeBenchTypeLabel = "kubebench.apecloud.io/type"
)

const (
	PgbenchType    = "pgbench"
	SysbenchType   = "sysbench"
	RedisBenchType = "redisbench"
	TpccType       = "tpcc"
	TpcdsType      = "tpcds"
	TpchType       = "tpch"
	YcsbType       = "ycsb"
	FioType        = "fio"
)

const (
	MySqlDriver      = "mysql"
	PostgreSqlDriver = "postgresql"
	MongoDbDriver    = "mongodb"
	RedisDriver      = "redis"
)

const (
	CleanupStep = "cleanup"
	PrepareStep = "prepare"
	RunStep     = "run"
	AllStep     = "all"
)
