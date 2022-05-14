package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	// SummaryObjectives defines the quantiles and their respective
	// absolute error.
	SummaryObjectives = map[float64]float64{
		0.1:  0.05,
		0.25: 0.05,
		0.5:  0.025,
		0.75: 0.025,
		0.9:  0.01,
		0.95: 0.005,
		0.99: 0.001,
	}

	registry *prometheus.Registry

	totalAPIRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "total_api_requests",
			Help: "Total API requests sent from workers to kube-apiserver during performance testing",
		},
		[]string{"verb", "latency", "percent"},
	)

	successfulAPIRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "successful_api_requests",
			Help: "API requests sent from workers to kube-apiserver during performance testing that does not get error response",
		},
		[]string{"verb", "latency", "percent"},
	)

	apiRequestLatencies = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "api_request_latencies",
			Help:       "The latency of API requests sent from workers to kube-apiserver during performance testing",
			Objectives: SummaryObjectives,
		},
		[]string{"verb", "latency", "percent"},
	)
)

func init() {
	registry = prometheus.NewRegistry()

	registry.MustRegister(totalAPIRequests)
	registry.MustRegister(successfulAPIRequests)
	registry.MustRegister(apiRequestLatencies)
}

// MetricSetID groups the metrics by latency label and percent label.
type MetricSetID struct {
	// Latency is the value of latency label.
	Latency string
	// Percent is the value of percent label.
	Percent string
}
