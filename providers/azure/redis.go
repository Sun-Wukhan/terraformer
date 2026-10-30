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
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package azure

import (
	"context"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type RedisGenerator struct {
	AzureService
}

func (g *RedisGenerator) listRedisServers() ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	client, err := armredis.NewClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	pager := client.NewListBySubscriptionPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return resources, err
		}
		for _, redisServer := range page.Value {
			resources = append(resources, terraformutils.NewSimpleResource(
				*redisServer.ID,
				*redisServer.Name,
				"azurerm_redis_cache",
				g.ProviderName,
				[]string{}))
		}
	}

	return resources, nil
}

// Sweep additions (data cluster). Skipped in this family:
//   - azurerm_redis_cache_access_policy and
//     azurerm_redis_cache_access_policy_assignment: armredis v1.0.0 has no
//     AccessPolicies / AccessPolicyAssignments client (the redis
//     accessPolicies API surface is not shipped in the track-2 Go SDK).

func (g *RedisGenerator) listRedisChildren() ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	client, err := armredis.NewClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	firewallClient, err := armredis.NewFirewallRulesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	linkedServerClient, err := armredis.NewLinkedServerClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	pager := client.NewListBySubscriptionPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return resources, err
		}
		for _, cache := range page.Value {
			id, err := ParseAzureResourceID(*cache.ID)
			if err != nil {
				return nil, err
			}
			firewallPager := firewallClient.NewListPager(id.ResourceGroup, *cache.Name, nil)
			for firewallPager.More() {
				firewallPage, err := firewallPager.NextPage(ctx)
				if err != nil {
					log.Println(err)
					return resources, err
				}
				for _, rule := range firewallPage.Value {
					resources = append(resources, terraformutils.NewSimpleResource(
						*rule.ID, *rule.Name, "azurerm_redis_firewall_rule", g.ProviderName, []string{}))
				}
			}
			linkedPager := linkedServerClient.NewListPager(id.ResourceGroup, *cache.Name, nil)
			for linkedPager.More() {
				linkedPage, err := linkedPager.NextPage(ctx)
				if err != nil {
					log.Println(err)
					return resources, err
				}
				for _, linked := range linkedPage.Value {
					resources = append(resources, terraformutils.NewSimpleResource(
						*linked.ID, *linked.Name, "azurerm_redis_linked_server", g.ProviderName, []string{}))
				}
			}
		}
	}

	return resources, nil
}

func (g *RedisGenerator) InitResources() error {
	functions := []func() ([]terraformutils.Resource, error){
		g.listRedisServers,
		g.listRedisChildren,
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
