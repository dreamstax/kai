/*
Copyright 2023 The Kai Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package app

import (
	"context"
	"fmt"

	aiv1alpha1 "github.com/dreamstax/kai/api/ai/v1alpha1"

	corev1alpha1 "github.com/dreamstax/kai/api/core/v1alpha1"
	"github.com/dreamstax/kai/internal/credentials"
	"github.com/dreamstax/kai/internal/inferenceservice/reconcilers/names"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/kmap"
	"knative.dev/pkg/kmeta"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	storageInitializerImage       = "kserve/storage-initializer:v0.10.1"
	modelVolumeName               = "kai-mount-location"
	inferenceServiceContainerName = "kai-container"
)

type Reconciler struct {
	client     kclient.Client
	credClient *credentials.Client
}

func NewReconciler(client kclient.Client) *Reconciler {
	return &Reconciler{
		client:     client,
		credClient: credentials.NewDefaultCredentialBuilder(client),
	}
}

func (r *Reconciler) Reconcile(ctx context.Context, is *aiv1alpha1.InferenceService) error {
	name := names.AppName(is)
	app := &corev1alpha1.App{}
	err := r.client.Get(ctx, name, app)
	if apierrs.IsNotFound(err) {
		// doesn't exist so create it
		_, err = r.createApp(ctx, name, is)
		if apierrs.IsAlreadyExists(err) {
			return fmt.Errorf("app already exists for inferenceservice %q: %w", name, err)
		} else if err != nil {
			return fmt.Errorf("failed to create app %q: %w", name, err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to get app %q: %w", name, err)
	} else {
		// update app
		_, err = r.updateApp(ctx, name, is, app)
		if err != nil {
			return fmt.Errorf("failed to update app %q: %w", name, err)
		}
	}

	// handle other things

	return nil
}

func (r *Reconciler) createApp(ctx context.Context, name types.NamespacedName, is *aiv1alpha1.InferenceService) (*corev1alpha1.App, error) {
	app, err := r.makeApp(ctx, name, is)
	if err != nil {
		return nil, fmt.Errorf("failed to make app %q: %w", name, err)
	}

	err = r.client.Create(ctx, app)
	if err != nil {
		return nil, fmt.Errorf("failed to create app %q: %w", name, err)
	}

	return app, nil
}

func (r *Reconciler) updateApp(ctx context.Context, name types.NamespacedName, is *aiv1alpha1.InferenceService, have *corev1alpha1.App) (*corev1alpha1.App, error) {
	app, err := r.makeApp(ctx, name, is)
	if err != nil {
		return nil, fmt.Errorf("failed to make app %q: %w", name, err)
	}

	if equality.Semantic.DeepEqual(have.Spec, app.Spec) {
		// no changes to make just return
		return have, nil
	}

	// update app
	want := have.DeepCopy()
	want.Spec = app.Spec
	want.Labels = kmap.Union(app.Labels, want.Labels)

	err = r.client.Update(ctx, want)
	if err != nil {
		return nil, fmt.Errorf("failed to update app %q: %w", name, err)
	}

	return want, nil
}

func (r *Reconciler) makeApp(ctx context.Context, name types.NamespacedName, is *aiv1alpha1.InferenceService) (*corev1alpha1.App, error) {
	isSpec := is.Spec.DeepCopy()
	versionSpec := isSpec.Template.Spec.ConfigSpec.Template.Spec
	routerSpec := isSpec.Template.Spec.RouterSpec

	// make init container
	initContainer, err := r.makeInitContainer(ctx, name, is)
	if err != nil {
		return nil, err
	}
	versionSpec.InitContainers = []corev1.Container{
		initContainer,
	}

	rt, err := r.getModelRuntime(ctx, isSpec)
	if err != nil {
		return nil, err
	}

	mergeRuntimeSpec(&versionSpec, rt)

	labels := names.MakeLabels(is)
	annotations := names.MakeAnnotations(is)

	return &corev1alpha1.App{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name.Name,
			Namespace:       name.Namespace,
			Labels:          labels,
			Annotations:     annotations,
			OwnerReferences: []metav1.OwnerReference{*kmeta.NewControllerRef(is)},
		},
		Spec: corev1alpha1.AppSpec{
			ConfigSpec: corev1alpha1.ConfigSpec{
				Template: corev1alpha1.VersionTemplateSpec{
					Spec: versionSpec,
				},
			},
			RouterSpec: routerSpec,
		},
	}, nil
}

func (r *Reconciler) getModelRuntime(ctx context.Context, is *aiv1alpha1.InferenceServiceSpec) (aiv1alpha1.ModelRuntime, error) {
	// only support cluster wide modelRuntimes atm can easily support namespaced ones
	runtimes := &aiv1alpha1.ModelRuntimeList{}
	err := r.client.List(ctx, runtimes)
	if err != nil {
		return aiv1alpha1.ModelRuntime{}, fmt.Errorf("failed to list modelruntimes %w", err)
	}

	// if modelRuntime is specified in spec use that ilo modelformat
	if is.Model.ModelRuntime != "" {
		for _, rt := range runtimes.Items {
			if rt.Name == is.Model.ModelRuntime {
				return rt, nil
			}
		}
	}

	// modelRuntime not specified so match based on modelFormat
	// first match wins, may need to sort this for consistency
	for _, rt := range runtimes.Items {
		for _, format := range rt.Spec.SupportedModelFormats {
			if format == is.Model.ModelFormat {
				return rt, nil
			}
		}
	}

	return aiv1alpha1.ModelRuntime{}, fmt.Errorf("no supporting modelruntime found")
}

func mergeRuntimeSpec(versionSpec *corev1alpha1.VersionSpec, rt aiv1alpha1.ModelRuntime) {
	// we don't allow container level overrides at the inference service level
	// so replace containers in podSpec with our modelRuntime containers
	versionSpec.Containers = rt.Spec.Containers
	for i, con := range versionSpec.Containers {
		if con.Name == inferenceServiceContainerName {
			versionSpec.Containers[i].VolumeMounts = append(con.VolumeMounts, corev1.VolumeMount{
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
	versionSpec.Volumes = append(versionSpec.Volumes, mountVolume)

	// set additional overrides if necessary
	// both values could be empty but always default to modelRuntime value
	if versionSpec.MinReplicas == nil || (*versionSpec.MinReplicas) < 1 {
		versionSpec.MinReplicas = rt.Spec.MinReplicas
	}

	if versionSpec.MaxReplicas == 0 {
		versionSpec.MaxReplicas = rt.Spec.MaxReplicas
	}

	if len(versionSpec.Metrics) == 0 {
		versionSpec.Metrics = rt.Spec.Metrics
	}

	if versionSpec.Behavior == nil {
		versionSpec.Behavior = rt.Spec.Behavior
	}
}

func (r *Reconciler) makeInitContainer(ctx context.Context, name types.NamespacedName, is *aiv1alpha1.InferenceService) (corev1.Container, error) {
	initContainer := corev1.Container{
		Args: []string{
			is.Spec.Model.URI,
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
	if is.Spec.Model.ServiceAccountRef != "" {
		err := r.credClient.BuildCredentials(ctx, types.NamespacedName{Name: is.Spec.Model.ServiceAccountRef, Namespace: name.Namespace}, &initContainer, &is.Spec.Template.Spec.Template.Spec.Volumes)
		if err != nil {
			return initContainer, err
		}
	}

	return initContainer, nil
}
