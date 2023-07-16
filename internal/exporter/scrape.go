package exporter

import (
	"fmt"
	"os"

	"github.com/prometheus/client_golang/prometheus"
)

// Scrape is a function to scrape benchmark result from log.
func Scrape(benchType string, file string, ch chan struct{}) {
	defer func() {
		// notify the channel
		ch <- struct{}{}
	}()

	switch benchType {
	case Pgbench:
		ScrapPgbench(file)
	case Sysbench:
		ScrapeSysbench(file)
	default:
		fmt.Printf("not support benchmark type: %s\n", benchType)
	}
}

func ScrapeSysbench(file string) {
	// read the file
	data, err := os.ReadFile(file)
	if err != nil {
		fmt.Printf("read file error: %s\n", err)
		return
	}

	// parse the file
	result := ParseSysBenchResult(string(data))
	SysbenchMetrics.With(prometheus.Labels{
		"query_read":            fmt.Sprintf("%d", result.SQL.Read),
		"query_write":           fmt.Sprintf("%d", result.SQL.Write),
		"query_other":           fmt.Sprintf("%d", result.SQL.Other),
		"query_total":           fmt.Sprintf("%d", result.SQL.Total),
		"transactions":          fmt.Sprintf("%d", result.SQL.Transactions),
		"queries":               fmt.Sprintf("%d", result.SQL.Queries),
		"ignored_errors":        fmt.Sprintf("%d", result.SQL.IgnoreErrors),
		"reconnects":            fmt.Sprintf("%d", result.SQL.Reconnects),
		"total_time":            fmt.Sprintf("%f", result.General.TotalTime),
		"total_events":          fmt.Sprintf("%d", result.General.TotalEvents),
		"latency_min":           fmt.Sprintf("%f", result.Latency.Min),
		"latency_avg":           fmt.Sprintf("%f", result.Latency.Avg),
		"latency_max":           fmt.Sprintf("%f", result.Latency.Max),
		"latency_95th":          fmt.Sprintf("%f", result.Latency.NinetyFifth),
		"latency_sum":           fmt.Sprintf("%f", result.Latency.Sum),
		"threads_events_avg":    fmt.Sprintf("%f", result.ThreadsFairness.EventsAvg),
		"threads_events_stddev": fmt.Sprintf("%f", result.ThreadsFairness.EventsStddev),
		"threads_exec_avg":      fmt.Sprintf("%f", result.ThreadsFairness.ExecTimeAvg),
		"threads_exec_stddev":   fmt.Sprintf("%f", result.ThreadsFairness.ExecTimeStd),
	}).Inc()

}

func ScrapPgbench(file string) {
	// read the file
	data, err := os.ReadFile(file)
	if err != nil {
		fmt.Printf("read file error: %s\n", err)
		return
	}

	// parse the file
	result := ParsePgbenchResult(string(data))
	PgbenchMetrics.With(prometheus.Labels{
		"scale":                    fmt.Sprintf("%d", result.Scale),
		"query_mode":               result.QueryMode,
		"clients":                  fmt.Sprintf("%d", result.Clients),
		"threads":                  fmt.Sprintf("%d", result.Threads),
		"maximum_try":              fmt.Sprintf("%d", result.MaximumTry),
		"transactions_per_client":  fmt.Sprintf("%d", result.TransactionsPerClient),
		"transactions_processed":   fmt.Sprintf("%d", result.TransactionsProcessed),
		"transactions_failed":      fmt.Sprintf("%d", result.TransactionsFailed),
		"avg_latency":              fmt.Sprintf("%f", result.AvgLatency),
		"std_latency":              fmt.Sprintf("%f", result.StdLatency),
		"initial_connections_time": fmt.Sprintf("%f", result.InitialConnectionsTime),
		"tps":                      fmt.Sprintf("%f", result.TPS),
	}).Inc()
}
