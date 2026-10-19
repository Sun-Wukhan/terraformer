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
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/analysisservices/armanalysisservices"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/databricks/armdatabricks"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dns/armdns"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// These tests prove the track-2 migration works with NO Azure credentials and
// NO network: a stub azcore.TokenCredential plus a mock HTTP transport serving
// canned ARM JSON exercise the real SDK client + pager + resource-mapping code.
// They run in CI on every push/PR (azure-unit job).

// stubCredential is a fake azcore.TokenCredential: no network, no secrets.
type stubCredential struct{}

func (stubCredential) GetToken(context.Context, policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{Token: "fake-token", ExpiresOn: time.Now().Add(time.Hour)}, nil
}

// cannedTransport serves body for every request and records the URLs seen,
// so tests can assert the SDK actually issued the expected ARM calls.
type cannedTransport struct {
	body string
	urls []string
}

func (c *cannedTransport) Do(r *http.Request) (*http.Response, error) {
	c.urls = append(c.urls, r.URL.String())
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(c.body)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Request:    r,
	}, nil
}

func mockClientOptions(tr *cannedTransport) *arm.ClientOptions {
	return &arm.ClientOptions{
		ClientOptions: azcore.ClientOptions{Transport: tr},
	}
}

const mockSubscriptionID = "00000000-0000-0000-0000-000000000000"

const mockResourceGroupsJSON = `{"value":[
	{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1","name":"rg-mock-1","location":"eastus","type":"Microsoft.Resources/resourceGroups"},
	{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-2","name":"rg-mock-2","location":"westus","type":"Microsoft.Resources/resourceGroups"}
]}`

const mockVnetsJSON = `{"value":[
	{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.Network/virtualNetworks/vnet-mock-1","name":"vnet-mock-1","location":"eastus","type":"Microsoft.Network/virtualNetworks","properties":{"addressSpace":{"addressPrefixes":["10.0.0.0/16"]}}}
]}`

func TestMockTrack2ResourceGroups(t *testing.T) {
	tr := &cannedTransport{body: mockResourceGroupsJSON}
	client, err := armresources.NewResourceGroupsClient(mockSubscriptionID, stubCredential{}, mockClientOptions(tr))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	pager := client.NewListPager(nil)
	resources := ResourceGroupGenerator{}.createResources(context.Background(), pager)
	if len(resources) != 2 {
		t.Fatalf("got %d resources, want 2", len(resources))
	}
	for i, want := range []string{"tfer--rg-mock-1", "tfer--rg-mock-2"} {
		if resources[i].ResourceName != want {
			t.Errorf("resource %d name = %q, want %q", i, resources[i].ResourceName, want)
		}
		if resources[i].InstanceInfo.Type != "azurerm_resource_group" {
			t.Errorf("resource %d type = %q, want azurerm_resource_group", i, resources[i].InstanceInfo.Type)
		}
	}
	if len(tr.urls) == 0 || !strings.Contains(tr.urls[0], "resourcegroups") {
		t.Errorf("expected an ARM resourcegroups call, got %v", tr.urls)
	}
}

func TestMockTrack2VirtualNetworks(t *testing.T) {
	tr := &cannedTransport{body: mockVnetsJSON}
	client, err := armnetwork.NewVirtualNetworksClient(mockSubscriptionID, stubCredential{}, mockClientOptions(tr))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	g := &VirtualNetworkGenerator{}
	pager := client.NewListAllPager(nil)
	ctx := context.Background()
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			t.Fatalf("next page: %v", err)
		}
		g.appendResources(page.Value)
	}
	if len(g.Resources) != 1 {
		t.Fatalf("got %d resources, want 1", len(g.Resources))
	}
	if g.Resources[0].ResourceName != "tfer--vnet-mock-1" {
		t.Errorf("name = %q, want tfer--vnet-mock-1", g.Resources[0].ResourceName)
	}
	if g.Resources[0].InstanceInfo.Type != "azurerm_virtual_network" {
		t.Errorf("type = %q, want azurerm_virtual_network", g.Resources[0].InstanceInfo.Type)
	}
	if len(tr.urls) == 0 || !strings.Contains(tr.urls[0], "virtualNetworks") {
		t.Errorf("expected an ARM virtualNetworks call, got %v", tr.urls)
	}
}

const mockDisksJSON = `{"value":[
	{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.Compute/disks/disk-mock-1","name":"disk-mock-1","location":"eastus","type":"Microsoft.Compute/disks","properties":{}}
]}`

const mockStorageAccountsJSON = `{"value":[
	{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.Storage/storageAccounts/stmock1","name":"stmock1","location":"eastus","type":"Microsoft.Storage/storageAccounts","kind":"StorageV2","sku":{"name":"Standard_LRS"}}
]}`

const mockNSGsJSON = `{"value":[
	{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.Network/networkSecurityGroups/nsg-mock-1","name":"nsg-mock-1","location":"eastus","type":"Microsoft.Network/networkSecurityGroups","properties":{}}
]}`

const mockDNSZonesJSON = `{"value":[
	{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.Network/dnszones/example.com","name":"example.com","type":"Microsoft.Network/dnszones","properties":{}}
]}`

func TestMockTrack2Disks(t *testing.T) {
	tr := &cannedTransport{body: mockDisksJSON}
	client, err := armcompute.NewDisksClient(mockSubscriptionID, stubCredential{}, mockClientOptions(tr))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	g := &DiskGenerator{}
	pager := client.NewListPager(nil)
	ctx := context.Background()
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			t.Fatalf("next page: %v", err)
		}
		g.appendResources(page.Value)
	}
	if len(g.Resources) != 1 {
		t.Fatalf("got %d resources, want 1", len(g.Resources))
	}
	if g.Resources[0].ResourceName != "tfer--disk-mock-1" {
		t.Errorf("name = %q, want tfer--disk-mock-1", g.Resources[0].ResourceName)
	}
	if g.Resources[0].InstanceInfo.Type != "azurerm_managed_disk" {
		t.Errorf("type = %q, want azurerm_managed_disk", g.Resources[0].InstanceInfo.Type)
	}
	if len(tr.urls) == 0 || !strings.Contains(tr.urls[0], "disks") {
		t.Errorf("expected an ARM disks call, got %v", tr.urls)
	}
}

func TestMockTrack2StorageAccounts(t *testing.T) {
	tr := &cannedTransport{body: mockStorageAccountsJSON}
	client, err := armstorage.NewAccountsClient(mockSubscriptionID, stubCredential{}, mockClientOptions(tr))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	pager := client.NewListPager(nil)
	ctx := context.Background()
	var resources []terraformutils.Resource
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			t.Fatalf("next page: %v", err)
		}
		resources = appendStorageAccounts(resources, page.Value)
	}
	if len(resources) != 1 {
		t.Fatalf("got %d resources, want 1", len(resources))
	}
	if resources[0].ResourceName != "tfer--stmock1" {
		t.Errorf("name = %q, want tfer--stmock1", resources[0].ResourceName)
	}
	if resources[0].InstanceInfo.Type != "azurerm_storage_account" {
		t.Errorf("type = %q, want azurerm_storage_account", resources[0].InstanceInfo.Type)
	}
	if len(tr.urls) == 0 || !strings.Contains(tr.urls[0], "storageAccounts") {
		t.Errorf("expected an ARM storageAccounts call, got %v", tr.urls)
	}
}

func TestMockTrack2NetworkSecurityGroups(t *testing.T) {
	tr := &cannedTransport{body: mockNSGsJSON}
	client, err := armnetwork.NewSecurityGroupsClient(mockSubscriptionID, stubCredential{}, mockClientOptions(tr))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	g := &NetworkSecurityGroupGenerator{}
	pager := client.NewListAllPager(nil)
	ctx := context.Background()
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			t.Fatalf("next page: %v", err)
		}
		for _, sg := range page.Value {
			g.appendResource(sg)
		}
	}
	if len(g.Resources) != 1 {
		t.Fatalf("got %d resources, want 1", len(g.Resources))
	}
	if !strings.HasPrefix(g.Resources[0].ResourceName, "tfer--nsg-mock-1") {
		t.Errorf("name = %q, want prefix tfer--nsg-mock-1", g.Resources[0].ResourceName)
	}
	if g.Resources[0].InstanceInfo.Type != "azurerm_network_security_group" {
		t.Errorf("type = %q, want azurerm_network_security_group", g.Resources[0].InstanceInfo.Type)
	}
	if len(tr.urls) == 0 || !strings.Contains(tr.urls[0], "networkSecurityGroups") {
		t.Errorf("expected an ARM networkSecurityGroups call, got %v", tr.urls)
	}
}

func TestMockTrack2DNSZones(t *testing.T) {
	tr := &cannedTransport{body: mockDNSZonesJSON}
	client, err := armdns.NewZonesClient(mockSubscriptionID, stubCredential{}, mockClientOptions(tr))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	var pageSize int32 = 50
	pager := client.NewListPager(&armdns.ZonesClientListOptions{Top: &pageSize})
	ctx := context.Background()
	if !pager.More() {
		t.Fatal("expected at least one page")
	}
	page, err := pager.NextPage(ctx)
	if err != nil {
		t.Fatalf("next page: %v", err)
	}
	if len(page.Value) != 1 || *page.Value[0].Name != "example.com" {
		t.Fatalf("unexpected zones page: %+v", page.Value)
	}
	if len(tr.urls) == 0 || !strings.Contains(tr.urls[0], "dnszones") {
		t.Errorf("expected an ARM dnszones call, got %v", tr.urls)
	}
}

const mockRedisJSON = `{"value":[
	{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.Cache/Redis/redis-mock-1","name":"redis-mock-1","location":"eastus","type":"Microsoft.Cache/Redis","properties":{}}
]}`

const mockAnalysisServersJSON = `{"value":[
	{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.AnalysisServices/servers/as-mock-1","name":"as-mock-1","location":"eastus","type":"Microsoft.AnalysisServices/servers","properties":{}}
]}`

const mockDatabricksJSON = `{"value":[
	{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.Databricks/workspaces/dbw-mock-1","name":"dbw-mock-1","location":"eastus","type":"Microsoft.Databricks/workspaces","properties":{}}
]}`

const mockKeyVaultsJSON = `{"value":[
	{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.KeyVault/vaults/kv-mock-1","name":"kv-mock-1","location":"eastus","type":"Microsoft.KeyVault/vaults","properties":{}}
]}`

func TestMockTrack2Redis(t *testing.T) {
	tr := &cannedTransport{body: mockRedisJSON}
	client, err := armredis.NewClient(mockSubscriptionID, stubCredential{}, mockClientOptions(tr))
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
		for _, s := range page.Value {
			if *s.Name != "redis-mock-1" {
				t.Errorf("name = %q, want redis-mock-1", *s.Name)
			}
			count++
		}
	}
	if count != 1 {
		t.Fatalf("got %d servers, want 1", count)
	}
	if len(tr.urls) == 0 || !strings.Contains(tr.urls[0], "Microsoft.Cache") {
		t.Errorf("expected a Microsoft.Cache call, got %v", tr.urls)
	}
}

func TestMockTrack2AnalysisServers(t *testing.T) {
	tr := &cannedTransport{body: mockAnalysisServersJSON}
	client, err := armanalysisservices.NewServersClient(mockSubscriptionID, stubCredential{}, mockClientOptions(tr))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	pager := client.NewListPager(nil)
	ctx := context.Background()
	if !pager.More() {
		t.Fatal("expected at least one page")
	}
	page, err := pager.NextPage(ctx)
	if err != nil {
		t.Fatalf("next page: %v", err)
	}
	if len(page.Value) != 1 || *page.Value[0].Name != "as-mock-1" {
		t.Fatalf("unexpected servers page: %+v", page.Value)
	}
	if len(tr.urls) == 0 || !strings.Contains(tr.urls[0], "Microsoft.AnalysisServices") {
		t.Errorf("expected a Microsoft.AnalysisServices call, got %v", tr.urls)
	}
}

func TestMockTrack2DatabricksWorkspaces(t *testing.T) {
	tr := &cannedTransport{body: mockDatabricksJSON}
	client, err := armdatabricks.NewWorkspacesClient(mockSubscriptionID, stubCredential{}, mockClientOptions(tr))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	pager := client.NewListBySubscriptionPager(nil)
	ctx := context.Background()
	if !pager.More() {
		t.Fatal("expected at least one page")
	}
	page, err := pager.NextPage(ctx)
	if err != nil {
		t.Fatalf("next page: %v", err)
	}
	if len(page.Value) != 1 || *page.Value[0].Name != "dbw-mock-1" {
		t.Fatalf("unexpected workspaces page: %+v", page.Value)
	}
	if len(tr.urls) == 0 || !strings.Contains(tr.urls[0], "Microsoft.Databricks") {
		t.Errorf("expected a Microsoft.Databricks call, got %v", tr.urls)
	}
}

func TestMockTrack2KeyVaults(t *testing.T) {
	tr := &cannedTransport{body: mockKeyVaultsJSON}
	client, err := armkeyvault.NewVaultsClient(mockSubscriptionID, stubCredential{}, mockClientOptions(tr))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	pager := client.NewListPager(nil)
	ctx := context.Background()
	if !pager.More() {
		t.Fatal("expected at least one page")
	}
	page, err := pager.NextPage(ctx)
	if err != nil {
		t.Fatalf("next page: %v", err)
	}
	if len(page.Value) != 1 || *page.Value[0].Name != "kv-mock-1" {
		t.Fatalf("unexpected vaults page: %+v", page.Value)
	}
	if len(tr.urls) == 0 || !strings.Contains(tr.urls[0], "Microsoft.KeyVault") {
		t.Errorf("expected a Microsoft.KeyVault call, got %v", tr.urls)
	}
}
