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
	corev1alpha1 "github.com/dreamstax/kai/api/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// InferenceServiceSpec defines the desired state of InferenceService
type InferenceServiceSpec struct {
	// +required
	Model *ModelSpec `json:"model,omitempty"`

	// +optional
	Template corev1alpha1.AppTemplateSpec `json:"template"`
}

// InferenceServiceStatus defines the observed state of InferenceService
type InferenceServiceStatus struct {
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

// InferenceService is the Schema for the inferenceservices API
type InferenceService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InferenceServiceSpec   `json:"spec,omitempty"`
	Status InferenceServiceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// InferenceServiceList contains a list of InferenceService
type InferenceServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InferenceService `json:"items"`
}

func (is *InferenceService) GetGroupVersionKind() schema.GroupVersionKind {
	return is.GroupVersionKind()
}

func init() {
	SchemeBuilder.Register(&InferenceService{}, &InferenceServiceList{})
}
