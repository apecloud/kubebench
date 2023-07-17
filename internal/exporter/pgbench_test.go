package exporter

import "testing"

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
				TransactionsPerClient:  22583,
				TransactionsProcessed:  22583,
				TransactionsFailed:     0,
				AvgLatency:             2.647,
				StdLatency:             11.912,
				InitialConnectionsTime: 70.124,
				TPS:                    754.431143,
				SecondResults: []*PgbenchSecondResult{
					{
						TPS:                   610.0,
						AvgLatency:            3.043,
						StdLatency:            8.900,
						FailedTransactionsSum: 0,
					},
				},
			},
		},
	}

	for _, tc := range testcase {
		for i, sr := range tc.expected.SecondResults {
			if sr.TPS != tc.expected.SecondResults[i].TPS {
				t.Errorf("Expected %f, got %f", tc.expected.SecondResults[i].TPS, sr.TPS)
			}

			if sr.AvgLatency != tc.expected.SecondResults[i].AvgLatency {
				t.Errorf("Expected %f, got %f", tc.expected.SecondResults[i].AvgLatency, sr.AvgLatency)
			}

			if sr.StdLatency != tc.expected.SecondResults[i].StdLatency {
				t.Errorf("Expected %f, got %f", tc.expected.SecondResults[i].StdLatency, sr.StdLatency)
			}

			if sr.FailedTransactionsSum != tc.expected.SecondResults[i].FailedTransactionsSum {
				t.Errorf("Expected %d, got %d", tc.expected.SecondResults[i].FailedTransactionsSum, sr.FailedTransactionsSum)
			}
		}

		if tc.expected.Scale != tc.expected.Scale {
			t.Errorf("Expected %d, got %d", tc.expected.Scale, tc.expected.Scale)
		}

		if tc.expected.QueryMode != tc.expected.QueryMode {
			t.Errorf("Expected %s, got %s", tc.expected.QueryMode, tc.expected.QueryMode)
		}

		if tc.expected.Clients != tc.expected.Clients {
			t.Errorf("Expected %d, got %d", tc.expected.Clients, tc.expected.Clients)
		}

		if tc.expected.Threads != tc.expected.Threads {
			t.Errorf("Expected %d, got %d", tc.expected.Threads, tc.expected.Threads)
		}

		if tc.expected.MaximumTry != tc.expected.MaximumTry {
			t.Errorf("Expected %d, got %d", tc.expected.MaximumTry, tc.expected.MaximumTry)
		}

		if tc.expected.TransactionsPerClient != tc.expected.TransactionsPerClient {
			t.Errorf("Expected %d, got %d", tc.expected.TransactionsPerClient, tc.expected.TransactionsPerClient)
		}

		if tc.expected.TransactionsProcessed != tc.expected.TransactionsProcessed {
			t.Errorf("Expected %d, got %d", tc.expected.TransactionsProcessed, tc.expected.TransactionsProcessed)
		}

		if tc.expected.TransactionsFailed != tc.expected.TransactionsFailed {
			t.Errorf("Expected %d, got %d", tc.expected.TransactionsFailed, tc.expected.TransactionsFailed)
		}

		if tc.expected.AvgLatency != tc.expected.AvgLatency {
			t.Errorf("Expected %f, got %f", tc.expected.AvgLatency, tc.expected.AvgLatency)
		}

		if tc.expected.StdLatency != tc.expected.StdLatency {
			t.Errorf("Expected %f, got %f", tc.expected.StdLatency, tc.expected.StdLatency)
		}

		if tc.expected.InitialConnectionsTime != tc.expected.InitialConnectionsTime {
			t.Errorf("Expected %f, got %f", tc.expected.InitialConnectionsTime, tc.expected.InitialConnectionsTime)
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
