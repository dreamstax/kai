package step

import (
	"context"
	"fmt"

	corev1alpha1 "github.com/dreamstax/kai/api/core/v1alpha1"
	"github.com/dreamstax/kai/internal/credentials"
	"github.com/dreamstax/kai/internal/step/reconcilers/deployment"
	"github.com/dreamstax/kai/internal/step/reconcilers/hpa"
	"github.com/dreamstax/kai/internal/step/reconcilers/service"
	corev1 "k8s.io/api/core/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	ctrl "sigs.k8s.io/controller-runtime"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	storageInitializerImage       = "kserve/storage-initializer:v0.10.1"
	modelVolumeName               = "kai-mount-location"
	inferenceServiceContainerName = "kai-container"
)

type Client struct {
	kclient    kclient.Client
	credClient *credentials.Client
}

func New(client kclient.Client) *Client {
	return &Client{
		kclient:    client,
		credClient: credentials.NewDefaultCredentialBuilder(client),
	}
}

func (c *Client) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	s := &corev1alpha1.Step{}
	err := c.kclient.Get(ctx, req.NamespacedName, s)
	if err != nil {
		if apierr.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to retrieve latest step %s: %w", req.NamespacedName, err)
	}

	// should we merge modelSpec and PodSpec here?
	// NOTE: since we're potentially merging values here higher level resources may not be aware of these changes
	// until after reconcile
	if s.Spec.Model != nil {
		ic, err := c.makeInitContainer(ctx, s.NamespacedName(), s.Spec.Model, &s.Spec.PodSpec)
		if err != nil {
			return ctrl.Result{}, err
		}
		s.Spec.InitContainers = []corev1.Container{
			ic,
		}

		rt, err := c.getModelRuntime(ctx, s)
		if err != nil {
			return ctrl.Result{}, err
		}

		mergeRuntimeSpec(&s.Spec, rt)
	}

	for _, rec := range []func(context.Context, *corev1alpha1.Step) error{
		deployment.NewReconciler(c.kclient).Reconcile,
		service.NewReconciler(c.kclient).Reconcile,
		hpa.NewReconciler(c.kclient).Reconcile,
	} {
		if err := rec(ctx, s); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (c *Client) makeInitContainer(ctx context.Context, name types.NamespacedName, m *corev1alpha1.ModelSpec, p *corev1alpha1.PodSpec) (corev1.Container, error) {
	initContainer := corev1.Container{
		Args: []string{
			m.URI,
			"/mnt/models",
		},
		Name:  "storage-initializer",
		Image: storageInitializerImage,
		VolumeMounts: []corev1.VolumeMount{
			{
				MountPath: "/mnt/models",
				Name:      modelVolumeName,
			},
		},
	}

	// add service account creds if present
	if m.ServiceAccountRef != "" {
		err := c.credClient.BuildCredentials(
			ctx,
			types.NamespacedName{Name: m.ServiceAccountRef, Namespace: name.Namespace},
			&initContainer,
			&p.Volumes,
		)
		if err != nil {
			return initContainer, err
		}
	}

	return initContainer, nil
}

func (c *Client) getModelRuntime(ctx context.Context, s *corev1alpha1.Step) (corev1alpha1.ModelRuntime, error) {
	// only support cluster wide modelRuntimes atm can easily support namespaced ones
	runtimes := &corev1alpha1.ModelRuntimeList{}
	err := c.kclient.List(ctx, runtimes)
	if err != nil {
		return corev1alpha1.ModelRuntime{}, fmt.Errorf("failed to list modelruntimes %w", err)
	}

	// if modelRuntime is specified in spec use that ilo modelformat
	if s.Spec.Model.ModelRuntime != "" {
		for _, rt := range runtimes.Items {
			if rt.Name == s.Spec.Model.ModelRuntime {
				return rt, nil
			}
		}
	}

	// modelRuntime not specified so match based on modelFormat
	// first match wins, may need to sort this for consistency
	for _, rt := range runtimes.Items {
		for _, format := range rt.Spec.SupportedModelFormats {
			if format == s.Spec.Model.ModelFormat {
				return rt, nil
			}
		}
	}

	return corev1alpha1.ModelRuntime{}, fmt.Errorf("no supporting modelruntime found")
}

func mergeRuntimeSpec(stepSpec *corev1alpha1.StepSpec, rt corev1alpha1.ModelRuntime) {
	// we don't allow container level overrides at the inference service level
	// so replace containers in podSpec with our modelRuntime containers
	stepSpec.Containers = rt.Spec.Containers
	for i, con := range stepSpec.Containers {
		if con.Name == inferenceServiceContainerName {
			stepSpec.Containers[i].VolumeMounts = append(con.VolumeMounts, corev1.VolumeMount{
				MountPath: "/mnt/models",
				Name:      modelVolumeName,
			})
		}
	}

	// add model mount volume
	mountVolume := corev1.Volume{
		Name: modelVolumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}
	stepSpec.Volumes = append(stepSpec.Volumes, mountVolume)

	// set additional overrides if necessary
	// both values could be empty but always default to modelRuntime value
	if stepSpec.MinReplicas == nil || (*stepSpec.MinReplicas) < 1 {
		stepSpec.MinReplicas = rt.Spec.MinReplicas
	}

	if stepSpec.MaxReplicas == 0 {
		stepSpec.MaxReplicas = rt.Spec.MaxReplicas
	}

	if len(stepSpec.Metrics) == 0 {
		stepSpec.Metrics = rt.Spec.Metrics
	}

	if stepSpec.Behavior == nil {
		stepSpec.Behavior = rt.Spec.Behavior
	}
}
