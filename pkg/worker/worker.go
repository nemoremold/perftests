package worker

import (
	"context"
	"sync"

	v1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/flowcontrol"
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
	Deployments []*v1.Deployment
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

	config.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(100, 50)

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
			klog.Errorf("[worker %v] has stopped performance testing due to error: %v", err)
		}
	}()

	klog.V(4).Infof("[worker %v] has started performance testing", w.ID)

	w.Deployments = nil
	w.testCreateDeployments(ctx, numberOfJobs, set)
	w.testGetDeployments(ctx, set)
	w.testUpdateDeployments(ctx, set)
	w.testPatchDeployments(ctx, set)
	w.testListDeployments(ctx, set)
	w.testDeleteDeployments(ctx, set)

	klog.V(4).Infof("[worker %v] performance testing done!", w.ID)
}

// Cleanup starts the clean up workflow of a worker.
func (w *Worker) Cleanup(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() {
		if err := recover(); err != nil {
			klog.Errorf("[worker %v] has stopped cleanup due to error: %v", err)
		}
	}()

	klog.V(4).Infof("[worker %v] has started cleanup", w.ID)

	w.cleanupDeployments(ctx)
	w.cleanupPods(ctx)

	klog.V(4).Infof("[worker %v] cleanup done!", w.ID)
}
