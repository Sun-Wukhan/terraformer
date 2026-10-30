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
	"fmt"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/digitaltwins/armdigitaltwins"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// DigitalTwinsGenerator covers Azure Digital Twins via armdigitaltwins:
// instances, their endpoints (split by endpoint type) and time series
// database connections.
type DigitalTwinsGenerator struct {
	AzureService
}

// digitalTwinsEndpointType maps an endpoint's concrete
// armdigitaltwins properties type to the azurerm endpoint resource type.
func digitalTwinsEndpointType(concreteType string) string {
	switch {
	case strings.Contains(concreteType, "EventHub"):
		return "azurerm_digital_twins_endpoint_eventhub"
	case strings.Contains(concreteType, "EventGrid"):
		return "azurerm_digital_twins_endpoint_eventgrid"
	case strings.Contains(concreteType, "ServiceBus"):
		return "azurerm_digital_twins_endpoint_servicebus"
	}
	return ""
}

func (g *DigitalTwinsGenerator) InitResources() error {
	ctx := context.Background()
	sub := g.getSubscriptionID()
	cred := g.getCredential()
	opts := g.armClientOptions()
	var resources []terraformutils.Resource
	add := func(id, name, typ string) {
		if id != "" && name != "" && typ != "" {
			resources = append(resources, terraformutils.NewSimpleResource(id, name, typ, "azurerm", []string{}))
		}
	}
	instancesClient, err := armdigitaltwins.NewClient(sub, cred, opts)
	if err != nil {
		return err
	}
	endpointsClient, err := armdigitaltwins.NewEndpointClient(sub, cred, opts)
	if err != nil {
		return err
	}
	tsdbClient, err := armdigitaltwins.NewTimeSeriesDatabaseConnectionsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	type instRef struct{ rg, name string }
	var instances []instRef
	if rg := g.getResourceGroup(); rg != "" {
		pager := instancesClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_digital_twins_instance")
					instances = append(instances, instRef{rg, *item.Name})
				}
			}
		}
	} else {
		pager := instancesClient.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				if item.ID == nil || item.Name == nil {
					continue
				}
				add(*item.ID, *item.Name, "azurerm_digital_twins_instance")
				if parsed, err := ParseAzureResourceID(*item.ID); err == nil {
					instances = append(instances, instRef{parsed.ResourceGroup, *item.Name})
				}
			}
		}
	}
	for _, inst := range instances {
		ePager := endpointsClient.NewListPager(inst.rg, inst.name, nil)
		for ePager.More() {
			page, err := ePager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID == nil || item.Name == nil || item.Properties == nil {
					continue
				}
				add(*item.ID, *item.Name, digitalTwinsEndpointType(fmt.Sprintf("%T", item.Properties)))
			}
		}
		tPager := tsdbClient.NewListPager(inst.rg, inst.name, nil)
		for tPager.More() {
			page, err := tPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_digital_twins_time_series_database_connection")
				}
			}
		}
	}
	g.Resources = resources
	return nil
}
