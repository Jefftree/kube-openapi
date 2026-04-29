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
  name: test.apidefinitions.k8s.io/v1
spec:
  modelPackage: io.k8s.api.apps.v1
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
	if got, want := av.Metadata.Name, "test.apidefinitions.k8s.io/v1"; got != want {
		t.Errorf("metadata.name = %q, want %q", got, want)
	}
	if got, want := av.Spec.ModelPackage, "io.k8s.api.apps.v1"; got != want {
		t.Errorf("spec.modelPackage = %q, want %q", got, want)
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
