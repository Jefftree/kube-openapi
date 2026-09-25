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

package openapiconv

import (
	"strings"

	klog "k8s.io/klog/v2"
	builderutil "k8s.io/kube-openapi/pkg/builder3/util"
	"k8s.io/kube-openapi/pkg/spec3"
	"k8s.io/kube-openapi/pkg/validation/spec"
)

var OpenAPIV2DefPrefix = "#/definitions/"
var OpenAPIV3DefPrefix = "#/components/schemas/"

// ConvertV2ToV3 converts an OpenAPI V2 object into V3.
// Certain references may be shared between the V2 and V3 objects in the conversion.
func ConvertV2ToV3(v2Spec *spec.Swagger) *spec3.OpenAPI {
	v3Spec := &spec3.OpenAPI{
		Version:      "3.0.0",
		Info:         v2Spec.Info,
		ExternalDocs: ConvertExternalDocumentation(v2Spec.ExternalDocs),
		Paths:        ConvertPaths(v2Spec.Paths),
		Components:   ConvertComponents(v2Spec.SecurityDefinitions, v2Spec.Definitions, v2Spec.Responses, v2Spec.Produces),
	}

	return v3Spec
}

func ConvertExternalDocumentation(v2ED *spec.ExternalDocumentation) *spec3.ExternalDocumentation {
	if v2ED == nil {
		return nil
	}
	return &spec3.ExternalDocumentation{
		ExternalDocumentationProps: spec3.ExternalDocumentationProps{
			Description: v2ED.Description,
			URL:         v2ED.URL,
		},
	}
}

func ConvertComponents(v2SecurityDefinitions spec.SecurityDefinitions, v2Definitions spec.Definitions, v2Responses map[string]spec.Response, produces []string) *spec3.Components {
	components := &spec3.Components{}

	if v2Definitions != nil {
		components.Schemas = make(map[string]*spec.Schema, len(v2Definitions))
	}
	for s, schema := range v2Definitions {
		components.Schemas[s] = ConvertSchema(&schema)
	}
	if v2SecurityDefinitions != nil {
		components.SecuritySchemes = make(spec3.SecuritySchemes, len(v2SecurityDefinitions))
	}
	for s, securityScheme := range v2SecurityDefinitions {
		components.SecuritySchemes[s] = ConvertSecurityScheme(securityScheme)
	}
	if v2Responses != nil {
		components.Responses = make(map[string]*spec3.Response, len(v2Responses))
	}
	for r, response := range v2Responses {
		components.Responses[r] = ConvertResponse(&response, produces)
	}

	return components
}

func ConvertSchema(v2Schema *spec.Schema) *spec.Schema {
	if v2Schema == nil {
		return nil
	}
	v3Schema := convertSchemaValue(*v2Schema)
	return builderutil.WrapRefs(&v3Schema)
}

func convertSchemaValue(v2Schema spec.Schema) spec.Schema {
	v3Schema := spec.Schema{
		VendorExtensible:   v2Schema.VendorExtensible,
		SchemaProps:        v2Schema.SchemaProps,
		SwaggerSchemaProps: v2Schema.SwaggerSchemaProps,
		ExtraProps:         v2Schema.ExtraProps,
	}

	if refString := v2Schema.Ref.String(); refString != "" {
		if idx := strings.Index(refString, OpenAPIV2DefPrefix); idx != -1 {
			v3Schema.Ref = spec.MustCreateRef(OpenAPIV3DefPrefix + refString[idx+len(OpenAPIV2DefPrefix):])
		} else {
			klog.Errorf("Error: Swagger V2 Ref %s does not contain #/definitions\n", refString)
		}
	}

	if v2Schema.Properties != nil {
		v3Schema.Properties = make(map[string]spec.Schema, len(v2Schema.Properties))
		for key, property := range v2Schema.Properties {
			v3Schema.Properties[key] = convertSchemaValue(property)
		}
	}
	if v2Schema.Items != nil {
		var itemSchema *spec.Schema
		if v2Schema.Items.Schema != nil {
			s := convertSchemaValue(*v2Schema.Items.Schema)
			itemSchema = &s
		}
		var itemSchemas []spec.Schema
		if v2Schema.Items.Schemas != nil {
			itemSchemas = make([]spec.Schema, len(v2Schema.Items.Schemas))
			for i, s := range v2Schema.Items.Schemas {
				itemSchemas[i] = convertSchemaValue(s)
			}
		}
		v3Schema.Items = &spec.SchemaOrArray{
			Schema:  itemSchema,
			Schemas: itemSchemas,
		}
	}

	if v2Schema.AdditionalProperties != nil {
		var addPropSchema *spec.Schema
		if v2Schema.AdditionalProperties.Schema != nil {
			s := convertSchemaValue(*v2Schema.AdditionalProperties.Schema)
			addPropSchema = &s
		}
		v3Schema.AdditionalProperties = &spec.SchemaOrBool{
			Schema: addPropSchema,
			Allows: v2Schema.AdditionalProperties.Allows,
		}
	}
	if v2Schema.AdditionalItems != nil {
		var addItemSchema *spec.Schema
		if v2Schema.AdditionalItems.Schema != nil {
			s := convertSchemaValue(*v2Schema.AdditionalItems.Schema)
			addItemSchema = &s
		}
		v3Schema.AdditionalItems = &spec.SchemaOrBool{
			Schema: addItemSchema,
			Allows: v2Schema.AdditionalItems.Allows,
		}
	}

	return v3Schema
}

func ConvertSchemaList(v2SchemaList []spec.Schema) []spec.Schema {
	if v2SchemaList == nil {
		return nil
	}
	v3SchemaList := make([]spec.Schema, len(v2SchemaList))
	for i, s := range v2SchemaList {
		v3SchemaList[i] = *ConvertSchema(&s)
	}
	return v3SchemaList
}

func ConvertSecurityScheme(v2securityScheme *spec.SecurityScheme) *spec3.SecurityScheme {
	if v2securityScheme == nil {
		return nil
	}
	securityScheme := &spec3.SecurityScheme{
		VendorExtensible: v2securityScheme.VendorExtensible,
		SecuritySchemeProps: spec3.SecuritySchemeProps{
			Description: v2securityScheme.Description,
			Type:        v2securityScheme.Type,
			Name:        v2securityScheme.Name,
			In:          v2securityScheme.In,
		},
	}

	if v2securityScheme.Flow != "" {
		securityScheme.Flows = map[string]*spec3.OAuthFlow{
			v2securityScheme.Flow: {
				OAuthFlowProps: spec3.OAuthFlowProps{
					AuthorizationUrl: v2securityScheme.AuthorizationURL,
					TokenUrl:         v2securityScheme.TokenURL,
					Scopes:           v2securityScheme.Scopes,
				},
			},
		}
	}
	return securityScheme
}

func ConvertPaths(v2Paths *spec.Paths) *spec3.Paths {
	if v2Paths == nil {
		return nil
	}
	paths := &spec3.Paths{
		VendorExtensible: v2Paths.VendorExtensible,
	}

	if v2Paths.Paths != nil {
		paths.Paths = make(map[string]*spec3.Path, len(v2Paths.Paths))
	}
	for k, v := range v2Paths.Paths {
		paths.Paths[k] = ConvertPathItem(v)
	}
	return paths
}

func ConvertPathItem(v2pathItem spec.PathItem) *spec3.Path {
	var params []*spec3.Parameter
	if len(v2pathItem.Parameters) > 0 {
		params = make([]*spec3.Parameter, len(v2pathItem.Parameters))
		for i, param := range v2pathItem.Parameters {
			params[i] = ConvertParameter(param)
		}
	}
	return &spec3.Path{
		Refable: v2pathItem.Refable,
		PathProps: spec3.PathProps{
			Get:        ConvertOperation(v2pathItem.Get),
			Put:        ConvertOperation(v2pathItem.Put),
			Post:       ConvertOperation(v2pathItem.Post),
			Delete:     ConvertOperation(v2pathItem.Delete),
			Options:    ConvertOperation(v2pathItem.Options),
			Head:       ConvertOperation(v2pathItem.Head),
			Patch:      ConvertOperation(v2pathItem.Patch),
			Parameters: params,
		},
		VendorExtensible: v2pathItem.VendorExtensible,
	}
}

func ConvertOperation(v2Operation *spec.Operation) *spec3.Operation {
	if v2Operation == nil {
		return nil
	}
	operation := &spec3.Operation{
		VendorExtensible: v2Operation.VendorExtensible,
		OperationProps: spec3.OperationProps{
			Description:  v2Operation.Description,
			ExternalDocs: ConvertExternalDocumentation(v2Operation.OperationProps.ExternalDocs),
			Tags:         v2Operation.Tags,
			Summary:      v2Operation.Summary,
			Deprecated:   v2Operation.Deprecated,
			OperationId:  v2Operation.ID,
		},
	}

	for _, param := range v2Operation.Parameters {
		if param.ParamProps.Name == "body" && param.ParamProps.Schema != nil {
			operation.OperationProps.RequestBody = &spec3.RequestBody{
				RequestBodyProps: spec3.RequestBodyProps{},
			}
			if v2Operation.Consumes != nil {
				operation.RequestBody.Content = make(map[string]*spec3.MediaType, len(v2Operation.Consumes))
				convertedSchema := ConvertSchema(param.ParamProps.Schema)
				for i, consumer := range v2Operation.Consumes {
					s := convertedSchema
					if i > 0 && convertedSchema != nil {
						schemaCopy := *convertedSchema
						s = &schemaCopy
					}
					operation.RequestBody.Content[consumer] = &spec3.MediaType{
						MediaTypeProps: spec3.MediaTypeProps{
							Schema: s,
						},
					}
				}
			}
		} else {
			if operation.Parameters == nil {
				operation.Parameters = make([]*spec3.Parameter, 0, len(v2Operation.Parameters))
			}
			operation.Parameters = append(operation.Parameters, ConvertParameter(param))
		}
	}

	operation.Responses = &spec3.Responses{ResponsesProps: spec3.ResponsesProps{
		Default: ConvertResponse(v2Operation.Responses.Default, v2Operation.Produces),
	},
		VendorExtensible: v2Operation.Responses.VendorExtensible,
	}

	if v2Operation.Responses.StatusCodeResponses != nil {
		operation.Responses.StatusCodeResponses = make(map[int]*spec3.Response, len(v2Operation.Responses.StatusCodeResponses))
	}
	for k, v := range v2Operation.Responses.StatusCodeResponses {
		operation.Responses.StatusCodeResponses[k] = ConvertResponse(&v, v2Operation.Produces)
	}
	return operation
}

func ConvertResponse(v2Response *spec.Response, produces []string) *spec3.Response {
	if v2Response == nil {
		return nil
	}
	response := &spec3.Response{
		Refable:          ConvertRefableResponse(v2Response.Refable),
		VendorExtensible: v2Response.VendorExtensible,
		ResponseProps: spec3.ResponseProps{
			Description: v2Response.Description,
		},
	}

	if v2Response.Schema != nil {
		if produces != nil {
			response.Content = make(map[string]*spec3.MediaType, len(produces))
			convertedSchema := ConvertSchema(v2Response.Schema)
			for i, producer := range produces {
				s := convertedSchema
				if i > 0 && convertedSchema != nil {
					schemaCopy := *convertedSchema
					s = &schemaCopy
				}
				response.ResponseProps.Content[producer] = &spec3.MediaType{
					MediaTypeProps: spec3.MediaTypeProps{
						Schema: s,
					},
				}
			}
		}
	}
	return response
}

func ConvertParameter(v2Param spec.Parameter) *spec3.Parameter {
	param := &spec3.Parameter{
		Refable:          ConvertRefableParameter(v2Param.Refable),
		VendorExtensible: v2Param.VendorExtensible,
		ParameterProps: spec3.ParameterProps{
			Name:            v2Param.Name,
			Description:     v2Param.Description,
			In:              v2Param.In,
			Required:        v2Param.Required,
			Schema:          ConvertSchema(v2Param.Schema),
			AllowEmptyValue: v2Param.AllowEmptyValue,
		},
	}
	// Convert SimpleSchema into Schema
	if param.Schema == nil {
		param.Schema = &spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type:        []string{v2Param.Type},
				Format:      v2Param.Format,
				UniqueItems: v2Param.UniqueItems,
			},
		}
	}

	return param
}

func ConvertRefableParameter(refable spec.Refable) spec.Refable {
	if refStr := refable.Ref.String(); refStr != "" {
		return spec.Refable{Ref: spec.MustCreateRef(strings.Replace(refStr, "#/parameters/", "#/components/parameters/", 1))}
	}
	return refable
}

func ConvertRefableResponse(refable spec.Refable) spec.Refable {
	if refStr := refable.Ref.String(); refStr != "" {
		return spec.Refable{Ref: spec.MustCreateRef(strings.Replace(refStr, "#/responses/", "#/components/responses/", 1))}
	}
	return refable
}
