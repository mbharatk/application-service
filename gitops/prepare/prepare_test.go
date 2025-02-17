/*
Copyright 2021 Red Hat, Inc.

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
package prepare

import (
	"context"
	"testing"

	appstudiov1alpha1 "github.com/redhat-appstudio/application-service/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestPrepareGitopsConfig(t *testing.T) {
	client := fake.NewClientBuilder().Build()

	component := appstudiov1alpha1.Component{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "appstudio.redhat.com/v1alpha1",
			Kind:       "Component",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myName",
			Namespace: "myNamespace",
		},
	}

	want := GitopsConfig{
		BuildBundle: FallbackBuildBundle,
	}

	t.Run("", func(t *testing.T) {
		if got := PrepareGitopsConfig(context.TODO(), client, component); got != want {
			t.Errorf("ResolveBuildBundle() = %v, want %v", got, want)
		}
	})

}

func TestResolveBuildBundle(t *testing.T) {
	ctx := context.TODO()

	component := appstudiov1alpha1.Component{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "appstudio.redhat.com/v1alpha1",
			Kind:       "Component",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myName",
			Namespace: "myNamespace",
		},
	}

	tests := []struct {
		name string
		data corev1.ConfigMap
		want string
	}{
		{
			name: "should resolve the build bundle in case a configmap exists in the component's namespace",
			data: corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "ConfigMap",
				},
				Data: map[string]string{
					BuildBundleConfigMapKey: "quay.io/foo/bar:1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      BuildBundleConfigMapName,
					Namespace: component.Namespace,
				},
			},
			want: "quay.io/foo/bar:1",
		},
		{
			name: "should resolve the build bundle in case a configmap exists in the default namespace",
			data: corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "ConfigMap",
				},
				Data: map[string]string{
					BuildBundleConfigMapKey: "quay.io/foo/bar:2",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      BuildBundleConfigMapName,
					Namespace: BuildBundleDefaultNamespace,
				},
			},
			want: "quay.io/foo/bar:2",
		},
		{
			name: "should fall back to the hard-coded bundle in case the resolution fails",
			data: corev1.ConfigMap{},
			want: FallbackBuildBundle,
		},
		{
			name: "should ignore malformed configmaps",
			data: corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "ConfigMap",
				},
				Data: map[string]string{
					"invalidKey": "quay.io/foo/bar:3",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      BuildBundleConfigMapName,
					Namespace: BuildBundleDefaultNamespace,
				},
			},
			want: FallbackBuildBundle,
		},
		{
			name: "should ignore configmaps with empty keys",
			data: corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "ConfigMap",
				},
				Data: map[string]string{
					BuildBundleConfigMapKey: "",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      BuildBundleConfigMapName,
					Namespace: BuildBundleDefaultNamespace,
				},
			},
			want: FallbackBuildBundle,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := fake.NewClientBuilder().WithRuntimeObjects(&tt.data).Build()

			if got := resolveBuildBundle(ctx, client, component); got != tt.want {
				t.Errorf("ResolveBuildBundle() = %v, want %v", got, tt.want)
			}
		})
	}
}
