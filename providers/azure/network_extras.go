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

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// ApplicationSecurityGroupGenerator imports application security groups
// via the track-2 armnetwork SDK.
type ApplicationSecurityGroupGenerator struct {
	AzureService
}

func (g *ApplicationSecurityGroupGenerator) InitResources() error {
	ctx := context.Background()
	client, err := armnetwork.NewApplicationSecurityGroupsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	var pager *runtime.Pager[armnetwork.ApplicationSecurityGroupsClientListAllResponse]
	if rg := g.getResourceGroup(); rg != "" {
		rgPager := client.NewListPager(rg, nil)
		resources, err := drainPager(ctx, rgPager,
			func(p armnetwork.ApplicationSecurityGroupsClientListResponse) []*armnetwork.ApplicationSecurityGroup {
				return p.Value
			},
			func(item *armnetwork.ApplicationSecurityGroup) (terraformutils.Resource, bool) {
				if item.ID == nil || item.Name == nil {
					return terraformutils.Resource{}, false
				}
				return terraformutils.NewSimpleResource(*item.ID, *item.Name, "azurerm_application_security_group", g.ProviderName, []string{}), true
			})
		if err != nil {
			return err
		}
		g.Resources = append(g.Resources, resources...)
		return nil
	}
	pager = client.NewListAllPager(nil)
	resources, err := drainPager(ctx, pager,
		func(p armnetwork.ApplicationSecurityGroupsClientListAllResponse) []*armnetwork.ApplicationSecurityGroup {
			return p.Value
		},
		func(item *armnetwork.ApplicationSecurityGroup) (terraformutils.Resource, bool) {
			if item.ID == nil || item.Name == nil {
				return terraformutils.Resource{}, false
			}
			return terraformutils.NewSimpleResource(*item.ID, *item.Name, "azurerm_application_security_group", g.ProviderName, []string{}), true
		})
	if err != nil {
		return err
	}
	g.Resources = append(g.Resources, resources...)
	return nil
}

// WebApplicationFirewallPolicyGenerator imports Web Application Firewall
// policies via the track-2 armnetwork SDK.
type WebApplicationFirewallPolicyGenerator struct {
	AzureService
}

func (g *WebApplicationFirewallPolicyGenerator) InitResources() error {
	ctx := context.Background()
	client, err := armnetwork.NewWebApplicationFirewallPoliciesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	var pager *runtime.Pager[armnetwork.WebApplicationFirewallPoliciesClientListAllResponse]
	if rg := g.getResourceGroup(); rg != "" {
		rgPager := client.NewListPager(rg, nil)
		resources, err := drainPager(ctx, rgPager,
			func(p armnetwork.WebApplicationFirewallPoliciesClientListResponse) []*armnetwork.WebApplicationFirewallPolicy {
				return p.Value
			},
			func(item *armnetwork.WebApplicationFirewallPolicy) (terraformutils.Resource, bool) {
				if item.ID == nil || item.Name == nil {
					return terraformutils.Resource{}, false
				}
				return terraformutils.NewSimpleResource(*item.ID, *item.Name, "azurerm_web_application_firewall_policy", g.ProviderName, []string{}), true
			})
		if err != nil {
			return err
		}
		g.Resources = append(g.Resources, resources...)
		return nil
	}
	pager = client.NewListAllPager(nil)
	resources, err := drainPager(ctx, pager,
		func(p armnetwork.WebApplicationFirewallPoliciesClientListAllResponse) []*armnetwork.WebApplicationFirewallPolicy {
			return p.Value
		},
		func(item *armnetwork.WebApplicationFirewallPolicy) (terraformutils.Resource, bool) {
			if item.ID == nil || item.Name == nil {
				return terraformutils.Resource{}, false
			}
			return terraformutils.NewSimpleResource(*item.ID, *item.Name, "azurerm_web_application_firewall_policy", g.ProviderName, []string{}), true
		})
	if err != nil {
		return err
	}
	g.Resources = append(g.Resources, resources...)
	return nil
}
