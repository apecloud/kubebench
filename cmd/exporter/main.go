package main

import (
	"flag"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/apecloud/kubebench/internal/exporter"
)

var (
	benchType string
	benchName string
	jobName   string
	file      string
)

func main() {
	// parse flags
	flag.StringVar(&benchType, "type", "", "benchmark type")
	flag.StringVar(&file, "file", "", "log file")
	flag.StringVar(&benchName, "bench", "", "benchmark name")
	flag.StringVar(&jobName, "job", "", "job name")
	flag.Parse()

	quit := make(chan struct{}, 1)

	r := gin.Default()
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	go r.Run(":9187")

	exporter.InitMetrics()
	exporter.Register()
	exporter.Scrape(benchType, file, benchName, jobName, quit)

	// get signal, exit
	<-quit

	// wait prometheus to collect data
	time.Sleep(30 * time.Second)
}
