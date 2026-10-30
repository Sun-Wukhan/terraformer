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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/servicefabric/armservicefabric"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/servicefabricmanagedclusters/armservicefabricmanagedclusters"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// ServiceFabricGenerator covers Microsoft.ServiceFabric clusters
// (armservicefabric ClustersClient, whose List/ListByResourceGroup are
// plain calls rather than pagers) and managed clusters
// (armservicefabricmanagedclusters).
type ServiceFabricGenerator struct {
	AzureService
}

func (g *ServiceFabricGenerator) appendSimple(id, name, resourceType string) {
	if id == "" || name == "" {
		return
	}
	g.Resources = append(g.Resources, terraformutils.NewSimpleResource(id, name, resourceType, "azurerm", []string{}))
}

func (g *ServiceFabricGenerator) InitResources() error {
	ctx := context.Background()
	subscriptionID := g.getSubscriptionID()
	cred := g.getCredential()
	opts := g.armClientOptions()
	rg := g.getResourceGroup()

	clusterClient, err := armservicefabric.NewClustersClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	if rg != "" {
		resp, err := clusterClient.ListByResourceGroup(ctx, rg, nil)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range resp.Value {
			g.appendSimple(*item.ID, *item.Name, "azurerm_service_fabric_cluster")
		}
	} else {
		resp, err := clusterClient.List(ctx, nil)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range resp.Value {
			g.appendSimple(*item.ID, *item.Name, "azurerm_service_fabric_cluster")
		}
	}

	managedClient, err := armservicefabricmanagedclusters.NewManagedClustersClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	if rg != "" {
		pager := managedClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_service_fabric_managed_cluster")
			}
		}
	} else {
		pager := managedClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_service_fabric_managed_cluster")
			}
		}
	}
	return nil
}
