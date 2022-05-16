package metrics

import (
	"sort"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	// SummaryObjectives defines the quantiles and their respective
	// absolute error.
	SummaryObjectives = map[float64]float64{
		0.1:  0.001,
		0.25: 0.001,
		0.5:  0.001,
		0.75: 0.001,
		0.9:  0.001,
		0.95: 0.001,
		0.99: 0.001,
	}

	// SortedQuantiles is the sorted array of summary objectives.
	SortedQuantiles []float64

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
			MaxAge:     60 * time.Minute, // Set a longer MaxAge because some test cases may take longer to finish.
		},
		[]string{"verb", "latency", "percent"},
	)
)

func init() {
	registry = prometheus.NewRegistry()

	registry.MustRegister(totalAPIRequests)
	registry.MustRegister(successfulAPIRequests)
	registry.MustRegister(apiRequestLatencies)

	SortedQuantiles = make([]float64, 0)
	for quantile := range SummaryObjectives {
		SortedQuantiles = append(SortedQuantiles, quantile)
	}
	sort.Float64s(SortedQuantiles)
}

// MetricSetID groups the metrics by latency label and percent label.
type MetricSetID struct {
	// Latency is the value of latency label.
	Latency string
	// Percent is the value of percent label.
	Percent string
}
