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

	//match "99th percentile:                       71.83"
	latencyNinetyNinthRegex = regexp.MustCompile(`99th\s+percentile:\s+(\d+\.\d+)`)

	//match "sum:                                39998.33"
	latencySumRegex = regexp.MustCompile(`sum:\s+(\d+\.\d+)`)

	//match "events (avg/stddev):           1344.2500/18.17"
	eventsRegex = regexp.MustCompile(`events\s+\(avg/stddev\):\s+(\d+\.\d+)/(\d+\.\d+)`)

	//match "execution time (avg/stddev):   9.9996/0.00"
	execTimeRegex = regexp.MustCompile(`execution\s+time\s+\(avg/stddev\):\s+(\d+\.\d+)/(\d+\.\d+)`)

	//match "[ 1s ] thds: 4 tps: 563.40 qps: 11319.87 (r/w/o: 7931.50/2257.58/1130.79) lat (ms,99%): 70.55 err/s: 0.00 reconn/s: 0.00"
	sysbenchSecondRegex = regexp.MustCompile(`\[ \d+s \]\s+thds:\s+(\d+)\s+tps:\s+(\d+\.\d+)\s+qps:\s+(\d+\.\d+)\s+\(r/w/o:\s+(\d+\.\d+)/(\d+\.\d+)/(\d+\.\d+)\)\s+lat\s+\(ms,99%\):\s+(\d+\.\d+)\s+err/s:\s+(\d+\.\d+)\s+reconn/s:\s+(\d+\.\d+)`)
)

const (
	SysbenchQueryReadName = "kubebench_sysbench_query_read"
	SysbenchQueryReadHelp = "Sysbench query read result"

	SysbenchQueryWriteName = "kubebench_sysbench_query_write"
	SysbenchQueryWriteHelp = "Sysbench query write result"

	SysbenchQueryOtherName = "kubebench_sysbench_query_other"
	SysbenchQueryOtherHelp = "Sysbench query other result"

	SysbenchQueryTotalName = "kubebench_sysbench_query_total"
	SysbenchQueryTotalHelp = "Sysbench query total result"

	SysbenchTransactionsName = "kubebench_sysbench_transactions"
	SysbenchTransactionsHelp = "Sysbench transactions"

	SysbenchQueriesName = "kubebench_sysbench_queries"
	SysbenchQueriesHelp = "Sysbench queries"

	SysbenchIgnoredErrorsName = "kubebench_sysbench_ignored_errors"
	SysbenchIgnoredErrorsHelp = "Sysbench ignored errors"

	SysbenchReconnectsName = "kubebench_sysbench_reconnects"
	SysbenchReconnectsHelp = "Sysbench reconnects"

	SysbenchTotalTimeName = "kubebench_sysbench_total_time"
	SysbenchTotalTimeHelp = "Sysbench total time"

	SysbenchTotalEventsName = "kubebench_sysbench_total_events"
	SysbenchTotalEventsHelp = "Sysbench total events"

	SysbenchLatencyMinName = "kubebench_sysbench_latency_min"
	SysbenchLatencyMinHelp = "Sysbench latency min"

	SysbenchLatencyAvgName = "kubebench_sysbench_latency_avg"
	SysbenchLatencyAvgHelp = "Sysbench latency avg"

	SysbenchLatencyMaxName = "kubebench_sysbench_latency_max"
	SysbenchLatencyMaxHelp = "Sysbench latency max"

	SysbenchLatencyNinetyNinthName = "kubebench_sysbench_latency_ninety_ninth"
	SysbenchLatencyNinetyNinthHelp = "Sysbench latency ninety ninth"

	SysbenchLatencySumName = "kubebench_sysbench_latency_sum"
	SysbenchLatencySumHelp = "Sysbench latency sum"

	SysbenchEventsAvgName = "kubebench_sysbench_events_avg"
	SysbenchEventsAvgHelp = "Sysbench events avg"

	SysbenchEventsStddevName = "kubebench_sysbench_events_stddev"
	SysbenchEventsStddevHelp = "Sysbench events stddev"

	SysbenchExecTimeAvgName = "kubebench_sysbench_exec_time_avg"
	SysbenchExecTimeAvgHelp = "Sysbench exec time avg"

	SysbenchExecTimeStddevName = "kubebench_sysbench_exec_time_stddev"
	SysbenchExecTimeStddevHelp = "Sysbench exec time stddev"

	SysbenchThreadsName = "kubebench_sysbench_threads"
	SysbenchThreadsHelp = "Sysbench threads"

	SysbenchTpsSecondName = "kubebench_sysbench_tps_second"
	SysbenchTpsSecondHelp = "Sysbench tps every second"

	SysbenchQpsSecondName = "kubebench_sysbench_qps_second"
	SysbenchQpsSecondHelp = "Sysbench qps every second"

	SysbenchReadQpsSecondName = "kubebench_sysbench_read_qps_second"
	SysbenchReadQpsSecondHelp = "Sysbench read qps every second"

	SysbenchWriteQpsSecondName = "kubebench_sysbench_write_qps_second"
	SysbenchWriteQpsSecondHelp = "Sysbench write qps every second"

	SysbenchOtherQpsSecondName = "kubebench_sysbench_other_qps_second"
	SysbenchOtherQpsSecondHelp = "Sysbench other qps every second"

	SysbenchLatencySecondName = "kubebench_sysbench_latency_second"
	SysbenchLatencySecondHelp = "Sysbench latency every second"

	SysbenchErrorsSecondName = "kubebench_sysbench_errors_second"
	SysbenchErrorsSecondHelp = "Sysbench errors every second"

	SysbenchReconnectsSecondName = "kubebench_sysbench_reconnects_second"
	SysbenchReconnectsSecondHelp = "Sysbench reconnects every second"
)

var (
	SysbenchLabels   = []string{"benchmark", "name"}
	SysbenchGaugeMap = map[string]*prometheus.GaugeVec{}
)

// InitSysbench init sysbench metrics
func InitSysbench() {
	// Init Gauge
	SysbenchGaugeMap[SysbenchQueryReadName] = NewGauge(SysbenchQueryReadName, SysbenchQueryReadHelp, SysbenchLabels)
	SysbenchGaugeMap[SysbenchQueryWriteName] = NewGauge(SysbenchQueryWriteName, SysbenchQueryWriteHelp, SysbenchLabels)
	SysbenchGaugeMap[SysbenchQueryOtherName] = NewGauge(SysbenchQueryOtherName, SysbenchQueryOtherHelp, SysbenchLabels)
	SysbenchGaugeMap[SysbenchQueryTotalName] = NewGauge(SysbenchQueryTotalName, SysbenchQueryTotalHelp, SysbenchLabels)
	SysbenchGaugeMap[SysbenchTransactionsName] = NewGauge(SysbenchTransactionsName, SysbenchTransactionsHelp, SysbenchLabels)
	SysbenchGaugeMap[SysbenchQueriesName] = NewGauge(SysbenchQueriesName, SysbenchQueriesHelp, SysbenchLabels)
	SysbenchGaugeMap[SysbenchIgnoredErrorsName] = NewGauge(SysbenchIgnoredErrorsName, SysbenchIgnoredErrorsHelp, SysbenchLabels)
	SysbenchGaugeMap[SysbenchReconnectsName] = NewGauge(SysbenchReconnectsName, SysbenchReconnectsHelp, SysbenchLabels)
	SysbenchGaugeMap[SysbenchTotalEventsName] = NewGauge(SysbenchTotalEventsName, SysbenchTotalEventsHelp, SysbenchLabels)
	SysbenchGaugeMap[SysbenchTotalTimeName] = NewGauge(SysbenchTotalTimeName, SysbenchTotalTimeHelp, SysbenchLabels)
	SysbenchGaugeMap[SysbenchLatencyMinName] = NewGauge(SysbenchLatencyMinName, SysbenchLatencyMinHelp, SysbenchLabels)
	SysbenchGaugeMap[SysbenchLatencyAvgName] = NewGauge(SysbenchLatencyAvgName, SysbenchLatencyAvgHelp, SysbenchLabels)
	SysbenchGaugeMap[SysbenchLatencyMaxName] = NewGauge(SysbenchLatencyMaxName, SysbenchLatencyMaxHelp, SysbenchLabels)
	SysbenchGaugeMap[SysbenchLatencyNinetyNinthName] = NewGauge(SysbenchLatencyNinetyNinthName, SysbenchLatencyNinetyNinthHelp, SysbenchLabels)
	SysbenchGaugeMap[SysbenchLatencySumName] = NewGauge(SysbenchLatencySumName, SysbenchLatencySumHelp, SysbenchLabels)
	SysbenchGaugeMap[SysbenchEventsAvgName] = NewGauge(SysbenchEventsAvgName, SysbenchEventsAvgHelp, SysbenchLabels)
	SysbenchGaugeMap[SysbenchEventsStddevName] = NewGauge(SysbenchEventsStddevName, SysbenchEventsStddevHelp, SysbenchLabels)
	SysbenchGaugeMap[SysbenchExecTimeAvgName] = NewGauge(SysbenchExecTimeAvgName, SysbenchExecTimeAvgHelp, SysbenchLabels)
	SysbenchGaugeMap[SysbenchExecTimeStddevName] = NewGauge(SysbenchExecTimeStddevName, SysbenchExecTimeStddevHelp, SysbenchLabels)
	SysbenchGaugeMap[SysbenchThreadsName] = NewGauge(SysbenchThreadsName, SysbenchThreadsHelp, SysbenchLabels)
	SysbenchGaugeMap[SysbenchTpsSecondName] = NewGauge(SysbenchTpsSecondName, SysbenchTpsSecondHelp, SysbenchLabels)
	SysbenchGaugeMap[SysbenchQpsSecondName] = NewGauge(SysbenchQpsSecondName, SysbenchQpsSecondHelp, SysbenchLabels)
	SysbenchGaugeMap[SysbenchReadQpsSecondName] = NewGauge(SysbenchReadQpsSecondName, SysbenchReadQpsSecondHelp, SysbenchLabels)
	SysbenchGaugeMap[SysbenchWriteQpsSecondName] = NewGauge(SysbenchWriteQpsSecondName, SysbenchWriteQpsSecondHelp, SysbenchLabels)
	SysbenchGaugeMap[SysbenchOtherQpsSecondName] = NewGauge(SysbenchOtherQpsSecondName, SysbenchOtherQpsSecondHelp, SysbenchLabels)
	SysbenchGaugeMap[SysbenchLatencySecondName] = NewGauge(SysbenchLatencySecondName, SysbenchLatencySecondHelp, SysbenchLabels)
	SysbenchGaugeMap[SysbenchErrorsSecondName] = NewGauge(SysbenchErrorsSecondName, SysbenchErrorsSecondHelp, SysbenchLabels)
	SysbenchGaugeMap[SysbenchReconnectsSecondName] = NewGauge(SysbenchReconnectsSecondName, SysbenchReconnectsSecondHelp, SysbenchLabels)
}

// RegisterSysbenchMetrics register sysbench metrics
func RegisterSysbenchMetrics() {
	for _, v := range SysbenchGaugeMap {
		prometheus.MustRegister(v)
	}
}

type SysbenchResult struct {
	SQL             SQLStatistics           `json:"sql"`
	General         GeneralStatistics       `json:"general"`
	Latency         Latency                 `json:"latency"`
	ThreadsFairness ThreadsFairness         `json:"threadsFairness"`
	Transactions    int                     `json:"transactions"`
	Queries         int                     `json:"queries"`
	IgnoreErrors    int                     `json:"ignoreErrors"`
	Reconnects      int                     `json:"reconnects"`
	SecondResults   []*SysbenchSecondResult `json:"secondResults"`
}

type SysbenchSecondResult struct {
	Threads     int     `json:"threads"`
	TPS         float64 `json:"tps"`
	QPS         float64 `json:"qps"`
	Read        float64 `json:"read"`
	Write       float64 `json:"write"`
	Other       float64 `json:"other"`
	NinetyNinth float64 `json:"ninetyNinth"`
	Errors      float64 `json:"errs"`
	Reconnects  float64 `json:"reconnects"`
}

type SQLStatistics struct {
	Read  int `json:"read"`
	Write int `json:"write"`
	Other int `json:"other"`
	Total int `json:"total"`
}

type GeneralStatistics struct {
	TotalTime   float64 `json:"totalTime"`
	TotalEvents int     `json:"totalEvents"`
}

type Latency struct {
	Min         float64 `json:"min"`
	Avg         float64 `json:"avg"`
	Max         float64 `json:"max"`
	NinetyNinth float64 `json:"ninetyNinth"`
	Sum         float64 `json:"sum"`
}

type ThreadsFairness struct {
	EventsAvg    float64 `json:"eventsAvg"`
	EventsStddev float64 `json:"eventsStddev"`
	ExecTimeAvg  float64 `json:"execTimeAvg"`
	ExecTimeStd  float64 `json:"execTimeStd"`
}

func ScrapeSysbench(file, benchName, jobName string) {
	// read the file
	klog.Info("read file: ", file)
	data, err := os.ReadFile(file)
	if err != nil {
		fmt.Printf("read file error: %s\n", err)
		return
	}

	// parse the file
	result := ParseSysBenchResult(string(data))
	UpdateSysbenchMetrics(benchName, jobName, result)
}

func ParseSysBenchResult(msg string) *SysbenchResult {
	result := new(SysbenchResult)
	lines := strings.Split(msg, "\n")

	for _, l := range lines {
		switch {
		case sysbenchSecondRegex.MatchString(l):
			secondResult := ParseSysbenchSecondResult(l)
			result.SecondResults = append(result.SecondResults, secondResult)
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
			result.Transactions, _ = strconv.Atoi(transactions)
		case queriesRegex.MatchString(l):
			query := strings.TrimSpace(strings.Split(l, ":")[1])
			query = strings.TrimSpace(strings.Split(query, "(")[0])
			result.Queries, _ = strconv.Atoi(query)
		case ignoredErrorsRegex.MatchString(l):
			ignoreErrors := strings.TrimSpace(strings.Split(l, ":")[1])
			ignoreErrors = strings.TrimSpace(strings.Split(ignoreErrors, "(")[0])
			result.IgnoreErrors, _ = strconv.Atoi(ignoreErrors)
		case reconnectsRegex.MatchString(l):
			reconnects := strings.TrimSpace(strings.Split(l, ":")[1])
			reconnects = strings.TrimSpace(strings.Split(reconnects, "(")[0])
			result.Reconnects, _ = strconv.Atoi(reconnects)
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
		case latencyNinetyNinthRegex.MatchString(l):
			latencyNinetyNinth := strings.TrimSpace(strings.Split(l, ":")[1])
			result.Latency.NinetyNinth, _ = strconv.ParseFloat(latencyNinetyNinth, 64)
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

	return result
}

func ParseSysbenchSecondResult(msg string) *SysbenchSecondResult {
	//parse string like "[ 1s ] thds: 4 tps: 563.40 qps: 11319.87 (r/w/o: 7931.50/2257.58/1130.79) lat (ms,99%): 70.55 err/s: 0.00 reconn/s: 0.00"
	result := new(SysbenchSecondResult)

	thdsIndex := strings.Index(msg, "thds:")
	tpsIndex := strings.Index(msg, "tps:")
	qpsIndex := strings.Index(msg, "qps:")
	rwoIndex := strings.Index(msg, "(r/w/o:")
	latIndex := strings.Index(msg, "lat (ms,99%):")
	errIndex := strings.Index(msg, "err/s:")
	reconnIndex := strings.Index(msg, "reconn/s:")

	// parse thds
	thdsMsg := msg[thdsIndex:tpsIndex]
	thdsMsg = strings.TrimSpace(strings.Split(thdsMsg, ":")[1])
	result.Threads, _ = strconv.Atoi(thdsMsg)

	// parse tps
	tpsMsg := msg[tpsIndex:qpsIndex]
	tpsMsg = strings.TrimSpace(strings.Split(tpsMsg, ":")[1])
	result.TPS, _ = strconv.ParseFloat(tpsMsg, 64)

	// parse qps
	qpsMsg := msg[qpsIndex:rwoIndex]
	qpsMsg = strings.TrimSpace(strings.Split(qpsMsg, ":")[1])
	result.QPS, _ = strconv.ParseFloat(qpsMsg, 64)

	// parse r/w/o
	rwoMsg := msg[rwoIndex:latIndex]
	rwoMsg = strings.TrimSpace(strings.Split(rwoMsg, ":")[1])
	rwoMsg = strings.TrimSpace(strings.Split(rwoMsg, ")")[0])
	rMsg := strings.TrimSpace(strings.Split(rwoMsg, "/")[0])
	wMsg := strings.TrimSpace(strings.Split(rwoMsg, "/")[1])
	oMsg := strings.TrimSpace(strings.Split(rwoMsg, "/")[2])
	result.Read, _ = strconv.ParseFloat(rMsg, 64)
	result.Write, _ = strconv.ParseFloat(wMsg, 64)
	result.Other, _ = strconv.ParseFloat(oMsg, 64)

	// parse ninety-fifth latency
	latMsg := msg[latIndex:errIndex]
	latMsg = strings.TrimSpace(strings.Split(latMsg, ":")[1])
	result.NinetyNinth, _ = strconv.ParseFloat(latMsg, 64)

	// parse errors
	errMsg := msg[errIndex:reconnIndex]
	errMsg = strings.TrimSpace(strings.Split(errMsg, ":")[1])
	result.Errors, _ = strconv.ParseFloat(errMsg, 64)

	// parse reconnects
	reconnMsg := msg[reconnIndex:]
	reconnMsg = strings.TrimSpace(strings.Split(reconnMsg, ":")[1])
	result.Reconnects, _ = strconv.ParseFloat(reconnMsg, 64)

	return result
}

func UpdateSysbenchMetrics(benchName, jobName string, result *SysbenchResult) {
	value := []string{benchName, jobName}

	CommonCounterInc(benchName, jobName, Sysbench)

	// update total metrics
	SysbenchGaugeMap[SysbenchQueryReadName].WithLabelValues(value...).Set(float64(result.SQL.Read))
	SysbenchGaugeMap[SysbenchQueryWriteName].WithLabelValues(value...).Set(float64(result.SQL.Write))
	SysbenchGaugeMap[SysbenchQueryOtherName].WithLabelValues(value...).Set(float64(result.SQL.Other))
	SysbenchGaugeMap[SysbenchQueryTotalName].WithLabelValues(value...).Set(float64(result.SQL.Total))
	SysbenchGaugeMap[SysbenchTransactionsName].WithLabelValues(value...).Set(float64(result.Transactions))
	SysbenchGaugeMap[SysbenchQueriesName].WithLabelValues(value...).Set(float64(result.Queries))
	SysbenchGaugeMap[SysbenchIgnoredErrorsName].WithLabelValues(value...).Set(float64(result.IgnoreErrors))
	SysbenchGaugeMap[SysbenchReconnectsName].WithLabelValues(value...).Set(float64(result.Reconnects))
	SysbenchGaugeMap[SysbenchTotalEventsName].WithLabelValues(value...).Set(float64(result.General.TotalEvents))
	SysbenchGaugeMap[SysbenchTotalTimeName].WithLabelValues(value...).Set(result.General.TotalTime)
	SysbenchGaugeMap[SysbenchLatencyMinName].WithLabelValues(value...).Set(result.Latency.Min)
	SysbenchGaugeMap[SysbenchLatencyAvgName].WithLabelValues(value...).Set(result.Latency.Avg)
	SysbenchGaugeMap[SysbenchLatencyMaxName].WithLabelValues(value...).Set(result.Latency.Max)
	SysbenchGaugeMap[SysbenchLatencyNinetyNinthName].WithLabelValues(value...).Set(result.Latency.NinetyNinth)
	SysbenchGaugeMap[SysbenchLatencySumName].WithLabelValues(value...).Set(result.Latency.Sum)
	SysbenchGaugeMap[SysbenchEventsAvgName].WithLabelValues(value...).Set(result.ThreadsFairness.EventsAvg)
	SysbenchGaugeMap[SysbenchEventsStddevName].WithLabelValues(value...).Set(result.ThreadsFairness.EventsStddev)
	SysbenchGaugeMap[SysbenchExecTimeAvgName].WithLabelValues(value...).Set(result.ThreadsFairness.ExecTimeAvg)
	SysbenchGaugeMap[SysbenchExecTimeStddevName].WithLabelValues(value...).Set(result.ThreadsFairness.ExecTimeStd)
	klog.Info("update sysbench total metrics")

	// update second metrics
	for _, secondResult := range result.SecondResults {
		SysbenchGaugeMap[SysbenchThreadsName].WithLabelValues(value...).Set(float64(secondResult.Threads))
		SysbenchGaugeMap[SysbenchTpsSecondName].WithLabelValues(value...).Set(secondResult.TPS)
		SysbenchGaugeMap[SysbenchQpsSecondName].WithLabelValues(value...).Set(secondResult.QPS)
		SysbenchGaugeMap[SysbenchReadQpsSecondName].WithLabelValues(value...).Set(secondResult.Read)
		SysbenchGaugeMap[SysbenchWriteQpsSecondName].WithLabelValues(value...).Set(secondResult.Write)
		SysbenchGaugeMap[SysbenchOtherQpsSecondName].WithLabelValues(value...).Set(secondResult.Other)
		SysbenchGaugeMap[SysbenchLatencySecondName].WithLabelValues(value...).Set(secondResult.NinetyNinth)
		SysbenchGaugeMap[SysbenchErrorsSecondName].WithLabelValues(value...).Set(secondResult.Errors)
		SysbenchGaugeMap[SysbenchReconnectsSecondName].WithLabelValues(value...).Set(secondResult.Reconnects)

		// sleep 1 second to mock metrics collected every second
		klog.Info("update sysbench second metrics")
		time.Sleep(1 * time.Second)
	}
}
