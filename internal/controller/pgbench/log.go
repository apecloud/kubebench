package pgbench

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
)

var (
	// match scaling factor: 10
	scaleRegex = regexp.MustCompile(`scaling factor: (\d+)`)

	// match query mode: simple
	queryModeRegex = regexp.MustCompile(`query mode: (\w+)`)

	// match number of clients: 10
	clientsRegex = regexp.MustCompile(`number of clients: (\d+)`)

	// match number of threads: 10
	threadsRegex = regexp.MustCompile(`number of threads: (\d+)`)

	// match maximum number of tries: 0
	maximumTryRegex = regexp.MustCompile(`maximum number of tries: (\d+)`)

	// match number of transactions per client: 100
	transactionsPerClientRegex = regexp.MustCompile(`number of transactions per client: (\d+)`)

	// match number of transactions actually processed: 1000/1000
	transactionsProcessedRegex = regexp.MustCompile(`number of transactions actually processed: (\d+)/(\d+)`)

	// match number of failed transactions: 0 (0.000%)
	transactionsFailedRegex = regexp.MustCompile(`number of failed transactions: (\d+) \((\d+\.\d+)%\)`)

	// match latency average = 0.000 ms
	avgLatencyRegex = regexp.MustCompile(`latency average = (\d+\.\d+) ms`)

	// match latency stddev = 0.000 ms
	stdLatencyRegex = regexp.MustCompile(`latency stddev = (\d+\.\d+) ms`)

	// match initial connection time = 0.000 ms
	initialConnectionsTimeRegex = regexp.MustCompile(`initial connection time = (\d+\.\d+) ms`)

	// match tps = 0.000000 (without initial connection time)
	tpsRegex = regexp.MustCompile(`tps = (\d+\.\d+) \(without initial connection time\)`)
)

type PgbenchResult struct {
	Scale                  int     `json:"scale"`
	QueryMode              string  `json:"queryMode"`
	Clients                int     `json:"clients"`
	Threads                int     `json:"threads"`
	MaximumTry             int     `json:"maximumTry"`
	TransactionsPerClient  int     `json:"transactionsPerClient"`
	TransactionsProcessed  int     `json:"transactionsSum"`
	TransactionsFailed     int     `json:"failedTransactionsSum"`
	AvgLatency             float64 `json:"avgLatency"`
	StdLatency             float64 `json:"stdLatency"`
	InitialConnectionsTime float64 `json:"initialConnectionsTime"`
	Tps                    float64 `json:"tps"`
}

func ParsePgbenchResult(msg string) *PgbenchResult {
	result := new(PgbenchResult)
	lines := strings.Split(msg, "\n")

	for _, l := range lines {
		switch {
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
			result.Tps, _ = strconv.ParseFloat(tps, 64)
		}
	}

	return result
}

func ParsePgbench(msg string) string {
	result := ParsePgbenchResult(msg)

	// if scale is 0, it means we don't parse the result,
	// so we return empty string
	if result.Scale == 0 {
		return ""
	}

	// return the result like table
	t := table.NewWriter()
	t.AppendHeader(table.Row{"Scale", "Query Mode", "Clients",
		"Threads", "Maximum Try", "Transactions Per Client",
		"Transactions Processed", "Transactions Failed", "Avg Latency(ms)",
		"Std Latency(ms)", "Initial Connections Time(ms)", "TPS"})
	t.AppendRow(table.Row{result.Scale, result.QueryMode, result.Clients,
		result.Threads, result.MaximumTry, result.TransactionsPerClient,
		result.TransactionsProcessed, result.TransactionsFailed, result.AvgLatency,
		result.StdLatency, result.InitialConnectionsTime, result.Tps,
	})
	t.SetStyle(table.StyleLight)

	return fmt.Sprintf("\n%s", t.Render())
}
