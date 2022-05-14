package main

import (
	"context"
	"flag"

	"github.com/spf13/pflag"
	"k8s.io/klog/v2"

	"github.com/nemoremold/perftests/pkg/options"
	"github.com/nemoremold/perftests/pkg/testflow"
)

func parseFlags(opts *options.Options) {
	pflag.StringVarP(&opts.ExportFolderPath, "export_folder_path", "f", opts.ExportFolderPath, "path to the folder where exported reports will be saved, only valid when '--write_to_csv' is true")
	pflag.StringVarP(&opts.IOChaosKubeconfigFilePath, "io_chaos_kubeconfig", "c", opts.IOChaosKubeconfigFilePath, "path to the kubeconfig file used by chaos agent")
	pflag.IntVarP(&opts.JobsPerWorker, "jobs", "j", opts.JobsPerWorker, "number of jobs to be done per worker")
	pflag.StringVarP(&opts.KubeconfigFilePath, "kubeconfig", "k", opts.KubeconfigFilePath, "path to the kubeconfig file")
	pflag.StringArrayVarP(&opts.Latencies, "latencies", "l", opts.Latencies, "latencies to be applied to IOChaos for performance testing")
	pflag.StringArrayVarP(&opts.PercentsStr, "percents", "p", opts.PercentsStr, "percents to be applied to IOChaos for performance testing")
	pflag.IntVarP(&opts.SleepTimeInSeconds, "sleep", "s", opts.SleepTimeInSeconds, "waiting time in seconds after performance testing and before cleanup")
	pflag.BoolVarP(&opts.Summarize, "summarize", "", opts.Summarize, "print the report of each test to stdout")
	pflag.IntVarP(&opts.WorkerNumber, "workers", "w", opts.WorkerNumber, "number of workers")
	pflag.BoolVarP(&opts.WriteToCSV, "write_to_csv", "", opts.WriteToCSV, "export the final testing report to a csv file")

	fs := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(fs)

	pflag.CommandLine.AddGoFlagSet(fs)
	pflag.Parse()
}

// TODO: enhance logging with customized logger.
func main() {
	// Parse flags and read configurations
	opts := options.NewOptions()
	parseFlags(opts)
	if err := opts.Parse(); err != nil {
		klog.Fatalf("failed to parse options: %v", err.Error())
	}

	// TODO: fix context.
	// Prepare global context.
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
	}()

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
