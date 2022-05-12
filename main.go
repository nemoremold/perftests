package main

import (
	"context"
	"flag"
	"sync"
	"time"

	"github.com/spf13/pflag"
	"k8s.io/klog/v2"

	"github.com/nemoremold/perftests/pkg/metrics"
	"github.com/nemoremold/perftests/pkg/worker"
)

var (
	// JobsPerWorker is the number of jobs to be done per worker.
	JobsPerWorker int
	// KubeconfigFilePath is the path to the kubeconfig file.
	KubeconfigFilePath string
	// SleepTimeInSeconds is the length of time before cleanup is carried out after performance testing finishes.
	SleepTimeInSeconds int
	// WorkerNumber is the number of workers.
	WorkerNumber int
)

func parseFlags() {
	pflag.IntVarP(&JobsPerWorker, "jobs", "j", 100, "number of jobs to be done per worker")
	pflag.StringVarP(&KubeconfigFilePath, "kubeconfig", "k", "kubeconfig", "path to the kubeconfig file")
	pflag.IntVarP(&SleepTimeInSeconds, "sleep", "s", 30, "waiting time in seconds after performance testing and before cleanup")
	pflag.IntVarP(&WorkerNumber, "workers", "w", 30, "number of workers")

	fs := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(fs)

	pflag.CommandLine.AddGoFlagSet(fs)
	pflag.Parse()
}

func main() {
	parseFlags()

	// Prepare global context.
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
	}()

	// Initialize workers.
	var workers []*worker.Worker
	for workerID := 0; workerID < WorkerNumber; workerID++ {
		worker, err := worker.NewWorker(workerID, KubeconfigFilePath)
		if err != nil {
			klog.Fatalf("failed to initialize workers: %v", err.Error())
		}
		workers = append(workers, worker)
	}

	// Prepare context dedicated for performance testing. When stop signal
	// is received, this dedicated context will be cancelled first, triggering
	// the clean up process before actually stopping the program.
	jobsCtx, jobsCancel := context.WithCancel(ctx)
	// Clean up workflow leverages global context.
	defer cleanup(ctx, jobsCancel, workers)

	// Performance testing workflow leverages dedicated context.
	startTime := time.Now()
	run(jobsCtx, workers)
	endTime := time.Now()

	metrics.Summary(WorkerNumber, JobsPerWorker, startTime, endTime)

	// Sleep 30 seconds before proceeding with cleanup, because the deletions triggered by
	// performance testing might still be ongoing.
	klog.V(4).Infof("sleep %v seconds before cleanup, waiting for deletions to be gracefully proceeded", SleepTimeInSeconds)
	time.Sleep(time.Second * time.Duration(SleepTimeInSeconds))
}

// run tells all workers to run performance testing workflow and waits for them to complete.
func run(ctx context.Context, workers []*worker.Worker) {
	klog.V(4).Info("performance testing has started")

	jobsWaitGroup := &sync.WaitGroup{}
	jobsWaitGroup.Add(len(workers))

	for _, worker := range workers {
		go worker.Run(ctx, JobsPerWorker, jobsWaitGroup)
	}

	klog.V(4).Info("waiting for all workers to complete performance testing... work! work!")
	jobsWaitGroup.Wait()
	klog.V(4).Info("performance testing complete!")
}

// cleanup tells all workers to run clean up workflow and waits for them to complete.
func cleanup(ctx context.Context, cancel context.CancelFunc, workers []*worker.Worker) {
	cancel()

	klog.V(4).Info("cleanup has started")

	cleanupWaitGroup := &sync.WaitGroup{}
	cleanupWaitGroup.Add(len(workers))

	for _, worker := range workers {
		go worker.Cleanup(ctx, cleanupWaitGroup)
	}

	klog.V(4).Info("waiting for all workers to complete clean up... work! work!")
	cleanupWaitGroup.Wait()
	klog.V(4).Info("cleanup complete!")
}
