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

type NatGatewayGenerator struct {
	AzureService
}

func (g *NatGatewayGenerator) listNatGateways() error {
	ctx := context.Background()
	client, err := armnetwork.NewNatGatewaysClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	handle := func(items []*armnetwork.NatGateway) {
		for _, item := range items {
			if item.ID == nil || item.Name == nil {
				continue
			}
			g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_nat_gateway")
			g.appendNatGatewayAssociations(item)
		}
	}
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			handle(page.Value)
		}
	} else {
		pager := client.NewListAllPager(nil)
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

// The azurerm association resources use Terraform-specific composite IDs of
// the form {natGatewayId}|{publicIpId} / {natGatewayId}|{publicIpPrefixId}
// (verified in the provider's Import documentation); the associations
// themselves are the publicIpAddresses / publicIpPrefixes references on the
// NAT gateway object, there is no separate ARM List for them.
func (g *NatGatewayGenerator) appendNatGatewayAssociations(gateway *armnetwork.NatGateway) {
	if gateway.Properties == nil {
		return
	}
	for _, pip := range gateway.Properties.PublicIPAddresses {
		if pip.ID == nil {
			continue
		}
		g.appendSimpleAssociation(*gateway.ID+"|"+*pip.ID, *gateway.Name, nil,
			"azurerm_nat_gateway_public_ip_association",
			map[string]string{
				"nat_gateway_id":       *gateway.ID,
				"public_ip_address_id": *pip.ID,
			})
	}
	for _, prefix := range gateway.Properties.PublicIPPrefixes {
		if prefix.ID == nil {
			continue
		}
		g.appendSimpleAssociation(*gateway.ID+"|"+*prefix.ID, *gateway.Name, nil,
			"azurerm_nat_gateway_public_ip_prefix_association",
			map[string]string{
				"nat_gateway_id":      *gateway.ID,
				"public_ip_prefix_id": *prefix.ID,
			})
	}
}

func (g *NatGatewayGenerator) InitResources() error {
	if err := g.listNatGateways(); err != nil {
		return err
	}
	return nil
}
