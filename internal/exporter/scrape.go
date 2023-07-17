package exporter

import (
	"fmt"
)

const (
	Sysbench = "sysbench"
	Pgbench  = "pgbench"
)

// Scrape is a function to scrape benchmark result from log.
func Scrape(benchType string, file string, ch chan struct{}) {
	defer func() {
		// notify the channel
		ch <- struct{}{}
	}()

	switch benchType {
	case Pgbench:
		ScrapPgbench(file)
	case Sysbench:
		ScrapeSysbench(file)
	default:
		fmt.Printf("not support benchmark type: %s\n", benchType)
	}
}
