package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/pflag"
	"k8s.io/klog/v2"

	"github.com/nemoremold/perftests/pkg/options"
	"github.com/nemoremold/perftests/pkg/testflow"
)

func parseFlags(opts *options.Options) {
	pflag.IntVarP(&opts.ChaosAgentPollIntervalInSeconds, "chaos_agent_poll_interval", "", opts.ChaosAgentPollIntervalInSeconds, "interval in seconds between polls when waiting for IOChaos status change")
	pflag.IntVarP(&opts.ChaosAgentPollTimeoutInSeconds, "chaos_agent_poll_timeout", "", opts.ChaosAgentPollTimeoutInSeconds, "timeout in seconds between polls when waiting for IOChaos status change")
	pflag.StringVarP(&opts.ChaosAgentIOChaosTemplateFilePath, "chaos_agent_template", "", opts.ChaosAgentIOChaosTemplateFilePath, "path to the template IOChaos file")
	pflag.StringVarP(&opts.ExportFolderPath, "export_folder_path", "f", opts.ExportFolderPath, "path to the folder where exported reports will be saved, only valid when '--write_to_csv' is true")
	pflag.StringVarP(&opts.IOChaosKubeconfigFilePath, "chaos_agent_kubeconfig", "c", opts.IOChaosKubeconfigFilePath, "path to the kubeconfig file used by chaos agent")
	pflag.IntVarP(&opts.JobsPerWorker, "jobs", "j", opts.JobsPerWorker, "number of jobs to be done per worker")
	pflag.StringVarP(&opts.KubeconfigFilePath, "kubeconfig", "k", opts.KubeconfigFilePath, "path to the kubeconfig file")
	pflag.StringSliceVarP(&opts.Latencies, "latencies", "l", opts.Latencies, "comma-separated latencies to be applied to IOChaos for performance testing")
	pflag.StringSliceVarP(&opts.PercentsStr, "percents", "p", opts.PercentsStr, "comma-separated percents to be applied to IOChaos for performance testing")
	pflag.IntVarP(&opts.SleepTimeInSeconds, "sleep", "s", opts.SleepTimeInSeconds, "waiting time in seconds after performance testing and before cleanup")
	pflag.BoolVarP(&opts.Summarize, "summarize", "", opts.Summarize, "print the report of each test to stdout")
	pflag.IntVarP(&opts.WorkerNumber, "workers", "w", opts.WorkerNumber, "number of workers")
	pflag.BoolVarP(&opts.WriteToCSV, "export_to_csv", "", opts.WriteToCSV, "export the final testing report to a csv file")

	fs := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(fs)

	pflag.CommandLine.AddGoFlagSet(fs)
	pflag.Parse()
}

// TODO: enhance logging with customized logger.
func main() {
	// Prepare global context and setup notifications for os signals.
	signals := make(chan os.Signal, 2)
	defer close(signals)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		sig := <-signals
		klog.Warningf("%v signal received, gracefully stopping the tests", sig)
		cancel()
		sig = <-signals
		klog.Warningf("%v signal received, forcefully quiting the cleanups", sig)
		os.Exit(1)
	}()

	// Parse flags and read configurations
	opts := options.NewOptions()
	parseFlags(opts)
	if err := opts.Parse(); err != nil {
		klog.Fatalf("failed to parse options: %v", err.Error())
	}

	// Initialize test flow.
	flow, err := testflow.NewTestFlow(opts)
	if err != nil {
		klog.Fatalf("failed to create test flow: %v", err.Error())
	}

	// Run test flow.
	if flow == nil {
		klog.Fatal("failed to create test flow, empty flow returned")
	} else {
		if err := flow.RunTestFlow(ctx); err != nil {
			klog.Fatalf("failed to run test flow: %v", err.Error())
		}
	}
}
