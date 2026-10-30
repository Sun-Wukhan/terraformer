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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/datamigration/armdatamigration"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// DatabaseMigrationGenerator covers Azure Database Migration Service via
// armdatamigration.
type DatabaseMigrationGenerator struct {
	AzureService
}

func (g *DatabaseMigrationGenerator) listServices() ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	servicesClient, err := armdatamigration.NewServicesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	projectsClient, err := armdatamigration.NewProjectsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	processServices := func(services []*armdatamigration.Service) error {
		for _, service := range services {
			resources = append(resources, terraformutils.NewSimpleResource(
				*service.ID, *service.Name, "azurerm_database_migration_service", g.ProviderName, []string{}))
			id, err := ParseAzureResourceID(*service.ID)
			if err != nil {
				return err
			}
			pager := projectsClient.NewListPager(id.ResourceGroup, *service.Name, nil)
			for pager.More() {
				page, err := pager.NextPage(ctx)
				if err != nil {
					log.Println(err)
					return err
				}
				for _, project := range page.Value {
					resources = append(resources, terraformutils.NewSimpleResource(
						*project.ID, *project.Name, "azurerm_database_migration_project", g.ProviderName, []string{}))
				}
			}
		}
		return nil
	}

	if rg := g.getResourceGroup(); rg != "" {
		pager := servicesClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			if err := processServices(page.Value); err != nil {
				return nil, err
			}
		}
	} else {
		pager := servicesClient.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			if err := processServices(page.Value); err != nil {
				return nil, err
			}
		}
	}

	return resources, nil
}

func (g *DatabaseMigrationGenerator) InitResources() error {
	resources, err := g.listServices()
	if err != nil {
		return err
	}
	g.Resources = append(g.Resources, resources...)
	return nil
}
