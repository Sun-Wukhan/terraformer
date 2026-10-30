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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/frontdoor/armfrontdoor"
)

// FrontDoorGenerator covers Front Door (classic): Microsoft.Network
// frontDoors via the track-2 armfrontdoor module, its rules engines and
// per-frontend-endpoint custom HTTPS configurations, plus the
// Microsoft.Network frontDoorWebApplicationFirewallPolicies objects that
// back both azurerm_frontdoor_firewall_policy (Classic_AzureFrontDoor
// SKU) and azurerm_cdn_frontdoor_firewall_policy (Standard/Premium
// AzureFrontDoor SKUs) — the SKU on each policy decides its azurerm type.
type FrontDoorGenerator struct {
	AzureService
}

func (g *FrontDoorGenerator) listFrontDoors() ([]netParent, error) {
	ctx := context.Background()
	var parents []netParent
	client, err := armfrontdoor.NewFrontDoorsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return parents, err
	}
	handle := func(items []*armfrontdoor.FrontDoor) {
		for _, item := range items {
			if item.ID != nil && item.Name != nil {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_frontdoor")
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

func (g *FrontDoorGenerator) listRulesEngines(parents []netParent) error {
	ctx := context.Background()
	client, err := armfrontdoor.NewRulesEnginesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	for _, p := range parents {
		pager := client.NewListByFrontDoorPager(p.ResourceGroup, p.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_frontdoor_rules_engine")
				}
			}
		}
	}
	return nil
}

// Custom HTTPS configuration is a singleton settings object on each
// frontend endpoint (no ARM List of its own); its importer ID is
// {frontDoorId}/customHttpsConfiguration/{frontendEndpointName}, verified
// in the provider's Import documentation.
func (g *FrontDoorGenerator) listCustomHTTPSConfigurations(parents []netParent) error {
	ctx := context.Background()
	client, err := armfrontdoor.NewFrontendEndpointsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	for _, p := range parents {
		pager := client.NewListByFrontDoorPager(p.ResourceGroup, p.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				if item.Name != nil {
					g.AppendSimpleResource(p.ID+"/customHttpsConfiguration/"+*item.Name, *item.Name, "azurerm_frontdoor_custom_https_configuration")
				}
			}
		}
	}
	return nil
}

func (g *FrontDoorGenerator) listFirewallPolicies() error {
	ctx := context.Background()
	// The armfrontdoor PoliciesClient lists the same Microsoft.Network
	// frontDoorWebApplicationFirewallPolicies resources and, unlike the
	// armnetwork v1 model, surfaces the SKU that distinguishes the two
	// azurerm types.
	client, err := armfrontdoor.NewPoliciesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	handle := func(items []*armfrontdoor.WebApplicationFirewallPolicy) {
		for _, item := range items {
			if item.ID == nil || item.Name == nil {
				continue
			}
			tfType := "azurerm_frontdoor_firewall_policy"
			if item.SKU != nil && item.SKU.Name != nil &&
				strings.Contains(string(*item.SKU.Name), "AzureFrontDoor") &&
				!strings.Contains(string(*item.SKU.Name), "Classic") {
				tfType = "azurerm_cdn_frontdoor_firewall_policy"
			}
			g.AppendSimpleResource(*item.ID, *item.Name, tfType)
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
		pager := client.NewListBySubscriptionPager(nil)
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

func (g *FrontDoorGenerator) InitResources() error {
	frontDoors, err := g.listFrontDoors()
	if err != nil {
		return err
	}
	if err := g.listRulesEngines(frontDoors); err != nil {
		return err
	}
	if err := g.listCustomHTTPSConfigurations(frontDoors); err != nil {
		return err
	}
	return g.listFirewallPolicies()
}
