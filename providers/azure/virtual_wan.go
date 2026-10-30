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
	armnetworkv2 "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v2"
)

type VirtualWANGenerator struct {
	AzureService
}

func (g *VirtualWANGenerator) listVirtualWans() error {
	ctx := context.Background()
	client, err := armnetwork.NewVirtualWansClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	handle := func(items []*armnetwork.VirtualWAN) {
		for _, item := range items {
			if item.ID != nil && item.Name != nil {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_virtual_wan")
			}
		}
	}
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			handle(page.Value)
		}
	} else {
		pager := client.NewListPager(nil)
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

func (g *VirtualWANGenerator) listVirtualHubs() ([]netParent, error) {
	ctx := context.Background()
	var parents []netParent
	client, err := armnetwork.NewVirtualHubsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return parents, err
	}
	handle := func(items []*armnetwork.VirtualHub) {
		for _, item := range items {
			if item.ID != nil && item.Name != nil {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_virtual_hub")
				if p, ok := makeParent(item.ID, item.Name); ok {
					parents = append(parents, p)
				}
			}
		}
	}
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return parents, err
			}
			handle(page.Value)
		}
	} else {
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

func (g *VirtualWANGenerator) listHubVirtualNetworkConnections(parents []netParent) error {
	ctx := context.Background()
	client, err := armnetwork.NewHubVirtualNetworkConnectionsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
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
					g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_virtual_hub_connection")
				}
			}
		}
	}
	return nil
}

func (g *VirtualWANGenerator) listVirtualHubBgpConnections(parents []netParent) error {
	ctx := context.Background()
	client, err := armnetwork.NewVirtualHubBgpConnectionsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
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
					g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_virtual_hub_bgp_connection")
				}
			}
		}
	}
	return nil
}

func (g *VirtualWANGenerator) listVirtualHubIPConfigurations(parents []netParent) error {
	ctx := context.Background()
	client, err := armnetwork.NewVirtualHubIPConfigurationClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
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
					g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_virtual_hub_ip")
				}
			}
		}
	}
	return nil
}

func (g *VirtualWANGenerator) listRoutingIntents(parents []netParent) error {
	ctx := context.Background()
	client, err := armnetwork.NewRoutingIntentClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
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
					g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_virtual_hub_routing_intent")
				}
			}
		}
	}
	return nil
}

func (g *VirtualWANGenerator) listHubRouteTables(parents []netParent) error {
	ctx := context.Background()
	client, err := armnetwork.NewHubRouteTablesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
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
				if item.ID == nil || item.Name == nil {
					continue
				}
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_virtual_hub_route_table")
				// Routes are embedded in the route table object; their
				// importer ID is {routeTableId}/routes/{routeName} (verified
				// in the provider's Import documentation).
				if item.Properties != nil {
					for _, route := range item.Properties.Routes {
						if route.Name != nil {
							g.AppendSimpleResource(*item.ID+"/routes/"+*route.Name, *route.Name, "azurerm_virtual_hub_route_table_route")
						}
					}
				}
			}
		}
	}
	return nil
}

func (g *VirtualWANGenerator) listRouteMaps(parents []netParent) error {
	ctx := context.Background()
	client, err := armnetworkv2.NewRouteMapsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
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
					g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_route_map")
				}
			}
		}
	}
	return nil
}

func (g *VirtualWANGenerator) listSecurityPartnerProviders() error {
	ctx := context.Background()
	client, err := armnetwork.NewSecurityPartnerProvidersClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	handle := func(items []*armnetwork.SecurityPartnerProvider) {
		for _, item := range items {
			if item.ID != nil && item.Name != nil {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_virtual_hub_security_partner_provider")
			}
		}
	}
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			handle(page.Value)
		}
	} else {
		pager := client.NewListPager(nil)
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

func (g *VirtualWANGenerator) listVPNGateways() ([]netParent, error) {
	ctx := context.Background()
	var parents []netParent
	client, err := armnetwork.NewVPNGatewaysClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return parents, err
	}
	handle := func(items []*armnetwork.VPNGateway) {
		for _, item := range items {
			if item.ID != nil && item.Name != nil {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_vpn_gateway")
				if p, ok := makeParent(item.ID, item.Name); ok {
					parents = append(parents, p)
				}
			}
		}
	}
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return parents, err
			}
			handle(page.Value)
		}
	} else {
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

func (g *VirtualWANGenerator) listVPNConnections(parents []netParent) error {
	ctx := context.Background()
	client, err := armnetwork.NewVPNConnectionsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	for _, p := range parents {
		pager := client.NewListByVPNGatewayPager(p.ResourceGroup, p.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_vpn_gateway_connection")
				}
			}
		}
	}
	return nil
}

func (g *VirtualWANGenerator) listVPNGatewayNatRules(parents []netParent) error {
	ctx := context.Background()
	client, err := armnetwork.NewNatRulesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	for _, p := range parents {
		pager := client.NewListByVPNGatewayPager(p.ResourceGroup, p.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_vpn_gateway_nat_rule")
				}
			}
		}
	}
	return nil
}

func (g *VirtualWANGenerator) listP2SVPNGateways() error {
	ctx := context.Background()
	client, err := armnetwork.NewP2SVPNGatewaysClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	handle := func(items []*armnetwork.P2SVPNGateway) {
		for _, item := range items {
			if item.ID != nil && item.Name != nil {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_point_to_site_vpn_gateway")
			}
		}
	}
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			handle(page.Value)
		}
	} else {
		pager := client.NewListPager(nil)
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

func (g *VirtualWANGenerator) listVPNSites() error {
	ctx := context.Background()
	client, err := armnetwork.NewVPNSitesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	handle := func(items []*armnetwork.VPNSite) {
		for _, item := range items {
			if item.ID != nil && item.Name != nil {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_vpn_site")
			}
		}
	}
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			handle(page.Value)
		}
	} else {
		pager := client.NewListPager(nil)
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

func (g *VirtualWANGenerator) listVPNServerConfigurations() ([]netParent, error) {
	ctx := context.Background()
	var parents []netParent
	client, err := armnetwork.NewVPNServerConfigurationsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return parents, err
	}
	handle := func(items []*armnetwork.VPNServerConfiguration) {
		for _, item := range items {
			if item.ID != nil && item.Name != nil {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_vpn_server_configuration")
				if p, ok := makeParent(item.ID, item.Name); ok {
					parents = append(parents, p)
				}
			}
		}
	}
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return parents, err
			}
			handle(page.Value)
		}
	} else {
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

func (g *VirtualWANGenerator) listConfigurationPolicyGroups(parents []netParent) error {
	ctx := context.Background()
	client, err := armnetwork.NewConfigurationPolicyGroupsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	for _, p := range parents {
		pager := client.NewListByVPNServerConfigurationPager(p.ResourceGroup, p.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_vpn_server_configuration_policy_group")
				}
			}
		}
	}
	return nil
}

func (g *VirtualWANGenerator) listVirtualRouters() ([]netParent, error) {
	ctx := context.Background()
	var parents []netParent
	client, err := armnetwork.NewVirtualRoutersClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return parents, err
	}
	handle := func(items []*armnetwork.VirtualRouter) {
		for _, item := range items {
			if item.ID != nil && item.Name != nil {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_route_server")
				if p, ok := makeParent(item.ID, item.Name); ok {
					parents = append(parents, p)
				}
			}
		}
	}
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return parents, err
			}
			handle(page.Value)
		}
	} else {
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

func (g *VirtualWANGenerator) listVirtualRouterPeerings(parents []netParent) error {
	ctx := context.Background()
	client, err := armnetwork.NewVirtualRouterPeeringsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
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
					g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_route_server_bgp_connection")
				}
			}
		}
	}
	return nil
}

func (g *VirtualWANGenerator) InitResources() error {
	if err := g.listVirtualWans(); err != nil {
		return err
	}
	if err := func() error {
		hubs, err := g.listVirtualHubs()
		if err != nil {
			return err
		}
		for _, f := range []func([]netParent) error{g.listHubVirtualNetworkConnections, g.listVirtualHubBgpConnections, g.listVirtualHubIPConfigurations, g.listRoutingIntents, g.listHubRouteTables, g.listRouteMaps} {
			if err := f(hubs); err != nil {
				return err
			}
		}
		return nil
	}(); err != nil {
		return err
	}
	if err := g.listSecurityPartnerProviders(); err != nil {
		return err
	}
	if err := func() error {
		gateways, err := g.listVPNGateways()
		if err != nil {
			return err
		}
		if err := g.listVPNConnections(gateways); err != nil {
			return err
		}
		return g.listVPNGatewayNatRules(gateways)
	}(); err != nil {
		return err
	}
	if err := g.listP2SVPNGateways(); err != nil {
		return err
	}
	if err := g.listVPNSites(); err != nil {
		return err
	}
	if err := func() error {
		configs, err := g.listVPNServerConfigurations()
		if err != nil {
			return err
		}
		return g.listConfigurationPolicyGroups(configs)
	}(); err != nil {
		return err
	}
	if err := func() error {
		routers, err := g.listVirtualRouters()
		if err != nil {
			return err
		}
		return g.listVirtualRouterPeerings(routers)
	}(); err != nil {
		return err
	}
	return nil
}
