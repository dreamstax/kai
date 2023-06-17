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

package kai

import "k8s.io/apimachinery/pkg/runtime/schema"

const (
	// GroupName is the group name for Kai labels and annotations
	GroupName = "core.kai.io"

	AIGroupName = "ai.kai.io"

	// AppLabelKey is the label key attached to k8s resources to indicate which app triggered their creation
	KaiInferenceServiceLabelKey = AIGroupName + "/inferenceService"

	// KaiAppUIDLabelKey is the label key attached to k8s resources to indicate which config triggered their creation
	KaiInferenceServiceUIDLabelKey = AIGroupName + "/inferenceServiceUID"

	// AppLabelKey is the label key attached to k8s resources to indicate which app triggered their creation
	KaiAppLabelKey = GroupName + "/app"

	// KaiAppUIDLabelKey is the label key attached to k8s resources to indicate which config triggered their creation
	KaiAppUIDLabelKey = GroupName + "/appUID"

	// ConfigLabelKey is the label key attached to k8s resources to indicate which config triggered their creation
	ConfigLabelKey = GroupName + "/config"

	// ConfigGenerationLabelKey is the label key attached to k8s resources to indicate which config triggered their creation
	ConfigGenerationLabelKey = GroupName + "/configGeneration"

	// ConfigUIDLabelKey is the label key attached to k8s resources to indicate which config triggered their creation
	ConfigUIDLabelKey = GroupName + "/configUID"

	// VersionLabelKey is the label key attached to k8s resources to indicate which version triggered their creation.
	VersionLabelKey = GroupName + "/version"

	// VersionUID is the label key attached to a version
	VersionUIDLabelKey = GroupName + "/versionUID"

	// RouterLabelKey is the annotation attched to a version to indicate whether it is referenced by a router
	RouterLabelKey = GroupName + "/router"

	// RouterLabelKey is the label key attached to k8s resources to indicate which router triggered their creation
	RouterUIDLabelKey = GroupName + "/routerUID"

	// RouterGenerationLabelKey is the label key attached to k8s resources to indicate which app triggered their creation
	RouterGenerationLabelKey = GroupName + "/routerGeneration"

	// RouterAnnotationKey is the annotation attched to a version to indicate whether it is referenced by a router
	RouterAnnotationKey = GroupName + "/router"

	// k8s app label key
	AppLabelKey = "app"
)

var (
	SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha1"}
)

func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}
