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

package azure

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cdn/armcdn"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/trafficmanager/armtrafficmanager"
)

const mockCDNProfilesJSON = `{"value":[
	{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.Cdn/profiles/afd-mock-1","name":"afd-mock-1","location":"global","type":"Microsoft.Cdn/profiles","sku":{"name":"Premium_AzureFrontDoor"},"properties":{}},
	{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.Cdn/profiles/cdn-mock-1","name":"cdn-mock-1","location":"global","type":"Microsoft.Cdn/profiles","sku":{"name":"Standard_Microsoft"},"properties":{}}
]}`

func TestMockTrack2CDNProfilesFrontDoorSplit(t *testing.T) {
	tr := &cannedTransport{body: mockCDNProfilesJSON}
	client, err := armcdn.NewProfilesClient(mockSubscriptionID, stubCredential{}, mockClientOptions(tr))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	pager := client.NewListPager(nil)
	ctx := context.Background()
	afd, cdn := 0, 0
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			t.Fatalf("next page: %v", err)
		}
		for _, profile := range page.Value {
			if isFrontDoorProfile(profile) {
				afd++
			} else {
				cdn++
			}
		}
	}
	if afd != 1 || cdn != 1 {
		t.Fatalf("got %d AFD / %d CDN profiles, want 1/1", afd, cdn)
	}
	if len(tr.urls) == 0 || !strings.Contains(tr.urls[0], "Microsoft.Cdn") {
		t.Errorf("expected a Microsoft.Cdn call, got %v", tr.urls)
	}
}

const mockTMProfileJSON = `{
	"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.Network/trafficManagerProfiles/tm-mock-1",
	"name":"tm-mock-1","location":"global","type":"Microsoft.Network/trafficManagerProfiles",
	"properties":{"profileStatus":"Enabled","endpoints":[
		{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.Network/trafficManagerProfiles/tm-mock-1/AzureEndpoints/ep-azure","name":"ep-azure","type":"Microsoft.Network/trafficManagerProfiles/AzureEndpoints","properties":{"endpointStatus":"Enabled"}},
		{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.Network/trafficManagerProfiles/tm-mock-1/ExternalEndpoints/ep-ext","name":"ep-ext","type":"Microsoft.Network/trafficManagerProfiles/ExternalEndpoints","properties":{"endpointStatus":"Enabled"}},
		{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.Network/trafficManagerProfiles/tm-mock-1/NestedEndpoints/ep-nested","name":"ep-nested","type":"Microsoft.Network/trafficManagerProfiles/NestedEndpoints","properties":{"endpointStatus":"Enabled"}}
	]}
}`

func TestTrafficManagerEndpointResources(t *testing.T) {
	var profile armtrafficmanager.Profile
	if err := json.Unmarshal([]byte(mockTMProfileJSON), &profile); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	resources := trafficManagerEndpointResources(&profile)
	if len(resources) != 3 {
		t.Fatalf("got %d endpoint resources, want 3", len(resources))
	}
	want := []string{
		"azurerm_traffic_manager_azure_endpoint",
		"azurerm_traffic_manager_external_endpoint",
		"azurerm_traffic_manager_nested_endpoint",
	}
	for i, w := range want {
		if resources[i].InstanceInfo.Type != w {
			t.Errorf("resource %d type = %q, want %q", i, resources[i].InstanceInfo.Type, w)
		}
	}
}

const mockLBJSON = `{
	"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.Network/loadBalancers/lb-mock-1",
	"name":"lb-mock-1","location":"eastus","type":"Microsoft.Network/loadBalancers",
	"properties":{"inboundNatPools":[
		{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.Network/loadBalancers/lb-mock-1/inboundNatPools/pool-mock-1","name":"pool-mock-1","properties":{}}
	]}
}`

const mockBackendPoolJSON = `{
	"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.Network/loadBalancers/lb-mock-1/backendAddressPools/bap-mock-1",
	"name":"bap-mock-1",
	"properties":{"loadBalancerBackendAddresses":[
		{"name":"addr-mock-1","properties":{"ipAddress":"10.0.0.4"}}
	]}
}`

func TestLoadBalancerEmbeddedChildren(t *testing.T) {
	var lb armnetwork.LoadBalancer
	if err := json.Unmarshal([]byte(mockLBJSON), &lb); err != nil {
		t.Fatalf("unmarshal lb: %v", err)
	}
	pools := loadBalancerNatPoolResources(&lb)
	if len(pools) != 1 || pools[0].InstanceInfo.Type != "azurerm_lb_nat_pool" {
		t.Fatalf("nat pools = %+v, want 1 azurerm_lb_nat_pool", pools)
	}
	var pool armnetwork.BackendAddressPool
	if err := json.Unmarshal([]byte(mockBackendPoolJSON), &pool); err != nil {
		t.Fatalf("unmarshal pool: %v", err)
	}
	addresses := backendAddressPoolAddressResources(&pool)
	if len(addresses) != 1 {
		t.Fatalf("got %d pool addresses, want 1", len(addresses))
	}
	wantID := *pool.ID + "/addresses/addr-mock-1"
	if addresses[0].InstanceState.ID != wantID {
		t.Errorf("address id = %q, want %q", addresses[0].InstanceState.ID, wantID)
	}
	if addresses[0].InstanceInfo.Type != "azurerm_lb_backend_address_pool_address" {
		t.Errorf("address type = %q", addresses[0].InstanceInfo.Type)
	}
}

const mockIPGroupJSON = `{
	"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.Network/ipGroups/ipg-mock-1",
	"name":"ipg-mock-1","location":"eastus","type":"Microsoft.Network/ipGroups",
	"properties":{"ipAddresses":["10.0.0.0/24","10.1.0.0/24"]}
}`

func TestIPGroupCIDRResources(t *testing.T) {
	var group armnetwork.IPGroup
	if err := json.Unmarshal([]byte(mockIPGroupJSON), &group); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	g := &FirewallGenerator{}
	g.appendIPGroupCIDRs(&group)
	if len(g.Resources) != 2 {
		t.Fatalf("got %d cidr resources, want 2", len(g.Resources))
	}
	wantID := *group.ID + "/cidrs/10.0.0.0_24"
	if g.Resources[0].InstanceState.ID != wantID {
		t.Errorf("cidr id = %q, want %q", g.Resources[0].InstanceState.ID, wantID)
	}
	if g.Resources[0].InstanceInfo.Type != "azurerm_ip_group_cidr" {
		t.Errorf("cidr type = %q", g.Resources[0].InstanceInfo.Type)
	}
}

const mockManagersJSON = `{"value":[
	{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.Network/networkManagers/nm-mock-1","name":"nm-mock-1","location":"eastus","type":"Microsoft.Network/networkManagers","properties":{"networkManagerScopes":{"subscriptions":["/subscriptions/00000000-0000-0000-0000-000000000000"]},"networkManagerScopeAccesses":["Connectivity"]}}
]}`

func TestMockTrack2NetworkManagers(t *testing.T) {
	tr := &cannedTransport{body: mockManagersJSON}
	client, err := armnetwork.NewManagersClient(mockSubscriptionID, stubCredential{}, mockClientOptions(tr))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	pager := client.NewListBySubscriptionPager(nil)
	ctx := context.Background()
	count := 0
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			t.Fatalf("next page: %v", err)
		}
		for _, manager := range page.Value {
			if p, ok := makeParent(manager.ID, manager.Name); ok && p.ResourceGroup == "rg-mock-1" {
				count++
			}
		}
	}
	if count != 1 {
		t.Fatalf("got %d managers, want 1", count)
	}
	if len(tr.urls) == 0 || !strings.Contains(tr.urls[0], "networkManagers") {
		t.Errorf("expected a networkManagers call, got %v", tr.urls)
	}
}
