package testflow

import (
	"context"
	"fmt"
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
	// Initialize ChaosAgent.
	agent, err := chaosmesh.NewChaosAgent(
		opts.IOChaosKubeconfigFilePath,
		opts.ChaosAgentIOChaosTemplateFilePath,
		opts.ChaosAgentPollIntervalInSeconds,
		opts.ChaosAgentPollTimeoutInSeconds,
	)
	if err != nil {
		return nil, err
	}

	// Initialize workers.
	var workers []*worker.Worker
	for workerID := 0; workerID < opts.WorkerNumber; workerID++ {
		w, err := worker.NewWorker(workerID, opts.KubeconfigFilePath)
		if err != nil {
			return nil, err
		}
		workers = append(workers, w)
	}

	// Initialize report exporter.
	var exporter *metrics.Exporter
	if opts.WriteToCSV {
		exporter = metrics.NewExporter(opts.Latencies, opts.PercentsStr)
	}

	return &TestFlow{
		Options:  opts,
		Agent:    agent,
		Workers:  workers,
		Exporter: exporter,
	}, nil
}

// RunTestFlow iterates the pre-defined range of percents and latencies to be applied
// to the IOChaos, and runs the test flow with every (latency, percent) pair setting.
func (flow *TestFlow) RunTestFlow(ctx context.Context) error {
	testFlowContext, testFlowCancel := context.WithCancel(ctx)
	// Use `context.Background` instead of `ctx` to be able to export current collected
	// metrics to a report when stop signal is received and the test has stopped.
	writerContext, writerCancel := context.WithCancel(context.Background())
	defer func() {
		if err := recover(); err != nil {
			klog.Errorf("recovered from error to safely terminate test flow: %v", err)
		}
		testFlowCancel()
		writerCancel()
	}()

	klog.V(2).Info("starting test flow")
	startTime := time.Now()
	cancelled := false
	for percentIndex := range flow.Percents {
		for latencyIndex := range flow.Latencies {
			select {
			case <-ctx.Done():
				klog.V(2).Info("stop signal received, stopping test flow")
				cancelled = true
				break
			default:
				if err := flow.startTestFlowWithIOChaos(testFlowContext, percentIndex, latencyIndex); err != nil {
					return err
				}
			}
			if cancelled {
				break
			}
		}
		if cancelled {
			break
		}
	}
	endTime := time.Now()
	klog.V(2).Info("finished test flow")
	klog.V(2).Infof("test flow started at %v", startTime.Local())
	klog.V(2).Infof("test flow finished at %v", endTime.Local())
	klog.V(2).Infof("test flow duration: %v", endTime.Sub(startTime).String())

	// Export the final report to a CSV file.
	if flow.WriteToCSV {
		// Export the report to a CSV file.
		flow.Exporter.WriteToCSV(writerContext, flow.Options, startTime)
	}
	return nil
}

// startTestFlowWithIOChaos prepares the IOChaos before running the actual tests and deletes
// it after the test has finished.
func (flow *TestFlow) startTestFlowWithIOChaos(ctx context.Context, percentIndex, latencyIndex int) (err error) {
	totalTests := len(flow.Latencies) * len(flow.Percents)
	currentTest := percentIndex*len(flow.Latencies) + latencyIndex + 1
	klog.V(2).Infof("starting tests (%v/%v) with IOChaos (latency: %v, percent: %v)",
		currentTest,
		totalTests,
		flow.Latencies[latencyIndex],
		flow.Percents[percentIndex],
	)

	// Prepare new IOChaos.
	ioChaos := flow.Agent.NewIOChaos(flow.Latencies[latencyIndex], flow.Percents[percentIndex])

	// TODO: cleanup tasks currently return no error info, so this process might still fail, causing the test to run in an unclean environment.
	// ALWAYS DO CLEANUP WITHOUT IOCHAOS! - RUN CLEANUP FIRST!
	// Ensure the environment is clean before testing.
	klog.V(4).Info("cleaning up testing environment before performance testing")
	flow.cleanup(context.Background())

	// ALWAYS DO CLEANUP WITHOUT IOCHAOS! - DEFER CLEANUP FIRST!
	// Prepare context dedicated for performance testing. When stop signal
	// is received, this dedicated context will be cancelled first, triggering
	// the clean up process before actually stopping the program. When stop
	// signal is received a second time, the cleanup will be forcefully quited.
	jobsCtx, jobsCancel := context.WithCancel(ctx)
	defer func() {
		jobsCancel()
		flow.cleanup(context.Background())
	}()

	// Ensure IOChaos is deleted after each test.
	defer func() {
		if deleteErr := flow.Agent.Delete(context.Background(), ioChaos); deleteErr != nil {
			err = fmt.Errorf("%v: %w", deleteErr.Error(), err)
		}
	}()

	// Create corresponding IOChaos before each test.
	if err = flow.Agent.Create(context.Background(), ioChaos); err != nil {
		return
	}

	// Run the actual test flow.
	if err = flow.startTestFlow(jobsCtx, percentIndex, latencyIndex); err == nil {
		klog.V(2).Infof("successfully finished tests with IOChaos (latency: %v, percent: %v)", flow.Latencies[latencyIndex], flow.Percents[percentIndex])
	}
	return
}

// startTestFlow does the actual performance testing, cleaning up the test environment before and
// after the tests.
func (flow *TestFlow) startTestFlow(ctx context.Context, percentIndex, latencyIndex int) error {
	set := metrics.MetricSetID{
		Latency: flow.Latencies[latencyIndex],
		Percent: flow.PercentsStr[percentIndex],
	}

	// Performance testing workflow leverages dedicated context.
	klog.V(4).Info("starting up testing environment before performance testing")
	startTime := time.Now()
	flow.performanceTest(ctx, set)
	endTime := time.Now()

	// Print summary for a single test.
	if flow.Summarize {
		// Print the report in stdout.
		metrics.Summary(set, flow.WorkerNumber, flow.JobsPerWorker, startTime, endTime)
	}
	// Collect metrics for final report right after a test has finished to avoid
	// the metrics from expiring (Prometheus Summary metrics has MaxAge).
	if flow.WriteToCSV {
		if err := flow.Exporter.Collect(percentIndex, latencyIndex); err != nil {
			klog.Errorf("failed to collect metrics for testing with IOChaos (latency: %v, percent: %v)", flow.Latencies[latencyIndex], flow.Percents[percentIndex])
		}
	}

	// Wait some time before proceeding with cleanup, because the deletions triggered by
	// performance testing might still be ongoing.
	klog.V(4).Infof("sleeping %v seconds before cleanup, waiting for deletions to be gracefully proceeded", flow.SleepTimeInSeconds)
	time.Sleep(time.Second * time.Duration(flow.SleepTimeInSeconds))
	return nil
}

// run tells all workers to run performance testing workflow and waits for them to complete.
func (flow *TestFlow) performanceTest(ctx context.Context, set metrics.MetricSetID) {
	klog.V(4).Info("performance testing has started")

	jobsWaitGroup := &sync.WaitGroup{}
	jobsWaitGroup.Add(len(flow.Workers))

	for _, w := range flow.Workers {
		go w.Run(ctx, flow.JobsPerWorker, jobsWaitGroup, set)
	}

	klog.V(4).Info("waiting for all workers to complete performance testing... work! work!")
	jobsWaitGroup.Wait()
	klog.V(4).Info("performance testing complete!")
}

// cleanup tells all workers to run clean up workflow and waits for them to complete.
func (flow *TestFlow) cleanup(ctx context.Context) {
	klog.V(4).Info("cleanup has started")

	cleanupWaitGroup := &sync.WaitGroup{}
	cleanupWaitGroup.Add(len(flow.Workers))

	for _, w := range flow.Workers {
		go w.Cleanup(ctx, cleanupWaitGroup)
	}

	klog.V(4).Info("waiting for all workers to complete clean up... work! work!")
	cleanupWaitGroup.Wait()
	klog.V(4).Info("cleanup complete!")
}
