// Copyright 2021 The Terraformer Authors.
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

type SubnetGenerator struct {
	AzureService
}

func (az *SubnetGenerator) listSubnets() ([]*armnetwork.Subnet, error) {
	ctx := context.Background()
	subnetClient, err := armnetwork.NewSubnetsClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
	if err != nil {
		return nil, err
	}
	vnetClient, err := armnetwork.NewVirtualNetworksClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
	if err != nil {
		return nil, err
	}
	var resources []*armnetwork.Subnet
	if rg := az.getResourceGroup(); rg != "" {
		pager := vnetClient.NewListPager(rg, nil)
		for pager.More() {
			vnetPage, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return resources, err
			}
			for _, vnet := range vnetPage.Value {
				vnetID, err := ParseAzureResourceID(*vnet.ID)
				if err != nil {
					return nil, err
				}
				subnetPager := subnetClient.NewListPager(vnetID.ResourceGroup, *vnet.Name, nil)
				for subnetPager.More() {
					subnetPage, err := subnetPager.NextPage(ctx)
					if err != nil {
						log.Println(err)
						return resources, err
					}
					resources = append(resources, subnetPage.Value...)
				}
			}
		}
	} else {
		pager := vnetClient.NewListAllPager(nil)
		for pager.More() {
			vnetPage, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return resources, err
			}
			for _, vnet := range vnetPage.Value {
				vnetID, err := ParseAzureResourceID(*vnet.ID)
				if err != nil {
					return nil, err
				}
				subnetPager := subnetClient.NewListPager(vnetID.ResourceGroup, *vnet.Name, nil)
				for subnetPager.More() {
					subnetPage, err := subnetPager.NextPage(ctx)
					if err != nil {
						log.Println(err)
						return resources, err
					}
					resources = append(resources, subnetPage.Value...)
				}
			}
		}
	}
	return resources, nil
}

func (az *SubnetGenerator) AppendSubnet(subnet *armnetwork.Subnet) {
	az.AppendSimpleResource(*subnet.ID, *subnet.Name, "azurerm_subnet")
}

func (az *SubnetGenerator) appendRouteTable(subnet *armnetwork.Subnet) {
	if props := subnet.Properties; props != nil {
		if prop := props.RouteTable; prop != nil {
			az.appendSimpleAssociation(
				*subnet.ID, *subnet.Name, prop.Name,
				"azurerm_subnet_route_table_association",
				map[string]string{
					"subnet_id":      *subnet.ID,
					"route_table_id": *prop.ID,
				})
		}
	}
}

func (az *SubnetGenerator) appendNetworkSecurityGroupAssociation(subnet *armnetwork.Subnet) {
	if props := subnet.Properties; props != nil {
		if prop := props.NetworkSecurityGroup; prop != nil {
			az.appendSimpleAssociation(
				*subnet.ID, *subnet.Name, prop.Name,
				"azurerm_subnet_network_security_group_association",
				map[string]string{
					"subnet_id":                 *subnet.ID,
					"network_security_group_id": *prop.ID,
				})
		}
	}
}

func (az *SubnetGenerator) appendNatGateway(subnet *armnetwork.Subnet) {
	if props := subnet.Properties; props != nil {
		if prop := props.NatGateway; prop != nil {
			az.appendSimpleAssociation(
				*subnet.ID, *subnet.Name, nil,
				"azurerm_subnet_nat_gateway_association",
				map[string]string{
					"subnet_id":      *subnet.ID,
					"nat_gateway_id": *prop.ID,
				})
		}
	}
}

func (az *SubnetGenerator) appendServiceEndpointPolicies() error {
	ctx := context.Background()
	client, err := armnetwork.NewServiceEndpointPoliciesClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
	if err != nil {
		return err
	}
	if rg := az.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				az.AppendSimpleResource(*item.ID, *item.Name, "azurerm_subnet_service_endpoint_storage_policy")
			}
		}
	} else {
		pager := client.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				az.AppendSimpleResource(*item.ID, *item.Name, "azurerm_subnet_service_endpoint_storage_policy")
			}
		}
	}
	return nil
}

func (az *SubnetGenerator) InitResources() error {

	subnets, err := az.listSubnets()
	if err != nil {
		return err
	}
	for _, subnet := range subnets {
		az.AppendSubnet(subnet)
		az.appendRouteTable(subnet)
		az.appendNetworkSecurityGroupAssociation(subnet)
		az.appendNatGateway(subnet)
	}
	if err := az.appendServiceEndpointPolicies(); err != nil {
		return err
	}
	return nil
}

func (az *SubnetGenerator) PostConvertHook() error {
	for _, resource := range az.Resources {
		if resource.InstanceInfo.Type != "azurerm_subnet" {
			continue
		}
		delete(resource.Item, "address_prefix")
	}
	return nil
}
