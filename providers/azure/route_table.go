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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

type RouteTableGenerator struct {
	AzureService
}

func (az *RouteTableGenerator) listResources() ([]*armnetwork.RouteTable, error) {
	ctx := context.Background()
	client, err := armnetwork.NewRouteTablesClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
	if err != nil {
		return nil, err
	}
	var resources []*armnetwork.RouteTable
	if rg := az.getResourceGroup(); rg != "" {
		pager := client.NewListPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return resources, err
			}
			resources = append(resources, page.Value...)
		}
	} else {
		pager := client.NewListAllPager(nil)
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

func (az *RouteTableGenerator) appendResource(resource *armnetwork.RouteTable) {
	az.AppendSimpleResourceWithDuplicateCheck(*resource.ID, *resource.Name, "azurerm_route_table")
}

func (az *RouteTableGenerator) appendRoutes(parent *armnetwork.RouteTable, resourceGroupID *ResourceID) error {
	ctx := context.Background()
	client, err := armnetwork.NewRoutesClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
	if err != nil {
		return err
	}
	pager := client.NewListPager(resourceGroupID.ResourceGroup, *parent.Name, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			az.AppendSimpleResourceWithDuplicateCheck(*item.ID, *item.Name, "azurerm_route")
		}
	}
	return nil
}

func (az *RouteTableGenerator) listRouteFilters() ([]*armnetwork.RouteFilter, error) {
	ctx := context.Background()
	client, err := armnetwork.NewRouteFiltersClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
	if err != nil {
		return nil, err
	}
	var resources []*armnetwork.RouteFilter
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

func (az *RouteTableGenerator) appendRouteFilters(resource *armnetwork.RouteFilter) {
	az.AppendSimpleResource(*resource.ID, *resource.Name, "azurerm_route_filter")
}

func (az *RouteTableGenerator) InitResources() error {

	resources, err := az.listResources()
	if err != nil {
		return err
	}
	for _, resource := range resources {
		az.appendResource(resource)
		resourceGroupID, err := ParseAzureResourceID(*resource.ID)
		if err != nil {
			return err
		}
		err = az.appendRoutes(resource, resourceGroupID)
		if err != nil {
			return err
		}
	}

	filters, err := az.listRouteFilters()
	if err != nil {
		return err
	}
	for _, resource := range filters {
		az.appendRouteFilters(resource)
		if err != nil {
			return err
		}
	}
	return nil
}
