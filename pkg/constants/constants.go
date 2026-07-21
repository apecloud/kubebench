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
	EsrallyType    = "esrally"
	TpccType       = "tpcc"
	TpcdsType      = "tpcds"
	TpchType       = "tpch"
	YcsbType       = "ycsb"
	FioType        = "fio"
)

const (
	MySqlDriver                 = "mysql"
	PostgreSqlDriver            = "postgresql"
	MongoDbDriver               = "mongodb"
	RedisDriver                 = "redis"
	OceanBaseOracleTenantDriver = "oceanbase-oracle"
	DamengDriver                = "dameng"
	MinioDriver                 = "minio"
	TidbDriver                  = "tidb"
	MssqlDriver                 = "mssql"
	ElasticsearchDriver         = "elasticsearch"
	GaussDBDriver              = "gaussdb"
)

const (
	CleanupStep = "cleanup"
	PrepareStep = "prepare"
	RunStep     = "run"
	AllStep     = "all"
)

const (
	EsrallyDataProfileLogs        = "logs"
	EsrallyDataProfileMetrics     = "metrics"
	EsrallyDataProfileHTTPLogs    = "http_logs"
	EsrallyDataProfileMetricbeat  = "metricbeat"
	EsrallyDataProfileGeonames    = "geonames"
	EsrallyDataProfileNYCTaxis    = "nyc_taxis"
	EsrallyDataProfileNOAA        = "noaa"
	EsrallyDataProfileNested      = "nested"
	EsrallyDataProfilePMC         = "pmc"
	EsrallyDataProfileSO          = "so"
	EsrallyDataProfileDenseVector = "dense_vector"
)

const (
	EsrallyWorkloadIndex  = "index"
	EsrallyWorkloadSearch = "search"
	EsrallyWorkloadMixed  = "mixed"
	EsrallyWorkloadAll    = "all"
)
