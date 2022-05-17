package worker

import (
	"context"
	"sync"

	v1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	"github.com/nemoremold/perftests/pkg/metrics"
)

// Worker does actual performance testing and resource cleanup.
type Worker struct {
	// TODO: make this actual unique, lol, e.g. via using a factory.
	// ID is the unique identity number for the worker.
	ID int

	// Client is the k8s client used to talk to the API server.
	Client *kubernetes.Clientset

	// Deployments is a list of deployments that the worker created.
	Deployment *v1.Deployment
}

// NewWorker initializes a new worker.
func NewWorker(workerId int, kubeconfig string) (*Worker, error) {
	var (
		config *rest.Config
		err    error
	)
	if len(kubeconfig) == 0 {
		config, err = rest.InClusterConfig()
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	if err != nil {
		return nil, err
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Worker{
		ID:     workerId,
		Client: client,
	}, nil
}

// Run starts the performance testing workflow of a worker.
func (w *Worker) Run(ctx context.Context, numberOfJobs int, wg *sync.WaitGroup, set metrics.MetricSetID) {
	defer wg.Done()
	defer func() {
		if err := recover(); err != nil {
			klog.Errorf("[worker %v] stopping work due to panics: %v", w.ID, err)
		}
	}()

	klog.V(4).Infof("[worker %v] has started", w.ID)
	for jobID, loop := 0, true; loop && (jobID < numberOfJobs || numberOfJobs == -1); jobID++ {
		select {
		case <-ctx.Done():
			klog.V(4).Infof("[worker %v] stop signal received, stopping worker", w.ID)
			loop = false
			break
		default:
			w.Deployment = nil
			w.testCreateDeployments(ctx, set)
			w.testGetDeployments(ctx, set)
			w.testUpdateDeployments(ctx, set)
			w.testPatchDeployments(ctx, set)
			w.testListDeployments(ctx, set)
			w.testDeleteDeployments(ctx, set)
		}
	}
	klog.V(4).Infof("[worker %v] has stopped!", w.ID)
}

// Cleanup starts the clean up workflow of a worker.
func (w *Worker) Cleanup(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() {
		if err := recover(); err != nil {
			klog.Errorf("[worker %v] stopping cleanup due to panics: %v", w.ID, err)
		}
	}()

	klog.V(4).Infof("[worker %v] has started cleanup", w.ID)
	w.cleanupDeployments(ctx)
	w.cleanupPods(ctx)
	klog.V(4).Infof("[worker %v] cleanup done!", w.ID)
}
