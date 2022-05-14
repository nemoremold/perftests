package worker

import (
	"encoding/json"

	v1 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

type patchValue struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

var (
	// DeploymentTemplate is the template workers use to create actual Deployment CRO from.
	DeploymentTemplate *v1.Deployment

	// PatchData is used to patch an existing deployment, replacing the annotations of it.
	PatchData []byte

	// PatchData is used to update an existing deployment, adding a new annotation to it.
	UpdateData = map[string]string{
		"updated": "true",
	}
)

func init() {
	DeploymentTemplate = &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				AppLabel: AppName,
			},
		},
		Spec: v1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(3),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					AppLabel: AppName,
				},
			},
			Template: v12.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						AppLabel: AppName,
					},
				},
				Spec: v12.PodSpec{
					Containers: []v12.Container{{
						Name:  AppName,
						Image: AppImage,
						Ports: []v12.ContainerPort{
							{
								ContainerPort: 80,
							},
						},
					}},
				},
			},
		},
	}

	Patch := []*patchValue{{
		Op:   "replace",
		Path: "/metadata/annotations",
		Value: map[string]string{
			"patched": "true",
		},
	}}
	PatchData, _ = json.Marshal(Patch)
}

const (
	// WorkerIDLabel is the label used on deployments and pods
	WorkerIDLabel = "workerId"
	// AppLabel is the label specifying the name of the app
	AppLabel = "app"
	// AppName is the name of the app
	AppName = "nginx"
	// AppImage is the image of the app
	AppImage = "nginx:1.14.2"
)
