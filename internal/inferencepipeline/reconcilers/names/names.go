/*
Copyright 2023 The Kai Authors.

Licensed under the Apache License, InferencePipeline 2.0 (the "License");
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

	aiv1alpha1 "github.com/dreamstax/kai/api/ai/v1alpha1"
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

	// annotations to exclude from inferencepipelines
	excludeAnnotations = sets.NewString()
)

func DeploymentName(i *aiv1alpha1.InferencePipeline) types.NamespacedName {
	return types.NamespacedName{
		Namespace: i.Namespace,
		Name:      fmt.Sprintf("%s-deployment", i.GetName()),
	}
}

func ServiceName(i *aiv1alpha1.InferencePipeline) types.NamespacedName {
	return types.NamespacedName{
		Namespace: i.Namespace,
		Name:      fmt.Sprintf("%s-service", i.GetName()),
	}
}

func HPAName(i *aiv1alpha1.InferencePipeline) types.NamespacedName {
	return types.NamespacedName{
		Namespace: i.Namespace,
		Name:      fmt.Sprintf("%s-hpa", i.GetName()),
	}
}

func MakeAnnotations(i *aiv1alpha1.InferencePipeline) map[string]string {
	return kmap.Filter(i.GetAnnotations(), excludeAnnotations.Has)
}

func MakeLabels(i *aiv1alpha1.InferencePipeline) map[string]string {
	labels := kmap.Filter(i.GetLabels(), excludeLabels.Has)
	labels = kmap.Union(labels, map[string]string{
		kai.InferencePipelineLabelKey:    i.Name,
		kai.InferencePipelineUIDLabelKey: string(i.UID),
	})

	return labels
}

func MakeSelector(i *aiv1alpha1.InferencePipeline) *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: map[string]string{
			kai.InferencePipelineUIDLabelKey: string(i.UID),
		},
	}
}

func MakeServiceSelector(i *aiv1alpha1.InferencePipeline) map[string]string {
	return map[string]string{
		kai.InferencePipelineUIDLabelKey: string(i.UID),
	}
}
