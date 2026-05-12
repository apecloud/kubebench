package exporter

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestParseEsrallyCSV(t *testing.T) {
	msg, err := os.ReadFile("testdata/esrally.csv")
	if err != nil {
		t.Fatal(err)
	}

	result := ParseEsrallyCSV(string(msg))
	if len(result) != 8 {
		t.Fatalf("expected 8 numeric metrics, got %d", len(result))
	}

	if result[0].Metric != "Min Throughput" || result[0].Task != "index-append" || result[0].Unit != "docs/s" || result[0].Value != 1000 {
		t.Fatalf("unexpected first metric: %#v", result[0])
	}
	if result[3].Metric != "Median Latency" || result[3].Value != 12.75 {
		t.Fatalf("unexpected latency metric: %#v", result[3])
	}
	if result[5].Metric != "Error rate" || result[5].Value != 0 {
		t.Fatalf("unexpected zero metric: %#v", result[5])
	}
	if result[6].Metric != "Total Young Gen GC Time" || result[6].Task != "" || result[6].Unit != "s" || result[6].Value != 3.4 {
		t.Fatalf("unexpected empty-task metric: %#v", result[6])
	}
}

func TestParseEsrallyCSVRows(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []EsrallyMetric
	}{
		{
			name: "skips empty unavailable and non numeric values",
			input: `Metric,Task,Value,Unit
Mean Throughput,index-append,42.5,docs/s
Service Time,default,N/A,ms
Warnings,default,,count
Processing time,default,not-a-number,ms`,
			want: []EsrallyMetric{
				{Metric: "Mean Throughput", Task: "index-append", Value: 42.5, Unit: "docs/s"},
			},
		},
		{
			name: "keeps task and unit labels from extra-column rally csv",
			input: `Metric,Task,Lap,Value,Unit
90th percentile latency,search,2,25.5,ms`,
			want: []EsrallyMetric{
				{Metric: "90th percentile latency", Task: "search", Value: 25.5, Unit: "ms"},
			},
		},
		{
			name:  "parses marker-prefixed rows with missing header",
			input: "Rally CSV report:\nMedian Latency,default,12.75,ms\nError rate,default,0,%",
			want: []EsrallyMetric{
				{Metric: "Median Latency", Task: "default", Value: 12.75, Unit: "ms"},
				{Metric: "Error rate", Task: "default", Value: 0, Unit: "%"},
			},
		},
		{
			name: "returns no metrics when required headers are absent in raw logs",
			input: `race id: abc
name,task,result,unit
Median Latency,default,12.75,ms`,
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseEsrallyCSV(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("expected %d metrics, got %d: %#v", len(tt.want), len(got), got)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Fatalf("metric %d: expected %#v, got %#v", i, tt.want[i], got[i])
				}
			}
		})
	}
}

func TestSummarizeEsrallyCSV(t *testing.T) {
	msg, err := os.ReadFile("testdata/esrally.csv")
	if err != nil {
		t.Fatal(err)
	}

	summary := SummarizeEsrallyCSV("Rally CSV report:\n"+string(msg), 2)
	if !strings.Contains(summary, "Min Throughput [index-append]: 1000 docs/s") {
		t.Fatalf("summary missing throughput: %s", summary)
	}
	if !strings.Contains(summary, "Mean Throughput [index-append]: 1200.5 docs/s") {
		t.Fatalf("summary missing mean throughput: %s", summary)
	}
	if strings.Contains(summary, "N/A") {
		t.Fatalf("summary should not include unavailable values: %s", summary)
	}
}

func TestSummarizeEsrallyCSVKeepsMetricsUnavailableReason(t *testing.T) {
	msg, err := os.ReadFile("testdata/esrally.csv")
	if err != nil {
		t.Fatal(err)
	}

	summary := SummarizeEsrallyCSV("Rally CSV report:\n"+string(msg)+"\nkubebench metrics unavailable: spec.metrics is false\n", 1)
	if !strings.Contains(summary, "Min Throughput [index-append]: 1000 docs/s") {
		t.Fatalf("summary missing throughput: %s", summary)
	}
	if !strings.Contains(summary, "kubebench metrics unavailable: spec.metrics is false") {
		t.Fatalf("summary missing metrics unavailable reason: %s", summary)
	}
}

func TestSummarizeEsrallyCSVMissingReportReturnsBoundedMessage(t *testing.T) {
	msg := `race id: 123
clientOptions=basic_auth_password:'super-secret',api_key:'abc123'
benchmark completed without shared csv report`

	summary := SummarizeEsrallyCSV(msg, 2)
	if !strings.Contains(summary, "No numeric Rally CSV summary was found.") {
		t.Fatalf("summary should explain the missing csv summary: %s", summary)
	}
	if strings.Contains(summary, "super-secret") || strings.Contains(summary, "abc123") {
		t.Fatalf("summary should not leak credential-like values: %s", summary)
	}
}

func TestSummarizeEsrallyCSVIgnoresMarkdownReport(t *testing.T) {
	msg := `Rally markdown report (kubebench metrics unavailable):

|   Metric |   Task |   Value |   Unit |
|---------:|-------:|--------:|-------:|
|      Min Throughput | index-append | 1000 | docs/s |
kubebench metrics unavailable: the exporter only supports reportFormat csv`

	result := ParseEsrallyCSV(msg)
	if len(result) != 0 {
		t.Fatalf("expected markdown report to produce no CSV metrics, got %#v", result)
	}

	summary := SummarizeEsrallyCSV(msg, 2)
	if !strings.Contains(summary, "No numeric Rally CSV summary was found.") {
		t.Fatalf("summary should explain the unsupported report summary: %s", summary)
	}
	if !strings.Contains(summary, "kubebench metrics unavailable: the exporter only supports reportFormat csv") {
		t.Fatalf("summary should keep explicit metrics unavailable message: %s", summary)
	}
}

func TestSummarizeEsrallyCSVUnparsableCSVReturnsBoundedMessage(t *testing.T) {
	msg := `Rally CSV report:
Metric,Task,Value,Unit
Mean Throughput,index-append,not-a-number,docs/s
Error rate,default,NaN?,"%"`

	summary := SummarizeEsrallyCSV(msg, 2)
	if !strings.Contains(summary, "No numeric Rally CSV summary was found.") {
		t.Fatalf("summary should explain the unparsable csv summary: %s", summary)
	}
	if strings.Contains(summary, "not-a-number") {
		t.Fatalf("summary should not echo unparsable csv values: %s", summary)
	}
}

func TestSummarizeEsrallyCSVOnlyKeepsKnownSafeUnavailableReasons(t *testing.T) {
	msg := `Rally CSV report:
Metric,Task,Value,Unit
Warnings,default,,count
kubebench metrics unavailable: reportFile must be under /var/log for the exporter shared volume
kubebench metrics unavailable: clientOptions basic_auth_password:'super-secret'`

	summary := SummarizeEsrallyCSV(msg, 2)
	if !strings.Contains(summary, "reportFile must be under /var/log for the exporter shared volume") {
		t.Fatalf("summary should keep the known-safe metrics unavailable reason: %s", summary)
	}
	if strings.Contains(summary, "super-secret") {
		t.Fatalf("summary should not keep unknown credential-bearing lines: %s", summary)
	}
}

func TestUpdateEsrallyMetricsSetsExpectedLabels(t *testing.T) {
	resetEsrallyTestMetrics()

	if !UpdateEsrallyMetrics("bench-a", "rally-run", `Metric,Task,Value,Unit
Mean Throughput,index-append,1200.5,docs/s`) {
		t.Fatal("expected metrics to be updated")
	}

	gauge := EsrallyGaugeMap[EsrallyMetricValueName].WithLabelValues("bench-a", "rally-run", "Mean Throughput", "index-append", "docs/s")
	if got := testutil.ToFloat64(gauge); got != 1200.5 {
		t.Fatalf("expected gauge value 1200.5, got %f", got)
	}

	counter := KubebenchCounter.WithLabelValues("bench-a", "rally-run", Esrally)
	if got := testutil.ToFloat64(counter); got != 1 {
		t.Fatalf("expected common counter value 1, got %f", got)
	}
}

func TestUpdateEsrallyMetricsReturnsFalseWithoutNumericRows(t *testing.T) {
	resetEsrallyTestMetrics()

	if UpdateEsrallyMetrics("bench-a", "rally-run", `Metric,Task,Value,Unit
Service Time,default,N/A,ms
Warnings,default,,count`) {
		t.Fatal("expected no update for csv without numeric rows")
	}
}

func TestScrapeEsrallyStopsWhenDoneFileExistsWithoutMetrics(t *testing.T) {
	resetEsrallyTestMetrics()
	dir := t.TempDir()
	reportFile := dir + "/esrally-report.csv"
	doneFile := dir + "/esrally.exit"
	if err := os.WriteFile(reportFile, []byte("Metric,Task,Value,Unit\nWarnings,default,,count\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(doneFile, []byte("1\n"), 0644); err != nil {
		t.Fatal(err)
	}

	start := time.Now()
	ScrapeEsrally(reportFile, doneFile, "bench-a", "rally-run")
	if elapsed := time.Since(start); elapsed > 200*time.Millisecond {
		t.Fatalf("expected done-file scrape to return without waiting for ticker, took %s", elapsed)
	}
	if got := testutil.ToFloat64(KubebenchCounter.WithLabelValues("bench-a", "rally-run", Esrally)); got != 0 {
		t.Fatalf("expected no common counter increment, got %f", got)
	}
}

func TestScrapeEsrallyUpdatesMetricsFromExistingReport(t *testing.T) {
	resetEsrallyTestMetrics()
	dir := t.TempDir()
	reportFile := dir + "/esrally-report.csv"
	if err := os.WriteFile(reportFile, []byte("Metric,Task,Value,Unit\nMean Throughput,index-append,99.9,docs/s\n"), 0644); err != nil {
		t.Fatal(err)
	}

	ScrapeEsrally(reportFile, "", "bench-a", "rally-run")

	gauge := EsrallyGaugeMap[EsrallyMetricValueName].WithLabelValues("bench-a", "rally-run", "Mean Throughput", "index-append", "docs/s")
	if got := testutil.ToFloat64(gauge); got != 99.9 {
		t.Fatalf("expected gauge value 99.9, got %f", got)
	}
}

func resetEsrallyTestMetrics() {
	KubebenchCounter = NewCounter(KubebenchTotalName, KubebenchTotalHelp, KubebenchTotalLabels)
	EsrallyGaugeMap = map[string]*prometheus.GaugeVec{}
	InitEsrally()
}
