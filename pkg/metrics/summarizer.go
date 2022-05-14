package metrics

import (
	"fmt"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"

	"github.com/nemoremold/perftests/pkg/constants"
	"github.com/nemoremold/perftests/pkg/utils/printer"
)

// summarizer summarizes the performance testing result of a single test set (latency, percent),
// generating the report and exporting it to a local file.
type summarizer struct {
	// set is the id of the metric set to be summarized.
	set MetricSetID
	// numberOfWorkers is the number of workers for the test.
	numberOfWorkers int
	// numberOfJobs is the number of jobs done per worker.
	numberOfJobs int
	// startTime is the time when the test starts.
	startTime time.Time
	// endTime is the time when the test finishes.
	endTime time.Time
}

func (s *summarizer) PrintSummary() {

}

// Summary prints out the analyzed result of the performance testing.
func Summary(set MetricSetID, numberOfWorkers, numberOfJobs int, start, end time.Time) {
	// Prepare summary sheet.
	sheet := printer.NewSheetPtr(0, printer.LineAlignCenter("Performance Testing Summary"))

	// Prepare sheet header.
	aligner := 0
	if numberOfWorkers < numberOfJobs {
		aligner = len(fmt.Sprint(numberOfJobs))
	} else {
		aligner = len(fmt.Sprint(numberOfWorkers))
	}
	sheet.SetHeader([]printer.Line{
		printer.LineAlignRight("Total number of workers: " + fmt.Sprintf("%*v", aligner, numberOfWorkers)),
		printer.LineAlignRight("Jobs done per worker: " + fmt.Sprintf("%*v", aligner, numberOfJobs)),
	})

	// Prepare sheet footer.
	sheet.SetFooter([]printer.Line{
		printer.LineAlignLeft("   Start time: " + start.Local().String()),
		printer.LineAlignLeft("     End time: " + end.Local().String()),
		printer.LineAlignLeft("Test duration: " + end.Sub(start).String()),
	})

	// Prepare tables.
	sheet.SetTables([]printer.Table{
		prepareSuccessRateTable(set),
		prepareLatencyTable(set),
	})

	// Print summary sheet.
	printer.PrintEmptyLine()
	sheet.Print()
	printer.PrintEmptyLine()
}

func prepareSuccessRateTable(set MetricSetID) printer.Table {
	// Prepare API request success rate table.
	indexRow := printer.TableRow{
		printer.LineAlignRight("Verb"),
		printer.LineAlignRight("Total"),
		printer.LineAlignRight("Successful"),
		printer.LineAlignRight("Percentage"),
	}
	table := printer.NewTable(0, indexRow.ColumnsCount(), printer.LineAlignCenter("API Request Success Rate"))

	// Prepare indexes.
	table.SetIndexes(indexRow)

	// Prepare values.
	var tableRows []printer.TableRow
	for _, verb := range constants.Verbs {
		tableRows = append(tableRows, prepareSuccessRateTableRow(verb, set))
	}
	table.SetDatum(tableRows)

	return table
}

func prepareSuccessRateTableRow(verb string, set MetricSetID) printer.TableRow {
	metric := &dto.Metric{}
	_ = totalAPIRequests.WithLabelValues(verb, set.Latency, set.Percent).Write(metric)
	allGets := metric.Counter.GetValue()
	_ = successfulAPIRequests.WithLabelValues(verb, set.Latency, set.Percent).Write(metric)
	allSuccessfulGets := metric.Counter.GetValue()
	percentage := allSuccessfulGets * 100 / allGets

	return printer.TableRow{
		printer.LineAlignRight(strings.ToUpper(verb)),
		printer.LineAlignRight(fmt.Sprint(allGets)),
		printer.LineAlignRight(fmt.Sprint(allSuccessfulGets)),
		printer.LineAlignRight(fmt.Sprintf("%.2f", percentage)),
	}
}

func prepareLatencyTable(set MetricSetID) printer.Table {
	// Prepare API request latency table.
	indexRow := printer.TableRow{
		printer.LineAlignRight("Verb"),
	}
	for quantile := range SummaryObjectives {
		indexRow.AddEntry(printer.LineAlignRight("P" + fmt.Sprint(int(quantile*100))))
	}
	table := printer.NewTable(0, indexRow.ColumnsCount(), printer.LineAlignCenter("API Request Latency"))

	// Prepare indexes.
	table.SetIndexes(indexRow)

	// Prepare values.
	var tableRows []printer.TableRow
	for _, verb := range constants.Verbs {
		tableRows = append(tableRows, prepareLatencyTableRow(verb, set))
	}
	table.SetDatum(tableRows)

	return table
}

func prepareLatencyTableRow(verb string, set MetricSetID) printer.TableRow {
	metric := &dto.Metric{}
	summary := apiRequestLatencies.WithLabelValues(verb, set.Latency, set.Percent).(prometheus.Summary)
	_ = summary.Write(metric)

	row := printer.TableRow{
		printer.LineAlignRight(strings.ToUpper(verb)),
	}
	for _, quantile := range metric.Summary.GetQuantile() {
		row.AddEntry(printer.LineAlignRight(fmt.Sprintf("%.5f", *quantile.Value)))
	}
	return row
}
