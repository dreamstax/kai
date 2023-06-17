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
	aiv1alpha1 "github.com/dreamstax/kai/api/ai/v1alpha1"
	"github.com/dreamstax/kai/api/kai"

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

func AppName(is *aiv1alpha1.InferenceService) types.NamespacedName {
	return types.NamespacedName{
		Namespace: is.GetNamespace(),
		Name:      is.GetName(),
	}
}

func MakeAnnotations(is *aiv1alpha1.InferenceService) map[string]string {
	return kmap.Filter(is.GetAnnotations(), excludeAnnotations.Has)
}

func MakeLabels(is *aiv1alpha1.InferenceService) map[string]string {
	labels := kmap.Filter(is.GetLabels(), excludeLabels.Has)
	labels = kmap.Union(labels, map[string]string{
		kai.KaiInferenceServiceLabelKey:    is.Name,
		kai.KaiInferenceServiceUIDLabelKey: string(is.UID),
	})

	return labels
}
