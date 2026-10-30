// Copyright 2021 The Terraformer Authors.
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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/eventhub/armeventhub"
)

// EventHubGenerator covers Event Hubs namespaces, hubs, consumer groups,
// authorization rules, dedicated clusters, disaster recovery aliases and
// schema groups.
// Skipped in this family:
//   - azurerm_eventhub_namespace_customer_managed_key: an embedded property
//     (properties.encryption) of the namespace itself, with no independent
//     ARM object or List.
type EventHubGenerator struct {
	AzureService
}

func (az *EventHubGenerator) listNamespaces() ([]*armeventhub.EHNamespace, error) {
	client, err := armeventhub.NewNamespacesClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	var resources []*armeventhub.EHNamespace
	if rg := az.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return resources, err
			}
			resources = append(resources, page.Value...)
		}
	} else {
		pager := client.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return resources, err
			}
			resources = append(resources, page.Value...)
		}
	}
	return resources, nil
}

func (az *EventHubGenerator) AppendNamespace(namespace *armeventhub.EHNamespace) {
	az.AppendSimpleResource(*namespace.ID, *namespace.Name, "azurerm_eventhub_namespace")
}

func (az *EventHubGenerator) appendEventHubs(namespace *armeventhub.EHNamespace, namespaceRg *ResourceID) error {
	client, err := armeventhub.NewEventHubsClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
	if err != nil {
		return err
	}
	ctx := context.Background()
	pager := client.NewListByNamespacePager(namespaceRg.ResourceGroup, *namespace.Name, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			az.AppendSimpleResource(*item.ID, *item.Name, "azurerm_eventhub")
			err = az.appendConsumerGroups(namespace, namespaceRg, *item.Name)
			if err != nil {
				return err
			}
			err = az.appendEventHubAuthorizationRules(namespace, namespaceRg, *item.Name)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (az *EventHubGenerator) appendConsumerGroups(namespace *armeventhub.EHNamespace, namespaceRg *ResourceID, eventHubName string) error {
	client, err := armeventhub.NewConsumerGroupsClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
	if err != nil {
		return err
	}
	ctx := context.Background()
	pager := client.NewListByEventHubPager(namespaceRg.ResourceGroup, *namespace.Name, eventHubName, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			if isDefaultConsumerGroup(item.Name) {
				continue
			}
			az.AppendSimpleResource(*item.ID, *item.Name, "azurerm_eventhub_consumer_group")
		}
	}
	return nil
}

// isDefaultConsumerGroup reports whether name is the `$Default` consumer
// group Azure creates in every event hub. It cannot be managed and azurerm
// rejects the name (it must start with a letter or digit), so importing it
// only yields invalid configuration.
func isDefaultConsumerGroup(name *string) bool {
	return name == nil || strings.EqualFold(*name, "$Default")
}

func (az *EventHubGenerator) appendAuthorizationRules(namespace *armeventhub.EHNamespace, namespaceRg *ResourceID) error {
	client, err := armeventhub.NewNamespacesClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
	if err != nil {
		return err
	}
	ctx := context.Background()
	pager := client.NewListAuthorizationRulesPager(namespaceRg.ResourceGroup, *namespace.Name, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			az.AppendSimpleResource(*item.ID, *item.Name, "azurerm_eventhub_namespace_authorization_rule")
		}
	}
	return nil
}

func (az *EventHubGenerator) appendEventHubAuthorizationRules(namespace *armeventhub.EHNamespace, namespaceRg *ResourceID, eventHubName string) error {
	client, err := armeventhub.NewEventHubsClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
	if err != nil {
		return err
	}
	ctx := context.Background()
	pager := client.NewListAuthorizationRulesPager(namespaceRg.ResourceGroup, *namespace.Name, eventHubName, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			az.AppendSimpleResource(*item.ID, *item.Name, "azurerm_eventhub_authorization_rule")
		}
	}
	return nil
}

func (az *EventHubGenerator) appendNamespaceChildren(namespace *armeventhub.EHNamespace, namespaceRg *ResourceID) error {
	ctx := context.Background()
	drClient, err := armeventhub.NewDisasterRecoveryConfigsClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
	if err != nil {
		return err
	}
	pager := drClient.NewListPager(namespaceRg.ResourceGroup, *namespace.Name, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			az.AppendSimpleResource(*item.ID, *item.Name, "azurerm_eventhub_namespace_disaster_recovery_config")
		}
	}
	// Basic namespaces have no schema registry; listing it returns a 400.
	if namespace.SKU != nil && namespace.SKU.Name != nil && *namespace.SKU.Name == armeventhub.SKUNameBasic {
		return nil
	}
	schemaClient, err := armeventhub.NewSchemaRegistryClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
	if err != nil {
		return err
	}
	schemaPager := schemaClient.NewListByNamespacePager(namespaceRg.ResourceGroup, *namespace.Name, nil)
	for schemaPager.More() {
		page, err := schemaPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			az.AppendSimpleResource(*item.ID, *item.Name, "azurerm_eventhub_namespace_schema_group")
		}
	}
	return nil
}

func (az *EventHubGenerator) appendClusters() error {
	client, err := armeventhub.NewClustersClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
	if err != nil {
		return err
	}
	ctx := context.Background()
	if rg := az.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				az.AppendSimpleResource(*item.ID, *item.Name, "azurerm_eventhub_cluster")
			}
		}
		return nil
	}
	pager := client.NewListBySubscriptionPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			az.AppendSimpleResource(*item.ID, *item.Name, "azurerm_eventhub_cluster")
		}
	}
	return nil
}

func (az *EventHubGenerator) InitResources() error {
	if err := az.appendClusters(); err != nil {
		return err
	}

	namespaces, err := az.listNamespaces()
	if err != nil {
		return err
	}
	for _, namespace := range namespaces {
		az.AppendNamespace(namespace)
		namespaceRg, err := ParseAzureResourceID(*namespace.ID)
		if err != nil {
			return err
		}
		err = az.appendEventHubs(namespace, namespaceRg)
		if err != nil {
			return err
		}
		err = az.appendAuthorizationRules(namespace, namespaceRg)
		if err != nil {
			return err
		}
		err = az.appendNamespaceChildren(namespace, namespaceRg)
		if err != nil {
			return err
		}
	}
	return nil
}
