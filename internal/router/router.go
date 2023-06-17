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
	"github.com/dreamstax/kai/internal/router/reconcilers/httproute"
	apierr "k8s.io/apimachinery/pkg/api/errors"

	ctrl "sigs.k8s.io/controller-runtime"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type Client struct {
	client kclient.Client
}

func New(client kclient.Client) *Client {
	return &Client{
		client: client,
	}
}

func (c *Client) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// fetch latest router
	router := &corev1alpha1.Router{}
	err := c.client.Get(ctx, req.NamespacedName, router)
	if err != nil {
		if apierr.IsNotFound(err) {
			// object garbage collected for w/e reason just return
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to retrieve latest router %s: %w", req.NamespacedName, err)
	}

	// resolve backendRefs to the config-version level // TODO?: What does this mean?

	// reconcile resources
	for _, rec := range []func(context.Context, *corev1alpha1.Router) error{
		httproute.NewReconciler(c.client).Reconcile,
	} {
		if err := rec(ctx, router); err != nil {
			// FIXME: We may not care that necessary resources aren't installed. Ignore and allow infinite retry?
			//if chkerr, ok := err.(*meta.NoKindMatchError); ok == true {
			//	if chkerr.GroupKind.Group == "gateway.networking.k8s.io" {
			//		fmt.Println("Missing Gateway resources")
			//		return ctrl.Result{}, nil
			//	}
			//}

			// FIXME: If error is not nil, this will continue to retry seemingly INDEFINITELY. Dumb.
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}
