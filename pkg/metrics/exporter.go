package metrics

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"time"

	dto "github.com/prometheus/client_model/go"
	"k8s.io/klog/v2"

	"github.com/nemoremold/perftests/pkg/constants"
	"github.com/nemoremold/perftests/pkg/options"
	"github.com/prometheus/client_golang/prometheus"
)

// exporter summarizes the final and overall performance testing result, generating
// the report and exporting it to a local file.
type exporter struct {
	// latencies are latency labels.
	latencies []string
	// percents are percent labels.
	percents []string

	// header is the header row for each table.
	header rowData
	// titles are the title of tables, its item index is also the table id.
	titles []string
	// datum are the tables of metrics, its item index is also the table id.
	datum []map[float64]rowData

	// numberOfTables is the number of table, equal to the number of percents.
	numberOfTables int
	// numberOfColumns is the number of table columns, equal to the number of latencies plus one.
	numberOfColumns int
	// numberOfColumns is the number of table rows, equal to the number of summary quantiles plus one.
	numberOfDataRowsPerTable int
}

// rowData is a string slice that corresponds to a line in the csv file.
// Each item is an entry in the table.
type rowData []string

// Init initializes the exporter, preparing metrics table.
func (e *exporter) Init() {
	e.numberOfTables = len(e.percents)
	e.numberOfColumns = len(e.latencies) + 1
	e.numberOfDataRowsPerTable = len(SummaryObjectives) + 1

	e.header = make(rowData, e.numberOfColumns)
	e.header[0] = "Quantile"
	for index, latency := range e.latencies {
		latency = strings.ReplaceAll(latency, "ms", "")
		e.header[index+1] = fmt.Sprintf("Latency(%v)", latency)
	}

	e.titles = make([]string, e.numberOfTables)
	for _, percent := range e.percents {
		e.titles = append(e.titles, percent+"% sample")
	}

	e.datum = make([]map[float64]rowData, e.numberOfTables)
	for index := range e.datum {
		e.datum[index] = make(map[float64]rowData)
		for quantile := range SummaryObjectives {
			e.datum[index][quantile] = make(rowData, e.numberOfColumns)
			e.datum[index][quantile][0] = fmt.Sprint(int(quantile*100)) + "%"
		}
		e.datum[index][0] = make(rowData, e.numberOfColumns)
		e.datum[index][0][0] = "Success Rate"
	}
}

// Summarize summarizes the overall latency metrics and success rate metrics.
func (e *exporter) Summarize() error {
	for tableID, percent := range e.percents {
		if err := e.summarize(tableID, percent); err != nil {
			return err
		}
	}
	return nil
}

// summarize summarizes the overall latency metrics and success rate metrics for a
// certain percent label in the registry.
func (e *exporter) summarize(tableID int, percent string) error {
	for index, latency := range e.latencies {
		latencyMetric, err := summarizeLatencyMetric(latency, percent)
		if err != nil {
			return err
		}
		for _, quantile := range latencyMetric.Summary.Quantile {
			quantileMark := *quantile.Quantile
			e.datum[tableID][quantileMark][index+1] = fmt.Sprintf("%.10f", *quantile.Value)
		}

		successRate, err := summarizeSuccessRate(latency, percent)
		if err != nil {
			return err
		}
		e.datum[tableID][0][index+1] = fmt.Sprintf("%.2f", successRate) + "%"
	}
	return nil
}

// Export exports the report to a file.
func (e *exporter) Export(ctx context.Context, filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	// TODO: is this the proper way?
	go func() {
		<-ctx.Done()
		_ = file.Close()
	}()

	writer := csv.NewWriter(file)
	for tableID := 0; tableID < e.numberOfTables; tableID++ {
		if err := writer.Write([]string{e.titles[tableID]}); err != nil {
			return err
		}
		if err := writer.Write(e.header); err != nil {
			return err
		}
		for quantile := range SummaryObjectives {
			if err := writer.Write(e.datum[tableID][quantile]); err != nil {
				return err
			}
		}
		if err := writer.Write(e.datum[tableID][0]); err != nil {
			return err
		}
		writer.Flush()
	}

	return nil
}

// summarizeLatencyMetric gets the overall API request latencies for a metric set.
func summarizeLatencyMetric(latency, percent string) (*dto.Metric, error) {
	metric := &dto.Metric{}
	summary := apiRequestLatencies.WithLabelValues(constants.ALL, latency, percent).(prometheus.Summary)
	if err := summary.Write(metric); err != nil {
		return nil, err
	}
	return metric, nil
}

// summarizeSuccessRate gets the overall API request success rate for a metric set.
func summarizeSuccessRate(latency, percent string) (float64, error) {
	metric := &dto.Metric{}
	if err := totalAPIRequests.WithLabelValues(constants.ALL, latency, percent).Write(metric); err != nil {
		return 0, err
	}
	allGets := metric.Counter.GetValue()
	if err := successfulAPIRequests.WithLabelValues(constants.ALL, latency, percent).Write(metric); err != nil {
		return 0, err
	}
	allSuccessfulGets := metric.Counter.GetValue()
	return allSuccessfulGets * 100 / allGets, nil
}

// WriteToCSV gathers the all-time metrics and summarizes them into an overall report,
// exporting it to the target folder.
func WriteToCSV(ctx context.Context, opts *options.Options, startTime time.Time) {
	// Create an exporter instance and initialize it.
	e := &exporter{
		latencies: opts.Latencies,
		percents:  opts.PercentsStr,
	}
	e.Init()

	// Generate the all-time final report.
	klog.V(4).Info("generating final performance testing report")
	if err := e.Summarize(); err != nil {
		klog.Errorf("failed to generate final report: %v", err.Error())
		return
	}
	klog.V(4).Info("successfully generated final performance testing report")

	// Determine export file path.
	// File name format: <formatted_test_start_date_time>_<number_of_workers>_<number_of_jobs_per_worker>.csv
	datetime := fmt.Sprint(startTime.Local())
	datetime = strings.ReplaceAll(datetime, ":", "-")
	datetime = strings.ReplaceAll(datetime, " ", "_")
	datetime = strings.ReplaceAll(datetime, "+", "")
	filepath := opts.ExportFolderPath + "/" + datetime + "_" + fmt.Sprint(opts.WorkerNumber) + "_" + fmt.Sprint(opts.JobsPerWorker) + ".csv"

	// Prepare file to export report to.
	klog.V(2).Infof("writing final performance testing report to %v", filepath)
	if err := e.Export(ctx, filepath); err != nil {
		klog.Errorf("failed to export to file %v: %v", filepath, err.Error())
		return
	}
	klog.V(2).Infof("successfully wrote final performance testing report to %v", filepath)
}
