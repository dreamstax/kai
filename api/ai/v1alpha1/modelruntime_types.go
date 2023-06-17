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
	autoscaling "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ModelRuntimeSpec defines the desired state of ModelRuntime
type ModelRuntimeSpec struct {
	SupportedModelFormats []ModelFormat `json:"supportedModelFormats"`

	Containers []corev1.Container `json:"containers"`

	// +optional
	MinReplicas *int32 `json:"minReplicas,omitempty"`

	// +optional
	MaxReplicas int32 `json:"maxReplicas"`

	// +optional
	Metrics []autoscaling.MetricSpec `json:"metrics,omitempty"`

	// +optional
	Behavior *autoscaling.HorizontalPodAutoscalerBehavior `json:"behavior,omitempty"`

	// TODO: determine need to set selectors and affinities here. it's possible in a multi-tenant
	// env this might be a convenient way to provide tenant specific settings or namespace specific settings
}

// ModelRuntimeStatus defines the observed state of ModelRuntime
type ModelRuntimeStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ModelRuntime is the Schema for the modelruntimes API
type ModelRuntime struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ModelRuntimeSpec   `json:"spec,omitempty"`
	Status ModelRuntimeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ModelRuntimeList contains a list of ModelRuntime
type ModelRuntimeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ModelRuntime `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ModelRuntime{}, &ModelRuntimeList{})
}
