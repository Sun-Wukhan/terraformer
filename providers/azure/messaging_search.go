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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/notificationhubs/armnotificationhubs"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/relay/armrelay"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/search/armsearch"
)

// NotificationHubGenerator covers Microsoft.NotificationHubs namespaces,
// notification hubs and hub authorization rules.
type NotificationHubGenerator struct {
	AzureService
}

func (g *NotificationHubGenerator) InitResources() error {
	ctx := context.Background()
	namespacesClient, err := armnotificationhubs.NewNamespacesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	var namespaces []*armnotificationhubs.NamespaceResource
	if rg := g.getResourceGroup(); rg != "" {
		pager := namespacesClient.NewListPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			namespaces = append(namespaces, page.Value...)
		}
	} else {
		pager := namespacesClient.NewListAllPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			namespaces = append(namespaces, page.Value...)
		}
	}

	hubsClient, err := armnotificationhubs.NewClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}

	for _, namespace := range namespaces {
		g.AppendSimpleResource(*namespace.ID, *namespace.Name, "azurerm_notification_hub_namespace")
		parsed, err := ParseAzureResourceID(*namespace.ID)
		if err != nil {
			return err
		}
		rg, namespaceName := parsed.ResourceGroup, *namespace.Name

		hubPager := hubsClient.NewListPager(rg, namespaceName, nil)
		for hubPager.More() {
			page, err := hubPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, hub := range page.Value {
				g.AppendSimpleResource(*hub.ID, *hub.Name, "azurerm_notification_hub")

				rulePager := hubsClient.NewListAuthorizationRulesPager(rg, namespaceName, *hub.Name, nil)
				for rulePager.More() {
					rulePage, err := rulePager.NextPage(ctx)
					if err != nil {
						log.Println(err)
						return err
					}
					for _, rule := range rulePage.Value {
						g.AppendSimpleResource(*rule.ID, *rule.Name, "azurerm_notification_hub_authorization_rule")
					}
				}
			}
		}
	}
	return nil
}

// RelayGenerator covers Microsoft.Relay namespaces, their authorization
// rules, hybrid connections and hybrid-connection authorization rules.
type RelayGenerator struct {
	AzureService
}

func (g *RelayGenerator) InitResources() error {
	ctx := context.Background()
	namespacesClient, err := armrelay.NewNamespacesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	var namespaces []*armrelay.Namespace
	if rg := g.getResourceGroup(); rg != "" {
		pager := namespacesClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			namespaces = append(namespaces, page.Value...)
		}
	} else {
		pager := namespacesClient.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			namespaces = append(namespaces, page.Value...)
		}
	}

	hybridClient, err := armrelay.NewHybridConnectionsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}

	for _, namespace := range namespaces {
		g.AppendSimpleResource(*namespace.ID, *namespace.Name, "azurerm_relay_namespace")
		parsed, err := ParseAzureResourceID(*namespace.ID)
		if err != nil {
			return err
		}
		rg, namespaceName := parsed.ResourceGroup, *namespace.Name

		rulePager := namespacesClient.NewListAuthorizationRulesPager(rg, namespaceName, nil)
		for rulePager.More() {
			page, err := rulePager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, rule := range page.Value {
				g.AppendSimpleResource(*rule.ID, *rule.Name, "azurerm_relay_namespace_authorization_rule")
			}
		}

		hybridPager := hybridClient.NewListByNamespacePager(rg, namespaceName, nil)
		for hybridPager.More() {
			page, err := hybridPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, hybrid := range page.Value {
				g.AppendSimpleResource(*hybrid.ID, *hybrid.Name, "azurerm_relay_hybrid_connection")

				hybridRulePager := hybridClient.NewListAuthorizationRulesPager(rg, namespaceName, *hybrid.Name, nil)
				for hybridRulePager.More() {
					rulePage, err := hybridRulePager.NextPage(ctx)
					if err != nil {
						log.Println(err)
						return err
					}
					for _, rule := range rulePage.Value {
						g.AppendSimpleResource(*rule.ID, *rule.Name, "azurerm_relay_hybrid_connection_authorization_rule")
					}
				}
			}
		}
	}
	return nil
}

// SearchGenerator covers Microsoft.Search search services and their shared
// private link resources.
type SearchGenerator struct {
	AzureService
}

func (g *SearchGenerator) InitResources() error {
	ctx := context.Background()
	servicesClient, err := armsearch.NewServicesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	var services []*armsearch.Service
	if rg := g.getResourceGroup(); rg != "" {
		pager := servicesClient.NewListByResourceGroupPager(rg, nil, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			services = append(services, page.Value...)
		}
	} else {
		pager := servicesClient.NewListBySubscriptionPager(nil, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			services = append(services, page.Value...)
		}
	}

	splClient, err := armsearch.NewSharedPrivateLinkResourcesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	for _, service := range services {
		g.AppendSimpleResource(*service.ID, *service.Name, "azurerm_search_service")
		parsed, err := ParseAzureResourceID(*service.ID)
		if err != nil {
			return err
		}
		splPager := splClient.NewListByServicePager(parsed.ResourceGroup, *service.Name, nil, nil)
		for splPager.More() {
			page, err := splPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_search_shared_private_link_service")
			}
		}
	}
	return nil
}
