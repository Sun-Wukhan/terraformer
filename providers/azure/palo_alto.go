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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/paloaltonetworksngfw/armpanngfw"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// PaloAltoGenerator covers PaloAltoNetworks.Cloudngfw local rulestacks
// (and their certificates, FQDN lists, prefix lists and rules) and
// firewalls. All six firewall Terraform types share the single ARM
// firewalls/{name} resource; they are distinguished by
// FirewallDeploymentProperties (see classifyPaloFirewall).
//
// Skipped in this service (see ledger):
//   - azurerm_palo_alto_local_rulestack_outbound_trust_certificate_association
//     and _outbound_untrust_certificate_association: armpanngfw v1.1.0 has
//     no client / List for associations; the association is the
//     outboundTrustCertificate / outboundUnTrustCertificate property of
//     the local rulestack itself and the provider ID is the rulestack ID
//     (embedded/singleton).
//   - azurerm_palo_alto_virtual_network_appliance: it shares the
//     Microsoft.Network/networkVirtualAppliances ARM resource with the
//     generic network virtual appliance type and armnetwork exposes no
//     vendor discriminator to identify Palo Alto appliances.
type PaloAltoGenerator struct {
	AzureService
}

// classifyPaloFirewall maps a firewall's deployment properties to one of
// the six azurerm_palo_alto_next_generation_firewall_* types: the
// NetworkProfile.NetworkType selects virtual_network (VNET) vs
// virtual_hub (VWAN); Panorama-managed firewalls (IsPanoramaManaged or a
// PanoramaConfig) are the panorama variants; firewalls with an associated
// local rulestack are the local_rulestack variants; the remainder are
// Strata Cloud Manager managed.
func classifyPaloFirewall(props *armpanngfw.FirewallDeploymentProperties) string {
	network := "virtual_network"
	management := "strata_cloud_manager"
	if props != nil {
		if props.NetworkProfile != nil && props.NetworkProfile.NetworkType != nil && *props.NetworkProfile.NetworkType == armpanngfw.NetworkTypeVWAN {
			network = "virtual_hub"
		}
		switch {
		case props.PanoramaConfig != nil, props.IsPanoramaManaged != nil && *props.IsPanoramaManaged == armpanngfw.BooleanEnumTRUE:
			management = "panorama"
		case props.AssociatedRulestack != nil:
			management = "local_rulestack"
		}
	}
	return "azurerm_palo_alto_next_generation_firewall_" + network + "_" + management
}

func (g *PaloAltoGenerator) listLocalRulestacks() ([]*armpanngfw.LocalRulestackResource, error) {
	client, err := armpanngfw.NewLocalRulestacksClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	var out []*armpanngfw.LocalRulestackResource
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return out, err
			}
			out = append(out, page.Value...)
		}
	} else {
		pager := client.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return out, err
			}
			out = append(out, page.Value...)
		}
	}
	return out, nil
}

func (g *PaloAltoGenerator) listRulestackChildren(rg, stack string) []terraformutils.Resource {
	var resources []terraformutils.Resource
	ctx := context.Background()
	sub, cred, opts := g.getSubscriptionID(), g.getCredential(), g.armClientOptions()
	if client, err := armpanngfw.NewCertificateObjectLocalRulestackClient(sub, cred, opts); err == nil {
		pager := client.NewListByLocalRulestacksPager(rg, stack, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					resources = append(resources, terraformutils.NewSimpleResource(*item.ID, *item.Name, "azurerm_palo_alto_local_rulestack_certificate", g.ProviderName, []string{}))
				}
			}
		}
	}
	if client, err := armpanngfw.NewFqdnListLocalRulestackClient(sub, cred, opts); err == nil {
		pager := client.NewListByLocalRulestacksPager(rg, stack, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					resources = append(resources, terraformutils.NewSimpleResource(*item.ID, *item.Name, "azurerm_palo_alto_local_rulestack_fqdn_list", g.ProviderName, []string{}))
				}
			}
		}
	}
	if client, err := armpanngfw.NewPrefixListLocalRulestackClient(sub, cred, opts); err == nil {
		pager := client.NewListByLocalRulestacksPager(rg, stack, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					resources = append(resources, terraformutils.NewSimpleResource(*item.ID, *item.Name, "azurerm_palo_alto_local_rulestack_prefix_list", g.ProviderName, []string{}))
				}
			}
		}
	}
	if client, err := armpanngfw.NewLocalRulesClient(sub, cred, opts); err == nil {
		pager := client.NewListByLocalRulestacksPager(rg, stack, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					resources = append(resources, terraformutils.NewSimpleResource(*item.ID, *item.Name, "azurerm_palo_alto_local_rulestack_rule", g.ProviderName, []string{}))
				}
			}
		}
	}
	return resources
}

func (g *PaloAltoGenerator) listFirewalls() ([]*armpanngfw.FirewallResource, error) {
	client, err := armpanngfw.NewFirewallsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	var out []*armpanngfw.FirewallResource
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return out, err
			}
			out = append(out, page.Value...)
		}
	} else {
		pager := client.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return out, err
			}
			out = append(out, page.Value...)
		}
	}
	return out, nil
}

func (g *PaloAltoGenerator) InitResources() error {
	stacks, err := g.listLocalRulestacks()
	if err != nil {
		return err
	}
	for _, s := range stacks {
		if s.ID == nil || s.Name == nil {
			continue
		}
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(*s.ID, *s.Name, "azurerm_palo_alto_local_rulestack", g.ProviderName, []string{}))
		parsed, err := ParseAzureResourceID(*s.ID)
		if err != nil {
			return err
		}
		g.Resources = append(g.Resources, g.listRulestackChildren(parsed.ResourceGroup, *s.Name)...)
	}
	firewalls, err := g.listFirewalls()
	if err != nil {
		return err
	}
	for _, fw := range firewalls {
		if fw.ID == nil || fw.Name == nil {
			continue
		}
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(*fw.ID, *fw.Name, classifyPaloFirewall(fw.Properties), g.ProviderName, []string{}))
	}
	return nil
}
