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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

type NetworkGatewayGenerator struct {
	AzureService
}

func (g *NetworkGatewayGenerator) listLocalNetworkGateways() error {
	ctx := context.Background()
	client, err := armnetwork.NewLocalNetworkGatewaysClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	handle := func(items []*armnetwork.LocalNetworkGateway) {
		for _, item := range items {
			if item.ID != nil && item.Name != nil {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_local_network_gateway")
			}
		}
	}
	rgs, err := g.resourceGroupNames()
	if err != nil {
		return err
	}
	for _, rg := range rgs {
		pager := client.NewListPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			handle(page.Value)
		}
	}
	return nil
}

func (g *NetworkGatewayGenerator) listVirtualNetworkGateways() ([]netParent, error) {
	ctx := context.Background()
	var parents []netParent
	client, err := armnetwork.NewVirtualNetworkGatewaysClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return parents, err
	}
	handle := func(items []*armnetwork.VirtualNetworkGateway) {
		for _, item := range items {
			if item.ID != nil && item.Name != nil {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_virtual_network_gateway")
				if p, ok := makeParent(item.ID, item.Name); ok {
					parents = append(parents, p)
				}
			}
		}
	}
	rgs, err := g.resourceGroupNames()
	if err != nil {
		return parents, err
	}
	for _, rg := range rgs {
		pager := client.NewListPager(rg, nil)
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

func (g *NetworkGatewayGenerator) listVirtualNetworkGatewayConnections() error {
	ctx := context.Background()
	client, err := armnetwork.NewVirtualNetworkGatewayConnectionsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	handle := func(items []*armnetwork.VirtualNetworkGatewayConnection) {
		for _, item := range items {
			if item.ID != nil && item.Name != nil {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_virtual_network_gateway_connection")
			}
		}
	}
	rgs, err := g.resourceGroupNames()
	if err != nil {
		return err
	}
	for _, rg := range rgs {
		pager := client.NewListPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			handle(page.Value)
		}
	}
	return nil
}

func (g *NetworkGatewayGenerator) listVirtualNetworkGatewayNatRules(parents []netParent) error {
	ctx := context.Background()
	client, err := armnetwork.NewVirtualNetworkGatewayNatRulesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	for _, p := range parents {
		pager := client.NewListByVirtualNetworkGatewayPager(p.ResourceGroup, p.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_virtual_network_gateway_nat_rule")
				}
			}
		}
	}
	return nil
}

func (g *NetworkGatewayGenerator) InitResources() error {
	if err := g.listLocalNetworkGateways(); err != nil {
		return err
	}
	if err := func() error {
		parents, err := g.listVirtualNetworkGateways()
		if err != nil {
			return err
		}
		return g.listVirtualNetworkGatewayNatRules(parents)
	}(); err != nil {
		return err
	}
	if err := g.listVirtualNetworkGatewayConnections(); err != nil {
		return err
	}
	return nil
}
