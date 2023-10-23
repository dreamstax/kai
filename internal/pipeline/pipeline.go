package pipeline

import (
	"context"
	"fmt"

	corev1alpha1 "github.com/dreamstax/kai/api/core/v1alpha1"
	"github.com/dreamstax/kai/internal/pipeline/reconcilers/step"
	apierr "k8s.io/apimachinery/pkg/api/errors"

	ctrl "sigs.k8s.io/controller-runtime"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// temporary usage of kserve initializer
	storageInitializerImage       = "kserve/storage-initializer:v0.10.1"
	modelVolumeName               = "kai-mount-location"
	inferenceServiceContainerName = "kai-container"
)

type Client struct {
	kclient kclient.Client
}

func New(client kclient.Client) *Client {
	return &Client{
		kclient: client,
	}
}

func (c *Client) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	p := &corev1alpha1.Pipeline{}
	err := c.kclient.Get(ctx, req.NamespacedName, p)
	if err != nil {
		if apierr.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to retrieve latest pipeline %s: %w", req.NamespacedName, err)
	}

	// Iterate over steps defined in the pipeline and reconcile them
	// TODO: find some way to pass pipeline object meta to step reconciler
	// TODO: steps in a pipeline need to be defined by order - 1,2,3,n...
	// TODO: actually need to remove objectMeta StepTemplateSpec - this was borrowed from Deployment
	//       but a pipeline won't create multiple copies of the same step like a deployment
	//       creates multiple pods from a single deployment.
	for _, rec := range []func(context.Context, *corev1alpha1.Pipeline) error{
		step.NewReconciler(c.kclient).Reconcile,
	} {
		if err := rec(ctx, p); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// TODO: make a step for each pipeline StepTemplateSpec
