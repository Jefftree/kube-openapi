// Package v1 verifies that the presence of apiversion.yaml (with a
// sibling parent apigroup.yaml supplying spec.modelPackage) enables
// openapi generation and drives the generated OpenAPIModelName methods.
// There is intentionally no +k8s:openapi-gen or +k8s:openapi-model-package
// tag on this package.
package v1
