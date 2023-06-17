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

func DeploymentName(v *corev1alpha1.Version) types.NamespacedName {
	return types.NamespacedName{
		Namespace: v.Namespace,
		Name:      fmt.Sprintf("%s-deployment", v.GetName()),
	}
}

func ServiceName(v *corev1alpha1.Version) types.NamespacedName {
	return types.NamespacedName{
		Namespace: v.Namespace,
		Name:      fmt.Sprintf("%s-service", v.GetName()),
	}
}

func HPAName(v *corev1alpha1.Version) types.NamespacedName {
	return types.NamespacedName{
		Namespace: v.Namespace,
		Name:      fmt.Sprintf("%s-hpa", v.GetName()),
	}
}

func MakeAnnotations(v *corev1alpha1.Version) map[string]string {
	return kmap.Filter(v.GetAnnotations(), excludeAnnotations.Has)
}

func MakeLabels(v *corev1alpha1.Version) map[string]string {
	labels := kmap.Filter(v.GetLabels(), excludeLabels.Has)
	labels = kmap.Union(labels, map[string]string{
		kai.VersionLabelKey:    v.Name,
		kai.VersionUIDLabelKey: string(v.UID),
		kai.AppLabelKey:        v.Name,
	})

	return labels
}

func MakeSelector(v *corev1alpha1.Version) *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: map[string]string{
			kai.VersionUIDLabelKey: string(v.UID),
		},
	}
}

func MakeServiceSelector(v *corev1alpha1.Version) map[string]string {
	return map[string]string{
		kai.VersionUIDLabelKey: string(v.UID),
	}
}
