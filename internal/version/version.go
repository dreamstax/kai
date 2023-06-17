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

package version

import (
	"context"
	"fmt"

	corev1alpha1 "github.com/dreamstax/kai/api/core/v1alpha1"
	"github.com/dreamstax/kai/internal/version/reconcilers/deployment"
	"github.com/dreamstax/kai/internal/version/reconcilers/hpa"
	"github.com/dreamstax/kai/internal/version/reconcilers/service"
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
	// fetch latest version instance
	v := &corev1alpha1.Version{}
	err := c.client.Get(ctx, req.NamespacedName, v)
	if err != nil {
		if apierr.IsNotFound(err) {
			// object garbage collected for w/e reason just return
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to retrieve latest version %s: %w", req.NamespacedName, err)
	}

	// reconcile resources
	for _, rec := range []func(context.Context, *corev1alpha1.Version) error{
		deployment.NewReconciler(c.client).Reconcile,
		service.NewReconciler(c.client).Reconcile,
		hpa.NewReconciler(c.client).Reconcile,
	} {
		if err := rec(ctx, v); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}
