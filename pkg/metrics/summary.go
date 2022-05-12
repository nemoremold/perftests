package metrics

import (
	"fmt"
	"strings"
	"time"

	dto "github.com/prometheus/client_model/go"
	"k8s.io/klog/v2"

	"github.com/nemoremold/perftests/pkg/utils"
	"github.com/prometheus/client_golang/prometheus"
)

// Summary prints out the analyzed result of the performance testing.
func Summary(numberOfWorkers, numberOfJobs int, start, end time.Time) {
	verbs := []string{"create", "get", "update", "patch", "list", "delete", "all"}

	// Print header
	utils.PrintEmptyLine(2)
	klog.V(2).Infof("Performance Testing Summary (%v workers, each operating on %v deployments)", numberOfWorkers, numberOfJobs)
	utils.PrintEmptyLine(2)

	// Summarize API request success rate
	utils.PrintDivider(2, 63)
	utils.PrintCenter(2, 63, "API Request Success Rate")
	utils.PrintDivider(2, 63)
	klog.V(2).Infof("|%15v|%15v|%15v|%15v|", "VERB", "TOTAL", "SUCCESSFUL", "PERCENTAGE")
	utils.PrintDivider(2, 63)
	for _, verb := range verbs {
		summarizeAPIRequestSuccessRate(verb)
	}
	utils.PrintDivider(2, 63)

	// Summarize API request latency
	utils.PrintDivider(2, 63)
	utils.PrintCenter(2, 63, "API Request Latency")
	utils.PrintDivider(2, 63)
	klog.V(2).Infof("|%15v|%15v|%15v|%15v|", "VERB", "P99", "P90", "P50")
	utils.PrintDivider(2, 63)
	for _, verb := range verbs {
		summarizeAPIRequestLatency(verb)
	}
	utils.PrintDivider(2, 63)

	// Print footer
	utils.PrintEmptyLine(2)
	klog.V(2).Info("Start time: ", start.Local().String())
	klog.V(2).Info("End time: ", end.Local().String())
	klog.V(2).Info("Test duration: ", end.Sub(start))
	utils.PrintEmptyLine(2)
}

func summarizeAPIRequestSuccessRate(verb string) {
	verb = strings.ToLower(verb)

	metric := &dto.Metric{}
	_ = totalAPIRequests.WithLabelValues(verb).Write(metric)
	allGets := metric.Counter.GetValue()
	_ = successfulAPIRequests.WithLabelValues(verb).Write(metric)
	allSuccessfulGets := metric.Counter.GetValue()
	percentage := allSuccessfulGets * 100 / allGets

	klog.V(2).Infof("|%15v|%15v|%15v|%15v|", strings.ToUpper(verb), fmt.Sprint(allGets), fmt.Sprint(allSuccessfulGets), fmt.Sprintf("%.2f", percentage))
}

func summarizeAPIRequestLatency(verb string) {
	verb = strings.ToLower(verb)

	var p99, p90, p50 float64
	metric := &dto.Metric{}
	summary := apiRequestLatencies.WithLabelValues(verb).(prometheus.Summary)
	_ = summary.Write(metric)
	for _, quantile := range metric.Summary.GetQuantile() {
		switch *quantile.Quantile {
		case 0.5:
			p50 = *quantile.Value
		case 0.9:
			p90 = *quantile.Value
		case 0.99:
			p99 = *quantile.Value
		}
	}

	klog.V(2).Infof("|%15v|%15v|%15v|%15v|", strings.ToUpper(verb), fmt.Sprintf("%.10f", p99), fmt.Sprintf("%.10f", p90), fmt.Sprintf("%.10f", p50))
}
