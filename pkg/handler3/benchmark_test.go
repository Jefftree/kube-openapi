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

package handler3

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"k8s.io/kube-openapi/pkg/spec3"
)

type benchResponseWriter struct {
	header http.Header
	code   int
	bytes  int
}

func newBenchResponseWriter() *benchResponseWriter {
	return &benchResponseWriter{header: make(http.Header, 8)}
}

func (w *benchResponseWriter) reset() {
	clear(w.header)
	w.code = http.StatusOK
	w.bytes = 0
}

func (w *benchResponseWriter) Header() http.Header {
	return w.header
}

func (w *benchResponseWriter) Write(b []byte) (int, error) {
	w.bytes += len(b)
	return len(b), nil
}

func (w *benchResponseWriter) WriteHeader(statusCode int) {
	w.code = statusCode
}

func loadAppsV1Spec(b *testing.B) *spec3.OpenAPI {
	b.Helper()
	data, err := os.ReadFile("../spec3/testdata/appsv1spec.json")
	if err != nil {
		b.Fatal(err)
	}
	var s spec3.OpenAPI
	if err := json.Unmarshal(data, &s); err != nil {
		b.Fatal(err)
	}
	return &s
}

func BenchmarkHandler3ColdServeJSON(b *testing.B) {
	spec := loadAppsV1Spec(b)
	svc := NewOpenAPIService()
	req := httptest.NewRequest(http.MethodGet, "/openapi/v3/apis/apps/v1", nil)
	req.Header.Set("Accept", "application/json")
	rw := newBenchResponseWriter()
	svc.UpdateGroupVersion("apis/apps/v1", spec)
	svc.HandleGroupVersion(rw, req)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.UpdateGroupVersion("apis/apps/v1", spec)
		rw.reset()
		svc.HandleGroupVersion(rw, req)
		if rw.code != http.StatusOK || rw.bytes == 0 {
			b.Fatalf("expected 200 and non-empty body, got %d (%d bytes)", rw.code, rw.bytes)
		}
	}
}

func BenchmarkHandler3WarmServeJSON(b *testing.B) {
	spec := loadAppsV1Spec(b)
	svc := NewOpenAPIService()
	svc.UpdateGroupVersion("apis/apps/v1", spec)
	req := httptest.NewRequest(http.MethodGet, "/openapi/v3/apis/apps/v1", nil)
	req.Header.Set("Accept", "application/json")
	rw := newBenchResponseWriter()
	svc.HandleGroupVersion(rw, req)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rw.reset()
		svc.HandleGroupVersion(rw, req)
		if rw.code != http.StatusOK || rw.bytes == 0 {
			b.Fatalf("expected 200 and non-empty body, got %d (%d bytes)", rw.code, rw.bytes)
		}
	}
}

func BenchmarkHandler3WarmDiscovery(b *testing.B) {
	spec := loadAppsV1Spec(b)
	svc := NewOpenAPIService()
	gvs := []string{
		"api/v1", "apis/apps/v1", "apis/batch/v1", "apis/autoscaling/v2",
		"apis/networking.k8s.io/v1", "apis/rbac.authorization.k8s.io/v1",
		"apis/storage.k8s.io/v1", "apis/policy/v1", "apis/certificates.k8s.io/v1",
		"apis/coordination.k8s.io/v1", "apis/discovery.k8s.io/v1", "apis/events.k8s.io/v1",
	}
	for _, gv := range gvs {
		svc.UpdateGroupVersion(gv, spec)
	}
	req := httptest.NewRequest(http.MethodGet, "/openapi/v3", nil)
	req.Header.Set("Accept", "application/json")
	rw := newBenchResponseWriter()
	for _, gv := range gvs {
		r := httptest.NewRequest(http.MethodGet, "/openapi/v3/"+gv, nil)
		r.Header.Set("Accept", "application/json")
		rw.reset()
		svc.HandleGroupVersion(rw, r)
	}
	rw.reset()
	svc.HandleDiscovery(rw, req)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rw.reset()
		svc.HandleDiscovery(rw, req)
		if rw.code != http.StatusOK || rw.bytes == 0 {
			b.Fatalf("expected 200 and non-empty body, got %d (%d bytes)", rw.code, rw.bytes)
		}
	}
}
