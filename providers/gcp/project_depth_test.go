// Copyright 2026 The Terraformer Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gcp

import (
	"context"
	"testing"

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
	serviceusagev1 "google.golang.org/api/serviceusage/v1"
)

// TestProjectDepthEnumeration asserts enabled services are emitted
// with the provider's {project}/{service} ID and the usage export
// bucket singleton with the project ID.
func TestProjectDepthEnumeration(t *testing.T) {
	ctx := context.Background()
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: "/services", body: `{"services":[{"name":"projects/test-project/services/compute.googleapis.com","config":{"name":"compute.googleapis.com"}}]}`},
		{substr: "/projects/test-project", body: `{"name":"test-project","usageExportLocation":{"bucketName":"usage-bucket","reportNamePrefix":"report"}}`},
	}}
	opts := []option.ClientOption{option.WithHTTPClient(transport.client()), option.WithoutAuthentication()}
	suSvc, err := serviceusagev1.NewService(ctx, opts...)
	if err != nil {
		t.Fatal(err)
	}
	computeSvc, err := compute.NewService(ctx, opts...)
	if err != nil {
		t.Fatal(err)
	}
	g := &ProjectGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project"})
	g.initProjectDepthResources(ctx, suSvc, computeSvc)
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_project_service":             {"test-project/compute.googleapis.com"},
		"google_project_usage_export_bucket": {"test-project"},
	})
}
