package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	counter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "golang",
			Name:      "my_counter",
			Help:      "This is my counter",
		})

	gauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "golang",
			Name:      "my_gauge",
			Help:      "This is my gauge",
		})

	histogram = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "golang",
			Name:      "my_histogram",
			Help:      "This is my histogram",
		})

	summary = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace: "golang",
			Name:      "my_summary",
			Help:      "This is my summary",
		})
)

// 测试：curl http://127.0.0.1:9100/metrics | grep golang
//
// 由 promethus 通过该 /metrics 接口拉取（pull）数据。
// 当使用 push gateway 中间件时，也可由 client 主动上报（push）数据。
//

func main() {
	http.Handle("/metrics", promhttp.Handler())
	prometheus.MustRegister(counter, gauge, histogram, summary)

	go func() {
		rand.Seed(time.Now().Unix())
		for {
			counter.Add(rand.Float64() * 5)
			gauge.Add(rand.Float64()*15 - 5)
			histogram.Observe(rand.Float64() * 10)
			summary.Observe(rand.Float64() * 10)
			time.Sleep(2 * time.Second)
		}
	}()

	log.Println("http server start")
	if err := http.ListenAndServe(":9100", nil); err != nil {
		log.Fatalln(err)
	}
}
