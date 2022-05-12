package worker

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"

	"github.com/nemoremold/perftests/pkg/constants"
	"github.com/nemoremold/perftests/pkg/metrics"
)

func getDuration(previousTime time.Time) time.Duration {
	return time.Since(previousTime)
}

func (w *Worker) testCreateDeployments(ctx context.Context, numberOfJobs int) {
	for jobId := 0; jobId < numberOfJobs; jobId++ {
		select {
		case <-ctx.Done():
			klog.V(2).Infof("[worker %v] has received stop signal, now exiting creation tests", w.ID)
			return
		default:
			deployment := constants.DeploymentTemplate.DeepCopy()
			deployment.Name = constants.AppName + "-" + fmt.Sprint(w.ID*numberOfJobs*10+jobId)
			deployment.Labels[constants.WorkerIDLabel] = fmt.Sprint(w.ID)
			deployment.Spec.Selector.MatchLabels[constants.WorkerIDLabel] = fmt.Sprint(w.ID)
			deployment.Spec.Template.Labels[constants.WorkerIDLabel] = fmt.Sprint(w.ID)

			startTime := time.Now()
			if createdDeployment, err := w.Client.AppsV1().Deployments("default").Create(ctx, deployment, metav1.CreateOptions{}); err != nil {
				if errors.IsAlreadyExists(err) {
					w.Deployments = append(w.Deployments, deployment)
					klog.V(4).Infof("[worker %v] finds that deployment %v already exists", w.ID, deployment.Name)
				} else {
					metrics.RecordAPIRequest("create", w.ID, false, getDuration(startTime))
					klog.Errorf("[worker %v] has failed creating deployment %v: %v", w.ID, deployment.Name, err.Error())
				}
			} else {
				metrics.RecordAPIRequest("create", w.ID, true, getDuration(startTime))
				w.Deployments = append(w.Deployments, deployment)
				klog.V(4).Infof("[worker %v] has successfully created deployment %v", w.ID, createdDeployment.Name)
			}
		}
	}
}

func (w *Worker) testGetDeployments(ctx context.Context) {
	for _, deployment := range w.Deployments {
		select {
		case <-ctx.Done():
			klog.V(2).Infof("[worker %v] has received stop signal, now exiting getting tests", w.ID)
			return
		default:
			startTime := time.Now()
			if gotDeployment, err := w.Client.AppsV1().Deployments("default").Get(ctx, deployment.Name, metav1.GetOptions{}); err != nil {
				metrics.RecordAPIRequest("get", w.ID, false, getDuration(startTime))
				klog.Errorf("[worker %v] has failed getting deployment %v: %v", w.ID, deployment.Name, err.Error())
			} else {
				metrics.RecordAPIRequest("get", w.ID, true, getDuration(startTime))
				klog.V(4).Infof("[worker %v] has successfully got deployment %v", w.ID, gotDeployment.Name)
			}
		}
	}
}

func (w *Worker) testUpdateDeployments(ctx context.Context) {
	for index, deployment := range w.Deployments {
		select {
		case <-ctx.Done():
			klog.V(2).Infof("[worker %v] has received stop signal, now exiting updating tests", w.ID)
			return
		default:
			deployment.Annotations = constants.UpdateData

			startTime := time.Now()
			if updatedDeployment, err := w.Client.AppsV1().Deployments("default").Update(ctx, deployment, metav1.UpdateOptions{}); err != nil {
				metrics.RecordAPIRequest("update", w.ID, false, getDuration(startTime))
				klog.Errorf("[worker %v] has failed updating deployment %v: %v", w.ID, deployment.Name, err.Error())
			} else {
				metrics.RecordAPIRequest("update", w.ID, true, getDuration(startTime))
				w.Deployments[index] = deployment
				klog.V(4).Infof("[worker %v] has successfully updated deployment %v", w.ID, updatedDeployment.Name)
			}
		}
	}
}

func (w *Worker) testPatchDeployments(ctx context.Context) {
	for _, deployment := range w.Deployments {
		select {
		case <-ctx.Done():
			klog.V(2).Infof("[worker %v] has received stop signal, now exiting patching tests", w.ID)
			return
		default:
			startTime := time.Now()
			if patchedDeployment, err := w.Client.AppsV1().Deployments("default").Patch(ctx, deployment.Name, types.JSONPatchType, constants.PatchData, metav1.PatchOptions{}); err != nil {
				metrics.RecordAPIRequest("patch", w.ID, false, getDuration(startTime))
				klog.Errorf("[worker %v] has failed patching deployment %v: %v", w.ID, deployment.Name, err.Error())
			} else {
				metrics.RecordAPIRequest("patch", w.ID, true, getDuration(startTime))
				klog.V(4).Infof("[worker %v] has successfully patched deployment %v", w.ID, patchedDeployment.Name)
			}
		}
	}
}

func (w *Worker) testListDeployments(ctx context.Context) {
	select {
	case <-ctx.Done():
		klog.V(2).Infof("[worker %v] has received stop signal, now exiting listing tests", w.ID)
		return
	default:
		startTime := time.Now()
		if _, err := w.Client.AppsV1().Deployments("default").List(ctx, metav1.ListOptions{
			LabelSelector: metav1.FormatLabelSelector(&metav1.LabelSelector{
				MatchLabels: map[string]string{
					constants.AppLabel:      constants.AppName,
					constants.WorkerIDLabel: fmt.Sprint(w.ID),
				},
			}),
		}); err != nil {
			metrics.RecordAPIRequest("list", w.ID, false, getDuration(startTime))
			klog.Errorf("[worker %v] has failed listing deployments: %v", w.ID, err.Error())
		} else {
			metrics.RecordAPIRequest("list", w.ID, true, getDuration(startTime))
			klog.V(4).Infof("[worker %v] has successfully listed deployments", w.ID)
		}
	}
}

func (w *Worker) testDeleteDeployments(ctx context.Context) {
	for index, deployment := range w.Deployments {
		select {
		case <-ctx.Done():
			klog.V(2).Infof("[worker %v] has received stop signal, now exiting deleting tests", w.ID)
			return
		default:
			startTime := time.Now()
			if err := w.Client.AppsV1().Deployments("default").Delete(ctx, deployment.Name, metav1.DeleteOptions{}); err != nil {
				metrics.RecordAPIRequest("delete", w.ID, false, getDuration(startTime))
				klog.Errorf("[worker %v] has failed deleting deployment %v: %v", w.ID, deployment.Name, err.Error())
			} else {
				metrics.RecordAPIRequest("delete", w.ID, true, getDuration(startTime))
				w.Deployments[index] = nil
				klog.V(4).Infof("[worker %v] has successfully deleted deployment %v", w.ID, deployment.Name)
			}
		}
	}
}
