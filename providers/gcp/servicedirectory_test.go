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
	svcdirv1 "google.golang.org/api/servicedirectory/v1"
)

// TestServiceDirectoryGenerator feeds canned Service Directory
// responses through the generator and verifies the emitted types and
// IDs: the namespace and its IAM trio, the service inside it with its
// IAM trio, and the endpoint inside the service.
func TestServiceDirectoryGenerator(t *testing.T) {
	namespace := "projects/test-project/locations/us-central1/namespaces/ns-1"
	sdService := namespace + "/services/svc-1"
	transport := &cannedTransport{routes: []cannedRoute{
		{"/services/svc-1/endpoints", `{"endpoints":[{"name":"` + sdService + `/endpoints/ep-1"}]}`},
		{":getIamPolicy", `{"bindings":[{"role":"roles/servicedirectory.viewer","members":["user:a@example.com"]}]}`},
		{"/namespaces/ns-1/services", `{"services":[{"name":"` + sdService + `"}]}`},
		{"/locations/us-central1/namespaces", `{"namespaces":[{"name":"` + namespace + `"}]}`},
		{"/v1/projects/test-project/locations", `{"locations":[{"name":"projects/test-project/locations/us-central1"}]}`},
	}}
	ctx := context.Background()
	service, err := svcdirv1.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("servicedirectory.NewService: %v", err)
	}
	g := &ServiceDirectoryGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project"})
	if err := g.initResources(ctx, service); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_service_directory_namespace":             {namespace},
		"google_service_directory_namespace_iam_policy":  {namespace},
		"google_service_directory_namespace_iam_binding": {namespace},
		"google_service_directory_namespace_iam_member":  {namespace},
		"google_service_directory_service":               {sdService},
		"google_service_directory_service_iam_policy":    {sdService},
		"google_service_directory_service_iam_binding":   {sdService},
		"google_service_directory_service_iam_member":    {sdService},
		"google_service_directory_endpoint":              {sdService + "/endpoints/ep-1"},
	})
}
