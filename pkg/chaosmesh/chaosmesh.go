package chaosmesh

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/nemoremold/perftests/pkg/utils"
)

// Blood for the Blood God!
type ChaosAgent struct {
	// Client is the k8s client from controller-runtime package used to operate IOChaos resource.
	Client client.Client

	// pollIntervalInSeconds is the interval between polls when waiting for IOChaos status change.
	pollIntervalInSeconds int
	// pollIntervalInSeconds is the timeout between polls when waiting for IOChaos status change.
	pollTimeoutInSeconds int
	// ioChaosTemplate is the template IOChaos CRO actual resources created from.
	ioChaosTemplate *v1alpha1.IOChaos
}

// NewChaosAgent instantiates a new ChaosAgent that operates on IOChaos resource.
func NewChaosAgent(kubeconfig string, templateFilePath string, interval, timeout int) (*ChaosAgent, error) {
	var (
		cfg *rest.Config
		err error
	)
	if len(kubeconfig) == 0 {
		cfg, err = rest.InClusterConfig()
	} else {
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	if err != nil {
		return nil, err
	}

	c, err := client.New(cfg, client.Options{})
	if err != nil {
		return nil, err
	}
	if err := v1alpha1.AddToScheme(c.Scheme()); err != nil {
		return nil, err
	}

	codecs := serializer.NewCodecFactory(c.Scheme())
	template, err := readIOChaosFromFile(codecs, templateFilePath)
	if err != nil {
		return nil, err
	}

	return &ChaosAgent{
		Client: c,

		pollIntervalInSeconds: interval,
		pollTimeoutInSeconds:  timeout,
		ioChaosTemplate:       template,
	}, nil
}

// readIOChaosFromFile reads a template file and instantiates a IOChaos object
// from it using a codec factory.
func readIOChaosFromFile(codecs serializer.CodecFactory, templateFilePath string) (*v1alpha1.IOChaos, error) {
	templateData, err := ioutil.ReadFile(templateFilePath)
	if err != nil {
		return nil, err
	}

	templateObj, gvk, err := codecs.UniversalDecoder(v1alpha1.GroupVersion).Decode(templateData, nil, nil)
	if err != nil {
		return nil, err
	}

	ioChaos, ok := templateObj.(*v1alpha1.IOChaos)
	if !ok {
		return nil, fmt.Errorf("got unexpected IOChaos group version kind: %v", gvk)
	}
	return ioChaos, nil
}

// NewIOChaos takes a delay and a percent and creates a new IOChaos CRO from
// the preconfigured template.
func (agent *ChaosAgent) NewIOChaos(delay string, percent int) *v1alpha1.IOChaos {
	ioChaos := agent.ioChaosTemplate.DeepCopy()
	ioChaos.Spec.Percent = percent
	ioChaos.Spec.Delay = delay
	return ioChaos
}

// Create creates a given IOChaos and wait until it is ready.
func (agent *ChaosAgent) Create(ctx context.Context, ioChaos *v1alpha1.IOChaos) error {
	namespacedName := utils.NamespacedName(ioChaos)
	klog.V(4).Infof("creating IOChaos %v", namespacedName)

	retryErr := retry.OnError(retry.DefaultBackoff, utils.AlwaysRetriable, func() error {
		if err := agent.Client.Create(ctx, ioChaos); err != nil {
			if !errors.IsAlreadyExists(err) {
				return err
			}
		}
		return nil
	})
	if retryErr != nil {
		return fmt.Errorf("failed creating IOChaos %v: %w", namespacedName, retryErr)
	}

	pollErr := wait.Poll(time.Second*time.Duration(agent.pollIntervalInSeconds), time.Second*time.Duration(agent.pollTimeoutInSeconds), func() (done bool, err error) {
		obj := v1alpha1.IOChaos{}
		if err := agent.Client.Get(ctx, client.ObjectKey{
			Namespace: ioChaos.Namespace,
			Name:      ioChaos.Name,
		}, &obj); err != nil {
			return false, nil
		}

		for _, record := range obj.Status.Experiment.Records {
			if record.Phase != v1alpha1.Injected {
				return false, nil
			}
		}

		return true, nil
	})
	if pollErr != nil {
		return fmt.Errorf("failed waiting for IOChaos %v to get ready: %w", namespacedName, pollErr)
	}

	klog.V(4).Infof("IOChaos %v successfully created", namespacedName)
	return nil
}

// Delete deletes a given IOChaos and wait until it is gone.
func (agent *ChaosAgent) Delete(ctx context.Context, ioChaos *v1alpha1.IOChaos) error {
	namespacedName := utils.NamespacedName(ioChaos)
	klog.V(4).Infof("deleting IOChaos %v", namespacedName)

	retryErr := retry.OnError(retry.DefaultBackoff, utils.AlwaysRetriable, func() error {
		if err := agent.Client.Delete(ctx, ioChaos); err != nil {
			if !errors.IsNotFound(err) {
				return err
			}
		}
		return nil
	})
	if retryErr != nil {
		return fmt.Errorf("failed deleting IOChaos %v: %w", namespacedName, retryErr)
	}

	pollErr := wait.Poll(time.Second*time.Duration(agent.pollIntervalInSeconds), time.Second*time.Duration(agent.pollTimeoutInSeconds), func() (done bool, err error) {
		dummy := v1alpha1.IOChaos{}
		if err := agent.Client.Get(ctx, client.ObjectKey{
			Namespace: ioChaos.Namespace,
			Name:      ioChaos.Name,
		}, &dummy); err != nil {
			if errors.IsNotFound(err) {
				return true, nil
			}
		}

		return false, nil
	})
	if pollErr != nil {
		return fmt.Errorf("failed waiting for IOChaos %v to be deleted: %w", namespacedName, pollErr)
	}

	klog.V(4).Infof("IOChaos %v successfully deleted", namespacedName)
	return nil
}
