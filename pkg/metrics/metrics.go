package metrics

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	registry *prometheus.Registry

	totalAPIRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "total_api_requests",
			Help: "Total API requests sent from workers to kube-apiserver during performance testing",
		},
		[]string{"verb"},
	)

	successfulAPIRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "successful_api_requests",
			Help: "API requests sent from workers to kube-apiserver during performance testing that does not get error response",
		},
		[]string{"verb"},
	)

	apiRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_requests",
			Help: "API requests sent from specific workers to kube-apiserver during performance testing",
		},
		[]string{"verb", "worker", "success"},
	)

	apiRequestLatencies = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "api_request_latencies",
			Help:       "The latency of API requests sent from workers to kube-apiserver during performance testing",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"verb"},
	)
)

func init() {
	registry = prometheus.NewRegistry()

	registry.MustRegister(totalAPIRequests)
	registry.MustRegister(successfulAPIRequests)
	registry.MustRegister(apiRequests)
	registry.MustRegister(apiRequestLatencies)
}

// RecordAPIRequest receive a API request report and store it in the Prometheus registry.
func RecordAPIRequest(verb string, workerId int, success bool, duration time.Duration) {
	totalAPIRequests.WithLabelValues(verb).Inc()
	totalAPIRequests.WithLabelValues("all").Inc()

	if success {
		successfulAPIRequests.WithLabelValues(verb).Inc()
		successfulAPIRequests.WithLabelValues("all").Inc()
	}

	apiRequests.WithLabelValues(verb, fmt.Sprint(workerId), fmt.Sprint(success)).Inc()
	apiRequests.WithLabelValues("all", fmt.Sprint(workerId), fmt.Sprint(success)).Inc()

	apiRequestLatencies.WithLabelValues(verb).Observe(duration.Seconds())
	apiRequestLatencies.WithLabelValues("all").Observe(duration.Seconds())
}
