/*
Copyright 2023.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

// StepTemplateSpec is a wrapper for resourcces embedding a StepSpec. This strategy is borrowed
// from k8s core (PodTemplate) and other popular projects like knative
type StepTemplateSpec struct {
	// +kubebuilder:pruning:PreserveUnknownFields
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	Spec StepSpec `json:"spec,omitempty"`
}

// StepSpec defines the desired state of Step
type StepSpec struct {
	// PodSpec can be used to define a standalone Pod for use within a step. Additionally
	// When Model is also provided it serves as means to override values set within the modelRuntime.
	PodSpec `json:",inline"`

	// +optional
	Model *ModelSpec `json:"model,omitempty"`

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
	// ModelFormat specifies the type of of the model e.g.; pytorch, onnx
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

// StepStatus defines the observed state of Step
type StepStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Step is the Schema for the steps API
type Step struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StepSpec   `json:"spec,omitempty"`
	Status StepStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// StepList contains a list of Step
type StepList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Step `json:"items"`
}

func (s *Step) GetGroupVersionKind() schema.GroupVersionKind {
	return s.GroupVersionKind()
}

func (s *Step) NamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: s.Namespace,
		Name:      s.Name,
	}
}

func init() {
	SchemeBuilder.Register(&Step{}, &StepList{})
}
