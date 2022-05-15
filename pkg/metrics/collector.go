package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// collectLatencyMetric gets the overall API request latencies for a metric set.
func collectLatencyMetric(verb string, set MetricSetID) (*dto.Metric, error) {
	metric := &dto.Metric{}
	summary := apiRequestLatencies.WithLabelValues(verb, set.Latency, set.Percent).(prometheus.Summary)
	if err := summary.Write(metric); err != nil {
		return nil, err
	}
	return metric, nil
}

// collectSuccessRate gets the overall API request success metrics for a metric set.
func collectSuccessRateMetrics(verb string, set MetricSetID) (float64, float64, float64, error) {
	metric := &dto.Metric{}
	if err := totalAPIRequests.WithLabelValues(verb, set.Latency, set.Percent).Write(metric); err != nil {
		return 0, 0, 0, err
	}
	allGets := metric.Counter.GetValue()
	if err := successfulAPIRequests.WithLabelValues(verb, set.Latency, set.Percent).Write(metric); err != nil {
		return 0, 0, 0, err
	}
	allSuccessfulGets := metric.Counter.GetValue()
	return allGets, allSuccessfulGets, allSuccessfulGets * 100 / allGets, nil
}
