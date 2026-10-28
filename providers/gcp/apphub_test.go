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

	apphub "google.golang.org/api/apphub/v1"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

func TestAppHubGeneratorEnumeration(t *testing.T) {
	ctx := context.Background()
	parent := "projects/test-project/locations/us-central1"
	app := parent + "/applications/app1"
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: "/boundary", body: `{}`},
		{substr: "/serviceProjectAttachments", body: `{"serviceProjectAttachments":[{"name":"` + parent + `/serviceProjectAttachments/sp1"}]}`},
		{substr: "/services", body: `{"services":[{"name":"` + app + `/services/svc1"}]}`},
		{substr: "/workloads", body: `{"workloads":[{"name":"` + app + `/workloads/wl1"}]}`},
		{substr: "/applications", body: `{"applications":[{"name":"` + app + `"}]}`},
	}}
	svc, err := apphub.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	g := &AppHubGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project", "region": compute.Region{Name: "us-central1"}})
	if err := g.initResources(ctx, svc, transport.client()); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_apphub_boundary":                   {parent + "/boundary"},
		"google_apphub_service_project_attachment": {parent + "/serviceProjectAttachments/sp1"},
		"google_apphub_application":                {app},
		"google_apphub_service":                    {app + "/services/svc1"},
		"google_apphub_workload":                   {app + "/workloads/wl1"},
	})
}
