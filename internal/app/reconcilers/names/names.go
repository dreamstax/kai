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
	corev1alpha1 "github.com/dreamstax/kai/api/core/v1alpha1"
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

func RouterName(app *corev1alpha1.App) types.NamespacedName {
	return types.NamespacedName{
		Namespace: app.GetNamespace(),
		Name:      app.GetName(),
	}
}

func ConfigName(app *corev1alpha1.App) types.NamespacedName {
	return types.NamespacedName{
		Namespace: app.GetNamespace(),
		Name:      app.GetName(),
	}
}

func MakeAnnotations(app *corev1alpha1.App) map[string]string {
	return kmap.Filter(app.GetAnnotations(), excludeAnnotations.Has)
}

func MakeLabels(app *corev1alpha1.App) map[string]string {
	labels := kmap.Filter(app.GetLabels(), excludeLabels.Has)
	labels = kmap.Union(labels, map[string]string{
		kai.KaiAppLabelKey:    app.Name,
		kai.KaiAppUIDLabelKey: string(app.UID),
	})

	return labels
}
