package sysbench

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

var (
	//match "read:                            75292"
	readRegex = regexp.MustCompile(`read:\s+(\d+)`)

	//match "write:                           21510"
	writeRegex = regexp.MustCompile(`write:\s+(\d+)`)

	//match "other:                           10756"
	otherRegex = regexp.MustCompile(`other:\s+(\d+)`)

	//match "total:                           107558"
	totalRegex = regexp.MustCompile(`total:\s+(\d+)`)

	//match "transactions:                        5377   (537.50 per sec.)"
	transactionsRegex = regexp.MustCompile(`transactions:\s+(\d+)\s+\((\d+\.\d+)\s+per\s+sec.\)`)

	//match "queries:                             107558 (10751.85 per sec.)"
	queriesRegex = regexp.MustCompile(`queries:\s+(\d+)\s+\((\d+\.\d+)\s+per\s+sec.\)`)

	//match "ignored errors:                      1      (0.10 per sec.)"
	ignoredErrorsRegex = regexp.MustCompile(`ignored\s+errors:\s+(\d+)\s+\((\d+\.\d+)\s+per\s+sec.\)`)

	//match "reconnects:                          0      (0.00 per sec.)"
	reconnectsRegex = regexp.MustCompile(`reconnects:\s+(\d+)\s+\((\d+\.\d+)\s+per\s+sec.\)`)

	//match "total time:                          10.0023s"
	totalTimeRegex = regexp.MustCompile(`total\s+time:\s+(\d+\.\d+)s`)

	//match "total number of events:              5377"
	totalEventsRegex = regexp.MustCompile(`total\s+number\s+of\s+events:\s+(\d+)`)

	//match "min:                                    1.14"
	latencyMinRegex = regexp.MustCompile(`min:\s+(\d+\.\d+)`)

	//match "avg:                                    7.44"
	latencyAvgRegex = regexp.MustCompile(`avg:\s+(\d+\.\d+)`)

	//match "max:                                   91.40"
	latencyMaxRegex = regexp.MustCompile(`max:\s+(\d+\.\d+)`)

	//match "95th percentile:                       71.83"
	latencyNinetyFifthRegex = regexp.MustCompile(`95th\s+percentile:\s+(\d+\.\d+)`)

	//match "sum:                                39998.33"
	latencySumRegex = regexp.MustCompile(`sum:\s+(\d+\.\d+)`)

	//match "events (avg/stddev):           1344.2500/18.17"
	eventsRegex = regexp.MustCompile(`events\s+\(avg/stddev\):\s+(\d+\.\d+)/(\d+\.\d+)`)

	//match "execution time (avg/stddev):   9.9996/0.00"
	execTimeRegex = regexp.MustCompile(`execution\s+time\s+\(avg/stddev\):\s+(\d+\.\d+)/(\d+\.\d+)`)
)

type SysbenchResult struct {
	SQL             SQLStatistics     `json:"sql"`
	General         GeneralStatistics `json:"general"`
	Latency         Latency           `json:"latency"`
	ThreadsFairness ThreadsFairness   `json:"threadsFairness"`
}

type SQLStatistics struct {
	Read         int `json:"read"`
	Write        int `json:"write"`
	Other        int `json:"other"`
	Total        int `json:"total"`
	Transactions int `json:"transactions"`
	Queries      int `json:"queries"`
	IgnoreErrors int `json:"ignoreErrors"`
	Reconnects   int `json:"reconnects"`
}

type GeneralStatistics struct {
	TotalTime   float64 `json:"totalTime"`
	TotalEvents int     `json:"totalEvents"`
}

type Latency struct {
	Min         float64 `json:"min"`
	Avg         float64 `json:"avg"`
	Max         float64 `json:"max"`
	NinetyFifth float64 `json:"ninetyFifth"`
	Sum         float64 `json:"sum"`
}

type ThreadsFairness struct {
	EventsAvg    float64 `json:"eventsAvg"`
	EventsStddev float64 `json:"eventsStddev"`
	ExecTimeAvg  float64 `json:"execTimeAvg"`
	ExecTimeStd  float64 `json:"execTimeStd"`
}

func ParseSysBenchResult(msg string) SysbenchResult {
	result := new(SysbenchResult)
	lines := strings.Split(msg, "\n")

	for _, l := range lines {
		switch {
		case readRegex.MatchString(l):
			read := strings.TrimSpace(strings.Split(l, ":")[1])
			result.SQL.Read, _ = strconv.Atoi(read)
		case writeRegex.MatchString(l):
			write := strings.TrimSpace(strings.Split(l, ":")[1])
			result.SQL.Write, _ = strconv.Atoi(write)
		case otherRegex.MatchString(l):
			other := strings.TrimSpace(strings.Split(l, ":")[1])
			result.SQL.Other, _ = strconv.Atoi(other)
		case totalRegex.MatchString(l):
			total := strings.TrimSpace(strings.Split(l, ":")[1])
			result.SQL.Total, _ = strconv.Atoi(total)
		case transactionsRegex.MatchString(l):
			transactions := strings.TrimSpace(strings.Split(l, ":")[1])
			transactions = strings.TrimSpace(strings.Split(transactions, "(")[0])
			result.SQL.Transactions, _ = strconv.Atoi(transactions)
		case queriesRegex.MatchString(l):
			query := strings.TrimSpace(strings.Split(l, ":")[1])
			query = strings.TrimSpace(strings.Split(query, "(")[0])
			result.SQL.Queries, _ = strconv.Atoi(query)
		case ignoredErrorsRegex.MatchString(l):
			ignoreErrors := strings.TrimSpace(strings.Split(l, ":")[1])
			ignoreErrors = strings.TrimSpace(strings.Split(ignoreErrors, "(")[0])
			result.SQL.IgnoreErrors, _ = strconv.Atoi(ignoreErrors)
		case reconnectsRegex.MatchString(l):
			reconnects := strings.TrimSpace(strings.Split(l, ":")[1])
			reconnects = strings.TrimSpace(strings.Split(reconnects, "(")[0])
			result.SQL.Reconnects, _ = strconv.Atoi(reconnects)
		case totalTimeRegex.MatchString(l):
			totalTime := strings.TrimSpace(strings.Split(l, ":")[1])
			totalTime = strings.Trim(totalTime, "s")
			result.General.TotalTime, _ = strconv.ParseFloat(totalTime, 64)
		case totalEventsRegex.MatchString(l):
			totalEvents := strings.TrimSpace(strings.Split(l, ":")[1])
			result.General.TotalEvents, _ = strconv.Atoi(totalEvents)
		case latencyMinRegex.MatchString(l):
			latencyMin := strings.TrimSpace(strings.Split(l, ":")[1])
			result.Latency.Min, _ = strconv.ParseFloat(latencyMin, 64)
		case latencyAvgRegex.MatchString(l):
			latencyAvg := strings.TrimSpace(strings.Split(l, ":")[1])
			result.Latency.Avg, _ = strconv.ParseFloat(latencyAvg, 64)
		case latencyMaxRegex.MatchString(l):
			latencyMax := strings.TrimSpace(strings.Split(l, ":")[1])
			result.Latency.Max, _ = strconv.ParseFloat(latencyMax, 64)
		case latencyNinetyFifthRegex.MatchString(l):
			latencyNinetyFifth := strings.TrimSpace(strings.Split(l, ":")[1])
			result.Latency.NinetyFifth, _ = strconv.ParseFloat(latencyNinetyFifth, 64)
		case latencySumRegex.MatchString(l):
			latencySum := strings.TrimSpace(strings.Split(l, ":")[1])
			result.Latency.Sum, _ = strconv.ParseFloat(latencySum, 64)
		case eventsRegex.MatchString(l):
			events := strings.TrimSpace(strings.Split(l, ":")[1])
			eventsAvg := strings.TrimSpace(strings.Split(events, "/")[0])
			eventsStddev := strings.TrimSpace(strings.Split(events, "/")[1])
			result.ThreadsFairness.EventsAvg, _ = strconv.ParseFloat(eventsAvg, 64)
			result.ThreadsFairness.EventsStddev, _ = strconv.ParseFloat(eventsStddev, 64)
		case execTimeRegex.MatchString(l):
			execTime := strings.TrimSpace(strings.Split(l, ":")[1])
			execTimeAvg := strings.TrimSpace(strings.Split(execTime, "/")[0])
			execTimeStd := strings.TrimSpace(strings.Split(execTime, "/")[1])
			result.ThreadsFairness.ExecTimeAvg, _ = strconv.ParseFloat(execTimeAvg, 64)
			result.ThreadsFairness.ExecTimeStd, _ = strconv.ParseFloat(execTimeStd, 64)
		}
	}

	return *result
}

func ParseSysBench(msg string) string {
	result := ParseSysBenchResult(msg)

	// if result.SQL.Total is 0, it means the parse failed
	if result.SQL.Total == 0 {
		return ""
	}

	t := table.NewWriter()
	t.SetTitle("SQL Statistics")
	t.AppendHeader(table.Row{"Read", "Write", "Other", "Total", "Transactions", "Queries", "IgnoreErrors", "Reconnects"})
	t.AppendRow(table.Row{
		result.SQL.Read, result.SQL.Write, result.SQL.Other, result.SQL.Total,
		fmt.Sprintf("%d (%.2f per sec.)", result.SQL.Transactions, float64(result.SQL.Transactions)/result.General.TotalTime),
		fmt.Sprintf("%d (%.2f per sec.)", result.SQL.Queries, float64(result.SQL.Queries)/result.General.TotalTime),
		fmt.Sprintf("%d (%.2f per sec.)", result.SQL.IgnoreErrors, float64(result.SQL.IgnoreErrors)/result.General.TotalTime),
		fmt.Sprintf("%d (%.2f per sec.)", result.SQL.Reconnects, float64(result.SQL.Reconnects)/result.General.TotalTime)})
	t.Style().Title.Align = text.AlignCenter
	sqlAnalysis := t.Render()

	t = table.NewWriter()
	t.SetTitle("General Statistics")
	t.AppendHeader(table.Row{"TotalTime", "TotalEvents"})
	t.AppendRow(table.Row{fmt.Sprintf("%.2fs", result.General.TotalTime), result.General.TotalEvents})
	t.Style().Title.Align = text.AlignCenter
	generalAnalysis := t.Render()

	t = table.NewWriter()
	t.SetTitle("Latency (ms)")
	t.AppendHeader(table.Row{"Min", "Avg", "Max", "NinetyFifth", "Sum"})
	t.AppendRow(table.Row{result.Latency.Min, result.Latency.Avg, result.Latency.Max, result.Latency.NinetyFifth, result.Latency.Sum})
	t.Style().Title.Align = text.AlignCenter
	latencyAnalysis := t.Render()

	t = table.NewWriter()
	t.SetTitle("Threads Fairness")
	t.AppendHeader(table.Row{"EventsAvg", "EventsStddev", "ExecTimeAvg", "ExecTimeStd"})
	t.AppendRow(table.Row{result.ThreadsFairness.EventsAvg, result.ThreadsFairness.EventsStddev, result.ThreadsFairness.ExecTimeAvg, result.ThreadsFairness.ExecTimeStd})
	t.Style().Title.Align = text.AlignCenter
	threadsFairnessAnalysis := t.Render()

	return fmt.Sprintf("\n%s\n%s\n%s\n%s\n", sqlAnalysis, generalAnalysis, latencyAnalysis, threadsFairnessAnalysis)
}
