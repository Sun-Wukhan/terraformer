// Copyright 2020 The Terraformer Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package azure

import (
	"context"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerinstance/armcontainerinstance"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry"
	armcontainerregistryv2 "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry/v2"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// Skipped in this generator (see results-containers.md):
//   - azurerm_container_registry_token_password: credential artefact, the
//     password is generated on demand via the tokens generateCredentials
//     action; there is no List surface for it in armcontainerregistry.
//   - azurerm_container_registry_task_schedule_run_now: action artefact
//     (scheduling a task run-now is a POST action, not an ARM resource).

type ContainerGenerator struct {
	AzureService
}

func (g *ContainerGenerator) listAndAddForContainerGroup() ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	ContainerGroupsClient, err := armcontainerinstance.NewContainerGroupsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	if rg := g.getResourceGroup(); rg != "" {
		pager := ContainerGroupsClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return resources, err
			}
			for _, containerGroup := range page.Value {
				resources = append(resources, terraformutils.NewSimpleResource(
					*containerGroup.ID,
					*containerGroup.Name,
					"azurerm_container_group",
					g.ProviderName,
					[]string{}))
			}
		}
	} else {
		pager := ContainerGroupsClient.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return resources, err
			}
			for _, containerGroup := range page.Value {
				resources = append(resources, terraformutils.NewSimpleResource(
					*containerGroup.ID,
					*containerGroup.Name,
					"azurerm_container_group",
					g.ProviderName,
					[]string{}))
			}
		}
	}

	return resources, nil
}

func (g *ContainerGenerator) listRegistryWebhooks(resourceGroupName string, registryName string) ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	WebhooksClient, err := armcontainerregistry.NewWebhooksClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	pager := WebhooksClient.NewListPager(resourceGroupName, registryName, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return resources, err
		}
		for _, webhook := range page.Value {
			resources = append(resources, terraformutils.NewSimpleResource(
				*webhook.ID,
				*webhook.Name,
				"azurerm_container_registry_webhook",
				g.ProviderName,
				[]string{}))
		}
	}
	return resources, nil
}

// listRegistryChildren enumerates the ACR child resources of one registry:
// agent pools, cache rules, credential sets, scope maps, tasks, tokens and
// (via the v2 module, the first with that client) connected registries.
func (g *ContainerGenerator) listRegistryChildren(resourceGroupName string, registryName string) ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	add := func(id, name, resourceType string) {
		resources = append(resources, terraformutils.NewSimpleResource(id, name, resourceType, g.ProviderName, []string{}))
	}

	agentPoolsClient, err := armcontainerregistry.NewAgentPoolsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	pager := agentPoolsClient.NewListPager(resourceGroupName, registryName, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return resources, err
		}
		for _, item := range page.Value {
			add(*item.ID, *item.Name, "azurerm_container_registry_agent_pool")
		}
	}

	cacheRulesClient, err := armcontainerregistry.NewCacheRulesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	cachePager := cacheRulesClient.NewListPager(resourceGroupName, registryName, nil)
	for cachePager.More() {
		page, err := cachePager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return resources, err
		}
		for _, item := range page.Value {
			add(*item.ID, *item.Name, "azurerm_container_registry_cache_rule")
		}
	}

	credentialSetsClient, err := armcontainerregistry.NewCredentialSetsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	credPager := credentialSetsClient.NewListPager(resourceGroupName, registryName, nil)
	for credPager.More() {
		page, err := credPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return resources, err
		}
		for _, item := range page.Value {
			add(*item.ID, *item.Name, "azurerm_container_registry_credential_set")
		}
	}

	scopeMapsClient, err := armcontainerregistry.NewScopeMapsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	scopePager := scopeMapsClient.NewListPager(resourceGroupName, registryName, nil)
	for scopePager.More() {
		page, err := scopePager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return resources, err
		}
		for _, item := range page.Value {
			add(*item.ID, *item.Name, "azurerm_container_registry_scope_map")
		}
	}

	tasksClient, err := armcontainerregistry.NewTasksClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	taskPager := tasksClient.NewListPager(resourceGroupName, registryName, nil)
	for taskPager.More() {
		page, err := taskPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return resources, err
		}
		for _, item := range page.Value {
			add(*item.ID, *item.Name, "azurerm_container_registry_task")
		}
	}

	tokensClient, err := armcontainerregistry.NewTokensClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	tokenPager := tokensClient.NewListPager(resourceGroupName, registryName, nil)
	for tokenPager.More() {
		page, err := tokenPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return resources, err
		}
		for _, item := range page.Value {
			add(*item.ID, *item.Name, "azurerm_container_registry_token")
		}
	}

	connectedClient, err := armcontainerregistryv2.NewConnectedRegistriesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	connectedPager := connectedClient.NewListPager(resourceGroupName, registryName, nil)
	for connectedPager.More() {
		page, err := connectedPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return resources, err
		}
		for _, item := range page.Value {
			add(*item.ID, *item.Name, "azurerm_container_connected_registry")
		}
	}
	return resources, nil
}

func (g *ContainerGenerator) listAndAddForContainerRegistry() ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	ContainerRegistriesClient, err := armcontainerregistry.NewRegistriesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	appendRegistry := func(registry *armcontainerregistry.Registry) error {
		resources = append(resources, terraformutils.NewSimpleResource(
			*registry.ID,
			*registry.Name,
			"azurerm_container_registry",
			g.ProviderName,
			[]string{}))

		id, err := ParseAzureResourceID(*registry.ID)
		if err != nil {
			return err
		}

		webhooks, err := g.listRegistryWebhooks(id.ResourceGroup, *registry.Name)
		if err != nil {
			return err
		}
		resources = append(resources, webhooks...)

		children, err := g.listRegistryChildren(id.ResourceGroup, *registry.Name)
		if err != nil {
			return err
		}
		resources = append(resources, children...)
		return nil
	}

	if rg := g.getResourceGroup(); rg != "" {
		pager := ContainerRegistriesClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return resources, err
			}
			for _, containerRegistry := range page.Value {
				if err := appendRegistry(containerRegistry); err != nil {
					return resources, err
				}
			}
		}
	} else {
		pager := ContainerRegistriesClient.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return resources, err
			}
			for _, containerRegistry := range page.Value {
				if err := appendRegistry(containerRegistry); err != nil {
					return resources, err
				}
			}
		}
	}

	return resources, nil
}

func (g *ContainerGenerator) InitResources() error {
	functions := []func() ([]terraformutils.Resource, error){
		g.listAndAddForContainerGroup,
		g.listAndAddForContainerRegistry,
	}

	for _, f := range functions {
		resources, err := f()
		if err != nil {
			return err
		}
		g.Resources = append(g.Resources, resources...)
	}

	return nil
}
