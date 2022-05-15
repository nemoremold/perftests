package metrics

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"time"

	"k8s.io/klog/v2"

	"github.com/nemoremold/perftests/pkg/constants"
	"github.com/nemoremold/perftests/pkg/options"
)

// Exporter summarizes the final and overall performance testing result, generating
// the report and exporting it to a local file.
type Exporter struct {
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

// NewExporter instantiates a new exporter instance.
func NewExporter(latencies, percents []string) *Exporter {
	e := &Exporter{
		latencies: latencies,
		percents:  percents,
	}
	e.init()
	return e
}

// init initializes the Exporter, preparing metrics table.
func (e *Exporter) init() {
	e.numberOfTables = len(e.percents)
	e.numberOfColumns = len(e.latencies) + 1
	e.numberOfDataRowsPerTable = len(SummaryObjectives) + 1

	// Set table headers.
	e.header = make(rowData, e.numberOfColumns)
	e.header[0] = "Quantile"
	for index, latency := range e.latencies {
		latency = strings.ReplaceAll(latency, "ms", "")
		e.header[index+1] = fmt.Sprintf("Latency(%v)", latency)
	}

	// Set title for each table.
	e.titles = make([]string, e.numberOfTables)
	for index, percent := range e.percents {
		e.titles[index] = percent + "% sample"
	}

	// Set table indexes.
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

// WriteToCSV gathers the all-time metrics and summarizes them into an overall report,
// exporting it to the target folder.
func (e *Exporter) WriteToCSV(ctx context.Context, opts *options.Options, startTime time.Time) {
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

// Export exports the report to a file.
func (e *Exporter) Export(ctx context.Context, filepath string) error {
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

		for _, quantile := range SortedQuantiles {
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

// Collect collects latency quantiles and success rate for a certain
// latency-percent pair.
func (e *Exporter) Collect(percentIndex, latencyIndex int) error {
	set := MetricSetID{
		Latency: e.latencies[latencyIndex],
		Percent: e.percents[percentIndex],
	}

	latencyMetric, err := collectLatencyMetric(constants.ALL, set)
	if err != nil {
		return err
	}
	for _, quantile := range latencyMetric.Summary.Quantile {
		e.datum[percentIndex][*quantile.Quantile][latencyIndex+1] = fmt.Sprintf("%.10f", *quantile.Value)
	}

	_, _, successRate, err := collectSuccessRateMetrics(constants.ALL, set)
	if err != nil {
		return err
	}
	e.datum[percentIndex][0][latencyIndex+1] = fmt.Sprintf("%.2f", successRate) + "%"

	return nil
}
