package testflow

import (
	"context"
	"sync"
	"time"

	"k8s.io/klog/v2"

	"github.com/nemoremold/perftests/pkg/chaosmesh"
	"github.com/nemoremold/perftests/pkg/metrics"
	"github.com/nemoremold/perftests/pkg/options"
	"github.com/nemoremold/perftests/pkg/worker"
)

// TestFlow defines the test flow of the performance testing.
type TestFlow struct {
	*options.Options

	// Agent is the ChaosAgent that operates on the IOChaos objects.
	Agent *chaosmesh.ChaosAgent

	// Workers do actual performance testing and resource cleanup.
	Workers []*worker.Worker

	// Exporter collects metrics data and generates the final report,
	// exporting it to a CSV file.
	Exporter *metrics.Exporter
}

// NewTestFlow instantiates a new performance testing test flow.
func NewTestFlow(opts *options.Options) (*TestFlow, error) {
	// Initialize workers.
	var workers []*worker.Worker
	for workerID := 0; workerID < opts.WorkerNumber; workerID++ {
		w, err := worker.NewWorker(workerID, opts.KubeconfigFilePath)
		if err != nil {
			return nil, err
		}
		workers = append(workers, w)
	}

	return &TestFlow{
		Options: opts,
		Workers: workers,
	}, nil
}

// RunTestFlow iterates the pre-defined range of percents and latencies to be applied
// to the IOChaos, and runs the test flow with every (latency, percent) pair setting.
func (flow *TestFlow) RunTestFlow(ctx context.Context) error {
	jobsCtx, jobsCancel := context.WithCancel(ctx)

	// Cleanup uses different context, `ctx` will get cancelled when stopping signal
	// is received for the first time, which will trigger the stopping of all running
	// jobs, starting the cleanup process. When the signal is received a second time,
	// cleanups will be forcefully quited.
	klog.V(2).Info("cleaning up at the start.")
	flow.cleanup(context.Background())

	defer func() {
		klog.V(2).Info("cleaning up at the end.")
		jobsCancel()
		flow.cleanup(context.Background())
	}()

	return flow.startTestFlow(jobsCtx, 0, 0)
}

// startTestFlow does the actual performance testing, cleaning up the test environment before and
// after the tests.
func (flow *TestFlow) startTestFlow(ctx context.Context, percentIndex, latencyIndex int) error {
	set := metrics.MetricSetID{
		Latency: flow.Latencies[latencyIndex],
		Percent: flow.PercentsStr[percentIndex],
	}

	startTime := time.Now()
	flow.performanceTest(ctx, set)
	endTime := time.Now()

	if flow.Summarize {
		metrics.Summary(set, flow.WorkerNumber, flow.JobsPerWorker, startTime, endTime)
	}

	klog.V(4).Infof("sleeping %v seconds before cleanup, waiting for deletions to be gracefully proceeded.", flow.SleepTimeInSeconds)
	time.Sleep(time.Second * time.Duration(flow.SleepTimeInSeconds))
	return nil
}

// run tells all workers to run performance testing workflow and waits for them to complete.
func (flow *TestFlow) performanceTest(ctx context.Context, set metrics.MetricSetID) {
	jobsWaitGroup := &sync.WaitGroup{}
	jobsWaitGroup.Add(len(flow.Workers))

	klog.V(4).Info("starting workers.")
	for _, w := range flow.Workers {
		go w.Run(ctx, flow.JobsPerWorker, jobsWaitGroup, set)
	}

	klog.V(4).Info("waiting for all workers to complete work... work! work!")
	jobsWaitGroup.Wait()
	klog.V(4).Info("work complete!")
}

// cleanup tells all workers to run clean up workflow and waits for them to complete.
func (flow *TestFlow) cleanup(ctx context.Context) {
	klog.V(4).Info("starting cleanup")
	cleanupWaitGroup := &sync.WaitGroup{}
	cleanupWaitGroup.Add(len(flow.Workers))

	for _, w := range flow.Workers {
		go w.Cleanup(ctx, cleanupWaitGroup)
	}

	klog.V(4).Info("waiting for all workers to complete clean up... work! work!")
	cleanupWaitGroup.Wait()
	klog.V(4).Info("cleanup complete!")
}
