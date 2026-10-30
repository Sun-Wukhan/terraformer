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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mongocluster/armmongocluster"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// MongoClusterGenerator covers Azure Cosmos DB for MongoDB (vCore)
// clusters (Microsoft.DocumentDB/mongoClusters) via armmongocluster.
type MongoClusterGenerator struct {
	AzureService
}

func (g *MongoClusterGenerator) listClusters() ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	clustersClient, err := armmongocluster.NewMongoClustersClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	firewallClient, err := armmongocluster.NewFirewallRulesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	usersClient, err := armmongocluster.NewUsersClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	processClusters := func(clusters []*armmongocluster.MongoCluster) error {
		for _, cluster := range clusters {
			resources = append(resources, terraformutils.NewResource(
				*cluster.ID, *cluster.Name, "azurerm_mongo_cluster", g.ProviderName,
				map[string]string{}, []string{}, map[string]interface{}{"administrator_password": ""}))
			id, err := ParseAzureResourceID(*cluster.ID)
			if err != nil {
				return err
			}

			firewallPager := firewallClient.NewListByMongoClusterPager(id.ResourceGroup, *cluster.Name, nil)
			for firewallPager.More() {
				page, err := firewallPager.NextPage(ctx)
				if err != nil {
					log.Println(err)
					return err
				}
				for _, rule := range page.Value {
					resources = append(resources, terraformutils.NewSimpleResource(
						*rule.ID, *rule.Name, "azurerm_mongo_cluster_firewall_rule", g.ProviderName, []string{}))
				}
			}

			usersPager := usersClient.NewListByMongoClusterPager(id.ResourceGroup, *cluster.Name, nil)
			for usersPager.More() {
				page, err := usersPager.NextPage(ctx)
				if err != nil {
					log.Println(err)
					return err
				}
				for _, user := range page.Value {
					resources = append(resources, terraformutils.NewResource(
						*user.ID, *user.Name, "azurerm_mongo_cluster_user", g.ProviderName,
						map[string]string{}, []string{}, map[string]interface{}{"password": ""}))
				}
			}
		}
		return nil
	}

	if rg := g.getResourceGroup(); rg != "" {
		pager := clustersClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			if err := processClusters(page.Value); err != nil {
				return nil, err
			}
		}
	} else {
		pager := clustersClient.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			if err := processClusters(page.Value); err != nil {
				return nil, err
			}
		}
	}

	return resources, nil
}

func (g *MongoClusterGenerator) InitResources() error {
	resources, err := g.listClusters()
	if err != nil {
		return err
	}
	g.Resources = append(g.Resources, resources...)
	return nil
}
