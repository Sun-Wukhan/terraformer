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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cdn/armcdn"
)

// CDNGenerator covers both classic CDN (Microsoft.Cdn/profiles with CDN
// SKUs) and Azure Front Door Standard/Premium (the same profiles resource
// with AzureFrontDoor SKUs, children under afdEndpoints / originGroups /
// ruleSets). The SKU on each profile decides which azurerm type it maps to
// and which child enumerations apply.
//
// Skipped in this generator:
//   - azurerm_cdn_frontdoor_batch_rule_set: it manages the exact same ARM
//     object and importer ID as azurerm_cdn_frontdoor_rule_set
//     (Microsoft.Cdn/profiles/{profile}/ruleSets/{name}); batch mode is a
//     provider-side write strategy that the ruleSets List cannot
//     distinguish, so emitting both would duplicate every rule set.
//   - azurerm_cdn_frontdoor_custom_domain_association: association glue;
//     armcdn has no Associations client/List — the association is stored
//     as the customDomains references on routes (the importer ID's
//     associations/{name} segment has no listable ARM counterpart).
type CDNGenerator struct {
	AzureService
}

func isFrontDoorProfile(profile *armcdn.Profile) bool {
	return profile.SKU != nil && profile.SKU.Name != nil &&
		strings.Contains(string(*profile.SKU.Name), "AzureFrontDoor")
}

func (g *CDNGenerator) listProfiles() (cdnProfiles, afdProfiles []netParent, err error) {
	ctx := context.Background()
	client, err := armcdn.NewProfilesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, nil, err
	}
	handle := func(items []*armcdn.Profile) {
		for _, item := range items {
			if item.ID == nil || item.Name == nil {
				continue
			}
			p, ok := makeParent(item.ID, item.Name)
			if !ok {
				continue
			}
			if isFrontDoorProfile(item) {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_cdn_frontdoor_profile")
				afdProfiles = append(afdProfiles, p)
			} else {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_cdn_profile")
				cdnProfiles = append(cdnProfiles, p)
			}
		}
	}
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, nil, err
			}
			handle(page.Value)
		}
	} else {
		pager := client.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, nil, err
			}
			handle(page.Value)
		}
	}
	return cdnProfiles, afdProfiles, nil
}

func (g *CDNGenerator) listCDNEndpoints(profiles []netParent) ([]netParent, error) {
	ctx := context.Background()
	var out []netParent
	client, err := armcdn.NewEndpointsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return out, err
	}
	for _, p := range profiles {
		pager := client.NewListByProfilePager(p.ResourceGroup, p.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return out, err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_cdn_endpoint")
					if cp, ok := makeParent(item.ID, item.Name); ok {
						cp.ParentName = p.Name
						out = append(out, cp)
					}
				}
			}
		}
	}
	return out, nil
}

func (g *CDNGenerator) listCDNEndpointCustomDomains(endpoints []netParent) error {
	ctx := context.Background()
	client, err := armcdn.NewCustomDomainsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	for _, p := range endpoints {
		pager := client.NewListByEndpointPager(p.ResourceGroup, p.ParentName, p.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_cdn_endpoint_custom_domain")
				}
			}
		}
	}
	return nil
}

func (g *CDNGenerator) listAFDEndpoints(profiles []netParent) ([]netParent, error) {
	ctx := context.Background()
	var out []netParent
	client, err := armcdn.NewAFDEndpointsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return out, err
	}
	for _, p := range profiles {
		pager := client.NewListByProfilePager(p.ResourceGroup, p.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return out, err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_cdn_frontdoor_endpoint")
					if cp, ok := makeParent(item.ID, item.Name); ok {
						cp.ParentName = p.Name
						out = append(out, cp)
					}
				}
			}
		}
	}
	return out, nil
}

func (g *CDNGenerator) listRoutes(endpoints []netParent) error {
	ctx := context.Background()
	client, err := armcdn.NewRoutesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	for _, p := range endpoints {
		pager := client.NewListByEndpointPager(p.ResourceGroup, p.ParentName, p.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_cdn_frontdoor_route")
				}
			}
		}
	}
	return nil
}

func (g *CDNGenerator) listAFDOriginGroups(profiles []netParent) ([]netParent, error) {
	ctx := context.Background()
	var out []netParent
	client, err := armcdn.NewAFDOriginGroupsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return out, err
	}
	for _, p := range profiles {
		pager := client.NewListByProfilePager(p.ResourceGroup, p.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return out, err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_cdn_frontdoor_origin_group")
					if cp, ok := makeParent(item.ID, item.Name); ok {
						cp.ParentName = p.Name
						out = append(out, cp)
					}
				}
			}
		}
	}
	return out, nil
}

func (g *CDNGenerator) listAFDOrigins(groups []netParent) error {
	ctx := context.Background()
	client, err := armcdn.NewAFDOriginsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	for _, p := range groups {
		pager := client.NewListByOriginGroupPager(p.ResourceGroup, p.ParentName, p.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_cdn_frontdoor_origin")
				}
			}
		}
	}
	return nil
}

func (g *CDNGenerator) listAFDCustomDomains(profiles []netParent) error {
	ctx := context.Background()
	client, err := armcdn.NewAFDCustomDomainsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	for _, p := range profiles {
		pager := client.NewListByProfilePager(p.ResourceGroup, p.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_cdn_frontdoor_custom_domain")
				}
			}
		}
	}
	return nil
}

func (g *CDNGenerator) listRuleSets(profiles []netParent) ([]netParent, error) {
	ctx := context.Background()
	var out []netParent
	client, err := armcdn.NewRuleSetsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return out, err
	}
	for _, p := range profiles {
		pager := client.NewListByProfilePager(p.ResourceGroup, p.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return out, err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_cdn_frontdoor_rule_set")
					if cp, ok := makeParent(item.ID, item.Name); ok {
						cp.ParentName = p.Name
						out = append(out, cp)
					}
				}
			}
		}
	}
	return out, nil
}

func (g *CDNGenerator) listRules(ruleSets []netParent) error {
	ctx := context.Background()
	client, err := armcdn.NewRulesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	for _, p := range ruleSets {
		pager := client.NewListByRuleSetPager(p.ResourceGroup, p.ParentName, p.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_cdn_frontdoor_rule")
				}
			}
		}
	}
	return nil
}

func (g *CDNGenerator) listSecrets(profiles []netParent) error {
	ctx := context.Background()
	client, err := armcdn.NewSecretsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	for _, p := range profiles {
		pager := client.NewListByProfilePager(p.ResourceGroup, p.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_cdn_frontdoor_secret")
				}
			}
		}
	}
	return nil
}

func (g *CDNGenerator) listSecurityPolicies(profiles []netParent) error {
	ctx := context.Background()
	client, err := armcdn.NewSecurityPoliciesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	for _, p := range profiles {
		pager := client.NewListByProfilePager(p.ResourceGroup, p.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_cdn_frontdoor_security_policy")
				}
			}
		}
	}
	return nil
}

func (g *CDNGenerator) InitResources() error {
	cdnProfiles, afdProfiles, err := g.listProfiles()
	if err != nil {
		return err
	}
	endpoints, err := g.listCDNEndpoints(cdnProfiles)
	if err != nil {
		return err
	}
	if err := g.listCDNEndpointCustomDomains(endpoints); err != nil {
		return err
	}
	afdEndpoints, err := g.listAFDEndpoints(afdProfiles)
	if err != nil {
		return err
	}
	if err := g.listRoutes(afdEndpoints); err != nil {
		return err
	}
	groups, err := g.listAFDOriginGroups(afdProfiles)
	if err != nil {
		return err
	}
	if err := g.listAFDOrigins(groups); err != nil {
		return err
	}
	if err := g.listAFDCustomDomains(afdProfiles); err != nil {
		return err
	}
	ruleSets, err := g.listRuleSets(afdProfiles)
	if err != nil {
		return err
	}
	if err := g.listRules(ruleSets); err != nil {
		return err
	}
	if err := g.listSecrets(afdProfiles); err != nil {
		return err
	}
	return g.listSecurityPolicies(afdProfiles)
}
