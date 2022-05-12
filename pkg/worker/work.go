package worker

import (
	"context"
	"sync"

	v1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

type Worker struct {
	// ID is the unique identity number for the worker.
	ID int

	// Client is the k8s client used to talk to the API server.
	Client *kubernetes.Clientset

	// Deployments is a list of deployments that the worker created.
	Deployments []*v1.Deployment
}

// NewWorker initializes a new worker.
func NewWorker(workerId int, kubeconfig string) (*Worker, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
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
func (w *Worker) Run(ctx context.Context, numberOfJobs int, wg *sync.WaitGroup) {
	defer wg.Done()

	klog.V(4).Infof("[worker %v] has started performance testing", w.ID)

	w.testCreateDeployments(ctx, numberOfJobs)
	w.testGetDeployments(ctx)
	w.testUpdateDeployments(ctx)
	w.testPatchDeployments(ctx)
	w.testListDeployments(ctx)
	w.testDeleteDeployments(ctx)

	klog.V(4).Infof("[worker %v] performance testing done!", w.ID)
}

// Cleanup starts the clean up workflow of a worker.
func (w *Worker) Cleanup(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	klog.V(4).Infof("[worker %v] has started cleanup", w.ID)

	w.cleanupDeployments(ctx)
	w.cleanupPods(ctx)

	klog.V(4).Infof("[worker %v] cleanup done!", w.ID)
}
