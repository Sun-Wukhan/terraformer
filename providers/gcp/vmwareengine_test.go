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
	vmwareenginev1 "google.golang.org/api/vmwareengine/v1"
)

// TestVmwareEngineGenerator feeds canned VMware Engine responses
// (generated client plus raw REST datastores) and verifies all nine
// resource types.
func TestVmwareEngineGenerator(t *testing.T) {
	loc := "projects/test-project/locations/us-central1"
	pc := loc + "/privateClouds/pc1"
	np := loc + "/networkPolicies/np1"
	transport := &cannedTransport{routes: []cannedRoute{
		{"/privateClouds/pc1/clusters", `{"clusters":[{"name":"` + pc + `/clusters/c1"}]}`},
		{"/privateClouds/pc1/subnets", `{"subnets":[{"name":"` + pc + `/subnets/s1"}]}`},
		{"/privateClouds/pc1/externalAddresses", `{"externalAddresses":[{"name":"` + pc + `/externalAddresses/e1"}]}`},
		{"/networkPolicies/np1/externalAccessRules", `{"externalAccessRules":[{"name":"` + np + `/externalAccessRules/r1"}]}`},
		{"/vmwareEngineNetworks", `{"vmwareEngineNetworks":[{"name":"` + loc + `/vmwareEngineNetworks/net1"}]}`},
		{"/networkPeerings", `{"networkPeerings":[{"name":"` + loc + `/networkPeerings/p1"}]}`},
		{"/networkPolicies", `{"networkPolicies":[{"name":"` + np + `"}]}`},
		{"/privateClouds", `{"privateClouds":[{"name":"` + pc + `"}]}`},
		{"/datastores", `{"datastores":[{"name":"` + loc + `/datastores/ds1"}]}`},
	}}
	ctx := context.Background()
	clientOpts := []option.ClientOption{option.WithHTTPClient(transport.client()), option.WithoutAuthentication()}
	engineService, err := vmwareenginev1.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("vmwareengine.NewService: %v", err)
	}
	g := &VmwareEngineGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project", "region": compute.Region{Name: "us-central1"}})
	if err := g.initResources(ctx, engineService, transport.client()); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_vmwareengine_network":              {loc + "/vmwareEngineNetworks/net1"},
		"google_vmwareengine_network_peering":      {loc + "/networkPeerings/p1"},
		"google_vmwareengine_network_policy":       {np},
		"google_vmwareengine_external_access_rule": {np + "/externalAccessRules/r1"},
		"google_vmwareengine_private_cloud":        {pc},
		"google_vmwareengine_cluster":              {pc + "/clusters/c1"},
		"google_vmwareengine_subnet":               {pc + "/subnets/s1"},
		"google_vmwareengine_external_address":     {pc + "/externalAddresses/e1"},
		"google_vmwareengine_datastore":            {loc + "/datastores/ds1"},
	})
}
