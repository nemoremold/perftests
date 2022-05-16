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
	"github.com/nemoremold/perftests/pkg/utils"
)

func (w *Worker) testCreateDeployments(ctx context.Context, set metrics.MetricSetID) {
	deployment := DeploymentTemplate.DeepCopy()
	deployment.GenerateName = AppName + "-" + fmt.Sprint(w.ID) + "-"
	deployment.Labels[WorkerIDLabel] = fmt.Sprint(w.ID)
	deployment.Spec.Selector.MatchLabels[WorkerIDLabel] = fmt.Sprint(w.ID)
	deployment.Spec.Template.Labels[WorkerIDLabel] = fmt.Sprint(w.ID)

	klog.V(4).Infof("[worker %v] has started to create deployment %v", w.ID, deployment.Name)
	startTime := time.Now()
	if createdDeployment, err := w.Client.AppsV1().Deployments("default").Create(ctx, deployment, metav1.CreateOptions{}); err != nil {
		if errors.IsAlreadyExists(err) {
			w.Deployment = deployment
			klog.V(4).Infof("[worker %v] finds that deployment %v already exists", w.ID, deployment.Name)
		} else {
			metrics.RecordAPIRequest(constants.CREATE, false, utils.GetDurationSince(startTime), set)
			klog.Errorf("[worker %v] has failed to create deployment %v: %v", w.ID, deployment.Name, err.Error())
		}
	} else {
		metrics.RecordAPIRequest(constants.CREATE, true, utils.GetDurationSince(startTime), set)
		w.Deployment = deployment
		w.Deployment.Name = createdDeployment.Name
		klog.V(4).Infof("[worker %v] has successfully created deployment %v", w.ID, createdDeployment.Name)
	}
}

func (w *Worker) testGetDeployments(ctx context.Context, set metrics.MetricSetID) {
	if w.Deployment != nil {
		klog.V(4).Infof("[worker %v] has started to get deployment %v", w.ID, w.Deployment.Name)
		startTime := time.Now()
		if gotDeployment, err := w.Client.AppsV1().Deployments("default").Get(ctx, w.Deployment.Name, metav1.GetOptions{}); err != nil {
			metrics.RecordAPIRequest(constants.GET, false, utils.GetDurationSince(startTime), set)
			klog.Errorf("[worker %v] has failed to get deployment %v: %v", w.ID, w.Deployment.Name, err.Error())
		} else {
			metrics.RecordAPIRequest(constants.GET, true, utils.GetDurationSince(startTime), set)
			klog.V(4).Infof("[worker %v] has successfully got deployment %v", w.ID, gotDeployment.Name)
		}
	}
}

func (w *Worker) testUpdateDeployments(ctx context.Context, set metrics.MetricSetID) {
	if w.Deployment != nil {
		klog.V(4).Infof("[worker %v] has started to update deployment %v", w.ID, w.Deployment.Name)
		w.Deployment.Annotations = UpdateData
		startTime := time.Now()
		if updatedDeployment, err := w.Client.AppsV1().Deployments("default").Update(ctx, w.Deployment, metav1.UpdateOptions{}); err != nil {
			metrics.RecordAPIRequest(constants.UPDATE, false, utils.GetDurationSince(startTime), set)
			klog.Errorf("[worker %v] has failed to update deployment %v: %v", w.ID, w.Deployment.Name, err.Error())
		} else {
			metrics.RecordAPIRequest(constants.UPDATE, true, utils.GetDurationSince(startTime), set)
			klog.V(4).Infof("[worker %v] has successfully updated deployment %v", w.ID, updatedDeployment.Name)
		}
	}
}

func (w *Worker) testPatchDeployments(ctx context.Context, set metrics.MetricSetID) {
	if w.Deployment != nil {
		klog.V(4).Infof("[worker %v] has started to patch deployment %v", w.ID, w.Deployment.Name)
		startTime := time.Now()
		if patchedDeployment, err := w.Client.AppsV1().Deployments("default").Patch(ctx, w.Deployment.Name, types.JSONPatchType, PatchData, metav1.PatchOptions{}); err != nil {
			metrics.RecordAPIRequest(constants.PATCH, false, utils.GetDurationSince(startTime), set)
			klog.Errorf("[worker %v] has failed to patch deployment %v: %v", w.ID, w.Deployment.Name, err.Error())
		} else {
			metrics.RecordAPIRequest(constants.PATCH, true, utils.GetDurationSince(startTime), set)
			klog.V(4).Infof("[worker %v] has successfully patched deployment %v", w.ID, patchedDeployment.Name)
		}
	}
}

func (w *Worker) testListDeployments(ctx context.Context, set metrics.MetricSetID) {
	klog.V(4).Infof("[worker %v] has started to list deployments", w.ID)
	startTime := time.Now()
	if _, err := w.Client.AppsV1().Deployments("default").List(ctx, metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(&metav1.LabelSelector{
			MatchLabels: map[string]string{
				AppLabel:      AppName,
				WorkerIDLabel: fmt.Sprint(w.ID),
			},
		}),
	}); err != nil {
		metrics.RecordAPIRequest(constants.LIST, false, utils.GetDurationSince(startTime), set)
		klog.Errorf("[worker %v] has failed to list deployments: %v", w.ID, err.Error())
	} else {
		metrics.RecordAPIRequest(constants.LIST, true, utils.GetDurationSince(startTime), set)
		klog.V(4).Infof("[worker %v] has successfully listed deployments", w.ID)
	}
}

func (w *Worker) testDeleteDeployments(ctx context.Context, set metrics.MetricSetID) {
	if w.Deployment != nil {
		klog.V(4).Infof("[worker %v] has started to delete deployment %v", w.ID, w.Deployment.Name)
		startTime := time.Now()
		if err := w.Client.AppsV1().Deployments("default").Delete(ctx, w.Deployment.Name, metav1.DeleteOptions{}); err != nil {
			metrics.RecordAPIRequest(constants.DELETE, false, utils.GetDurationSince(startTime), set)
			klog.Errorf("[worker %v] has failed to delete deployment %v: %v", w.ID, w.Deployment.Name, err.Error())
		} else {
			metrics.RecordAPIRequest(constants.DELETE, true, utils.GetDurationSince(startTime), set)
			klog.V(4).Infof("[worker %v] has successfully deleted deployment %v", w.ID, w.Deployment.Name)
			w.Deployment = nil
		}
	}
}
