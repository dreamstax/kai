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

// StepName returns the name for a step based on pipeline and index of step within a pipeline.
// This is somewhat crude currently. Users may want to name these steps something more
// descriptive and maybe we should support that?
func StepName(p *corev1alpha1.Pipeline, idx int) types.NamespacedName {
	return types.NamespacedName{
		Namespace: p.Namespace,
		Name:      fmt.Sprintf("%s-step-%d", p.GetName(), idx),
	}
}

func MakeLabels(p *corev1alpha1.Pipeline) map[string]string {
	labels := kmap.Filter(p.GetLabels(), excludeLabels.Has)
	labels = kmap.Union(labels, map[string]string{
		kai.PipelineLabelKey:    p.Name,
		kai.PipelineUIDLabelKey: string(p.UID),
	})

	return labels
}

func MakeSelector(p *corev1alpha1.Pipeline) *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: map[string]string{
			kai.PipelineUIDLabelKey: string(p.UID),
		},
	}
}
