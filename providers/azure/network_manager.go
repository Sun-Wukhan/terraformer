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

type NetworkManagerGenerator struct {
	AzureService
}

// Skipped in this generator (no usable track-2 List surface):
//   - azurerm_network_manager_deployment: side-effect-only commit;
//     ManagerCommitsClient exposes only BeginPost (no List/Get) and its
//     importer ID is the synthetic {managerId}/commit|{location}|{scopeAccess}.
//   - azurerm_network_manager_ipam_pool / _ipam_pool_static_cidr: no IPAM
//     pool client exists in armnetwork v1/v2/v4.
//   - azurerm_network_manager_routing_configuration / _routing_rule /
//     _routing_rule_collection: no routing configuration client exists in
//     armnetwork v1/v2/v4.
//   - azurerm_network_manager_verifier_workspace /
//     _verifier_workspace_reachability_analysis_intent: no verifier
//     workspace client exists in armnetwork v1/v2/v4.
//   - azurerm_network_manager_management_group_connection: its only List
//     (ManagementGroupNetworkManagerConnectionsClient) is keyed by a
//     management group ID, a tenant-scope parent that cannot be enumerated
//     from this subscription-scoped generator.

func (g *NetworkManagerGenerator) listNetworkManagers() ([]netParent, error) {
	ctx := context.Background()
	var parents []netParent
	client, err := armnetwork.NewManagersClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return parents, err
	}
	handle := func(items []*armnetwork.Manager) {
		for _, item := range items {
			if item.ID != nil && item.Name != nil {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_network_manager")
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
		pager := client.NewListBySubscriptionPager(nil)
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

func (g *NetworkManagerGenerator) listConnectivityConfigurations(parents []netParent) error {
	ctx := context.Background()
	client, err := armnetwork.NewConnectivityConfigurationsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
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
					g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_network_manager_connectivity_configuration")
				}
			}
		}
	}
	return nil
}

func (g *NetworkManagerGenerator) listSecurityAdminConfigurations(parents []netParent) ([]netParent, error) {
	var out []netParent
	ctx := context.Background()
	client, err := armnetwork.NewSecurityAdminConfigurationsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return out, err
	}
	for _, p := range parents {
		pager := client.NewListPager(p.ResourceGroup, p.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return out, err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_network_manager_security_admin_configuration")
					if cp, ok := makeParent(item.ID, item.Name); ok {
						cp.ParentName = p.Name
						cp.GrandparentName = p.ParentName
						out = append(out, cp)
					}
				}
			}
		}
	}
	return out, nil
}

func (g *NetworkManagerGenerator) listAdminRuleCollections(parents []netParent) ([]netParent, error) {
	var out []netParent
	ctx := context.Background()
	client, err := armnetwork.NewAdminRuleCollectionsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return out, err
	}
	for _, p := range parents {
		pager := client.NewListPager(p.ResourceGroup, p.ParentName, p.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return out, err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_network_manager_admin_rule_collection")
					if cp, ok := makeParent(item.ID, item.Name); ok {
						cp.ParentName = p.Name
						cp.GrandparentName = p.ParentName
						out = append(out, cp)
					}
				}
			}
		}
	}
	return out, nil
}

func (g *NetworkManagerGenerator) listAdminRules(parents []netParent) error {
	ctx := context.Background()
	client, err := armnetwork.NewAdminRulesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	for _, p := range parents {
		pager := client.NewListPager(p.ResourceGroup, p.GrandparentName, p.ParentName, p.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				// The List returns the polymorphic BaseAdminRuleClassification:
				// custom rules come back as *AdminRule (these are the
				// azurerm_network_manager_admin_rule objects); built-in
				// *DefaultAdminRule entries are not that resource.
				if rule, ok := item.(*armnetwork.AdminRule); ok && rule.ID != nil && rule.Name != nil {
					g.AppendSimpleResource(*rule.ID, *rule.Name, "azurerm_network_manager_admin_rule")
				}
			}
		}
	}
	return nil
}

func (g *NetworkManagerGenerator) listNetworkGroups(parents []netParent) ([]netParent, error) {
	var out []netParent
	ctx := context.Background()
	client, err := armnetwork.NewGroupsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return out, err
	}
	for _, p := range parents {
		pager := client.NewListPager(p.ResourceGroup, p.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return out, err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_network_manager_network_group")
					if cp, ok := makeParent(item.ID, item.Name); ok {
						cp.ParentName = p.Name
						cp.GrandparentName = p.ParentName
						out = append(out, cp)
					}
				}
			}
		}
	}
	return out, nil
}

func (g *NetworkManagerGenerator) listStaticMembers(parents []netParent) error {
	ctx := context.Background()
	client, err := armnetwork.NewStaticMembersClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	for _, p := range parents {
		pager := client.NewListPager(p.ResourceGroup, p.ParentName, p.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_network_manager_static_member")
				}
			}
		}
	}
	return nil
}

func (g *NetworkManagerGenerator) listScopeConnections(parents []netParent) error {
	ctx := context.Background()
	client, err := armnetwork.NewScopeConnectionsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
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
					g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_network_manager_scope_connection")
				}
			}
		}
	}
	return nil
}

func (g *NetworkManagerGenerator) listSubscriptionConnections() error {
	ctx := context.Background()
	client, err := armnetwork.NewSubscriptionNetworkManagerConnectionsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	pager := client.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			if item.ID != nil && item.Name != nil {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_network_manager_subscription_connection")
			}
		}
	}
	return nil
}

func (g *NetworkManagerGenerator) InitResources() error {
	if err := func() error {
		managers, err := g.listNetworkManagers()
		if err != nil {
			return err
		}
		if err := g.listConnectivityConfigurations(managers); err != nil {
			return err
		}
		configs, err := g.listSecurityAdminConfigurations(managers)
		if err != nil {
			return err
		}
		collections, err := g.listAdminRuleCollections(configs)
		if err != nil {
			return err
		}
		if err := g.listAdminRules(collections); err != nil {
			return err
		}
		groups, err := g.listNetworkGroups(managers)
		if err != nil {
			return err
		}
		if err := g.listStaticMembers(groups); err != nil {
			return err
		}
		return g.listScopeConnections(managers)
	}(); err != nil {
		return err
	}
	if err := g.listSubscriptionConnections(); err != nil {
		return err
	}
	return nil
}
