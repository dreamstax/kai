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
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type InferencePipelineSpec struct {
	// +required
	Steps []*Step `json:"steps"`
}

// InferencePipelineStatus defines the observed state of InferencePipeline
type InferencePipelineStatus struct{}

type Step struct {
	corev1.PodSpec `json:",inline"`

	// ModelSpec defines the model to use for a prediction step. If both ModelSpec and PodSpec are defined
	// PodSpec will override any default values that have been defined within the ModelRuntime.
	// +optional
	ModelSpec *ModelSpec `json:"modelSpec,omitempty"`

	// minReplicas is the lower limit for the number of replicas to which the autoscaler
	// can scale down.  It defaults to 1 pod.  minReplicas is allowed to be 0 if the
	// alpha feature gate HPAScaleToZero is enabled and at least one Object or External
	// metric is configured.  Scaling is active as long as at least one metric value is
	// available.
	// +optional
	MinReplicas *int32 `json:"minReplicas,omitempty"`

	// maxReplicas is the upper limit for the number of replicas to which the autoscaler can scale up.
	// It cannot be less that minReplicas.
	// +optional
	MaxReplicas int32 `json:"maxReplicas"`

	// metrics contains the specifications for which to use to calculate the
	// desired replica count (the maximum replica count across all metrics will
	// be used).  The desired replica count is calculated multiplying the
	// ratio between the target value and the current value by the current
	// number of pods.  Ergo, metrics used must decrease as the pod count is
	// increased, and vice-versa.  See the individual metric source types for
	// more information about how each type of metric must respond.
	// If not set, the default metric will be set to 80% average CPU utilization.
	// +listType=atomic
	// +optional
	Metrics []autoscaling.MetricSpec `json:"metrics,omitempty"`

	// behavior configures the scaling behavior of the target
	// in both Up and Down directions (scaleUp and scaleDown fields respectively).
	// If not set, the default HPAScalingRules for scale up and scale down are used.
	// +optional
	Behavior *autoscaling.HorizontalPodAutoscalerBehavior `json:"behavior,omitempty"`
}

type ModelSpec struct {
	// +required
	ModelFormat ModelFormat `json:"modelFormat,omitempty"`

	// +required
	URI string `json:"uri,omitempty"`

	// optionally set a modelRuntime - if modelRuntime is specified the inferenceContainer
	// specified within it will be used regardless of modelFormat see ModelRuntime for more info
	// +optional
	ModelRuntime string `json:"modelRuntime,omitempty"`

	// +optional
	ServiceAccountRef string `json:"servicecAccountRef,omitempty"`
}

type ModelFormat string

// supported model formats
const (
	PytorchModelFormat ModelFormat = "pytorch"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// InferencePipeline is the Schema for the inferencepipelines API
type InferencePipeline struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InferencePipelineSpec   `json:"spec,omitempty"`
	Status InferencePipelineStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// InferencePipelineList contains a list of InferencePipeline
type InferencePipelineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InferencePipeline `json:"items"`
}

func (is *InferencePipeline) GetGroupVersionKind() schema.GroupVersionKind {
	return is.GroupVersionKind()
}

func init() {
	SchemeBuilder.Register(&InferencePipeline{}, &InferencePipelineList{})
}
