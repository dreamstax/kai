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

package names

import (
	"fmt"

	corev1alpha1 "github.com/dreamstax/kai/api/core/v1alpha1"
	"github.com/dreamstax/kai/api/kai"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"knative.dev/pkg/kmap"
)

var (
	// there might be labels we want to exclude on subresources
	// knative excludes route labels in revisions
	excludeLabels = sets.NewString()

	// annotations to exclude from versions
	excludeAnnotations = sets.NewString()
)

func DeploymentName(s *corev1alpha1.Step) types.NamespacedName {
	return types.NamespacedName{
		Namespace: s.Namespace,
		Name:      fmt.Sprintf("%s-deployment", s.GetName()),
	}
}

func ServiceName(s *corev1alpha1.Step) types.NamespacedName {
	return types.NamespacedName{
		Namespace: s.Namespace,
		Name:      fmt.Sprintf("%s-service", s.GetName()),
	}
}

func HPAName(s *corev1alpha1.Step) types.NamespacedName {
	return types.NamespacedName{
		Namespace: s.Namespace,
		Name:      fmt.Sprintf("%s-hpa", s.GetName()),
	}
}

func MakeAnnotations(s *corev1alpha1.Step) map[string]string {
	return kmap.Filter(s.GetAnnotations(), excludeAnnotations.Has)
}

func MakeLabels(s *corev1alpha1.Step) map[string]string {
	labels := kmap.Filter(s.GetLabels(), excludeLabels.Has)
	labels = kmap.Union(labels, map[string]string{
		kai.StepLabelKey:    s.Name,
		kai.StepUIDLabelKey: string(s.UID),
	})

	return labels
}

func MakeSelector(s *corev1alpha1.Step) *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: map[string]string{
			kai.StepUIDLabelKey: string(s.UID),
		},
	}
}

func MakeServiceSelector(s *corev1alpha1.Step) map[string]string {
	return map[string]string{
		kai.StepUIDLabelKey: string(s.UID),
	}
}
