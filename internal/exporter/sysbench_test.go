package exporter

import (
	"os"
	"testing"
)

func TestParseSysBenchResult(t *testing.T) {
	testcase := []struct {
		path     string
		expected *SysbenchResult
	}{
		{
			path: "testdata/sysbench.txt",
			expected: &SysbenchResult{
				SQL: SQLStatistics{
					Read:  110110,
					Write: 31450,
					Other: 15734,
					Total: 157294,
				},
				General: GeneralStatistics{
					TotalTime:   20.0241,
					TotalEvents: 7862,
				},
				Latency: Latency{
					Min:         0.96,
					Avg:         10.17,
					Max:         184.54,
					NinetyFifth: 75.82,
					Sum:         79990.39,
				},
				ThreadsFairness: ThreadsFairness{
					EventsAvg:    1965.5000,
					EventsStddev: 19.50,
					ExecTimeAvg:  19.9976,
					ExecTimeStd:  0.00,
				},
				Transactions: 7862,
				Queries:      157294,
				IgnoreErrors: 3,
				Reconnects:   0,
				SecondResults: []*SysbenchSecondResult{
					{
						Threads:     4,
						TPS:         563.40,
						QPS:         11319.87,
						Read:        7931.50,
						Write:       2257.58,
						Other:       1130.79,
						Errors:      0.00,
						NinetyFifth: 70.55,
						Reconnects:  0.00,
					},
				},
			},
		},
	}

	for _, tc := range testcase {
		msg, _ := os.ReadFile(tc.path)
		result := ParseSysBenchResult(string(msg))

		if result.SQL.Read != tc.expected.SQL.Read {
			t.Errorf("Expected %d, got %d", tc.expected.SQL.Read, result.SQL.Read)
		}

		if result.SQL.Write != tc.expected.SQL.Write {
			t.Errorf("Expected %d, got %d", tc.expected.SQL.Write, result.SQL.Write)
		}

		if result.SQL.Other != tc.expected.SQL.Other {
			t.Errorf("Expected %d, got %d", tc.expected.SQL.Other, result.SQL.Other)
		}

		if result.SQL.Total != tc.expected.SQL.Total {
			t.Errorf("Expected %d, got %d", tc.expected.SQL.Total, result.SQL.Total)
		}

		if result.General.TotalTime != tc.expected.General.TotalTime {
			t.Errorf("Expected %f, got %f", tc.expected.General.TotalTime, result.General.TotalTime)
		}

		if result.General.TotalEvents != tc.expected.General.TotalEvents {
			t.Errorf("Expected %d, got %d", tc.expected.General.TotalEvents, result.General.TotalEvents)
		}

		if result.Latency.Min != tc.expected.Latency.Min {
			t.Errorf("Expected %f, got %f", tc.expected.Latency.Min, result.Latency.Min)
		}

		if result.Latency.Avg != tc.expected.Latency.Avg {
			t.Errorf("Expected %f, got %f", tc.expected.Latency.Avg, result.Latency.Avg)
		}

		if result.Latency.Max != tc.expected.Latency.Max {
			t.Errorf("Expected %f, got %f", tc.expected.Latency.Max, result.Latency.Max)
		}

		if result.Latency.NinetyFifth != tc.expected.Latency.NinetyFifth {
			t.Errorf("Expected %f, got %f", tc.expected.Latency.NinetyFifth, result.Latency.NinetyFifth)
		}

		if result.Latency.Sum != tc.expected.Latency.Sum {
			t.Errorf("Expected %f, got %f", tc.expected.Latency.Sum, result.Latency.Sum)
		}

		if result.ThreadsFairness.EventsAvg != tc.expected.ThreadsFairness.EventsAvg {
			t.Errorf("Expected %f, got %f", tc.expected.ThreadsFairness.EventsAvg, result.ThreadsFairness.EventsAvg)
		}

		if result.ThreadsFairness.EventsStddev != tc.expected.ThreadsFairness.EventsStddev {
			t.Errorf("Expected %f, got %f", tc.expected.ThreadsFairness.EventsStddev, result.ThreadsFairness.EventsStddev)
		}

		if result.ThreadsFairness.ExecTimeAvg != tc.expected.ThreadsFairness.ExecTimeAvg {
			t.Errorf("Expected %f, got %f", tc.expected.ThreadsFairness.ExecTimeAvg, result.ThreadsFairness.ExecTimeAvg)
		}

		if result.ThreadsFairness.ExecTimeStd != tc.expected.ThreadsFairness.ExecTimeStd {
			t.Errorf("Expected %f, got %f", tc.expected.ThreadsFairness.ExecTimeStd, result.ThreadsFairness.ExecTimeStd)
		}

		if result.Transactions != tc.expected.Transactions {
			t.Errorf("Expected %d, got %d", tc.expected.Transactions, result.Transactions)
		}

		if result.Queries != tc.expected.Queries {
			t.Errorf("Expected %d, got %d", tc.expected.Queries, result.Queries)
		}

		if result.IgnoreErrors != tc.expected.IgnoreErrors {
			t.Errorf("Expected %d, got %d", tc.expected.IgnoreErrors, result.IgnoreErrors)
		}

		if result.Reconnects != tc.expected.Reconnects {
			t.Errorf("Expected %d, got %d", tc.expected.Reconnects, result.Reconnects)
		}

		if len(result.SecondResults) != len(tc.expected.SecondResults) {
			t.Errorf("Expected %d, got %d", len(tc.expected.SecondResults), len(result.SecondResults))
		}

		for i, sr := range result.SecondResults {
			if sr.Threads != tc.expected.SecondResults[i].Threads {
				t.Errorf("Expected %d, got %d", tc.expected.SecondResults[i].Threads, sr.Threads)
			}

			if sr.TPS != tc.expected.SecondResults[i].TPS {
				t.Errorf("Expected %f, got %f", tc.expected.SecondResults[i].TPS, sr.TPS)
			}

			if sr.QPS != tc.expected.SecondResults[i].QPS {
				t.Errorf("Expected %f, got %f", tc.expected.SecondResults[i].QPS, sr.QPS)
			}

			if sr.Read != tc.expected.SecondResults[i].Read {
				t.Errorf("Expected %f, got %f", tc.expected.SecondResults[i].Read, sr.Read)
			}

			if sr.Write != tc.expected.SecondResults[i].Write {
				t.Errorf("Expected %f, got %f", tc.expected.SecondResults[i].Write, sr.Write)
			}

			if sr.Other != tc.expected.SecondResults[i].Other {
				t.Errorf("Expected %f, got %f", tc.expected.SecondResults[i].Other, sr.Other)
			}

			if sr.Errors != tc.expected.SecondResults[i].Errors {
				t.Errorf("Expected %f, got %f", tc.expected.SecondResults[i].Errors, sr.Errors)
			}

			if sr.NinetyFifth != tc.expected.SecondResults[i].NinetyFifth {
				t.Errorf("Expected %f, got %f", tc.expected.SecondResults[i].NinetyFifth, sr.NinetyFifth)
			}

			if sr.Reconnects != tc.expected.SecondResults[i].Reconnects {
				t.Errorf("Expected %f, got %f", tc.expected.SecondResults[i].Reconnects, sr.Reconnects)
			}
		}
	}
}

func TestParseSecondsResult(t *testing.T) {
	testcase := []struct {
		input    string
		expected *SysbenchSecondResult
	}{
		{
			input: "[ 1s ] thds: 4 tps: 563.40 qps: 11319.87 (r/w/o: 7931.50/2257.58/1130.79) lat (ms,95%): 70.55 err/s: 0.00 reconn/s: 0.00",
			expected: &SysbenchSecondResult{
				Threads:     4,
				TPS:         563.40,
				QPS:         11319.87,
				Read:        7931.50,
				Write:       2257.58,
				Other:       1130.79,
				Errors:      0.00,
				NinetyFifth: 70.55,
				Reconnects:  0.00,
			},
		},
		{
			input: "[ 20s ] thds: 10 tps: 368.98 qps: 7389.62 (r/w/o: 5171.74/1479.92/737.96) lat (ms,95%): 74.46 err/s: 0.00 reconn/s: 0.00",
			expected: &SysbenchSecondResult{
				Threads:     10,
				TPS:         368.98,
				QPS:         7389.62,
				Read:        5171.74,
				Write:       1479.92,
				Other:       737.96,
				Errors:      0.00,
				NinetyFifth: 74.46,
				Reconnects:  0.00,
			},
		},
	}

	for _, tc := range testcase {
		result := ParseSysbenchSecondResult(tc.input)
		if result.Threads != tc.expected.Threads {
			t.Errorf("expected %d, got %d", tc.expected.Threads, result.Threads)
		}
		if result.TPS != tc.expected.TPS {
			t.Errorf("expected %f, got %f", tc.expected.TPS, result.TPS)
		}
		if result.QPS != tc.expected.QPS {
			t.Errorf("expected %f, got %f", tc.expected.QPS, result.QPS)
		}
		if result.Read != tc.expected.Read {
			t.Errorf("expected %f, got %f", tc.expected.Read, result.Read)
		}
		if result.Write != tc.expected.Write {
			t.Errorf("expected %f, got %f", tc.expected.Write, result.Write)
		}
		if result.Other != tc.expected.Other {
			t.Errorf("expected %f, got %f", tc.expected.Other, result.Other)
		}
		if result.Errors != tc.expected.Errors {
			t.Errorf("expected %f, got %f", tc.expected.Errors, result.Errors)
		}
		if result.NinetyFifth != tc.expected.NinetyFifth {
			t.Errorf("expected %f, got %f", tc.expected.NinetyFifth, result.NinetyFifth)
		}
		if result.Reconnects != tc.expected.Reconnects {
			t.Errorf("expected %f, got %f", tc.expected.Reconnects, result.Reconnects)
		}
	}
}
