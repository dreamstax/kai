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

package httproute

import (
	"context"
	"fmt"

	"github.com/dreamstax/kai/internal/version/reconcilers/names"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/labels"

	corev1alpha1 "github.com/dreamstax/kai/api/core/v1alpha1"
	"github.com/dreamstax/kai/api/kai"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/kmap"
	"knative.dev/pkg/kmeta"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
	v1beta1gateway "sigs.k8s.io/gateway-api/apis/v1beta1"
)

type Reconciler struct {
	client kclient.Client
}

func NewReconciler(client kclient.Client) *Reconciler {
	return &Reconciler{
		client: client,
	}
}

func (r *Reconciler) Reconcile(ctx context.Context, router *corev1alpha1.Router) error {
	// fetch what should be this router's httproute
	name := httpRouteName(router)
	httpr := &v1beta1gateway.HTTPRoute{}
	err := r.client.Get(ctx, name, httpr)

	if apierrs.IsNotFound(err) {
		// doesn't exist so create it
		_, err = r.createHTTPRoute(ctx, router)
		if apierrs.IsAlreadyExists(err) {
			// TODO: set failed status on router
			return fmt.Errorf("httproute already exists for router %q: %w", name, err)
		} else if err != nil {
			// TODO: set failed status on router
			return fmt.Errorf("failed to create httproute %q: %w", name, err)
		}
	} else if apierrs.IsAlreadyExists(err) {
		// TODO: set failed status on router
		return fmt.Errorf("httproute already exists for router %q: %w", name, err)
	} else if meta.IsNoMatchError(err) {
		// Check if Gateway resources are installed
		if chkerr, ok := err.(*meta.NoKindMatchError); ok == true {
			missingType := ""
			if chkerr.GroupKind.Group == "gateway.networking.k8s.io" {
				missingType = "Gateway "
			}
			// TODO: set failed status on router
			return fmt.Errorf("failed to create httproute %q due to missing "+missingType+"resources: %w", name, err)
		}
	} else if err != nil {
		// TODO: set failed status on router
		return fmt.Errorf("failed to get httproute for router %q: %w", name, err)
	} else {
		// Found an HTTPRoute. Update it.
		_, err = r.updateHTTPRoute(ctx, name, router, httpr)

		if err != nil {
			// TODO: set failed status on router
			return fmt.Errorf("failed to update httproute %q: %w", name, err)
		}
	}

	return nil
}

func (r *Reconciler) createHTTPRoute(ctx context.Context, router *corev1alpha1.Router) (*v1beta1gateway.HTTPRoute, error) {
	// FIXME?: Combine this with updateHTTPRoute?
	var httpr *v1beta1gateway.HTTPRoute

	if ver, err := r.getCurrentVersion(ctx, router); ver != nil && err == nil {
		httpr = makeHTTPRoute(router, ver)
	} else {
		return nil, err
	}

	err := r.client.Create(ctx, httpr)
	if err != nil {
		// TODO: set failed status on router
		return nil, err
	}

	return httpr, nil
}

func (r *Reconciler) updateHTTPRoute(ctx context.Context, name types.NamespacedName, router *corev1alpha1.Router, httpr *v1beta1gateway.HTTPRoute) (*v1beta1gateway.HTTPRoute, error) {
	// FIXME?: Combine this with createHTTPRoute?
	var want *v1beta1gateway.HTTPRoute

	if ver, err := r.getCurrentVersion(ctx, router); ver != nil && err == nil {
		httpr = makeHTTPRoute(router, ver)
	} else {
		// TODO: set failed status on router
		return nil, err
	}

	if equality.Semantic.DeepEqual(httpr, want) {
		return httpr, nil
	}

	out := httpr.DeepCopy()
	out.Spec = want.Spec
	out.Labels = kmap.Union(want.Labels, out.Labels)

	err := r.client.Update(ctx, out)
	if err != nil {
		// TODO: set failed status on router
		return nil, fmt.Errorf("failed to update httproute %q: %w", name, err)
	}

	return out, nil
}

// TODO?: Place this into a general function so that we can obtain a specific version from other resources? "Base class" or whatever?
// Returns the Version currently associated with this Router
func (r *Reconciler) getCurrentVersion(ctx context.Context, router *corev1alpha1.Router) (*corev1alpha1.Version, error) {
	// Get the AppID of the router
	routeAppID := router.GetLabels()[kai.KaiAppUIDLabelKey] // core.kai.io/appUID

	// Get all the Versions currently deployed on the cluster in this Router's namespace
	versions := &corev1alpha1.VersionList{}
	err := r.client.List(ctx, versions, &kclient.ListOptions{
		LabelSelector: labels.Everything(), // FIXME: There is probably a quicker way to find just the AppIDs. Might not need this at all.
		Namespace:     router.GetNamespace(),
	})

	if err != nil {
		// TODO: set failed status on router
		return nil, fmt.Errorf("failed to retrieve version list %q: %w", router.GetNamespace(), err)
	}

	// FIXME?: O(n) the best we can do here?
	// Match the Router AppID to the Version AppID to find the correct Version
	for _, ver := range versions.Items {
		if routeAppID == ver.GetLabels()[kai.KaiAppUIDLabelKey] {
			return &ver, nil
		}
	}

	// TODO: set failed status on router
	return nil, fmt.Errorf("router could not be matched to version %q: %w", router.GetNamespace(), err)
}

func makeHTTPRoute(router *corev1alpha1.Router, version *corev1alpha1.Version) *v1beta1gateway.HTTPRoute {
	httpr := &v1beta1gateway.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      httpRouteName(router).Name,
			Namespace: router.Namespace,
		},
		Spec: router.Spec.DeepCopy().Route.ToK8sGatewaySpec(),
	}

	setHTTPRouteLabels(httpr, router)
	setHTTPRouteAnnotations(httpr, router)
	httpr.OwnerReferences = append(httpr.OwnerReferences, *kmeta.NewControllerRef(router))

	if httpr.Spec.ParentRefs == nil || len(httpr.Spec.ParentRefs) == 0 {
		group := v1beta1gateway.Group("gateway.networking.k8s.io")
		kind := v1beta1gateway.Kind("Gateway")
		// FIXME: Should the ParentRef name/other data be pulled from the GatewayClass?
		httpr.Spec.ParentRefs = []v1beta1gateway.ParentReference{{
			Name:  "kai-gateway",
			Group: &group,
			Kind:  &kind,
		}}
	}

	serviceName := names.ServiceName(version).Name
	port := getContainerPort(version)

	if httpr.Spec.Rules == nil {
		// No Route found in spec. Make one
		httpr.Spec.Rules = []v1beta1gateway.HTTPRouteRule{{}}
		makeRouteMatch(&httpr.Spec.Rules[0], version.Name)
		makeRouteFilter(&httpr.Spec.Rules[0])
		makeBackendRef(&httpr.Spec.Rules[0], serviceName, port)
	} else {
		// Check for missing BackendRefs
		for _, rule := range httpr.Spec.Rules {
			// FIXME: Do both Matches *and* Filters need to be present for a valid HTTPRoute?
			//        If not, we could rewrite this entire block to just be: "if resource == nil -> makeResource"
			if rule.Matches != nil { //|| rule.Filters != nil {
				if rule.BackendRefs == nil {
					makeBackendRef(&rule, serviceName, port)
				}
			}
		}
	}

	return httpr
}

// Make an HTTPRouteMatch and append it to rule
func makeRouteMatch(rule *v1beta1gateway.HTTPRouteRule, match string) {
	pathMatchType := v1beta1gateway.PathMatchPathPrefix
	pathMatchValue := "/" + match

	routeMatch := v1beta1gateway.HTTPRouteMatch{
		Path: &v1beta1gateway.HTTPPathMatch{
			Type:  &pathMatchType,
			Value: &pathMatchValue,
		},
		//Headers:
		//QueryParams:
		//Method:
	}

	rule.Matches = append(rule.Matches, routeMatch)
}

// Make an HTTPRouteFilter and append it to rule
func makeRouteFilter(rule *v1beta1gateway.HTTPRouteRule, prefix ...string) {
	var prefixMatch string

	if len(prefix) > 0 {
		prefixMatch = prefix[0]
	} else {
		prefixMatch = "/"
	}

	routeFilter := v1beta1gateway.HTTPRouteFilter{
		Type: v1beta1gateway.HTTPRouteFilterURLRewrite,
		URLRewrite: &v1beta1gateway.HTTPURLRewriteFilter{
			Path: &v1beta1gateway.HTTPPathModifier{
				Type:               v1beta1gateway.PrefixMatchHTTPPathModifier,
				ReplacePrefixMatch: &prefixMatch,
			},
		},
	}

	rule.Filters = append(rule.Filters, routeFilter)
}

// Make an HTTPBackendRef and append it to rule
func makeBackendRef(rule *v1beta1gateway.HTTPRouteRule, serviceName string, conPort int32) {
	if conPort < 0 {
		// FIXME: What to do when there is no valid container port? Create default? Empty? Fail?
		return
	}

	port := v1beta1gateway.PortNumber(conPort)

	routeBackend := v1beta1gateway.HTTPBackendRef{
		BackendRef: v1beta1gateway.BackendRef{
			BackendObjectReference: v1beta1gateway.BackendObjectReference{
				Name: v1beta1gateway.ObjectName(serviceName),
				Port: &port,
			},
		},
	}

	rule.BackendRefs = append(rule.BackendRefs, routeBackend)
}

func httpRouteName(router *corev1alpha1.Router) types.NamespacedName {
	return types.NamespacedName{
		Namespace: router.GetNamespace(),
		Name:      fmt.Sprintf("%s-route", router.GetName()),
	}
}

func setHTTPRouteLabels(httpRoute, router metav1.Object) {
	rlabels := httpRoute.GetLabels()
	if rlabels == nil {
		rlabels = map[string]string{}
	}

	for _, key := range []string{
		kai.KaiAppLabelKey,
		kai.KaiAppUIDLabelKey,
		kai.RouterLabelKey,
		kai.RouterUIDLabelKey,
		kai.RouterGenerationLabelKey,
	} {
		if value := getHTTPRouteLabelValue(key, router); value != "" {
			rlabels[key] = value
		}
	}

	httpRoute.SetLabels(rlabels)
}

func setHTTPRouteAnnotations(httpRoute, router metav1.Object) {
	annotations := httpRoute.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	routerAnnotations := router.GetAnnotations()
	if v, ok := routerAnnotations[kai.RouterAnnotationKey]; ok {
		annotations[kai.RouterAnnotationKey] = v

	}

	httpRoute.SetAnnotations(annotations)
}

func getHTTPRouteLabelValue(key string, router metav1.Object) string {
	switch key {
	case kai.KaiAppLabelKey:
		return router.GetLabels()[kai.KaiAppLabelKey]
	case kai.KaiAppUIDLabelKey:
		return router.GetLabels()[kai.KaiAppUIDLabelKey]
	case kai.RouterLabelKey:
		return router.GetName()
	case kai.RouterUIDLabelKey:
		return string(router.GetUID())
	case kai.RouterGenerationLabelKey:
		return fmt.Sprint(router.GetGeneration())
	}
	return ""
}

func getContainerPort(version *corev1alpha1.Version) int32 {
	// FIXME: What to do if there are multiple containers?
	// FIXME: How to easily determine if all these data are not nil?
	if version.Spec.Containers != nil && version.Spec.Containers[0].Ports != nil {
		return version.Spec.Containers[0].Ports[0].ContainerPort
	}

	//fmt.Errorf("failed to retrieve container port number %q: %w", version.GetNamespace(), err)
	return -1
}
