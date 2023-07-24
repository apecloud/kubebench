package exporter

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/klog/v2"
)

var (
	// match "scaling factor: 10"
	scaleRegex = regexp.MustCompile(`scaling factor: (\d+)`)

	// match "query mode: simple"
	queryModeRegex = regexp.MustCompile(`query mode: (\w+)`)

	// match "number of clients: 10"
	clientsRegex = regexp.MustCompile(`number of clients: (\d+)`)

	// match "number of threads: 10"
	threadsRegex = regexp.MustCompile(`number of threads: (\d+)`)

	// match "maximum number of tries: 0"
	maximumTryRegex = regexp.MustCompile(`maximum number of tries: (\d+)`)

	// match "number of transactions per client: 100"
	transactionsPerClientRegex = regexp.MustCompile(`number of transactions per client: (\d+)`)

	// match // "number of transactions actually processed: 1000/1000" // or "number of transactions actually processed: 1000"
	transactionsProcessedRegex = regexp.MustCompile(`number of transactions actually processed: (\d+)(/\d+)?`)

	// match "number of failed transactions: 0 (0.000%)"
	transactionsFailedRegex = regexp.MustCompile(`number of failed transactions: (\d+) \((\d+\.\d+)%\)`)

	// match "latency average = 0.000 ms"
	avgLatencyRegex = regexp.MustCompile(`latency average = (\d+\.\d+) ms`)

	// match "latency stddev = 0.000 ms"
	stdLatencyRegex = regexp.MustCompile(`latency stddev = (\d+\.\d+) ms`)

	// match "initial connection time = 0.000 ms"
	initialConnectionsTimeRegex = regexp.MustCompile(`initial connection time = (\d+\.\d+) ms`)

	// match "tps = 0.000000 (without initial connection time)"
	tpsRegex = regexp.MustCompile(`tps = (\d+\.\d+) \(without initial connection time\)`)

	// matc "progress: 1.0 s, 610.0 tps, lat 3.043 ms stddev 8.900, 0 failed"
	pgbenchSecondRegex = regexp.MustCompile(`progress: (\d+\.\d+) s, (\d+\.\d+) tps, lat (\d+\.\d+) ms stddev (\d+\.\d+), (\d+) failed`)
)

const (
	PgbenchScaleName = "kubebench_pgbench_scale"
	PgbenchScaleHelp = "The scale of pgbench"

	PgbenchClientsName = "kubebench_pgbench_clients"
	PgbenchClientsHelp = "The clients of pgbench"

	PgbenchThreadsName = "kubebench_pgbench_threads"
	PgbenchThreadsHelp = "The threads of pgbench"

	PgbenchMaximumTryName = "kubebench_pgbench_maximum_try"
	PgbenchMaximumTryHelp = "The maximum try of pgbench"

	PgbenchTransactionsPerClientName = "kubebench_pgbench_transactions_per_client"
	PgbenchTransactionsPerClientHelp = "The transactions per client of pgbench"

	PgbenchTransactionsProcessedName = "kubebench_pgbench_transactions_processed"
	PgbenchTransactionsProcessedHelp = "The transactions processed of pgbench"

	PgbenchTransactionsFailedName = "kubebench_pgbench_transactions_failed"
	PgbenchTransactionsFailedHelp = "The transactions failed of pgbench"

	PgbenchAvgLatencyName = "kubebench_pgbench_avg_latency"
	PgbenchAvgLatencyHelp = "The avg latency of pgbench"

	PgbenchStdLatencyName = "kubebench_pgbench_std_latency"
	PgbenchStdLatencyHelp = "The std latency of pgbench"

	PgbenchInitialConnectionsTimeName = "kubebench_pgbench_initial_connections_time"
	PgbenchInitialConnectionsTimeHelp = "The initial connections time of pgbench"

	PgbenchTpsName = "kubebench_pgbench_tps"
	PgbenchTpsHelp = "The tps of pgbench"

	PgbenchTpsSecondName = "kubebench_pgbench_tps_second"
	PgbenchTpsSecondHelp = "The tps of pgbench per second"

	PgbenchAvgLatencySecondName = "kubebench_pgbench_avg_latency_second"
	PgbenchAvgLatencySecondHelp = "The avg latency of pgbench per second"

	PgbenchStdLatencySecondName = "kubebench_pgbench_std_latency_second"
	PgbenchStdLatencySecondHelp = "The std latency of pgbench per second"

	PgbenchTransactionsFailedSecondName = "kubebench_pgbench_transactions_failed_second"
	PgbenchTransactionsFailedSecondHelp = "The transactions failed of pgbench per second"
)

var (
	PgbenchLabels   = []string{"benchmark", "name", "mode"}
	PgbenchGaugeMap = map[string]*prometheus.GaugeVec{}
)

// InitPgbench init the pgbench metrics
func InitPgbench() {
	// Init Gauge
	PgbenchGaugeMap[PgbenchScaleName] = NewGauge(PgbenchScaleName, PgbenchScaleHelp, PgbenchLabels)
	PgbenchGaugeMap[PgbenchClientsName] = NewGauge(PgbenchClientsName, PgbenchClientsHelp, PgbenchLabels)
	PgbenchGaugeMap[PgbenchThreadsName] = NewGauge(PgbenchThreadsName, PgbenchThreadsHelp, PgbenchLabels)
	PgbenchGaugeMap[PgbenchMaximumTryName] = NewGauge(PgbenchMaximumTryName, PgbenchMaximumTryHelp, PgbenchLabels)
	PgbenchGaugeMap[PgbenchTransactionsPerClientName] = NewGauge(PgbenchTransactionsPerClientName, PgbenchTransactionsPerClientHelp, PgbenchLabels)
	PgbenchGaugeMap[PgbenchTransactionsProcessedName] = NewGauge(PgbenchTransactionsProcessedName, PgbenchTransactionsProcessedHelp, PgbenchLabels)
	PgbenchGaugeMap[PgbenchTransactionsFailedName] = NewGauge(PgbenchTransactionsFailedName, PgbenchTransactionsFailedHelp, PgbenchLabels)
	PgbenchGaugeMap[PgbenchAvgLatencyName] = NewGauge(PgbenchAvgLatencyName, PgbenchAvgLatencyHelp, PgbenchLabels)
	PgbenchGaugeMap[PgbenchStdLatencyName] = NewGauge(PgbenchStdLatencyName, PgbenchStdLatencyHelp, PgbenchLabels)
	PgbenchGaugeMap[PgbenchInitialConnectionsTimeName] = NewGauge(PgbenchInitialConnectionsTimeName, PgbenchInitialConnectionsTimeHelp, PgbenchLabels)
	PgbenchGaugeMap[PgbenchTpsName] = NewGauge(PgbenchTpsName, PgbenchTpsHelp, PgbenchLabels)
	PgbenchGaugeMap[PgbenchTpsSecondName] = NewGauge(PgbenchTpsSecondName, PgbenchTpsSecondHelp, PgbenchLabels)
	PgbenchGaugeMap[PgbenchAvgLatencySecondName] = NewGauge(PgbenchAvgLatencySecondName, PgbenchAvgLatencySecondHelp, PgbenchLabels)
	PgbenchGaugeMap[PgbenchStdLatencySecondName] = NewGauge(PgbenchStdLatencySecondName, PgbenchStdLatencySecondHelp, PgbenchLabels)
	PgbenchGaugeMap[PgbenchTransactionsFailedSecondName] = NewGauge(PgbenchTransactionsFailedSecondName, PgbenchTransactionsFailedSecondHelp, PgbenchLabels)
}

// RegisterPgbenchMetrics registers the pgbench metrics
func RegisterPgbenchMetrics() {
	for _, gague := range PgbenchGaugeMap {
		prometheus.MustRegister(gague)
	}
}

type PgbenchResult struct {
	Scale                  int                    `json:"scale"`
	QueryMode              string                 `json:"queryMode"`
	Clients                int                    `json:"clients"`
	Threads                int                    `json:"threads"`
	MaximumTry             int                    `json:"maximumTry"`
	TransactionsPerClient  int                    `json:"transactionsPerClient"`
	TransactionsProcessed  int                    `json:"transactionsSum"`
	TransactionsFailed     int                    `json:"failedTransactionsSum"`
	AvgLatency             float64                `json:"avgLatency"`
	StdLatency             float64                `json:"stdLatency"`
	InitialConnectionsTime float64                `json:"initialConnectionsTime"`
	TPS                    float64                `json:"tps"`
	SecondResults          []*PgbenchSecondResult `json:"secondResults"`
}

type PgbenchSecondResult struct {
	TPS                   float64 `json:"tps"`
	AvgLatency            float64 `json:"avgLatency"`
	StdLatency            float64 `json:"stdLatency"`
	FailedTransactionsSum int     `json:"failedTransactionsSum"`
}

func ParsePgbenchResult(msg string) *PgbenchResult {
	result := new(PgbenchResult)
	lines := strings.Split(msg, "\n")

	for _, l := range lines {
		switch {
		case pgbenchSecondRegex.MatchString(l):
			secondResult := ParsePgbenchSecondResult(l)
			result.SecondResults = append(result.SecondResults, secondResult)
		case scaleRegex.MatchString(l):
			scale := strings.TrimSpace(strings.Split(l, ":")[1])
			result.Scale, _ = strconv.Atoi(scale)
		case queryModeRegex.MatchString(l):
			result.QueryMode = strings.TrimSpace(strings.Split(l, ":")[1])
		case clientsRegex.MatchString(l):
			clients := strings.TrimSpace(strings.Split(l, ":")[1])
			result.Clients, _ = strconv.Atoi(clients)
		case threadsRegex.MatchString(l):
			threads := strings.TrimSpace(strings.Split(l, ":")[1])
			result.Threads, _ = strconv.Atoi(threads)
		case maximumTryRegex.MatchString(l):
			maximumTry := strings.TrimSpace(strings.Split(l, ":")[1])
			result.MaximumTry, _ = strconv.Atoi(maximumTry)
		case transactionsPerClientRegex.MatchString(l):
			transactionsPerClient := strings.TrimSpace(strings.Split(l, ":")[1])
			result.TransactionsPerClient, _ = strconv.Atoi(transactionsPerClient)
		case transactionsProcessedRegex.MatchString(l):
			transactionsProcessed := strings.TrimSpace(strings.Split(l, ":")[1])
			transactionsProcessed = strings.Split(transactionsProcessed, "/")[0]
			result.TransactionsProcessed, _ = strconv.Atoi(transactionsProcessed)
		case transactionsFailedRegex.MatchString(l):
			transactionsFailed := strings.TrimSpace(strings.Split(l, ":")[1])
			transactionsFailed = strings.TrimSpace(strings.Split(transactionsFailed, "(")[0])
			result.TransactionsFailed, _ = strconv.Atoi(transactionsFailed)
		case avgLatencyRegex.MatchString(l):
			avgLatency := strings.TrimSpace(strings.Split(l, "=")[1])
			avgLatency = strings.TrimSpace(strings.Split(avgLatency, "ms")[0])
			result.AvgLatency, _ = strconv.ParseFloat(avgLatency, 64)
		case stdLatencyRegex.MatchString(l):
			stdLatency := strings.TrimSpace(strings.Split(l, "=")[1])
			stdLatency = strings.TrimSpace(strings.Split(stdLatency, "ms")[0])
			result.StdLatency, _ = strconv.ParseFloat(stdLatency, 64)
		case initialConnectionsTimeRegex.MatchString(l):
			initialConnectionsTime := strings.TrimSpace(strings.Split(l, "=")[1])
			initialConnectionsTime = strings.TrimSpace(strings.Split(initialConnectionsTime, "ms")[0])
			result.InitialConnectionsTime, _ = strconv.ParseFloat(initialConnectionsTime, 64)
		case tpsRegex.MatchString(l):
			tps := strings.TrimSpace(strings.Split(l, "=")[1])
			tps = strings.TrimSpace(strings.Split(tps, "(")[0])
			result.TPS, _ = strconv.ParseFloat(tps, 64)
		}
	}

	// if use pgbench -T, we need to calculate the transactions per client
	if result.TransactionsPerClient == 0 {
		result.TransactionsPerClient = result.TransactionsProcessed / result.Clients
	}

	return result
}

func ParsePgbenchSecondResult(msg string) *PgbenchSecondResult {
	// parse string like "progress: 1.0 s, 610.0 tps, lat 3.043 ms stddev 8.900, 0 failed"
	result := new(PgbenchSecondResult)

	// split by comma
	elements := strings.Split(msg, ",")

	// parse tps
	tpsMsg := strings.TrimSpace(elements[1])
	tpsMsg = strings.TrimSpace(strings.Split(tpsMsg, " ")[0])
	result.TPS, _ = strconv.ParseFloat(tpsMsg, 64)

	// parse latency
	latencyMsg := strings.TrimSpace(elements[2])
	avglatencyMsg := strings.TrimSpace(strings.Split(latencyMsg, " ")[1])
	stdLatencyMsg := strings.TrimSpace(strings.Split(latencyMsg, " ")[4])
	result.AvgLatency, _ = strconv.ParseFloat(avglatencyMsg, 64)
	result.StdLatency, _ = strconv.ParseFloat(stdLatencyMsg, 64)

	// parse failed
	failedMsg := strings.TrimSpace(elements[3])
	failedMsg = strings.TrimSpace(strings.Split(failedMsg, " ")[0])
	result.FailedTransactionsSum, _ = strconv.Atoi(failedMsg)

	return result
}

func ScrapPgbench(file, benchName, jobName string) {
	// read the file
	klog.Infof("read file %s", file)
	data, err := os.ReadFile(file)
	if err != nil {
		fmt.Printf("read file error: %s\n", err)
		return
	}

	// parse the file
	result := ParsePgbenchResult(string(data))
	UpdatePgbenchMetrics(benchName, jobName, result)
}

func UpdatePgbenchMetrics(benchName, jobName string, result *PgbenchResult) {
	queryMode := result.QueryMode
	values := []string{benchName, jobName, queryMode}

	CommonCounterInc(benchName, jobName, Pgbench)

	// update total metrics
	PgbenchGaugeMap[PgbenchScaleName].WithLabelValues(values...).Set(float64(result.Scale))
	PgbenchGaugeMap[PgbenchClientsName].WithLabelValues(values...).Set(float64(result.Clients))
	PgbenchGaugeMap[PgbenchThreadsName].WithLabelValues(values...).Set(float64(result.Threads))
	PgbenchGaugeMap[PgbenchMaximumTryName].WithLabelValues(values...).Set(float64(result.MaximumTry))
	PgbenchGaugeMap[PgbenchTransactionsPerClientName].WithLabelValues(values...).Set(float64(result.TransactionsPerClient))
	PgbenchGaugeMap[PgbenchTransactionsProcessedName].WithLabelValues(values...).Set(float64(result.TransactionsProcessed))
	PgbenchGaugeMap[PgbenchTransactionsFailedName].WithLabelValues(values...).Set(float64(result.TransactionsFailed))
	PgbenchGaugeMap[PgbenchAvgLatencyName].WithLabelValues(values...).Set(result.AvgLatency)
	PgbenchGaugeMap[PgbenchStdLatencyName].WithLabelValues(values...).Set(result.StdLatency)
	PgbenchGaugeMap[PgbenchInitialConnectionsTimeName].WithLabelValues(values...).Set(result.InitialConnectionsTime)
	PgbenchGaugeMap[PgbenchTpsName].WithLabelValues(values...).Set(result.TPS)
	klog.Info("UpdatePgbenchTotalMetrics result")

	// update second metrics
	for _, secondResult := range result.SecondResults {
		PgbenchGaugeMap[PgbenchTpsSecondName].WithLabelValues(values...).Set(secondResult.TPS)
		PgbenchGaugeMap[PgbenchAvgLatencySecondName].WithLabelValues(values...).Set(secondResult.AvgLatency)
		PgbenchGaugeMap[PgbenchStdLatencySecondName].WithLabelValues(values...).Set(secondResult.StdLatency)
		PgbenchGaugeMap[PgbenchTransactionsFailedSecondName].WithLabelValues(values...).Set(float64(secondResult.FailedTransactionsSum))

		// sleep 1 second to mock metrics collected every second
		klog.Info("update pgbench second metrics")
		time.Sleep(1 * time.Second)
	}
}
