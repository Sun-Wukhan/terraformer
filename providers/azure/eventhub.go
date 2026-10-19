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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/eventhub/armeventhub"
)

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
			az.AppendSimpleResource(*item.ID, *item.Name, "azurerm_eventhub_consumer_group")
		}
	}
	return nil
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

func (az *EventHubGenerator) InitResources() error {

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
	}
	return nil
}
