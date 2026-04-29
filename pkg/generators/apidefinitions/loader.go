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
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"sigs.k8s.io/yaml"
)

const (
	// We define an apiVersion and kinds for defining APIs
	// in the source tree like we do for everything else.
	schemeGroupVersion = "apidefinitions.k8s.io/v1alpha1"
	kindAPIVersion     = "APIVersion"

	// We have a naming convention for the files used to define
	// APIs in the source tree.
	apiVersionFile = "apiversion.yaml"
)

// LoadAPIVersion reads an apiversion.yaml file, returning nil if absent.
func LoadAPIVersion(dir string) (*APIVersion, error) {
	data, err := readManifest(dir, apiVersionFile)
	if err != nil || data == nil {
		return nil, err
	}
	av := &APIVersion{}
	if err := yaml.Unmarshal(data, av); err != nil {
		return nil, fmt.Errorf("%s: %w", filepath.Join(dir, apiVersionFile), err)
	}
	if err := validateTypeMeta(av.APIVersion, av.Kind, kindAPIVersion); err != nil {
		return nil, fmt.Errorf("%s: %w", filepath.Join(dir, apiVersionFile), err)
	}
	if err := validateName(av); err != nil {
		return nil, fmt.Errorf("%s: %w", filepath.Join(dir, apiVersionFile), err)
	}
	return av, nil
}

func readManifest(dir, filename string) ([]byte, error) {
	data, err := os.ReadFile(filepath.Join(dir, filename))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	return data, err
}

func validateTypeMeta(actualAPIVersion, actualKind, expectedKind string) error {
	if actualAPIVersion != schemeGroupVersion {
		return fmt.Errorf("expected apiVersion %s but got %s", schemeGroupVersion, actualAPIVersion)
	}
	if actualKind != expectedKind {
		return fmt.Errorf("expected kind %s but got %s", expectedKind, actualKind)
	}
	return nil
}

var (
	groupRegexp = regexp.MustCompile(`^[a-z0-9]+(\.[a-z0-9]+)*$`)
)

func validateName(av *APIVersion) error {
	if av.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}
	g, v, err := splitGroupVersion(av.Metadata.Name)
	if err != nil {
		return fmt.Errorf("metadata.name: %w", err)
	}
	if g != "" && !groupRegexp.MatchString(g) {
		return fmt.Errorf("metadata.name: group %q must be lowercase letters, optionally dot-separated", g)
	}
	if len(v) < 1 {
		return fmt.Errorf("metadata.name: version is required")
	}
	return nil
}

// splitGroupVersion parses "<group>/<version>" or "<version>" (core group).
func splitGroupVersion(name string) (string, string, error) {
	if i := strings.LastIndex(name, "/"); i >= 0 {
		return name[:i], name[i+1:], nil
	}
	if name == "" {
		return "", "", fmt.Errorf("name is empty")
	}
	return "", name, nil
}
