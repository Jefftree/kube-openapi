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
	"k8s.io/kube-openapi/pkg/validation/spec"
)

// hasSiblingFields returns true if s has any fields set other than Ref.
func hasSiblingFields(s *spec.Schema) bool {
	return len(s.Extensions) > 0 ||
		s.ID != "" ||
		s.Schema != "" ||
		s.Description != "" ||
		len(s.Type) > 0 ||
		s.Nullable ||
		s.Format != "" ||
		s.Title != "" ||
		s.Default != nil ||
		s.Maximum != nil ||
		s.ExclusiveMaximum ||
		s.Minimum != nil ||
		s.ExclusiveMinimum ||
		s.MaxLength != nil ||
		s.MinLength != nil ||
		s.Pattern != "" ||
		s.MaxItems != nil ||
		s.MinItems != nil ||
		s.UniqueItems ||
		s.MultipleOf != nil ||
		len(s.Enum) > 0 ||
		s.MaxProperties != nil ||
		s.MinProperties != nil ||
		len(s.Required) > 0 ||
		s.Items != nil ||
		len(s.AllOf) > 0 ||
		len(s.OneOf) > 0 ||
		len(s.AnyOf) > 0 ||
		s.Not != nil ||
		len(s.Properties) > 0 ||
		s.AdditionalProperties != nil ||
		len(s.PatternProperties) > 0 ||
		len(s.Dependencies) > 0 ||
		s.AdditionalItems != nil ||
		len(s.Definitions) > 0 ||
		s.Discriminator != "" ||
		s.ReadOnly ||
		s.ExternalDocs != nil ||
		s.Example != nil ||
		len(s.ExtraProps) > 0
}

func needsWrapRefAtRoot(s *spec.Schema) bool {
	return s.Ref.GetURL() != nil && s.Ref.String() != "" && hasSiblingFields(s)
}

func wrapRefsValue(schema spec.Schema) (spec.Schema, bool) {
	changed := false
	if needsWrapRefAtRoot(&schema) {
		schema.AllOf = []spec.Schema{{SchemaProps: spec.SchemaProps{Ref: schema.Ref}}}
		schema.Ref = spec.Ref{}
		changed = true
	}

	if len(schema.Definitions) > 0 {
		defsCloned := false
		for k, v := range schema.Definitions {
			if nv, c := wrapRefsValue(v); c {
				if !defsCloned {
					defsCloned = true
					changed = true
					cloned := make(spec.Definitions, len(schema.Definitions))
					for k2, v2 := range schema.Definitions {
						cloned[k2] = v2
					}
					schema.Definitions = cloned
				}
				schema.Definitions[k] = nv
			}
		}
	}

	if len(schema.Properties) > 0 {
		propsCloned := false
		for k, v := range schema.Properties {
			if nv, c := wrapRefsValue(v); c {
				if !propsCloned {
					propsCloned = true
					changed = true
					cloned := make(map[string]spec.Schema, len(schema.Properties))
					for k2, v2 := range schema.Properties {
						cloned[k2] = v2
					}
					schema.Properties = cloned
				}
				schema.Properties[k] = nv
			}
		}
	}

	if len(schema.PatternProperties) > 0 {
		patCloned := false
		for k, v := range schema.PatternProperties {
			if nv, c := wrapRefsValue(v); c {
				if !patCloned {
					patCloned = true
					changed = true
					cloned := make(map[string]spec.Schema, len(schema.PatternProperties))
					for k2, v2 := range schema.PatternProperties {
						cloned[k2] = v2
					}
					schema.PatternProperties = cloned
				}
				schema.PatternProperties[k] = nv
			}
		}
	}

	if len(schema.AllOf) > 0 {
		allOfCloned := false
		for i := range schema.AllOf {
			if nv, c := wrapRefsValue(schema.AllOf[i]); c {
				if !allOfCloned {
					allOfCloned = true
					changed = true
					cloned := make([]spec.Schema, len(schema.AllOf))
					copy(cloned, schema.AllOf)
					schema.AllOf = cloned
				}
				schema.AllOf[i] = nv
			}
		}
	}

	if len(schema.AnyOf) > 0 {
		anyOfCloned := false
		for i := range schema.AnyOf {
			if nv, c := wrapRefsValue(schema.AnyOf[i]); c {
				if !anyOfCloned {
					anyOfCloned = true
					changed = true
					cloned := make([]spec.Schema, len(schema.AnyOf))
					copy(cloned, schema.AnyOf)
					schema.AnyOf = cloned
				}
				schema.AnyOf[i] = nv
			}
		}
	}

	if len(schema.OneOf) > 0 {
		oneOfCloned := false
		for i := range schema.OneOf {
			if nv, c := wrapRefsValue(schema.OneOf[i]); c {
				if !oneOfCloned {
					oneOfCloned = true
					changed = true
					cloned := make([]spec.Schema, len(schema.OneOf))
					copy(cloned, schema.OneOf)
					schema.OneOf = cloned
				}
				schema.OneOf[i] = nv
			}
		}
	}

	if schema.Not != nil {
		if nv, c := wrapRefsValue(*schema.Not); c {
			changed = true
			schema.Not = &nv
		}
	}

	if schema.AdditionalProperties != nil && schema.AdditionalProperties.Schema != nil {
		if nv, c := wrapRefsValue(*schema.AdditionalProperties.Schema); c {
			changed = true
			schema.AdditionalProperties = &spec.SchemaOrBool{Schema: &nv, Allows: schema.AdditionalProperties.Allows}
		}
	}

	if schema.AdditionalItems != nil && schema.AdditionalItems.Schema != nil {
		if nv, c := wrapRefsValue(*schema.AdditionalItems.Schema); c {
			changed = true
			schema.AdditionalItems = &spec.SchemaOrBool{Schema: &nv, Allows: schema.AdditionalItems.Allows}
		}
	}

	if schema.Items != nil {
		if schema.Items.Schema != nil {
			if nv, c := wrapRefsValue(*schema.Items.Schema); c {
				changed = true
				schema.Items = &spec.SchemaOrArray{Schema: &nv}
			}
		} else if len(schema.Items.Schemas) > 0 {
			itemsCloned := false
			for i := range schema.Items.Schemas {
				if nv, c := wrapRefsValue(schema.Items.Schemas[i]); c {
					if !itemsCloned {
						itemsCloned = true
						changed = true
						cloned := make([]spec.Schema, len(schema.Items.Schemas))
						copy(cloned, schema.Items.Schemas)
						schema.Items = &spec.SchemaOrArray{Schemas: cloned}
					}
					schema.Items.Schemas[i] = nv
				}
			}
		}
	}

	return schema, changed
}

// WrapRefs wraps OpenAPI V3 Schema refs that contain sibling elements.
// AllOf is used to wrap the Ref to prevent references from having sibling elements
// Please see https://github.com/kubernetes/kubernetes/issues/106387#issuecomment-967640388
func WrapRefs(schema *spec.Schema) *spec.Schema {
	if schema == nil {
		return nil
	}
	if nv, changed := wrapRefsValue(*schema); changed {
		return &nv
	}
	return schema
}
