package metrics

import (
	"time"
)

// RecordAPIRequest receives a API request report and stores it in the Prometheus registry.
// In addition to storing with the original verb, it all stores it with verb `all`.
func RecordAPIRequest(verb string, success bool, duration time.Duration, set MetricSetID) {
	recordAPIRequest(verb, success, duration, set)
	recordAPIRequest("all", success, duration, set)
}

// recordAPIRequest receives a API request report and stores it in the Prometheus registry.
func recordAPIRequest(verb string, success bool, duration time.Duration, set MetricSetID) {
	totalAPIRequests.WithLabelValues(verb, set.Latency, set.Percent).Inc()

	if success {
		successfulAPIRequests.WithLabelValues(verb, set.Latency, set.Percent).Inc()
	}

	apiRequestLatencies.WithLabelValues(verb, set.Latency, set.Percent).Observe(duration.Seconds())
}
