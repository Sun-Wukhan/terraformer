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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/servicebus/armservicebus"
)

// ServiceBusGenerator covers Service Bus namespaces, queues, topics,
// topic subscriptions and their authorization rules / disaster recovery
// aliases.
// Skipped in this family:
//   - azurerm_servicebus_namespace_customer_managed_key: an embedded
//     property (properties.encryption) of the namespace itself, with no
//     independent ARM object or List.
type ServiceBusGenerator struct {
	AzureService
}

func (g *ServiceBusGenerator) InitResources() error {
	ctx := context.Background()
	sub := g.getSubscriptionID()
	cred := g.getCredential()
	opts := g.armClientOptions()

	nsClient, err := armservicebus.NewNamespacesClient(sub, cred, opts)
	if err != nil {
		return err
	}
	var namespaces []*armservicebus.SBNamespace
	if rg := g.getResourceGroup(); rg != "" {
		pager := nsClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			namespaces = append(namespaces, page.Value...)
		}
	} else {
		pager := nsClient.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			namespaces = append(namespaces, page.Value...)
		}
	}

	queuesClient, err := armservicebus.NewQueuesClient(sub, cred, opts)
	if err != nil {
		return err
	}
	topicsClient, err := armservicebus.NewTopicsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	subsClient, err := armservicebus.NewSubscriptionsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	rulesClient, err := armservicebus.NewRulesClient(sub, cred, opts)
	if err != nil {
		return err
	}
	drClient, err := armservicebus.NewDisasterRecoveryConfigsClient(sub, cred, opts)
	if err != nil {
		return err
	}

	for _, ns := range namespaces {
		g.AppendSimpleResource(*ns.ID, *ns.Name, "azurerm_servicebus_namespace")
		parsed, err := ParseAzureResourceID(*ns.ID)
		if err != nil {
			return err
		}
		rg := parsed.ResourceGroup
		nsName := *ns.Name

		authPager := nsClient.NewListAuthorizationRulesPager(rg, nsName, nil)
		for authPager.More() {
			page, err := authPager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_servicebus_namespace_authorization_rule")
			}
		}

		drPager := drClient.NewListPager(rg, nsName, nil)
		for drPager.More() {
			page, err := drPager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_servicebus_namespace_disaster_recovery_config")
			}
		}

		queuePager := queuesClient.NewListByNamespacePager(rg, nsName, nil)
		for queuePager.More() {
			page, err := queuePager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, queue := range page.Value {
				g.AppendSimpleResource(*queue.ID, *queue.Name, "azurerm_servicebus_queue")
				qAuthPager := queuesClient.NewListAuthorizationRulesPager(rg, nsName, *queue.Name, nil)
				for qAuthPager.More() {
					qPage, err := qAuthPager.NextPage(ctx)
					if err != nil {
						return err
					}
					for _, item := range qPage.Value {
						g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_servicebus_queue_authorization_rule")
					}
				}
			}
		}

		topicPager := topicsClient.NewListByNamespacePager(rg, nsName, nil)
		for topicPager.More() {
			page, err := topicPager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, topic := range page.Value {
				g.AppendSimpleResource(*topic.ID, *topic.Name, "azurerm_servicebus_topic")
				tAuthPager := topicsClient.NewListAuthorizationRulesPager(rg, nsName, *topic.Name, nil)
				for tAuthPager.More() {
					tPage, err := tAuthPager.NextPage(ctx)
					if err != nil {
						return err
					}
					for _, item := range tPage.Value {
						g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_servicebus_topic_authorization_rule")
					}
				}
				subPager := subsClient.NewListByTopicPager(rg, nsName, *topic.Name, nil)
				for subPager.More() {
					sPage, err := subPager.NextPage(ctx)
					if err != nil {
						return err
					}
					for _, sbSub := range sPage.Value {
						g.AppendSimpleResource(*sbSub.ID, *sbSub.Name, "azurerm_servicebus_subscription")
						rulePager := rulesClient.NewListBySubscriptionsPager(rg, nsName, *topic.Name, *sbSub.Name, nil)
						for rulePager.More() {
							rPage, err := rulePager.NextPage(ctx)
							if err != nil {
								return err
							}
							for _, item := range rPage.Value {
								g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_servicebus_subscription_rule")
							}
						}
					}
				}
			}
		}
	}
	return nil
}
