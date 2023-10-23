/*
Copyright 2023 The Kai Authors.

Licensed under the Apache License, step 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package hpa

import (
	"context"
	"fmt"

	corev1alpha1 "github.com/dreamstax/kai/api/core/v1alpha1"
	"github.com/dreamstax/kai/internal/step/reconcilers/names"
	autoscaling "k8s.io/api/autoscaling/v2"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/kmap"
	"knative.dev/pkg/kmeta"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciler struct {
	client kclient.Client
}

func NewReconciler(client kclient.Client) *Reconciler {
	return &Reconciler{
		client: client,
	}
}

func (r *Reconciler) Reconcile(ctx context.Context, s *corev1alpha1.Step) error {
	hpaName := names.HPAName(s)
	hpa := &autoscaling.HorizontalPodAutoscaler{}
	err := r.client.Get(ctx, hpaName, hpa)
	if apierrs.IsNotFound(err) {
		// hpa doesn't exist so create it.
		// TODO: set step status (we should expose enough info to users for proper visibility)
		_, err = r.createHPA(ctx, hpaName, s)
		if err != nil {
			return fmt.Errorf("failed to create hpa %q: %w", hpaName, err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to get hpa %q: %w", hpa, err)
	} else {
		// deployment exists
		_, err = r.updateHPA(ctx, hpaName, s, hpa)
		if err != nil {
			return fmt.Errorf("failed to update hpa %q: %w", hpaName, err)
		}

		// TODO: Surface hpa status to step
	}

	return nil
}

func (r *Reconciler) createHPA(ctx context.Context, name types.NamespacedName, s *corev1alpha1.Step) (*autoscaling.HorizontalPodAutoscaler, error) {
	hpa, err := makeHPA(name, s)
	if err != nil {
		return nil, fmt.Errorf("failed to make hpa %q: %w", name, err)
	}

	err = r.client.Create(ctx, hpa)
	if err != nil {
		return nil, fmt.Errorf("failed to create hpa %q: %w", name, err)
	}

	return hpa, nil
}

func (r *Reconciler) updateHPA(ctx context.Context, name types.NamespacedName, s *corev1alpha1.Step, in *autoscaling.HorizontalPodAutoscaler) (*autoscaling.HorizontalPodAutoscaler, error) {
	hpa, err := makeHPA(name, s)
	if err != nil {
		return nil, fmt.Errorf("failed to make hpa %q: %w", name, err)
	}

	// ignore labels
	hpa.Labels = in.Labels

	if equality.Semantic.DeepEqual(in.Spec, hpa.Spec) {
		// no changes to make just return
		return in, nil
	}

	// update hpa
	out := in.DeepCopy()
	out.Spec = hpa.Spec
	out.Labels = kmap.Union(hpa.Labels, out.Labels)

	err = r.client.Update(ctx, out)
	if err != nil {
		return nil, fmt.Errorf("failed to update hpa %q: %w", name, err)
	}

	return out, nil
}

func makeHPA(name types.NamespacedName, s *corev1alpha1.Step) (*autoscaling.HorizontalPodAutoscaler, error) {
	labels := names.MakeLabels(s)
	annotations := names.MakeAnnotations(s)

	hpa := &autoscaling.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name.Name,
			Namespace:       s.Namespace,
			Labels:          labels,
			Annotations:     annotations,
			OwnerReferences: []metav1.OwnerReference{*kmeta.NewControllerRef(s)},
		},
	}

	hpaSpec := &autoscaling.HorizontalPodAutoscalerSpec{
		ScaleTargetRef: autoscaling.CrossVersionObjectReference{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       names.DeploymentName(s).Name,
		},
	}

	sSpec := s.Spec.DeepCopy()

	// TODO: support scale to zero with hpa alpha feature gate
	var minReplicas int32
	if sSpec.MinReplicas == nil || (*sSpec.MinReplicas) < 1 {
		minReplicas = 1
	} else {
		minReplicas = (*sSpec.MinReplicas)
	}
	hpaSpec.MinReplicas = &minReplicas

	if sSpec.MaxReplicas < minReplicas {
		hpaSpec.MaxReplicas = minReplicas
	} else {
		hpaSpec.MaxReplicas = sSpec.MaxReplicas
	}

	if sSpec.Metrics == nil {
		hpaSpec.Metrics = []autoscaling.MetricSpec{}
	} else {
		hpaSpec.Metrics = sSpec.Metrics
	}

	hpaSpec.Behavior = &autoscaling.HorizontalPodAutoscalerBehavior{}

	hpa.Spec = *hpaSpec

	return hpa, nil
}
