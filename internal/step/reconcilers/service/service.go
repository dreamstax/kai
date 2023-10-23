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

package service

import (
	"context"
	"fmt"

	corev1alpha1 "github.com/dreamstax/kai/api/core/v1alpha1"
	"github.com/dreamstax/kai/internal/step/reconcilers/names"
	v1 "k8s.io/api/core/v1"
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
	serviceName := names.ServiceName(s)
	service := &v1.Service{}
	err := r.client.Get(ctx, serviceName, service)
	if apierrs.IsNotFound(err) {
		// service doesn't exist so create it
		// TODO: set step status
		_, err = r.createService(ctx, serviceName, s)
		if err != nil {
			return fmt.Errorf("faield to create service %q: %w", serviceName, err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to get service %q: %w", serviceName, err)
	} else {
		// service exists
		_, err = r.updateService(ctx, serviceName, s, service)
		if err != nil {
			return fmt.Errorf("failed to update service %q: %w", serviceName, err)
		}

		// TODO: surface service status to step
	}

	return nil
}

func (r *Reconciler) createService(ctx context.Context, name types.NamespacedName, s *corev1alpha1.Step) (*v1.Service, error) {
	service, err := makeService(name, s)
	if err != nil {
		return nil, fmt.Errorf("failed to make service %q: %w", name, err)
	}

	err = r.client.Create(ctx, service)
	if err != nil {
		return nil, fmt.Errorf("failed to create service %q: %w", name, err)
	}

	return service, nil
}

func (r *Reconciler) updateService(ctx context.Context, name types.NamespacedName, s *corev1alpha1.Step, in *v1.Service) (*v1.Service, error) {
	service, err := makeService(name, s)
	if err != nil {
		return nil, fmt.Errorf("failed to update service %q: %w", name, err)
	}

	if equality.Semantic.DeepEqual(in.Spec, service.Spec) {
		return in, nil
	}

	out := in.DeepCopy()
	out.Spec = service.Spec
	out.Labels = kmap.Union(service.Labels, out.Labels)

	err = r.client.Update(ctx, out)
	if err != nil {
		return nil, fmt.Errorf("failed to update service %q: %w", name, err)
	}

	return out, nil
}

func makeService(name types.NamespacedName, s *corev1alpha1.Step) (*v1.Service, error) {
	labels := names.MakeLabels(s)
	annotations := names.MakeAnnotations(s)

	ports := makePorts(s)
	// if user didn't specify a port default to 80 so we don't fail to create a service
	if len(ports) == 0 {
		ports = append(ports, v1.ServicePort{
			Port: int32(80),
		})
	}

	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name.Name,
			Namespace:       s.Namespace,
			Labels:          labels,
			Annotations:     annotations,
			OwnerReferences: []metav1.OwnerReference{*kmeta.NewControllerRef(s)},
		},
		Spec: v1.ServiceSpec{
			Selector: names.MakeServiceSelector(s),
			Ports:    ports,
			Type:     v1.ServiceTypeClusterIP,
		},
	}, nil
}

func makePorts(s *corev1alpha1.Step) []v1.ServicePort {
	out := []v1.ServicePort{}
	podSpec := s.Spec.PodSpec.DeepCopy()

	for _, container := range podSpec.Containers {
		for _, port := range container.Ports {
			out = append(out, v1.ServicePort{
				Port:       port.ContainerPort,
				TargetPort: intstr.FromInt(int(port.ContainerPort)),
			})
		}
	}

	return out
}
