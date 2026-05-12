package exporter

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/klog/v2"
)

const (
	EsrallyMetricValueName = "kubebench_esrally_metric_value"
	EsrallyMetricValueHelp = "Numeric Elastic Rally summary report value"
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
	return msg
}

func SummarizeEsrallyCSV(msg string, limit int) string {
	metrics := ParseEsrallyCSV(msg)
	if len(metrics) == 0 {
		return strings.TrimSpace(msg)
	}
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
	return strings.Join(lines, "\n")
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
