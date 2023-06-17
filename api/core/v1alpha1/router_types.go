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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	v1beta1gateway "sigs.k8s.io/gateway-api/apis/v1beta1"
)

// RouterSpec defines the desired state of Router
type RouterSpec struct {
	// +kubebuilder:pruning:PreserveUnknownFields
	// +optional
	Route HTTPRouteSpec `json:"route,omitempty"`
}

// NOTE: The following was copied from sigs.k8s.io/gateway-api@v0.6.0-rc2.0.20230421073315-b8b01612a092/apis/v1beta1/httproute_types.go
//       The default route rules and matches have been removed because they were interfering with our implementation.

// HTTPRoute provides a way to route HTTP requests. This includes the capability
// to match requests by hostname, path, header, or query param. Filters can be
// used to specify additional processing steps. Backends specify where matching
// requests should be routed.
type HTTPRoute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of HTTPRoute.
	Spec HTTPRouteSpec `json:"spec"`

	// Status defines the current state of HTTPRoute.
	Status v1beta1gateway.HTTPRouteStatus `json:"status,omitempty"`
}

// NOTE: The following was copied from sigs.k8s.io/gateway-api@v0.6.0-rc2.0.20230421073315-b8b01612a092/apis/v1beta1/httproute_types.go
//       The default route rules and matches have been removed because they were interfering with our implementation.

// HTTPRouteSpec defines the desired state of HTTPRoute
type HTTPRouteSpec struct {
	v1beta1gateway.CommonRouteSpec `json:",inline"`

	// Hostnames defines a set of hostname that should match against the HTTP
	// Host header to select a HTTPRoute to process the request. This matches
	// the RFC 1123 definition of a hostname with 2 notable exceptions:
	//
	// 1. IPs are not allowed.
	// 2. A hostname may be prefixed with a wildcard label (`*.`). The wildcard
	//    label must appear by itself as the first label.
	//
	// If a hostname is specified by both the Listener and HTTPRoute, there
	// must be at least one intersecting hostname for the HTTPRoute to be
	// attached to the Listener. For example:
	//
	// * A Listener with `test.example.com` as the hostname matches HTTPRoutes
	//   that have either not specified any hostnames, or have specified at
	//   least one of `test.example.com` or `*.example.com`.
	// * A Listener with `*.example.com` as the hostname matches HTTPRoutes
	//   that have either not specified any hostnames or have specified at least
	//   one hostname that matches the Listener hostname. For example,
	//   `*.example.com`, `test.example.com`, and `foo.test.example.com` would
	//   all match. On the other hand, `example.com` and `test.example.net` would
	//   not match.
	//
	// Hostnames that are prefixed with a wildcard label (`*.`) are interpreted
	// as a suffix match. That means that a match for `*.example.com` would match
	// both `test.example.com`, and `foo.test.example.com`, but not `example.com`.
	//
	// If both the Listener and HTTPRoute have specified hostnames, any
	// HTTPRoute hostnames that do not match the Listener hostname MUST be
	// ignored. For example, if a Listener specified `*.example.com`, and the
	// HTTPRoute specified `test.example.com` and `test.example.net`,
	// `test.example.net` must not be considered for a match.
	//
	// If both the Listener and HTTPRoute have specified hostnames, and none
	// match with the criteria above, then the HTTPRoute is not accepted. The
	// implementation must raise an 'Accepted' Condition with a status of
	// `False` in the corresponding RouteParentStatus.
	//
	// In the event that multiple HTTPRoutes specify intersecting hostnames (e.g.
	// overlapping wildcard matching and exact matching hostnames), precedence must
	// be given to rules from the HTTPRoute with the largest number of:
	//
	// * Characters in a matching non-wildcard hostname.
	// * Characters in a matching hostname.
	//
	// If ties exist across multiple Routes, the matching precedence rules for
	// HTTPRouteMatches takes over.
	//
	// Support: Core
	//
	// +optional
	// +kubebuilder:validation:MaxItems=16
	Hostnames []v1beta1gateway.Hostname `json:"hostnames,omitempty"`

	// Rules are a list of HTTP matchers, filters and actions.
	//
	// +optional
	// +kubebuilder:validation:MaxItems=16
	Rules []v1beta1gateway.HTTPRouteRule `json:"rules,omitempty"`
}

// Converts a kai HTTPRouteSpec into a K8s Gateway HTTPRouteSpec
func (rs *HTTPRouteSpec) ToK8sGatewaySpec() v1beta1gateway.HTTPRouteSpec {
	spec := rs.DeepCopy()

	return v1beta1gateway.HTTPRouteSpec{
		CommonRouteSpec: spec.CommonRouteSpec,
		Hostnames:       spec.Hostnames,
		Rules:           spec.toK8sGatewayRules(),
	}
}

// Converts a kai HTTPRouteRules into K8s Gateway HTTPRouteRules
func (rs *HTTPRouteSpec) toK8sGatewayRules() []v1beta1gateway.HTTPRouteRule {
	var out []v1beta1gateway.HTTPRouteRule

	for _, v := range rs.Rules {
		out = append(out, v1beta1gateway.HTTPRouteRule{
			Matches:     v.Matches,
			Filters:     v.Filters,
			BackendRefs: v.BackendRefs,
		})
	}

	return out
}

// NOTE: The following was copied from sigs.k8s.io/gateway-api@v0.6.0-rc2.0.20230421073315-b8b01612a092/apis/v1beta1/httproute_types.go
//       The default route rules and matches have been removed because they were interfering with our implementation.

// HTTPRouteRule defines semantics for matching an HTTP request based on
// conditions (matches), processing it (filters), and forwarding the request to
// an API object (backendRefs).
type HTTPRouteRule struct {
	// Matches define conditions used for matching the rule against incoming
	// HTTP requests. Each match is independent, i.e. this rule will be matched
	// if **any** one of the matches is satisfied.
	//
	// For example, take the following matches configuration:
	//
	// ```
	// matches:
	// - path:
	//     value: "/foo"
	//   headers:
	//   - name: "version"
	//     value: "v2"
	// - path:
	//     value: "/v2/foo"
	// ```
	//
	// For a request to match against this rule, a request must satisfy
	// EITHER of the two conditions:
	//
	// - path prefixed with `/foo` AND contains the header `version: v2`
	// - path prefix of `/v2/foo`
	//
	// See the documentation for HTTPRouteMatch on how to specify multiple
	// match conditions that should be ANDed together.
	//
	// If no matches are specified, the default is a prefix
	// path match on "/", which has the effect of matching every
	// HTTP request.
	//
	// Proxy or Load Balancer routing configuration generated from HTTPRoutes
	// MUST prioritize matches based on the following criteria, continuing on
	// ties. Across all rules specified on applicable Routes, precedence must be
	// given to the match with the largest number of:
	//
	// * Characters in a matching "Exact" path match
	// * Characters in a matching "Prefix" path match
	// * Header matches.
	// * Query param matches.
	//
	// Note: The precedence of RegularExpression path matches are implementation-specific.
	//
	// If ties still exist across multiple Routes, matching precedence MUST be
	// determined in order of the following criteria, continuing on ties:
	//
	// * The oldest Route based on creation timestamp.
	// * The Route appearing first in alphabetical order by
	//   "{namespace}/{name}".
	//
	// If ties still exist within an HTTPRoute, matching precedence MUST be granted
	// to the FIRST matching rule (in list order) with a match meeting the above
	// criteria.
	//
	// When no rules matching a request have been successfully attached to the
	// parent a request is coming from, a HTTP 404 status code MUST be returned.
	//
	// +optional
	// +kubebuilder:validation:MaxItems=8
	Matches []v1beta1gateway.HTTPRouteMatch `json:"matches,omitempty"`

	// Filters define the filters that are applied to requests that match
	// this rule.
	//
	// The effects of ordering of multiple behaviors are currently unspecified.
	// This can change in the future based on feedback during the alpha stage.
	//
	// Conformance-levels at this level are defined based on the type of filter:
	//
	// - ALL core filters MUST be supported by all implementations.
	// - Implementers are encouraged to support extended filters.
	// - Implementation-specific custom filters have no API guarantees across
	//   implementations.
	//
	// Specifying a core filter multiple times has unspecified or
	// implementation-specific conformance.
	//
	// All filters are expected to be compatible with each other except for the
	// URLRewrite and RequestRedirect filters, which may not be combined. If an
	// implementation can not support other combinations of filters, they must clearly
	// document that limitation. In all cases where incompatible or unsupported
	// filters are specified, implementations MUST add a warning condition to status.
	//
	// Support: Core
	//
	// +optional
	// +kubebuilder:validation:MaxItems=16
	Filters []v1beta1gateway.HTTPRouteFilter `json:"filters,omitempty"`

	// BackendRefs defines the backend(s) where matching requests should be
	// sent.
	//
	// Failure behavior here depends on how many BackendRefs are specified and
	// how many are invalid.
	//
	// If *all* entries in BackendRefs are invalid, and there are also no filters
	// specified in this route rule, *all* traffic which matches this rule MUST
	// receive a 500 status code.
	//
	// See the HTTPBackendRef definition for the rules about what makes a single
	// HTTPBackendRef invalid.
	//
	// When a HTTPBackendRef is invalid, 500 status codes MUST be returned for
	// requests that would have otherwise been routed to an invalid backend. If
	// multiple backends are specified, and some are invalid, the proportion of
	// requests that would otherwise have been routed to an invalid backend
	// MUST receive a 500 status code.
	//
	// For example, if two backends are specified with equal weights, and one is
	// invalid, 50 percent of traffic must receive a 500. Implementations may
	// choose how that 50 percent is determined.
	//
	// Support: Core for Kubernetes Service
	//
	// Support: Extended for Kubernetes ServiceImport
	//
	// Support: Implementation-specific for any other resource
	//
	// Support for weight: Core
	//
	// +optional
	// +kubebuilder:validation:MaxItems=16
	BackendRefs []v1beta1gateway.HTTPBackendRef `json:"backendRefs,omitempty"`
}

// RouterStatus defines the observed state of Router
type RouterStatus struct {
	// TODO add relevant status fields
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Router is the Schema for the routers API
type Router struct {
	metav1.TypeMeta `json:",inline"`

	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec RouterSpec `json:"spec,omitempty"`

	Status RouterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RouterList contains a list of Router
type RouterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Router `json:"items"`
}

func (r *Router) GetGroupVersionKind() schema.GroupVersionKind {
	return r.GroupVersionKind()
}

func init() {
	SchemeBuilder.Register(&Router{}, &RouterList{})
}
