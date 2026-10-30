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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/eventgrid/armeventgrid"
)

// EventGridGenerator covers Event Grid domains, topics, system topics and
// event subscriptions.
// Skipped in this family (no client/List in the track-2 armeventgrid
// v1.0.0 module, which predates those resource types):
//   - azurerm_eventgrid_namespace and azurerm_eventgrid_namespace_topic
//   - azurerm_eventgrid_partner_configuration, azurerm_eventgrid_partner_namespace
//     and azurerm_eventgrid_partner_registration
type EventGridGenerator struct {
	AzureService
}

func (g *EventGridGenerator) InitResources() error {
	ctx := context.Background()
	sub := g.getSubscriptionID()
	cred := g.getCredential()
	opts := g.armClientOptions()
	rgFilter := g.getResourceGroup()

	domainsClient, err := armeventgrid.NewDomainsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	var domains []*armeventgrid.Domain
	if rgFilter != "" {
		pager := domainsClient.NewListByResourceGroupPager(rgFilter, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			domains = append(domains, page.Value...)
		}
	} else {
		pager := domainsClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			domains = append(domains, page.Value...)
		}
	}
	domainTopicsClient, err := armeventgrid.NewDomainTopicsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	for _, domain := range domains {
		g.AppendSimpleResource(*domain.ID, *domain.Name, "azurerm_eventgrid_domain")
		parsed, err := ParseAzureResourceID(*domain.ID)
		if err != nil {
			return err
		}
		pager := domainTopicsClient.NewListByDomainPager(parsed.ResourceGroup, *domain.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_eventgrid_domain_topic")
			}
		}
	}

	topicsClient, err := armeventgrid.NewTopicsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	if rgFilter != "" {
		pager := topicsClient.NewListByResourceGroupPager(rgFilter, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_eventgrid_topic")
			}
		}
	} else {
		pager := topicsClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_eventgrid_topic")
			}
		}
	}

	systemTopicsClient, err := armeventgrid.NewSystemTopicsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	var systemTopics []*armeventgrid.SystemTopic
	if rgFilter != "" {
		pager := systemTopicsClient.NewListByResourceGroupPager(rgFilter, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			systemTopics = append(systemTopics, page.Value...)
		}
	} else {
		pager := systemTopicsClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			systemTopics = append(systemTopics, page.Value...)
		}
	}
	sysSubClient, err := armeventgrid.NewSystemTopicEventSubscriptionsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	for _, topic := range systemTopics {
		g.AppendSimpleResource(*topic.ID, *topic.Name, "azurerm_eventgrid_system_topic")
		parsed, err := ParseAzureResourceID(*topic.ID)
		if err != nil {
			return err
		}
		pager := sysSubClient.NewListBySystemTopicPager(parsed.ResourceGroup, *topic.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_eventgrid_system_topic_event_subscription")
			}
		}
	}

	subsClient, err := armeventgrid.NewEventSubscriptionsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	if rgFilter != "" {
		pager := subsClient.NewListGlobalByResourceGroupPager(rgFilter, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_eventgrid_event_subscription")
			}
		}
	} else {
		pager := subsClient.NewListGlobalBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_eventgrid_event_subscription")
			}
		}
	}
	return nil
}
