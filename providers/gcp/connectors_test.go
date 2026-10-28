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
	connectors "google.golang.org/api/connectors/v1"
	"google.golang.org/api/option"
)

func TestIntegrationConnectorsGeneratorEnumeration(t *testing.T) {
	ctx := context.Background()
	parent := "projects/test-project/locations/us-central1"
	globalParent := "projects/test-project/locations/global"
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: "/connections", body: `{"connections":[{"name":"` + parent + `/connections/conn1"}]}`},
		{substr: "/endpointAttachments", body: `{"endpointAttachments":[{"name":"` + parent + `/endpointAttachments/ea1"}]}`},
		{substr: "/managedZones", body: `{"managedZones":[{"name":"` + globalParent + `/managedZones/zone1"}]}`},
	}}
	svc, err := connectors.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	g := &IntegrationConnectorsGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project", "region": compute.Region{Name: "us-central1"}})
	if err := g.initResources(ctx, svc); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_integration_connectors_connection":          {parent + "/connections/conn1"},
		"google_integration_connectors_endpoint_attachment": {parent + "/endpointAttachments/ea1"},
		"google_integration_connectors_managed_zone":        {globalParent + "/managedZones/zone1"},
	})
}
