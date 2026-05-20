package exporter

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/klog/v2"
)

const (
	EsrallyMetricValueName   = "kubebench_esrally_metric_value"
	EsrallyMetricValueHelp   = "Numeric Elastic Rally summary report value"
	esrallyUnavailablePrefix = "kubebench metrics unavailable:"
	esrallySummaryFallback   = "No numeric Rally CSV summary was found. Inspect the Esrally pod logs and configured report file for full output."
)

var (
	EsrallyLabels   = []string{"benchmark", "name", "metric", "task", "unit"}
	EsrallyGaugeMap = map[string]*prometheus.GaugeVec{}
)

type EsrallyMetric struct {
	Metric string
	Task   string
	Unit   string
	Value  float64
}

func InitEsrally() {
	EsrallyGaugeMap[EsrallyMetricValueName] = NewGauge(EsrallyMetricValueName, EsrallyMetricValueHelp, EsrallyLabels)
}

func RegisterEsrallyMetrics() {
	for _, gauge := range EsrallyGaugeMap {
		prometheus.MustRegister(gauge)
	}
}

func ParseEsrallyCSV(msg string) []EsrallyMetric {
	reader := csv.NewReader(strings.NewReader(extractEsrallyCSV(msg)))
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true

	rows, err := reader.ReadAll()
	if err != nil || len(rows) == 0 {
		return nil
	}

	metricIdx, taskIdx, valueIdx, unitIdx, start := esrallyCSVIndexes(rows[0])
	metrics := make([]EsrallyMetric, 0)
	for _, row := range rows[start:] {
		if len(row) <= metricIdx || len(row) <= valueIdx {
			continue
		}
		valueText := strings.TrimSpace(row[valueIdx])
		if valueText == "" || strings.EqualFold(valueText, "n/a") || valueText == "-" {
			continue
		}
		value, err := strconv.ParseFloat(valueText, 64)
		if err != nil {
			continue
		}

		metric := EsrallyMetric{
			Metric: strings.TrimSpace(row[metricIdx]),
			Value:  value,
		}
		if taskIdx >= 0 && len(row) > taskIdx {
			metric.Task = strings.TrimSpace(row[taskIdx])
		}
		if unitIdx >= 0 && len(row) > unitIdx {
			metric.Unit = strings.TrimSpace(row[unitIdx])
		}
		if metric.Metric == "" {
			continue
		}
		metrics = append(metrics, metric)
	}

	return metrics
}

func esrallyCSVIndexes(header []string) (metricIdx, taskIdx, valueIdx, unitIdx, start int) {
	metricIdx, taskIdx, valueIdx, unitIdx = 0, 1, 2, 3
	start = 0
	for i, h := range header {
		switch strings.ToLower(strings.TrimSpace(h)) {
		case "metric":
			metricIdx = i
			start = 1
		case "task":
			taskIdx = i
			start = 1
		case "value":
			valueIdx = i
			start = 1
		case "unit":
			unitIdx = i
			start = 1
		}
	}
	return metricIdx, taskIdx, valueIdx, unitIdx, start
}

func extractEsrallyCSV(msg string) string {
	const marker = "Rally CSV report:"
	if idx := strings.Index(msg, marker); idx >= 0 {
		return msg[idx+len(marker):]
	}

	lines := strings.Split(msg, "\n")
	for i, line := range lines {
		if isEsrallyCSVHeader(line) {
			return strings.Join(lines[i:], "\n")
		}
	}
	return ""
}

func SummarizeEsrallyCSV(msg string, limit int) string {
	metrics := ParseEsrallyCSV(msg)
	if len(metrics) == 0 {
		return strings.Join(append([]string{esrallySummaryFallback}, esrallyMetricsUnavailableMessages(msg)...), "\n")
	}
	metrics = prioritizeEsrallyMetrics(metrics)
	if limit <= 0 || limit > len(metrics) {
		limit = len(metrics)
	}

	lines := make([]string, 0, limit)
	for _, metric := range metrics[:limit] {
		label := metric.Metric
		if metric.Task != "" {
			label = fmt.Sprintf("%s [%s]", label, metric.Task)
		}
		value := strconv.FormatFloat(metric.Value, 'f', -1, 64)
		if metric.Unit != "" {
			value = fmt.Sprintf("%s %s", value, metric.Unit)
		}
		lines = append(lines, fmt.Sprintf("%s: %s", label, value))
	}
	lines = append(lines, esrallyMetricsUnavailableMessages(msg)...)
	return strings.Join(lines, "\n")
}

func esrallyMetricsUnavailableMessages(msg string) []string {
	messages := make([]string, 0)
	for _, line := range strings.Split(msg, "\n") {
		line = strings.TrimSpace(line)
		if isKnownEsrallyUnavailableMessage(line) {
			messages = append(messages, line)
		}
	}
	return messages
}

func isEsrallyCSVHeader(line string) bool {
	fields := strings.Split(strings.TrimSpace(line), ",")
	if len(fields) < 3 {
		return false
	}

	hasMetric := false
	hasValue := false
	for _, field := range fields {
		switch strings.ToLower(strings.TrimSpace(field)) {
		case "metric":
			hasMetric = true
		case "value":
			hasValue = true
		}
	}
	return hasMetric && hasValue
}

func prioritizeEsrallyMetrics(metrics []EsrallyMetric) []EsrallyMetric {
	prioritized := append([]EsrallyMetric(nil), metrics...)
	sort.SliceStable(prioritized, func(i, j int) bool {
		return esrallyMetricPriority(prioritized[i]) < esrallyMetricPriority(prioritized[j])
	})
	return prioritized
}

func esrallyMetricPriority(metric EsrallyMetric) int {
	name := strings.ToLower(metric.Metric)
	switch {
	case strings.Contains(name, "throughput"):
		return 0
	case strings.Contains(name, "latency"):
		return 1
	case strings.Contains(name, "service time"):
		return 2
	case strings.Contains(name, "processing time"):
		return 3
	case strings.Contains(name, "error rate"):
		return 4
	default:
		return 5
	}
}

func isKnownEsrallyUnavailableMessage(line string) bool {
	if !strings.HasPrefix(line, esrallyUnavailablePrefix) {
		return false
	}

	switch line {
	case "kubebench metrics unavailable: spec.metrics is false",
		"kubebench metrics unavailable: the exporter only supports reportFormat csv",
		"kubebench metrics unavailable: reportFile must be under /var/log for the exporter shared volume":
		return true
	default:
		return false
	}
}

func ScrapeEsrally(file, doneFile, benchName, jobName string) {
	klog.Infof("read esrally report file %s", file)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		if content, err := os.ReadFile(file); err == nil && len(content) > 0 {
			if UpdateEsrallyMetrics(benchName, jobName, string(content)) {
				return
			}
		}
		if doneFile != "" {
			if _, err := os.Stat(doneFile); err == nil {
				klog.Infof("esrally done marker found before report metrics were parsed: %s", doneFile)
				return
			}
		}
		<-ticker.C
	}
}

func UpdateEsrallyMetrics(benchName, jobName, msg string) bool {
	metrics := ParseEsrallyCSV(msg)
	if len(metrics) == 0 {
		return false
	}

	CommonCounterInc(benchName, jobName, Esrally)
	for _, metric := range metrics {
		EsrallyGaugeMap[EsrallyMetricValueName].
			WithLabelValues(benchName, jobName, metric.Metric, metric.Task, metric.Unit).
			Set(metric.Value)
	}
	klog.Info("update esrally metrics")
	return true
}
