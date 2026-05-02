/*
Copyright The Kubernetes Authors.

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

package apidefinitions

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAPIVersion(t *testing.T) {
	const manifest = `apiVersion: apidefinitions.k8s.io/v1alpha1
kind: APIVersion
metadata:
  name: apps/v1
`
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "apiversion.yaml"), []byte(manifest), 0644); err != nil {
		t.Fatal(err)
	}

	av, err := LoadAPIVersion(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if av == nil {
		t.Fatal("expected APIVersion, got nil")
	}
	if got, want := av.Metadata.Name, "apps/v1"; got != want {
		t.Errorf("metadata.name = %q, want %q", got, want)
	}
	if got, want := av.VersionFromName(), "v1"; got != want {
		t.Errorf("VersionFromName() = %q, want %q", got, want)
	}
}

func TestLoadAPIVersion_Missing(t *testing.T) {
	av, err := LoadAPIVersion(t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if av != nil {
		t.Errorf("expected nil APIVersion, got %+v", av)
	}
}

func TestLoadAPIGroup(t *testing.T) {
	const manifest = `apiVersion: apidefinitions.k8s.io/v1alpha1
kind: APIGroup
metadata:
  name: apps
spec:
  modelPackage: io.k8s.api.apps
`
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "apigroup.yaml"), []byte(manifest), 0644); err != nil {
		t.Fatal(err)
	}

	g, err := LoadAPIGroup(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g == nil {
		t.Fatal("expected APIGroup, got nil")
	}
	if got, want := g.Metadata.Name, "apps"; got != want {
		t.Errorf("metadata.name = %q, want %q", got, want)
	}
	if got, want := g.Spec.ModelPackage, "io.k8s.api.apps"; got != want {
		t.Errorf("spec.modelPackage = %q, want %q", got, want)
	}
	if got, want := g.ModelPackageFor("v1"), "io.k8s.api.apps.v1"; got != want {
		t.Errorf("ModelPackageFor(v1) = %q, want %q", got, want)
	}
}

func TestLoadAPIGroup_Missing(t *testing.T) {
	g, err := LoadAPIGroup(t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g != nil {
		t.Errorf("expected nil APIGroup, got %+v", g)
	}
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name string
		err  bool
	}{
		{name: "apps/v1"},
		{name: "v1"},
		{name: "apidefinitions.k8s.io/v1alpha1"},
		{name: "code-generator.k8s.io/v2beta2"},
		{name: "foo/bar/v1", err: true},
		{name: "", err: true},
		{name: "apps/", err: true},
		{name: "/v1", err: true},
		{name: "Apps/v1", err: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateName(&APIVersion{Metadata: Metadata{Name: tc.name}})
			if tc.err && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.err && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
