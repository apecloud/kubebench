package exporter

import (
	"os"
	"testing"
)

func TestParsePgbenchResult(t *testing.T) {
	testcase := []struct {
		path     string
		expected *PgbenchResult
	}{
		{
			path: "testdata/pgbench.txt",
			expected: &PgbenchResult{
				Scale:                  100,
				QueryMode:              "simple",
				Clients:                2,
				Threads:                1,
				MaximumTry:             1,
				TransactionsPerClient:  11291,
				TransactionsProcessed:  22583,
				TransactionsFailed:     0,
				AvgLatency:             2.647,
				StdLatency:             11.912,
				InitialConnectionsTime: 70.124,
				TPS:                    754.431143,
			},
		},
	}

	for _, tc := range testcase {
		msg, _ := os.ReadFile(tc.path)
		reuslt := ParsePgbenchResult(string(msg))

		if reuslt.Scale != tc.expected.Scale {
			t.Errorf("Expected scale is %d, got %d", tc.expected.Scale, reuslt.Scale)
		}

		if reuslt.QueryMode != tc.expected.QueryMode {
			t.Errorf("Expected query mode is %s, got %s", tc.expected.QueryMode, reuslt.QueryMode)
		}

		if reuslt.Clients != tc.expected.Clients {
			t.Errorf("Expected clients is %d, got %d", tc.expected.Clients, reuslt.Clients)
		}

		if reuslt.Threads != tc.expected.Threads {
			t.Errorf("Expected threads is %d, got %d", tc.expected.Threads, reuslt.Threads)
		}

		if reuslt.MaximumTry != tc.expected.MaximumTry {
			t.Errorf("Expected maximum try is %d, got %d", tc.expected.MaximumTry, reuslt.MaximumTry)
		}

		if reuslt.TransactionsPerClient != tc.expected.TransactionsPerClient {
			t.Errorf("Expected transcations per client is %d, got %d", tc.expected.TransactionsPerClient, reuslt.TransactionsPerClient)
		}

		if reuslt.TransactionsProcessed != tc.expected.TransactionsProcessed {
			t.Errorf("Expected transcations processed is %d, got %d", tc.expected.TransactionsProcessed, reuslt.TransactionsProcessed)
		}

		if reuslt.TransactionsFailed != tc.expected.TransactionsFailed {
			t.Errorf("Expected transcations failed is %d, got %d", tc.expected.TransactionsFailed, reuslt.TransactionsFailed)
		}

		if reuslt.AvgLatency != tc.expected.AvgLatency {
			t.Errorf("Expected avg latency is %f, got %f", tc.expected.AvgLatency, reuslt.AvgLatency)
		}

		if reuslt.StdLatency != tc.expected.StdLatency {
			t.Errorf("Expected std latency is %f, got %f", tc.expected.StdLatency, reuslt.StdLatency)
		}

		if reuslt.InitialConnectionsTime != tc.expected.InitialConnectionsTime {
			t.Errorf("Expected initial connections time is %f, got %f", tc.expected.InitialConnectionsTime, reuslt.InitialConnectionsTime)
		}

		if reuslt.TPS != tc.expected.TPS {
			t.Errorf("Expected tps is %f, got %f", tc.expected.TPS, reuslt.TPS)
		}

	}
}

func TestParsePgbenchSecondResult(t *testing.T) {
	testcase := []struct {
		input    string
		expected *PgbenchSecondResult
	}{
		{
			input: "progress: 1.0 s, 610.0 tps, lat 3.043 ms stddev 8.900, 0 failed",
			expected: &PgbenchSecondResult{
				TPS:                   610.0,
				AvgLatency:            3.043,
				StdLatency:            8.900,
				FailedTransactionsSum: 0,
			},
		},
		{
			input: "progress: 24.0 s, 200.3 tps, lat 16.081 ms stddev 94.222, 10 failed",
			expected: &PgbenchSecondResult{
				TPS:                   200.3,
				AvgLatency:            16.081,
				StdLatency:            94.222,
				FailedTransactionsSum: 10,
			},
		},
	}

	for _, tc := range testcase {
		result := ParsePgbenchSecondResult(tc.input)
		if result.TPS != tc.expected.TPS {
			t.Errorf("Expected %f, got %f", tc.expected.TPS, result.TPS)
		}

		if result.AvgLatency != tc.expected.AvgLatency {
			t.Errorf("Expected %f, got %f", tc.expected.AvgLatency, result.AvgLatency)
		}

		if result.StdLatency != tc.expected.StdLatency {
			t.Errorf("Expected %f, got %f", tc.expected.StdLatency, result.StdLatency)
		}

		if result.FailedTransactionsSum != tc.expected.FailedTransactionsSum {
			t.Errorf("Expected %d, got %d", tc.expected.FailedTransactionsSum, result.FailedTransactionsSum)
		}
	}
}
