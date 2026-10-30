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

type NetworkInterfaceGenerator struct {
	AzureService
}

func (g *NetworkInterfaceGenerator) appendResources(interfaces []*armnetwork.Interface) {
	for _, networkInterface := range interfaces {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			*networkInterface.ID,
			*networkInterface.Name,
			"azurerm_network_interface",
			"azurerm",
			[]string{}))
		g.appendAssociations(networkInterface)
	}
}

// The azurerm NIC association resources use Terraform-specific composite
// IDs: {networkInterfaceId}|{networkSecurityGroupId} for the NSG
// association and {networkInterfaceId}/ipConfigurations/{name}|{linkedId}
// for the per-IP-configuration associations (verified in the providers'
// Import documentation). The associations themselves are the references
// on the NIC / IP configuration objects; there is no separate ARM List.
func (g *NetworkInterfaceGenerator) appendAssociations(nic *armnetwork.Interface) {
	if nic.ID == nil || nic.Name == nil || nic.Properties == nil {
		return
	}
	if nsg := nic.Properties.NetworkSecurityGroup; nsg != nil && nsg.ID != nil {
		name := *nic.Name + "-security-group"
		g.appendSimpleAssociation(
			*nic.ID+"|"+*nsg.ID, *nic.Name, &name,
			"azurerm_network_interface_security_group_association",
			map[string]string{
				"network_interface_id":      *nic.ID,
				"network_security_group_id": *nsg.ID,
			})
	}
	for _, ipConfig := range nic.Properties.IPConfigurations {
		if ipConfig.Name == nil || ipConfig.Properties == nil {
			continue
		}
		base := *nic.ID + "/ipConfigurations/" + *ipConfig.Name
		props := ipConfig.Properties
		for _, pool := range props.LoadBalancerBackendAddressPools {
			if pool.ID == nil {
				continue
			}
			name := *nic.Name + "-" + *ipConfig.Name + "-backend-pool"
			g.appendSimpleAssociation(
				base+"|"+*pool.ID, *nic.Name, &name,
				"azurerm_network_interface_backend_address_pool_association",
				map[string]string{
					"network_interface_id":    *nic.ID,
					"ip_configuration_name":   *ipConfig.Name,
					"backend_address_pool_id": *pool.ID,
				})
		}
		for _, rule := range props.LoadBalancerInboundNatRules {
			if rule.ID == nil {
				continue
			}
			name := *nic.Name + "-" + *ipConfig.Name + "-nat-rule"
			g.appendSimpleAssociation(
				base+"|"+*rule.ID, *nic.Name, &name,
				"azurerm_network_interface_nat_rule_association",
				map[string]string{
					"network_interface_id":  *nic.ID,
					"ip_configuration_name": *ipConfig.Name,
					"nat_rule_id":           *rule.ID,
				})
		}
		for _, pool := range props.ApplicationGatewayBackendAddressPools {
			if pool.ID == nil {
				continue
			}
			name := *nic.Name + "-" + *ipConfig.Name + "-appgw-pool"
			g.appendSimpleAssociation(
				base+"|"+*pool.ID, *nic.Name, &name,
				"azurerm_network_interface_application_gateway_backend_address_pool_association",
				map[string]string{
					"network_interface_id":    *nic.ID,
					"ip_configuration_name":   *ipConfig.Name,
					"backend_address_pool_id": *pool.ID,
				})
		}
		for _, asg := range props.ApplicationSecurityGroups {
			if asg.ID == nil {
				continue
			}
			name := *nic.Name + "-" + *ipConfig.Name + "-asg"
			g.appendSimpleAssociation(
				base+"|"+*asg.ID, *nic.Name, &name,
				"azurerm_network_interface_application_security_group_association",
				map[string]string{
					"network_interface_id":          *nic.ID,
					"ip_configuration_name":         *ipConfig.Name,
					"application_security_group_id": *asg.ID,
				})
		}
	}
}

func (g *NetworkInterfaceGenerator) InitResources() error {
	ctx := context.Background()
	client, err := armnetwork.NewInterfacesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
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
	return nil
}
