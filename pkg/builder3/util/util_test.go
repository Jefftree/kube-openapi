/*
Copyright 2022 The Kubernetes Authors.

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

package util

import (
	"reflect"
	"testing"

	"k8s.io/kube-openapi/pkg/validation/spec"
)

func TestHasSiblingFieldsExhaustive(t *testing.T) {
	if hasSiblingFields(&spec.Schema{}) {
		t.Fatal("expected empty spec.Schema to have no sibling fields")
	}
	if hasSiblingFields(&spec.Schema{SchemaProps: spec.SchemaProps{Ref: spec.MustCreateRef("#/components/schemas/Foo")}}) {
		t.Fatal("expected Ref-only spec.Schema to have no sibling fields")
	}

	schemaType := reflect.TypeOf(spec.Schema{})
	const expectedTopFields = 4
	if schemaType.NumField() != expectedTopFields {
		t.Fatalf("spec.Schema top-level field count changed from %d to %d; update hasSiblingFields", expectedTopFields, schemaType.NumField())
	}

	leafCount := 0
	var checkStruct func(path []int, st reflect.Type)
	checkStruct = func(path []int, st reflect.Type) {
		for i := 0; i < st.NumField(); i++ {
			sf := st.Field(i)
			idx := append(append([]int(nil), path...), i)
			if sf.Anonymous && sf.Type.Kind() == reflect.Struct {
				checkStruct(idx, sf.Type)
				continue
			}
			leafCount++
			if sf.Name == "Ref" {
				continue
			}

			var s spec.Schema
			fv := reflect.ValueOf(&s).Elem().FieldByIndex(idx)
			switch fv.Kind() {
			case reflect.String:
				fv.SetString("x")
			case reflect.Bool:
				fv.SetBool(true)
			case reflect.Ptr:
				fv.Set(reflect.New(fv.Type().Elem()))
			case reflect.Slice:
				fv.Set(reflect.MakeSlice(fv.Type(), 1, 1))
			case reflect.Map:
				m := reflect.MakeMapWithSize(fv.Type(), 1)
				m.SetMapIndex(reflect.ValueOf("k"), reflect.Zero(fv.Type().Elem()))
				fv.Set(m)
			case reflect.Interface:
				fv.Set(reflect.ValueOf("x"))
			default:
				t.Fatalf("unhandled field kind %v for field %s", fv.Kind(), sf.Name)
			}

			if !hasSiblingFields(&s) {
				t.Errorf("hasSiblingFields returned false when field %s was set", sf.Name)
			}
		}
	}
	checkStruct(nil, schemaType)

	const expectedLeafFields = 41
	if leafCount != expectedLeafFields {
		t.Fatalf("spec.Schema leaf field count changed from %d to %d; update hasSiblingFields", expectedLeafFields, leafCount)
	}
}
