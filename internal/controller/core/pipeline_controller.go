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

package core

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1alpha1 "github.com/dreamstax/kai/api/core/v1alpha1"
	"github.com/dreamstax/kai/internal/pipeline"
)

// PipelineReconciler reconciles a Pipeline object
type PipelineReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	pipec  *pipeline.Client
}

//+kubebuilder:rbac:groups=core.kai.io,resources=steps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.kai.io,resources=steps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.kai.io,resources=steps/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
//+kubebuilder:rbac:groups=core.kai.io,resources=pipelines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.kai.io,resources=pipelines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.kai.io,resources=pipelines/finalizers,verbs=update

func (r *PipelineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.pipec.Reconcile(ctx, req)
}

// SetupWithManager sets up the controller with the Manager.
func (r *PipelineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.pipec = pipeline.New(r.Client)
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.Pipeline{}).
		Complete(r)
}
