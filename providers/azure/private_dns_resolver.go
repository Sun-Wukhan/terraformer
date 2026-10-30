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

	armdnsresolver "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dnsresolver/armdnsresolver"
)

type PrivateDNSResolverGenerator struct {
	AzureService
}

func (g *PrivateDNSResolverGenerator) listDNSResolvers() ([]netParent, error) {
	ctx := context.Background()
	var parents []netParent
	client, err := armdnsresolver.NewDNSResolversClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return parents, err
	}
	handle := func(items []*armdnsresolver.DNSResolver) {
		for _, item := range items {
			if item.ID != nil && item.Name != nil {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_private_dns_resolver")
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

func (g *PrivateDNSResolverGenerator) listInboundEndpoints(parents []netParent) error {
	ctx := context.Background()
	client, err := armdnsresolver.NewInboundEndpointsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
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
					g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_private_dns_resolver_inbound_endpoint")
				}
			}
		}
	}
	return nil
}

func (g *PrivateDNSResolverGenerator) listOutboundEndpoints(parents []netParent) error {
	ctx := context.Background()
	client, err := armdnsresolver.NewOutboundEndpointsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
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
					g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_private_dns_resolver_outbound_endpoint")
				}
			}
		}
	}
	return nil
}

func (g *PrivateDNSResolverGenerator) listDNSForwardingRulesets() ([]netParent, error) {
	ctx := context.Background()
	var parents []netParent
	client, err := armdnsresolver.NewDNSForwardingRulesetsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return parents, err
	}
	handle := func(items []*armdnsresolver.DNSForwardingRuleset) {
		for _, item := range items {
			if item.ID != nil && item.Name != nil {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_private_dns_resolver_dns_forwarding_ruleset")
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

func (g *PrivateDNSResolverGenerator) listForwardingRules(parents []netParent) error {
	ctx := context.Background()
	client, err := armdnsresolver.NewForwardingRulesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
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
					g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_private_dns_resolver_forwarding_rule")
				}
			}
		}
	}
	return nil
}

func (g *PrivateDNSResolverGenerator) listVirtualNetworkLinks(parents []netParent) error {
	ctx := context.Background()
	client, err := armdnsresolver.NewVirtualNetworkLinksClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
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
					g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_private_dns_resolver_virtual_network_link")
				}
			}
		}
	}
	return nil
}

func (g *PrivateDNSResolverGenerator) InitResources() error {
	if err := func() error {
		resolvers, err := g.listDNSResolvers()
		if err != nil {
			return err
		}
		if err := g.listInboundEndpoints(resolvers); err != nil {
			return err
		}
		return g.listOutboundEndpoints(resolvers)
	}(); err != nil {
		return err
	}
	if err := func() error {
		rulesets, err := g.listDNSForwardingRulesets()
		if err != nil {
			return err
		}
		if err := g.listForwardingRules(rulesets); err != nil {
			return err
		}
		return g.listVirtualNetworkLinks(rulesets)
	}(); err != nil {
		return err
	}
	return nil
}
