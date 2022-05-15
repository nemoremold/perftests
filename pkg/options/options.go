package options

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Options is the configuration of the perftests program.
type Options struct {
	// ChaosAgentPollIntervalInSeconds is the interval between polls when waiting for IOChaos status change.
	ChaosAgentPollIntervalInSeconds int
	// ChaosAgentPollTimeoutInSeconds is the timeout between polls when waiting for IOChaos status change.
	ChaosAgentPollTimeoutInSeconds int
	// ChaosAgentIOChaosTemplateFilePath is the path to the template IOChaos file.
	ChaosAgentIOChaosTemplateFilePath string
	// ExportFolderPath is the path to the folder where exported reports will be saved,
	// only valid when `WriteToCSV` is set to `true`.
	ExportFolderPath string
	// IOChaosKubeconfigFilePath is the the path to the kubeconfig file used by chaos agent.
	IOChaosKubeconfigFilePath string
	// JobsPerWorker is the number of jobs to be done per worker.
	JobsPerWorker int
	// KubeconfigFilePath is the path to the kubeconfig file.
	KubeconfigFilePath string
	// Latencies are a list of latencies to be applied to IOChaos.
	Latencies []string
	// PercentsStr are a list of percents in string format, should be converted in to integers before use.
	PercentsStr []string
	// SleepTimeInSeconds is the length of time before cleanup is carried out after performance testing finishes.
	SleepTimeInSeconds int
	// Summarize when set to true, prints the report of each test in stdout.
	Summarize bool
	// WorkerNumber is the number of workers.
	WorkerNumber int
	// WriteToCSV when set to true, exports the final report to a csv file.
	WriteToCSV bool

	// Percents are a list of percents to be applied to IOChaos.
	Percents []int
}

// NewOptions instantiates a new Options object with default values.
func NewOptions() *Options {
	return &Options{
		ChaosAgentPollIntervalInSeconds:   2,
		ChaosAgentPollTimeoutInSeconds:    60,
		ChaosAgentIOChaosTemplateFilePath: "",
		ExportFolderPath:                  "",
		IOChaosKubeconfigFilePath:         "",
		JobsPerWorker:                     100,
		KubeconfigFilePath:                "kubeconfig",
		Latencies:                         []string{"0ms", "10ms", "20ms", "30ms", "40ms", "50ms", "60ms", "70ms", "100ms", "200ms", "300ms"},
		PercentsStr:                       []string{"10", "20", "30", "40", "50", "60", "70"},
		SleepTimeInSeconds:                60,
		Summarize:                         true,
		WorkerNumber:                      30,
		WriteToCSV:                        false,
	}
}

// Parse parses an option and checks whether it is valid.
func (o *Options) Parse() error {
	// Convert percent strings to integers.
	for _, percentStr := range o.PercentsStr {
		percent, err := strconv.Atoi(percentStr)
		if err != nil {
			return err
		}

		if percent < 0 || percent > 100 {
			return fmt.Errorf("%v is not a valid percentile (should be in range [0, 100])", percentStr)
		}
		o.Percents = append(o.Percents, percent)
	}

	// Sort the percents by ascending order.
	sort.Ints(o.Percents)
	for index, percent := range o.Percents {
		o.PercentsStr[index] = fmt.Sprint(percent)
	}

	// Ensure latencies are valid values.
	// Valid: 0ms, 1ms, 10ms...
	// Invalid: '01ms', '1 ms', ' 20ms', '9ms '...
	var latencies []int
	r, _ := regexp.Compile("^(0ms|[1-9]([0-9]*)ms)$")
	for _, latency := range o.Latencies {
		if !r.MatchString(latency) {
			return fmt.Errorf("%v is not a valid latency (valid format regex: ^(0ms|[1-9]([0-9]*)ms)$)", latency)
		}

		latencyStr := strings.Replace(latency, "ms", "", 1)
		latencyInt, err := strconv.Atoi(latencyStr)
		if err != nil {
			return err
		}
		latencies = append(latencies, latencyInt)
	}

	// Sort the latencies by ascending order.
	sort.Ints(latencies)
	for index, latencyInt := range latencies {
		o.Latencies[index] = fmt.Sprint(latencyInt) + "ms"
	}

	// Ensure `ExportFolderPath` is a folder.
	if o.WriteToCSV && len(o.ExportFolderPath) > 0 {
		info, err := os.Stat(o.ExportFolderPath)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return fmt.Errorf("export destination %v is not a valid directory", o.ExportFolderPath)
		}
	}

	return nil
}
