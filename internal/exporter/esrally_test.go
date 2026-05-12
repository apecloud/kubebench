package exporter

import (
	"os"
	"strings"
	"testing"
)

func TestParseEsrallyCSV(t *testing.T) {
	msg, err := os.ReadFile("testdata/esrally.csv")
	if err != nil {
		t.Fatal(err)
	}

	result := ParseEsrallyCSV(string(msg))
	if len(result) != 4 {
		t.Fatalf("expected 4 numeric metrics, got %d", len(result))
	}

	if result[0].Metric != "Min Throughput" || result[0].Task != "index-append" || result[0].Unit != "docs/s" || result[0].Value != 1000 {
		t.Fatalf("unexpected first metric: %#v", result[0])
	}
	if result[2].Metric != "Median Latency" || result[2].Value != 12.75 {
		t.Fatalf("unexpected latency metric: %#v", result[2])
	}
	if result[3].Metric != "Error rate" || result[3].Value != 0 {
		t.Fatalf("unexpected zero metric: %#v", result[3])
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
