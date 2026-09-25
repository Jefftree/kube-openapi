/*
Copyright 2021 The Kubernetes Authors.

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

package builder3

import (
	"cmp"
	"slices"

	"k8s.io/kube-openapi/pkg/common"
	"k8s.io/kube-openapi/pkg/spec3"
)

type paramKey struct {
	Name string
	Kind common.ParameterKind
}

func mapKeyFromParam(param common.Parameter) paramKey {
	return paramKey{
		Name: param.Name(),
		Kind: param.Kind(),
	}
}

func (s parameters) Len() int      { return len(s) }
func (s parameters) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type parameters []*spec3.Parameter

type byNameIn struct {
	parameters
}

func (s byNameIn) Less(i, j int) bool {
	return s.parameters[i].Name < s.parameters[j].Name || (s.parameters[i].Name == s.parameters[j].Name && s.parameters[i].In < s.parameters[j].In)
}

func hasParamKey(keys []paramKey, target paramKey) bool {
	for i := range keys {
		if keys[i] == target {
			return true
		}
	}
	return false
}

// SortParameters sorts parameters by Name and In fields.
func sortParameters(p []*spec3.Parameter) {
	if len(p) <= 1 {
		return
	}
	slices.SortFunc(p, func(a, b *spec3.Parameter) int {
		if a.Name != b.Name {
			return cmp.Compare(a.Name, b.Name)
		}
		return cmp.Compare(a.In, b.In)
	})
}
