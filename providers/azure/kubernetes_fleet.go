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

	armcontainerservicefleet "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservicefleet/armcontainerservicefleet/v2"
)

// KubernetesFleetGenerator covers Microsoft.ContainerService fleets and
// their members, update runs and update strategies.
type KubernetesFleetGenerator struct {
	AzureService
}

func (g *KubernetesFleetGenerator) InitResources() error {
	ctx := context.Background()
	client, err := armcontainerservicefleet.NewFleetsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	var fleets []*armcontainerservicefleet.Fleet
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			fleets = append(fleets, page.Value...)
		}
	} else {
		pager := client.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			fleets = append(fleets, page.Value...)
		}
	}

	membersClient, err := armcontainerservicefleet.NewFleetMembersClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	runsClient, err := armcontainerservicefleet.NewUpdateRunsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	strategiesClient, err := armcontainerservicefleet.NewFleetUpdateStrategiesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}

	for _, fleet := range fleets {
		g.AppendSimpleResource(*fleet.ID, *fleet.Name, "azurerm_kubernetes_fleet_manager")
		parsed, err := ParseAzureResourceID(*fleet.ID)
		if err != nil {
			return err
		}
		rg, fleetName := parsed.ResourceGroup, *fleet.Name

		memberPager := membersClient.NewListByFleetPager(rg, fleetName, nil)
		for memberPager.More() {
			page, err := memberPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_kubernetes_fleet_member")
			}
		}

		runPager := runsClient.NewListByFleetPager(rg, fleetName, nil)
		for runPager.More() {
			page, err := runPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_kubernetes_fleet_update_run")
			}
		}

		strategyPager := strategiesClient.NewListByFleetPager(rg, fleetName, nil)
		for strategyPager.More() {
			page, err := strategyPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_kubernetes_fleet_update_strategy")
			}
		}
	}
	return nil
}
