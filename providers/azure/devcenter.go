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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/devcenter/armdevcenter"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// DevCenterGenerator covers Microsoft.DevCenter: dev centers with their
// attached networks, catalogs, dev box definitions, environment types and
// galleries; network connections; and projects with their environment
// types and pools.
type DevCenterGenerator struct {
	AzureService
}

func (g *DevCenterGenerator) appendSimple(id, name, resourceType string) {
	if id == "" || name == "" {
		return
	}
	g.Resources = append(g.Resources, terraformutils.NewSimpleResource(id, name, resourceType, "azurerm", []string{}))
}

func (g *DevCenterGenerator) InitResources() error {
	ctx := context.Background()
	subscriptionID := g.getSubscriptionID()
	cred := g.getCredential()
	opts := g.armClientOptions()
	rg := g.getResourceGroup()

	// Network connections (top-level)
	ncClient, err := armdevcenter.NewNetworkConnectionsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	if rg != "" {
		pager := ncClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_dev_center_network_connection")
			}
		}
	} else {
		pager := ncClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_dev_center_network_connection")
			}
		}
	}

	// Dev centers + children
	dcClient, err := armdevcenter.NewDevCentersClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	anClient, err := armdevcenter.NewAttachedNetworksClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	catClient, err := armdevcenter.NewCatalogsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	dbdClient, err := armdevcenter.NewDevBoxDefinitionsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	etClient, err := armdevcenter.NewEnvironmentTypesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	galClient, err := armdevcenter.NewGalleriesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	var centers []*armdevcenter.DevCenter
	if rg != "" {
		pager := dcClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			centers = append(centers, page.Value...)
		}
	} else {
		pager := dcClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			centers = append(centers, page.Value...)
		}
	}
	for _, center := range centers {
		g.appendSimple(*center.ID, *center.Name, "azurerm_dev_center")
		centerRG := resourceGroupFromID(*center.ID)

		pager := anClient.NewListByDevCenterPager(centerRG, *center.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_dev_center_attached_network")
			}
		}
		cpager := catClient.NewListByDevCenterPager(centerRG, *center.Name, nil)
		for cpager.More() {
			page, err := cpager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_dev_center_catalog")
			}
		}
		dpager := dbdClient.NewListByDevCenterPager(centerRG, *center.Name, nil)
		for dpager.More() {
			page, err := dpager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_dev_center_dev_box_definition")
			}
		}
		epager := etClient.NewListByDevCenterPager(centerRG, *center.Name, nil)
		for epager.More() {
			page, err := epager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_dev_center_environment_type")
			}
		}
		gpager := galClient.NewListByDevCenterPager(centerRG, *center.Name, nil)
		for gpager.More() {
			page, err := gpager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_dev_center_gallery")
			}
		}
	}

	// Projects + environment types + pools
	projClient, err := armdevcenter.NewProjectsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	petClient, err := armdevcenter.NewProjectEnvironmentTypesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	poolClient, err := armdevcenter.NewPoolsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	var projects []*armdevcenter.Project
	if rg != "" {
		pager := projClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			projects = append(projects, page.Value...)
		}
	} else {
		pager := projClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			projects = append(projects, page.Value...)
		}
	}
	for _, project := range projects {
		g.appendSimple(*project.ID, *project.Name, "azurerm_dev_center_project")
		projectRG := resourceGroupFromID(*project.ID)
		pager := petClient.NewListPager(projectRG, *project.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_dev_center_project_environment_type")
			}
		}
		ppager := poolClient.NewListByProjectPager(projectRG, *project.Name, nil)
		for ppager.More() {
			page, err := ppager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_dev_center_project_pool")
			}
		}
	}
	return nil
}
