package pkg

import "github.com/prometheus/client_golang/prometheus"

var (
	GaugeVecApiDuration = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "GinMonitor",
		Name:      "apiDuration",
		Help:      "api耗时单位ms",
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
	prometheus.MustRegister(GaugeVecApiDuration, GaugeVecApiMethod, GaugeVecApiError)
}
