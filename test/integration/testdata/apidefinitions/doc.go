// Package apidefinitions verifies that the presence of apiversion.yaml
// enables openapi generation, and that spec.modelPackage drives the
// generated OpenAPIModelName methods. There is intentionally no
// +k8s:openapi-gen or +k8s:openapi-model-package tag on this package.
package apidefinitions
