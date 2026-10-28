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

	"google.golang.org/api/option"
	svcmgmtv1 "google.golang.org/api/servicemanagement/v1"
)

// TestEndpointsGenerator feeds canned Service Management responses
// through the generator and verifies the emitted types and IDs: the
// endpoints service produced by the project and its IAM trio. A
// second canned service produced by another project is filtered out.
func TestEndpointsGenerator(t *testing.T) {
	serviceName := "echo-api.endpoints.test-project.cloud.goog"
	transport := &cannedTransport{routes: []cannedRoute{
		{":getIamPolicy", `{"bindings":[{"role":"roles/servicemanagement.serviceController","members":["serviceAccount:sa@test-project.iam.gserviceaccount.com"]}]}`},
		{"/v1/services", `{"services":[{"serviceName":"` + serviceName + `","producerProjectId":"test-project"},{"serviceName":"other.googleapis.com","producerProjectId":"other-project"}]}`},
	}}
	ctx := context.Background()
	service, err := svcmgmtv1.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("servicemanagement.NewService: %v", err)
	}
	g := &EndpointsGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project"})
	if err := g.initResources(ctx, service); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_endpoints_service":             {serviceName},
		"google_endpoints_service_iam_policy":  {serviceName},
		"google_endpoints_service_iam_binding": {serviceName},
		"google_endpoints_service_iam_member":  {serviceName},
	})
}
