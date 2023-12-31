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

package deployment

import (
	"context"
	"encoding/json"
	"fmt"

	corev1alpha1 "github.com/dreamstax/kai/api/core/v1alpha1"
	"github.com/dreamstax/kai/internal/step/reconcilers/names"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
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
	deploymentName := names.DeploymentName(s)
	deployment := &appsv1.Deployment{}
	err := r.client.Get(ctx, deploymentName, deployment)
	if apierrs.IsNotFound(err) {
		// deplyoment doesn't exist so create it.
		_, err = r.createDeployment(ctx, deploymentName, s)
		if err != nil {
			return fmt.Errorf("failed to create deployment %q: %w", deploymentName, err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to get deployment %q: %w", deploymentName, err)
	} else {
		// deployment exists
		_, err = r.updateDeployment(ctx, deploymentName, s, deployment)
		if err != nil {
			return fmt.Errorf("failed to update deplyoment %q: %w", deploymentName, err)
		}

		// TODO: Surface deployment status to step
	}

	// TODO: handle failing pods

	return nil
}

func (r *Reconciler) createDeployment(ctx context.Context, name types.NamespacedName, step *corev1alpha1.Step) (*appsv1.Deployment, error) {
	deployment, err := r.makeDeployment(ctx, name, step)
	if err != nil {
		return nil, fmt.Errorf("failed to make deployment %q: %w", name, err)
	}

	err = r.client.Create(ctx, deployment)
	if err != nil {
		return nil, fmt.Errorf("failed to create deployment %q: %w", name, err)
	}

	return deployment, nil
}

func (r *Reconciler) updateDeployment(ctx context.Context, name types.NamespacedName, step *corev1alpha1.Step, in *appsv1.Deployment) (*appsv1.Deployment, error) {
	deployment, err := r.makeDeployment(ctx, name, step)
	if err != nil {
		return nil, fmt.Errorf("failed to update deployment %q: %w", name, err)
	}

	// ignore replicas and labels
	deployment.Spec.Replicas = in.Spec.Replicas
	deployment.Spec.Selector = in.Spec.Selector

	if equality.Semantic.DeepEqual(in.Spec, deployment.Spec) {
		// no changes to make just return
		return in, nil
	}

	// update deployment
	out := in.DeepCopy()
	out.Spec = deployment.Spec
	out.Labels = kmap.Union(deployment.Labels, out.Labels)

	err = r.client.Update(ctx, out)
	if err != nil {
		return nil, fmt.Errorf("failed to update deployment %q: %w", name, err)
	}

	return out, nil
}

func (r *Reconciler) makeDeployment(ctx context.Context, name types.NamespacedName, step *corev1alpha1.Step) (*appsv1.Deployment, error) {
	stepSpec := step.Spec.DeepCopy()

	var initialReplicaCount int32
	if stepSpec.MinReplicas == nil || (*stepSpec.MinReplicas) < 1 {
		initialReplicaCount = 1
	} else {
		initialReplicaCount = *stepSpec.MinReplicas
	}
	// TODO: get deadline from step
	var progressDeadline int32 = 60
	maxUnavailable := intstr.FromInt(0)

	labels := names.MakeLabels(step)
	annotations := names.MakeAnnotations(step)

	podSpecJson, err := json.Marshal(stepSpec.PodSpec)
	if err != nil {
		return nil, err
	}
	corePodSpec := corev1.PodSpec{}
	err = json.Unmarshal(podSpecJson, &corePodSpec)
	if err != nil {
		return nil, err
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name.Name,
			Namespace:       step.Namespace,
			Labels:          labels,
			Annotations:     annotations,
			OwnerReferences: []metav1.OwnerReference{*kmeta.NewControllerRef(step)},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas:                &initialReplicaCount,
			Selector:                names.MakeSelector(step),
			ProgressDeadlineSeconds: &progressDeadline,
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxUnavailable: &maxUnavailable,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: annotations,
				},
				Spec: corePodSpec,
			},
		},
	}, nil
}
