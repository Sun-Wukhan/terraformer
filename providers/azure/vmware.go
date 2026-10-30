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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/avs/armavs"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// VMwareGenerator covers Azure VMware Solution (Microsoft.AVS): private
// clouds, their clusters and ExpressRoute authorizations, and datastores
// (the azurerm_vmware_netapp_volume_attachment resource is a datastore
// object under a cluster).
type VMwareGenerator struct {
	AzureService
}

func (g *VMwareGenerator) appendSimple(id, name, resourceType string) {
	if id == "" || name == "" {
		return
	}
	g.Resources = append(g.Resources, terraformutils.NewSimpleResource(id, name, resourceType, "azurerm", []string{}))
}

func (g *VMwareGenerator) InitResources() error {
	ctx := context.Background()
	subscriptionID := g.getSubscriptionID()
	cred := g.getCredential()
	opts := g.armClientOptions()
	rg := g.getResourceGroup()

	pcClient, err := armavs.NewPrivateCloudsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	clusterClient, err := armavs.NewClustersClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	authClient, err := armavs.NewAuthorizationsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	dsClient, err := armavs.NewDatastoresClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}

	var clouds []*armavs.PrivateCloud
	if rg != "" {
		pager := pcClient.NewListPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			clouds = append(clouds, page.Value...)
		}
	} else {
		pager := pcClient.NewListInSubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			clouds = append(clouds, page.Value...)
		}
	}
	for _, cloud := range clouds {
		g.appendSimple(*cloud.ID, *cloud.Name, "azurerm_vmware_private_cloud")
		cloudRG := resourceGroupFromID(*cloud.ID)

		apager := authClient.NewListPager(cloudRG, *cloud.Name, nil)
		for apager.More() {
			page, err := apager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_vmware_express_route_authorization")
			}
		}

		cpager := clusterClient.NewListPager(cloudRG, *cloud.Name, nil)
		for cpager.More() {
			page, err := cpager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, cluster := range page.Value {
				g.appendSimple(*cluster.ID, *cluster.Name, "azurerm_vmware_cluster")
				dpager := dsClient.NewListPager(cloudRG, *cloud.Name, *cluster.Name, nil)
				for dpager.More() {
					dpage, err := dpager.NextPage(ctx)
					if err != nil {
						log.Println(err)
						return err
					}
					for _, item := range dpage.Value {
						g.appendSimple(*item.ID, *item.Name, "azurerm_vmware_netapp_volume_attachment")
					}
				}
			}
		}
	}
	return nil
}
