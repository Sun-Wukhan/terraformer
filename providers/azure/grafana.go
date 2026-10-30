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
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dashboard/armdashboard"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// DashboardGrafanaGenerator covers Azure Managed Grafana via armdashboard:
// workspaces and their managed private endpoints, both by ARM ID.
type DashboardGrafanaGenerator struct {
	AzureService
}

func (g *DashboardGrafanaGenerator) InitResources() error {
	ctx := context.Background()
	sub := g.getSubscriptionID()
	cred := g.getCredential()
	opts := g.armClientOptions()
	var resources []terraformutils.Resource
	add := func(id, name, typ string) {
		if id != "" && name != "" {
			resources = append(resources, terraformutils.NewSimpleResource(id, name, typ, "azurerm", []string{}))
		}
	}
	grafanaClient, err := armdashboard.NewGrafanaClient(sub, cred, opts)
	if err != nil {
		return err
	}
	endpointsClient, err := armdashboard.NewManagedPrivateEndpointsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	type wsRef struct{ rg, name string }
	var workspaces []wsRef
	if rg := g.getResourceGroup(); rg != "" {
		pager := grafanaClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_dashboard_grafana")
					workspaces = append(workspaces, wsRef{rg, *item.Name})
				}
			}
		}
	} else {
		pager := grafanaClient.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				if item.ID == nil || item.Name == nil {
					continue
				}
				add(*item.ID, *item.Name, "azurerm_dashboard_grafana")
				if parsed, err := ParseAzureResourceID(*item.ID); err == nil {
					workspaces = append(workspaces, wsRef{parsed.ResourceGroup, *item.Name})
				}
			}
		}
	}
	for _, ws := range workspaces {
		pager := endpointsClient.NewListPager(ws.rg, ws.name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_dashboard_grafana_managed_private_endpoint")
				}
			}
		}
	}
	g.Resources = resources
	return nil
}
