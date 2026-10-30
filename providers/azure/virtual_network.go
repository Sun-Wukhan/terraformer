// Copyright 2019 The Terraformer Authors.
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
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type VirtualNetworkGenerator struct {
	AzureService
}

func (g *VirtualNetworkGenerator) appendResources(vnets []*armnetwork.VirtualNetwork) {
	for _, virtualNetwork := range vnets {
		tferName := terraformutils.TfSanitize(*virtualNetwork.Name)
		for _, resource := range g.Resources {
			if tferName == resource.ResourceName {
				*virtualNetwork.Name = *virtualNetwork.Name + "_" + *virtualNetwork.ID
			}
		}

		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			*virtualNetwork.ID,
			*virtualNetwork.Name,
			"azurerm_virtual_network",
			g.ProviderName,
			[]string{}))
	}
}

func (g *VirtualNetworkGenerator) InitResources() error {
	ctx := context.Background()
	client, err := armnetwork.NewVirtualNetworksClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}

	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			g.appendResources(page.Value)
		}
	} else {
		pager := client.NewListAllPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			g.appendResources(page.Value)
		}
	}
	return g.appendPeerings()
}

// appendPeerings enumerates the peerings of every virtual network just
// listed. The peering importer ID is the peering's own ARM ID
// ({vnetId}/virtualNetworkPeerings/{name}), verified in the provider's
// Import documentation.
//
// Skipped here: azurerm_virtual_network_dns_servers — it is a PUT-only
// sub-resource of the virtual network (the dnsServers array on the VNet
// object itself); there is no independent ARM object or List operation
// for it in armnetwork.
func (g *VirtualNetworkGenerator) appendPeerings() error {
	ctx := context.Background()
	vnetClient, err := armnetwork.NewVirtualNetworksClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	peeringClient, err := armnetwork.NewVirtualNetworkPeeringsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	var vnets []*armnetwork.VirtualNetwork
	if rg := g.getResourceGroup(); rg != "" {
		pager := vnetClient.NewListPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			vnets = append(vnets, page.Value...)
		}
	} else {
		pager := vnetClient.NewListAllPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			vnets = append(vnets, page.Value...)
		}
	}
	for _, vnet := range vnets {
		if vnet.ID == nil || vnet.Name == nil {
			continue
		}
		id, err := ParseAzureResourceID(*vnet.ID)
		if err != nil {
			return err
		}
		pager := peeringClient.NewListPager(id.ResourceGroup, *vnet.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, peering := range page.Value {
				if peering.ID != nil && peering.Name != nil {
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						*peering.ID,
						*peering.Name,
						"azurerm_virtual_network_peering",
						g.ProviderName,
						[]string{}))
				}
			}
		}
	}
	return nil
}

// NOTE on Virtual Networks and Subnet's:
// Terraform currently provides both a standalone Subnet resource, and allows for Subnets to be defined in-line within the Virtual Network
// resource. At this time you cannot use a Virtual Network with in-line Subnets in conjunction with any Subnet resources.
// Doing so will cause a conflict of Subnet configurations and will overwrite Subnet's.
func (g *VirtualNetworkGenerator) PostConvertHook() error {
	for _, resource := range g.Resources {
		if resource.InstanceInfo.Type != "azurerm_virtual_network" {
			continue
		}
		delete(resource.Item, "subnet")
	}
	return nil
}
