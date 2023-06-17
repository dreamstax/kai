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

package router

import (
	"context"
	"fmt"

	corev1alpha1 "github.com/dreamstax/kai/api/core/v1alpha1"
	"github.com/dreamstax/kai/internal/app/reconcilers/names"
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

func (r *Reconciler) Reconcile(ctx context.Context, app *corev1alpha1.App) error {
	name := names.RouterName(app)
	router := &corev1alpha1.Router{}
	err := r.client.Get(ctx, name, router)
	if apierrs.IsNotFound(err) {
		// doesn't exist so create it
		_, err = r.createRouter(ctx, name, app)
		if apierrs.IsAlreadyExists(err) {
			return fmt.Errorf("router already exists for app %q: %w", name, err)
		} else if err != nil {
			return fmt.Errorf("failed to create router %q: %w", name, err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to get router %q: %w", name, err)
	} else {
		// update router
		_, err = r.updateRouter(ctx, name, app, router)
		if err != nil {
			return fmt.Errorf("failed to update router %q: %w", name, err)
		}
	}

	// handle other things like rollout

	return nil
}

func (r *Reconciler) createRouter(ctx context.Context, name types.NamespacedName, app *corev1alpha1.App) (*corev1alpha1.Router, error) {
	router, err := makeRouter(name, app)
	if err != nil {
		return nil, fmt.Errorf("failed to make router %q: %w", name, err)
	}

	err = r.client.Create(ctx, router)
	if err != nil {
		return nil, fmt.Errorf("failed to create router %q: %w", name, err)
	}

	return router, nil
}

func (r *Reconciler) updateRouter(ctx context.Context, name types.NamespacedName, app *corev1alpha1.App, have *corev1alpha1.Router) (*corev1alpha1.Router, error) {
	router, err := makeRouter(name, app)
	if err != nil {
		return nil, fmt.Errorf("failed to make rotuer %q: %w", name, err)
	}

	if equality.Semantic.DeepEqual(have.Spec, router.Spec) {
		// no changes to make just return
		return have, nil
	}

	// update router
	want := have.DeepCopy()
	want.Spec = router.Spec
	want.Labels = kmap.Union(router.Labels, want.Labels)

	err = r.client.Update(ctx, want)
	if err != nil {
		return nil, fmt.Errorf("failed to update router %q: %w", name, err)
	}

	return want, nil
}

func makeRouter(name types.NamespacedName, app *corev1alpha1.App) (*corev1alpha1.Router, error) {
	appSpec := app.Spec.DeepCopy()
	// Even with no Route explicitly included in the spec file, this will NOT be nil, and the Spec.Route will NOT be nil
	// However, the Spec.Route.Hostnames and the Spec.Route.Rules WILL be nil
	routerSpec := appSpec.RouterSpec

	labels := names.MakeLabels(app)
	annotations := names.MakeAnnotations(app)

	return &corev1alpha1.Router{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name.Name,
			Namespace:       name.Namespace,
			Labels:          labels,
			Annotations:     annotations,
			OwnerReferences: []metav1.OwnerReference{*kmeta.NewControllerRef(app)},
		},
		Spec: routerSpec,
	}, nil
}
