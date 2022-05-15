package metrics

import (
	"fmt"
	"strings"
	"time"

	"github.com/nemoremold/perftests/pkg/constants"
	"github.com/nemoremold/perftests/pkg/utils/printer"
)

// Summary prints out the analyzed result of the performance testing.
func Summary(set MetricSetID, numberOfWorkers, numberOfJobs int, start, end time.Time) {
	// Prepare summary sheet.
	sheet := printer.NewSheet(0, printer.LineAlignCenter("Performance Testing Summary"))

	// Prepare sheet header.
	aligner := 0
	for _, candidate := range []int{len(fmt.Sprint(numberOfJobs)), len(fmt.Sprint(numberOfWorkers)), len(set.Latency), len(set.Percent)} {
		if aligner < candidate {
			aligner = candidate
		}
	}
	sheet.SetHeader([]printer.Line{
		printer.LineAlignRight("Latency: " + fmt.Sprintf("%*v", aligner, set.Latency)),
		printer.LineAlignRight("Percent: " + fmt.Sprintf("%*v", aligner, set.Percent)),
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

// prepareSuccessRateTable generates the success rate table.
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
	table.SetHeaders(indexRow)

	// Prepare values.
	var tableRows []printer.TableRow
	for _, verb := range constants.Verbs {
		tableRows = append(tableRows, prepareSuccessRateTableRow(verb, set))
	}
	table.SetDatum(tableRows)

	return *table
}

// prepareSuccessRateTableRow collects success rate related metrics from a specific
// metric set and insert its values to a table row.
func prepareSuccessRateTableRow(verb string, set MetricSetID) printer.TableRow {
	allGets, allSuccessfulGets, percentage, _ := collectSuccessRateMetrics(verb, set)

	return printer.TableRow{
		// Row index.
		printer.LineAlignRight(strings.ToUpper(verb)),
		// Row values.
		printer.LineAlignRight(fmt.Sprint(allGets)),
		printer.LineAlignRight(fmt.Sprint(allSuccessfulGets)),
		printer.LineAlignRight(fmt.Sprintf("%.2f", percentage)),
	}
}

// prepareLatencyTable generates the latency table.
func prepareLatencyTable(set MetricSetID) printer.Table {
	// Prepare API request latency table.
	headerRow := printer.TableRow{
		printer.LineAlignRight("Verb"),
	}
	for _, quantile := range SortedQuantiles {
		headerRow.AddEntry(printer.LineAlignRight("P" + fmt.Sprint(int(quantile*100))))
	}
	table := printer.NewTable(0, headerRow.ColumnsCount(), printer.LineAlignCenter("API Request Latency"))

	// Prepare indexes.
	table.SetHeaders(headerRow)

	// Prepare values.
	var tableRows []printer.TableRow
	for _, verb := range constants.Verbs {
		tableRows = append(tableRows, prepareLatencyTableRow(verb, set))
	}
	table.SetDatum(tableRows)

	return *table
}

// prepareLatencyTableRow collects latency metrics from a specific metric set and
// insert its values to a table row.
func prepareLatencyTableRow(verb string, set MetricSetID) printer.TableRow {
	metric, _ := collectLatencyMetric(verb, set)
	quantileMap := make(map[float64]float64)
	for _, quantile := range metric.Summary.GetQuantile() {
		quantileMap[*quantile.Quantile] = *quantile.Value
	}

	// Set row index.
	row := printer.TableRow{
		printer.LineAlignRight(strings.ToUpper(verb)),
	}
	// Set row values.
	for _, quantile := range SortedQuantiles {
		row.AddEntry(printer.LineAlignRight(fmt.Sprintf("%.5f", quantileMap[quantile])))
	}
	return row
}
