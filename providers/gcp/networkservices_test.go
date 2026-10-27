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
	"sort"
	"testing"

	"google.golang.org/api/compute/v1"
	netsvcv1 "google.golang.org/api/networkservices/v1"
	netsvcbeta "google.golang.org/api/networkservices/v1beta1"
	"google.golang.org/api/option"
)

// TestNetworkServicesGenerator feeds canned Network Services (v1 and
// v1beta1) responses through the real enumeration code and verifies the
// emitted resource types and IDs at the configured region and the global
// location.
func TestNetworkServicesGenerator(t *testing.T) {
	regional := "projects/test-project/locations/us-central1/"
	global := "projects/test-project/locations/global/"
	transport := &cannedTransport{routes: []cannedRoute{
		{"/locations/us-central1/authzExtensions", `{"authzExtensions":[{"name":"` + regional + `authzExtensions/authz-1"}]}`},
		{"/locations/us-central1/endpointPolicies", `{"endpointPolicies":[{"name":"` + regional + `endpointPolicies/ep-1"}]}`},
		{"/locations/us-central1/gateways", `{"gateways":[{"name":"` + regional + `gateways/gw-1"}]}`},
		{"/locations/us-central1/grpcRoutes", `{"grpcRoutes":[{"name":"` + regional + `grpcRoutes/grpc-1"}]}`},
		{"/locations/us-central1/httpRoutes", `{"httpRoutes":[{"name":"` + regional + `httpRoutes/http-1"}]}`},
		{"/locations/us-central1/lbRouteExtensions", `{"lbRouteExtensions":[{"name":"` + regional + `lbRouteExtensions/lbre-1"}]}`},
		{"/locations/us-central1/lbTrafficExtensions", `{"lbTrafficExtensions":[{"name":"` + regional + `lbTrafficExtensions/lbte-1"}]}`},
		{"/locations/us-central1/lbEdgeExtensions", `{"lbEdgeExtensions":[{"name":"` + regional + `lbEdgeExtensions/lbee-1"}]}`},
		{"/locations/us-central1/meshes", `{"meshes":[{"name":"` + regional + `meshes/mesh-1"}]}`},
		{"/locations/us-central1/serviceBindings", `{"serviceBindings":[{"name":"` + regional + `serviceBindings/sb-1"}]}`},
		{"/locations/us-central1/serviceLbPolicies", `{"serviceLbPolicies":[{"name":"` + regional + `serviceLbPolicies/slp-1"}]}`},
		{"/locations/us-central1/tcpRoutes", `{"tcpRoutes":[{"name":"` + regional + `tcpRoutes/tcp-1"}]}`},
		{"/locations/us-central1/tlsRoutes", `{"tlsRoutes":[{"name":"` + regional + `tlsRoutes/tls-1"}]}`},
		{"/locations/us-central1/wasmPlugins", `{"wasmPlugins":[{"name":"` + regional + `wasmPlugins/wp-1"}]}`},
		{"/locations/global/gateways", `{"gateways":[{"name":"` + global + `gateways/gw-global"}]}`},
		{"/locations/global/httpRoutes", `{"httpRoutes":[{"name":"` + global + `httpRoutes/http-global"}]}`},
		{"/locations/global/meshes", `{"meshes":[{"name":"` + global + `meshes/mesh-global"}]}`},
		{"/locations/global/lbEdgeExtensions", `{"lbEdgeExtensions":[{"name":"` + global + `lbEdgeExtensions/lbee-global"}]}`},
	}}

	ctx := context.Background()
	clientOpts := []option.ClientOption{
		option.WithHTTPClient(transport.client()),
		option.WithoutAuthentication(),
	}
	v1Service, err := netsvcv1.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("netsvcv1.NewService: %v", err)
	}
	betaService, err := netsvcbeta.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("netsvcbeta.NewService: %v", err)
	}

	g := &NetworkServicesGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	if err := g.initResources(ctx, v1Service, betaService); err != nil {
		t.Fatalf("initResources: %v", err)
	}

	expected := map[string][]string{
		"google_network_services_authz_extension":      {regional + "authzExtensions/authz-1"},
		"google_network_services_endpoint_policy":      {regional + "endpointPolicies/ep-1"},
		"google_network_services_gateway":              {regional + "gateways/gw-1", global + "gateways/gw-global"},
		"google_network_services_grpc_route":           {regional + "grpcRoutes/grpc-1"},
		"google_network_services_http_route":           {regional + "httpRoutes/http-1", global + "httpRoutes/http-global"},
		"google_network_services_lb_edge_extension":    {regional + "lbEdgeExtensions/lbee-1", global + "lbEdgeExtensions/lbee-global"},
		"google_network_services_lb_route_extension":   {regional + "lbRouteExtensions/lbre-1"},
		"google_network_services_lb_traffic_extension": {regional + "lbTrafficExtensions/lbte-1"},
		"google_network_services_mesh":                 {regional + "meshes/mesh-1", global + "meshes/mesh-global"},
		"google_network_services_service_binding":      {regional + "serviceBindings/sb-1"},
		"google_network_services_service_lb_policies":  {regional + "serviceLbPolicies/slp-1"},
		"google_network_services_tcp_route":            {regional + "tcpRoutes/tcp-1"},
		"google_network_services_tls_route":            {regional + "tlsRoutes/tls-1"},
		"google_network_services_wasm_plugin":          {regional + "wasmPlugins/wp-1"},
	}

	got := map[string]map[string]bool{}
	for _, resource := range g.Resources {
		if got[resource.InstanceInfo.Type] == nil {
			got[resource.InstanceInfo.Type] = map[string]bool{}
		}
		got[resource.InstanceInfo.Type][resource.InstanceState.ID] = true
	}
	if len(got) != len(expected) {
		t.Errorf("resource types = %d, want %d; got %v", len(got), len(expected), got)
	}
	for resourceType, ids := range expected {
		gotIDs := []string{}
		for id := range got[resourceType] {
			gotIDs = append(gotIDs, id)
		}
		sort.Strings(gotIDs)
		wantIDs := append([]string{}, ids...)
		sort.Strings(wantIDs)
		if len(gotIDs) != len(wantIDs) {
			t.Errorf("%s IDs = %v, want %v", resourceType, gotIDs, wantIDs)
			continue
		}
		for i := range wantIDs {
			if gotIDs[i] != wantIDs[i] {
				t.Errorf("%s IDs = %v, want %v", resourceType, gotIDs, wantIDs)
				break
			}
		}
	}
}
