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

	armnetworkfunction "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/networkfunction/armnetworkfunction"
)

type NetworkFunctionGenerator struct {
	AzureService
}

func (g *NetworkFunctionGenerator) listAzureTrafficCollectors() ([]netParent, error) {
	ctx := context.Background()
	var parents []netParent
	handle := func(items []*armnetworkfunction.AzureTrafficCollector) {
		for _, item := range items {
			if item.ID != nil && item.Name != nil {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_network_function_azure_traffic_collector")
				if p, ok := makeParent(item.ID, item.Name); ok {
					parents = append(parents, p)
				}
			}
		}
	}
	if rg := g.getResourceGroup(); rg != "" {
		client, err := armnetworkfunction.NewAzureTrafficCollectorsByResourceGroupClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
		if err != nil {
			return parents, err
		}
		pager := client.NewListPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return parents, err
			}
			handle(page.Value)
		}
	} else {
		client, err := armnetworkfunction.NewAzureTrafficCollectorsBySubscriptionClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
		if err != nil {
			return parents, err
		}
		pager := client.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return parents, err
			}
			handle(page.Value)
		}
	}
	return parents, nil
}

func (g *NetworkFunctionGenerator) listCollectorPolicies(parents []netParent) error {
	ctx := context.Background()
	client, err := armnetworkfunction.NewCollectorPoliciesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	for _, p := range parents {
		pager := client.NewListPager(p.ResourceGroup, p.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_network_function_collector_policy")
				}
			}
		}
	}
	return nil
}

func (g *NetworkFunctionGenerator) InitResources() error {
	if err := func() error {
		collectors, err := g.listAzureTrafficCollectors()
		if err != nil {
			return err
		}
		return g.listCollectorPolicies(collectors)
	}(); err != nil {
		return err
	}
	return nil
}
