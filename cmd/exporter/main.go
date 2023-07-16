package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"time"

	"github.com/apecloud/kubebench/internal/exporter"
)

var (
	benchType string
	file      string
)

func main() {
	// parse flags
	flag.StringVar(&benchType, "type", "", "benchmark type")
	flag.StringVar(&file, "file", "", "log file")
	flag.Parse()

	quit := make(chan struct{}, 1)

	exporter.Register(benchType)
	exporter.Scrape(benchType, file, quit)

	r := gin.Default()
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	go r.Run(":9187")

	// get signal, exit
	<-quit

	// wait prometheus to collect data
	time.Sleep(30 * time.Second)
}
