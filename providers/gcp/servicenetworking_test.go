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
	svcnetv1 "google.golang.org/api/servicenetworking/v1"
)

// TestServiceNetworkingGenerator feeds canned Service Networking
// responses through the generator and verifies the emitted types and
// composite IDs: the connection (network:service), the peered DNS
// domain under the connection's network, and the VPC Service
// Controls resource (enabled in the canned response, with the
// provider's /global-less ID form).
func TestServiceNetworkingGenerator(t *testing.T) {
	network := "projects/12345/global/networks/net-1"
	transport := &cannedTransport{routes: []cannedRoute{
		{"/peeredDnsDomains", `{"peeredDnsDomains":[{"name":"domain-1","dnsSuffix":"example.com."}]}`},
		{"/vpcServiceControls", `{"enabled":true}`},
		{"/connections", `{"connections":[{"network":"` + network + `","peering":"servicenetworking-googleapis-com"}]}`},
	}}
	ctx := context.Background()
	service, err := svcnetv1.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("servicenetworking.NewService: %v", err)
	}
	g := &ServiceNetworkingGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project"})
	if err := g.initResources(ctx, service); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_service_networking_connection":           {network + ":servicenetworking.googleapis.com"},
		"google_service_networking_peered_dns_domain":    {"services/servicenetworking.googleapis.com/" + network + "/peeredDnsDomains/domain-1"},
		"google_service_networking_vpc_service_controls": {"services/servicenetworking.googleapis.com/projects/12345/networks/net-1"},
	})
}
