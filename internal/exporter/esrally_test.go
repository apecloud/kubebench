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

func TestParseEsrallyCSVIgnoresMarkdownReport(t *testing.T) {
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
	if !strings.Contains(summary, "kubebench metrics unavailable") {
		t.Fatalf("summary should keep explicit metrics unavailable message: %s", summary)
	}
}
