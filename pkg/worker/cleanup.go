package worker

import (
	"context"
	"fmt"

	v1 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
)

func (w *Worker) cleanupDeployments(ctx context.Context) {
	klog.V(4).Infof("[worker %v] has started to clean up left-over deployments", w.ID)

	var remainingDeployments []v1.Deployment

	// TODO: fix context.
	if err := retry.OnError(retry.DefaultRetry, func(_ error) bool {
		select {
		case <-ctx.Done():
			return false
		default:
			return true
		}
	}, func() error {
		deploymentList, err := w.Client.AppsV1().Deployments("default").List(ctx, metav1.ListOptions{
			LabelSelector: metav1.FormatLabelSelector(&metav1.LabelSelector{
				MatchLabels: map[string]string{
					AppLabel:      AppName,
					WorkerIDLabel: fmt.Sprint(w.ID),
				},
			})})
		if err == nil {
			remainingDeployments = deploymentList.Items
		}
		return err
	}); err != nil {
		klog.Errorf("[worker %v] has failed to list remaining deployments for cleanup: %v", w.ID, err.Error())
	}

	if len(remainingDeployments) > 0 {
		klog.V(4).Infof("[worker %v] has found %v remaining deployments, starting cleanup", w.ID, len(remainingDeployments))
	} else {
		klog.V(4).Infof("[worker %v] has found no remaining deployments", w.ID)
	}

	for _, deployment := range remainingDeployments {
		select {
		case <-ctx.Done():
			klog.V(2).Infof("[worker %v] has received stop signal, now exiting cleanup", w.ID)
			return
		default:
			if err := w.Client.AppsV1().Deployments("default").Delete(ctx, deployment.Name, metav1.DeleteOptions{}); err != nil {
				if !errors.IsNotFound(err) {
					klog.Errorf("[worker %v] has failed to delete deployment %v", w.ID, deployment.Name)
				}
			} else {
				klog.V(4).Infof("[worker %v] has successfully deleted deployment %v", w.ID, deployment.Name)
			}
		}
	}

	klog.V(4).Infof("[worker %v] has finished cleaning up left-over deployments", w.ID)
}

func (w *Worker) cleanupPods(ctx context.Context) {
	klog.V(4).Infof("[worker %v] has started to clean up left-over pods", w.ID)

	var remainingPods []v12.Pod

	// TODO: fix context.
	if err := retry.OnError(retry.DefaultRetry, func(_ error) bool {
		select {
		case <-ctx.Done():
			return false
		default:
			return true
		}
	}, func() error {
		podList, err := w.Client.CoreV1().Pods("default").List(ctx, metav1.ListOptions{
			LabelSelector: metav1.FormatLabelSelector(&metav1.LabelSelector{
				MatchLabels: map[string]string{
					AppLabel:      AppName,
					WorkerIDLabel: fmt.Sprint(w.ID),
				},
			})})
		if err == nil {
			remainingPods = podList.Items
		}
		return err
	}); err != nil {
		klog.Errorf("[worker %v] has failed to list remaining pods for cleanup: %v", w.ID, err.Error())
	}

	if len(remainingPods) > 0 {
		klog.V(4).Infof("[worker %v] has found %v remaining pods, starting cleanup", w.ID, len(remainingPods))
	} else {
		klog.V(4).Infof("[worker %v] has found no remaining pods", w.ID)
	}

	for _, pod := range remainingPods {
		select {
		case <-ctx.Done():
			klog.V(2).Infof("[worker %v] has received stop signal, now exiting cleanup", w.ID)
			return
		default:
			if err := w.Client.CoreV1().Pods("default").Delete(ctx, pod.Name, metav1.DeleteOptions{}); err != nil {
				if !errors.IsNotFound(err) {
					klog.Errorf("[worker %v] has failed to delete deployment %v", w.ID, pod.Name)
				}
			} else {
				klog.V(4).Infof("[worker %v] has successfully deleted pod %v", w.ID, pod.Name)
			}
		}
	}

	klog.V(4).Infof("[worker %v] has finished cleaning up left-over pods", w.ID)
}
