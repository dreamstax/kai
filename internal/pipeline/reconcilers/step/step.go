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

package step

import (
	"context"
	"fmt"

	corev1alpha1 "github.com/dreamstax/kai/api/core/v1alpha1"
	"github.com/dreamstax/kai/internal/pipeline/reconcilers/names"
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

// TODO: accept step number as arg to identify appropriate step
func (r *Reconciler) Reconcile(ctx context.Context, p *corev1alpha1.Pipeline) error {
	for i, s := range p.Spec.Steps {
		stepName := names.StepName(p, i)
		step := &corev1alpha1.Step{}
		err := r.client.Get(ctx, stepName, step)
		if apierrs.IsNotFound(err) {
			// step doesn't exist so create it.
			objMeta := metav1.ObjectMeta{
				Name:            stepName.Name,
				Namespace:       step.Namespace,
				Labels:          names.MakeLabels(p),
				Annotations:     names.MakeAnnotations(p),
				OwnerReferences: []metav1.OwnerReference{*kmeta.NewControllerRef(p)},
			}
			_, err = r.createStep(ctx, stepName, objMeta, s)
			if err != nil {
				return fmt.Errorf("failed to create step %q: %w", stepName, err)
			}
		} else if err != nil {
			return fmt.Errorf("failed to get step %q: %w", stepName, err)
		} else {
			// step exists
			objMeta := metav1.ObjectMeta{
				Name:            stepName.Name,
				Namespace:       step.Namespace,
				Labels:          names.MakeLabels(p),
				Annotations:     names.MakeAnnotations(p),
				OwnerReferences: []metav1.OwnerReference{*kmeta.NewControllerRef(p)},
			}
			_, err = r.updateStep(ctx, stepName, objMeta, s, step)
			if err != nil {
				return fmt.Errorf("failed to update deplyoment %q: %w", stepName, err)
			}

			// TODO: Surface step status to pipeline
		}
	}

	// TODO: handle failing pods

	return nil
}

func (r *Reconciler) createStep(ctx context.Context, name types.NamespacedName, objMeta metav1.ObjectMeta, stepTpl *corev1alpha1.StepTemplateSpec) (*corev1alpha1.Step, error) {
	step, err := r.makeStep(ctx, name, objMeta, stepTpl)
	if err != nil {
		return nil, fmt.Errorf("failed to make step %q: %w", name, err)
	}

	err = r.client.Create(ctx, step)
	if err != nil {
		return nil, fmt.Errorf("failed to create step %q: %w", name, err)
	}

	return step, nil
}

func (r *Reconciler) updateStep(ctx context.Context, name types.NamespacedName, objMeta metav1.ObjectMeta, stepTpl *corev1alpha1.StepTemplateSpec, in *corev1alpha1.Step) (*corev1alpha1.Step, error) {
	step, err := r.makeStep(ctx, name, objMeta, stepTpl)
	if err != nil {
		return nil, fmt.Errorf("failed to update step %q: %w", name, err)
	}

	if equality.Semantic.DeepEqual(in.Spec, step.Spec) {
		// no changes to make just return
		return in, nil
	}

	// update step
	out := in.DeepCopy()
	out.Spec = step.Spec
	out.Labels = kmap.Union(step.Labels, out.Labels)

	err = r.client.Update(ctx, out)
	if err != nil {
		return nil, fmt.Errorf("failed to update step %q: %w", name, err)
	}

	return out, nil
}

func (r *Reconciler) makeStep(ctx context.Context, name types.NamespacedName, objMeta metav1.ObjectMeta, step *corev1alpha1.StepTemplateSpec) (*corev1alpha1.Step, error) {
	stepc := step.DeepCopy()
	return &corev1alpha1.Step{
		ObjectMeta: objMeta,
		Spec:       stepc.Spec,
	}, nil
}
