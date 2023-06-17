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

package version

import (
	"context"
	"fmt"

	corev1alpha1 "github.com/dreamstax/kai/api/core/v1alpha1"
	"github.com/dreamstax/kai/api/kai"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/kmeta"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciler struct {
	client kclient.Client
}

func NewReconciler(client kclient.Client) *Reconciler {
	return &Reconciler{
		client: client,
	}
}

func (r *Reconciler) Reconcile(ctx context.Context, config *corev1alpha1.Config) error {
	// fetch what should be this configs version
	_, err := r.fetchVersion(ctx, config)
	if apierrs.IsNotFound(err) {
		// doesn't exist so create it now
		v, err := r.createVersion(ctx, config)
		if apierrs.IsAlreadyExists(err) {
			return fmt.Errorf("version already exists for config %q: %w", v.Name, err)
		} else if err != nil {
			// TODO: Set failed status on config
			return fmt.Errorf("failed to create version %q: %w", config.Name, err)
		}
	} else if apierrs.IsAlreadyExists(err) {
		// TODO: Set failed status on config
		return fmt.Errorf("failed to create version for config %q: %w", config.Name, err)
	} else if err != nil {
		return fmt.Errorf("failed to get version %q: %w", config.Name, err)
	}

	// TODO: Set version as latest on config for visibility

	// TODO: Set status on config for router visibility to determine active state

	return nil
}

func (r *Reconciler) createVersion(ctx context.Context, config *corev1alpha1.Config) (*corev1alpha1.Version, error) {
	v := makeVersion(config)

	err := r.client.Create(ctx, v)
	if err != nil {
		return nil, err
	}

	return v, nil
}

// fetchVersion attempts to retrieve what should be be the current config's version resource
func (r *Reconciler) fetchVersion(ctx context.Context, config *corev1alpha1.Config) (*corev1alpha1.Version, error) {
	name := determineName(config)

	v := &corev1alpha1.Version{}
	err := r.client.Get(ctx, name, v)
	if err != nil {
		if apierrs.IsNotFound(err) {
			// we expect not found on new versions
			return nil, err
		}

		return nil, fmt.Errorf("failed to get version %q: %w", name, err)
	}

	return v, apierrs.NewAlreadyExists(kai.Resource("versions"), name.Name)
}

// determineName returns the relevant name for a version resource based on config
// if user specifies name in template we use that otherwise use our consistent generated name
func determineName(config *corev1alpha1.Config) types.NamespacedName {
	if name := config.Spec.Template.Name; name != "" {
		return types.NamespacedName{
			Namespace: config.Namespace,
			Name:      name,
		}
	}

	// knative has a neat way of generating names for child resources so use that
	// TODO: may want to copy and freeze this later
	name := kmeta.ChildName(config.Name, fmt.Sprintf("-%05d", config.Generation))

	return types.NamespacedName{Name: name, Namespace: config.Namespace}
}

func makeVersion(config *corev1alpha1.Config) *corev1alpha1.Version {
	// TODO: determine if we need to deep copy here
	spec := config.Spec.DeepCopy()
	v := &corev1alpha1.Version{
		ObjectMeta: spec.Template.ObjectMeta,
		Spec:       spec.Template.Spec,
	}

	v.Namespace = config.Namespace
	v.Name = determineName(config).Name

	setVersionLabels(v, config)
	setVersionAnnotations(v, config)

	v.OwnerReferences = append(v.OwnerReferences, *kmeta.NewControllerRef(config))

	return v
}

func setVersionLabels(v, config metav1.Object) {
	labels := v.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}

	for _, key := range []string{
		kai.KaiAppLabelKey,
		kai.KaiAppUIDLabelKey,
		kai.ConfigLabelKey,
		kai.ConfigUIDLabelKey,
		kai.ConfigGenerationLabelKey,
	} {
		if value := getVersionLabelValue(key, config); value != "" {
			labels[key] = value
		}
	}

	v.SetLabels(labels)
}

func setVersionAnnotations(v, config metav1.Object) {
	annotations := v.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	cfgAnnotations := config.GetAnnotations()
	if v, ok := cfgAnnotations[kai.RouterAnnotationKey]; ok {
		annotations[kai.RouterAnnotationKey] = v
		// TODO: version is referenced in a router so will be active
		// set annotations/labels to signify this state
	}

	v.SetAnnotations(annotations)
}

func getVersionLabelValue(key string, config metav1.Object) string {
	switch key {
	case kai.KaiAppLabelKey:
		return config.GetLabels()[kai.KaiAppLabelKey]
	case kai.KaiAppUIDLabelKey:
		return config.GetLabels()[kai.KaiAppUIDLabelKey]
	case kai.ConfigLabelKey:
		return config.GetName()
	case kai.ConfigUIDLabelKey:
		return string(config.GetUID())
	case kai.ConfigGenerationLabelKey:
		return fmt.Sprint(config.GetGeneration())
	}
	return ""
}
