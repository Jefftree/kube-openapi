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

package spec3

import (
	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"

	"k8s.io/kube-openapi/pkg/internal"
	"k8s.io/kube-openapi/pkg/validation/spec"
)

// RequestBody describes a single request body, more at https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.0.md#requestBodyObject
//
// Note that this struct is actually a thin wrapper around RequestBodyProps to make it referable and extensible
type RequestBody struct {
	spec.Refable
	RequestBodyProps
	spec.VendorExtensible
}

// MarshalJSON is a custom marshal function that knows how to encode RequestBody as JSON
func (r *RequestBody) MarshalJSON() ([]byte, error) {
	return internal.DeterministicMarshal(r)
}

func (r *RequestBody) MarshalJSONTo(enc *jsontext.Encoder) error {
	var x struct {
		Ref              string                   `json:"$ref,omitempty"`
		RequestBodyProps requestBodyPropsOmitZero `json:",embed"`
		Extensions       spec.Extensions          `json:",embed"`
	}
	x.Ref = r.Refable.Ref.String()
	x.Extensions = internal.SanitizeExtensions(r.Extensions)
	x.RequestBodyProps = requestBodyPropsOmitZero(r.RequestBodyProps)
	return jsonv2.MarshalEncode(enc, &x)
}

func (r *RequestBody) UnmarshalJSON(data []byte) error {
	return jsonv2.Unmarshal(data, r)
}

// RequestBodyProps describes a single request body, more at https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.0.md#requestBodyObject
type RequestBodyProps struct {
	// Description holds a brief description of the request body
	Description string `json:"description,omitempty"`
	// Content is the content of the request body. The key is a media type or media type range and the value describes it
	Content map[string]*MediaType `json:"content,omitempty"`
	// Required determines if the request body is required in the request
	Required bool `json:"required,omitempty"`
}

type requestBodyPropsOmitZero struct {
	Description string                `json:"description,omitempty"`
	Content     map[string]*MediaType `json:"content,omitempty"`
	Required    bool                  `json:"required,omitzero"`
}

func (r *RequestBody) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var x struct {
		Ref        internal.StringOrAny `json:"$ref,omitempty"`
		Extensions spec.Extensions      `json:",embed"`
		RequestBodyProps
	}
	if err := jsonv2.UnmarshalDecode(dec, &x); err != nil {
		return err
	}
	if x.Ref.Set {
		if x.Ref.IsStr {
			ref, err := spec.NewRef(x.Ref.Str)
			if err != nil {
				return err
			}
			r.Ref = ref
		} else {
			if x.Extensions == nil {
				x.Extensions = make(spec.Extensions, 1)
			}
			x.Extensions["$ref"] = x.Ref.Other
		}
	}
	r.Extensions = internal.SanitizeExtensions(x.Extensions)
	r.RequestBodyProps = x.RequestBodyProps
	return nil
}
