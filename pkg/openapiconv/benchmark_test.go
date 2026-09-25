/*
Copyright 2026 The Kubernetes Authors.

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
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"k8s.io/kube-openapi/pkg/validation/spec"
)

func BenchmarkConvertV2ToV3(b *testing.B) {
	for _, gv := range []string{"api.v1", "batch.v1", "apiextensions.k8s.io.v1"} {
		spec2JSON, err := os.ReadFile(filepath.Join("testdata_generated_from_k8s", "v2_"+gv+".json"))
		if err != nil {
			b.Fatal(err)
		}
		var swaggerSpec spec.Swagger
		if err := json.Unmarshal(spec2JSON, &swaggerSpec); err != nil {
			b.Fatal(err)
		}
		b.Run(gv, func(b *testing.B) {
			_ = ConvertV2ToV3(&swaggerSpec)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				v3 := ConvertV2ToV3(&swaggerSpec)
				if v3 == nil {
					b.Fatal("unexpected nil OpenAPI v3 spec")
				}
			}
		})
	}
}
