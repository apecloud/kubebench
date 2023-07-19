package exporter

import (
	"fmt"

	"k8s.io/klog/v2"
)

const (
	Sysbench = "sysbench"
	Pgbench  = "pgbench"
)

// Scrape is a function to scrape benchmark result from log.
func Scrape(benchType, file, benchName, jobName string, ch chan struct{}) {
	defer func() {
		// notify the channel
		ch <- struct{}{}
	}()

	switch benchType {
	case Pgbench:
		klog.Info("scrape pgbench result")
		ScrapPgbench(file, benchName, jobName)
	case Sysbench:
		klog.Info("scrape sysbench result")
		ScrapeSysbench(file, benchName, jobName)
	default:
		fmt.Printf("not support benchmark type: %s\n", benchType)
	}
}
