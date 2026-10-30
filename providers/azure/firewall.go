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
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

type FirewallGenerator struct {
	AzureService
}

func (g *FirewallGenerator) listFirewalls() error {
	ctx := context.Background()
	client, err := armnetwork.NewAzureFirewallsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	handle := func(items []*armnetwork.AzureFirewall) {
		for _, item := range items {
			if item.ID == nil || item.Name == nil {
				continue
			}
			g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_firewall")
			g.appendFirewallRuleCollections(item)
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

// Classic Azure Firewall rule collections are embedded in the firewall
// object (properties.applicationRuleCollections / natRuleCollections /
// networkRuleCollections). Their importer IDs are
// {firewallId}/applicationRuleCollections/{name} (and nat/network
// equivalents), verified in the provider's Import documentation; the API
// returns an ID on each collection, with the documented form as fallback.
func (g *FirewallGenerator) appendFirewallRuleCollections(firewall *armnetwork.AzureFirewall) {
	if firewall.Properties == nil {
		return
	}
	for _, c := range firewall.Properties.ApplicationRuleCollections {
		if c.Name != nil {
			id := *firewall.ID + "/applicationRuleCollections/" + *c.Name
			if c.ID != nil {
				id = *c.ID
			}
			g.AppendSimpleResource(id, *c.Name, "azurerm_firewall_application_rule_collection")
		}
	}
	for _, c := range firewall.Properties.NatRuleCollections {
		if c.Name != nil {
			id := *firewall.ID + "/natRuleCollections/" + *c.Name
			if c.ID != nil {
				id = *c.ID
			}
			g.AppendSimpleResource(id, *c.Name, "azurerm_firewall_nat_rule_collection")
		}
	}
	for _, c := range firewall.Properties.NetworkRuleCollections {
		if c.Name != nil {
			id := *firewall.ID + "/networkRuleCollections/" + *c.Name
			if c.ID != nil {
				id = *c.ID
			}
			g.AppendSimpleResource(id, *c.Name, "azurerm_firewall_network_rule_collection")
		}
	}
}

func (g *FirewallGenerator) listFirewallPolicies() ([]netParent, error) {
	ctx := context.Background()
	var parents []netParent
	client, err := armnetwork.NewFirewallPoliciesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return parents, err
	}
	handle := func(items []*armnetwork.FirewallPolicy) {
		for _, item := range items {
			if item.ID != nil && item.Name != nil {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_firewall_policy")
				if p, ok := makeParent(item.ID, item.Name); ok {
					parents = append(parents, p)
				}
			}
		}
	}
	if rg := g.getResourceGroup(); rg != "" {
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
		pager := client.NewListAllPager(nil)
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

func (g *FirewallGenerator) listFirewallPolicyRuleCollectionGroups(parents []netParent) error {
	ctx := context.Background()
	client, err := armnetwork.NewFirewallPolicyRuleCollectionGroupsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
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
					g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_firewall_policy_rule_collection_group")
				}
			}
		}
	}
	return nil
}

func (g *FirewallGenerator) listIPGroups() error {
	ctx := context.Background()
	client, err := armnetwork.NewIPGroupsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	handle := func(items []*armnetwork.IPGroup) {
		for _, item := range items {
			if item.ID == nil || item.Name == nil {
				continue
			}
			g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_ip_group")
			g.appendIPGroupCIDRs(item)
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

// IP group CIDRs are the ipAddresses strings on the IP group object; the
// azurerm_ip_group_cidr importer ID is {ipGroupId}/cidrs/{cidr} with "/"
// replaced by "_" (verified in the provider's Import documentation).
func (g *FirewallGenerator) appendIPGroupCIDRs(group *armnetwork.IPGroup) {
	if group.Properties == nil {
		return
	}
	for _, cidr := range group.Properties.IPAddresses {
		if cidr == nil {
			continue
		}
		id := *group.ID + "/cidrs/" + strings.ReplaceAll(*cidr, "/", "_")
		g.AppendSimpleResource(id, *group.Name+"-"+*cidr, "azurerm_ip_group_cidr")
	}
}

func (g *FirewallGenerator) InitResources() error {
	if err := g.listFirewalls(); err != nil {
		return err
	}
	if err := func() error {
		policies, err := g.listFirewallPolicies()
		if err != nil {
			return err
		}
		return g.listFirewallPolicyRuleCollectionGroups(policies)
	}(); err != nil {
		return err
	}
	if err := g.listIPGroups(); err != nil {
		return err
	}
	return nil
}
