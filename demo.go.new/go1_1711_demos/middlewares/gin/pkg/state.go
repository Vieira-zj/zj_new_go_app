package pkg

import "github.com/prometheus/client_golang/prometheus"

var (
	HistogramVecApiDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "GinMonitor",
		Name:      "apiDuration",
		Help:      "api耗时单位ms",
		Buckets:   []float64{30.0, 50.0, 100.0, 200.0, 300.0},
	}, []string{"path"})

	GaugeVecApiMethod = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "GinMonitor",
		Name:      "apiCount",
		Help:      "网络请求总次数",
	}, []string{"method"})

	GaugeVecApiError = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "GinMonitor",
		Name:      "apiErrorCount",
		Help:      "api请求错误次数",
	}, []string{"path"})
)

func init() {
	prometheus.MustRegister(HistogramVecApiDuration, GaugeVecApiMethod, GaugeVecApiError)
}
