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
	"strings"

	armtrafficmanager "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/trafficmanager/armtrafficmanager"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type TrafficManagerGenerator struct {
	AzureService
}

func (g *TrafficManagerGenerator) listTrafficManagerProfiles() error {
	ctx := context.Background()
	client, err := armtrafficmanager.NewProfilesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	handle := func(items []*armtrafficmanager.Profile) {
		for _, item := range items {
			if item.ID == nil || item.Name == nil {
				continue
			}
			g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_traffic_manager_profile")
			g.Resources = append(g.Resources, trafficManagerEndpointResources(item)...)
		}
	}
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			handle(page.Value)
		}
	} else {
		pager := client.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			handle(page.Value)
		}
	}
	return nil
}

// Traffic Manager endpoints have no List operation of their own in the
// track-2 SDK (EndpointsClient exposes only Get/CreateOrUpdate/Delete):
// they are returned embedded in the profile object, each with its ARM ID
// and a Type of Microsoft.Network/trafficManagerProfiles/{Azure,External,
// Nested}Endpoints, which maps 1:1 onto the three azurerm endpoint types.
func trafficManagerEndpointResources(profile *armtrafficmanager.Profile) []terraformutils.Resource {
	var resources []terraformutils.Resource
	if profile.Properties == nil {
		return resources
	}
	for _, endpoint := range profile.Properties.Endpoints {
		if endpoint.ID == nil || endpoint.Name == nil || endpoint.Type == nil {
			continue
		}
		var tfType string
		switch {
		case strings.HasSuffix(*endpoint.Type, "/AzureEndpoints"):
			tfType = "azurerm_traffic_manager_azure_endpoint"
		case strings.HasSuffix(*endpoint.Type, "/ExternalEndpoints"):
			tfType = "azurerm_traffic_manager_external_endpoint"
		case strings.HasSuffix(*endpoint.Type, "/NestedEndpoints"):
			tfType = "azurerm_traffic_manager_nested_endpoint"
		default:
			continue
		}
		resources = append(resources, terraformutils.NewSimpleResource(
			*endpoint.ID, *endpoint.Name, tfType, "azurerm", []string{}))
	}
	return resources
}

func (g *TrafficManagerGenerator) InitResources() error {
	if err := g.listTrafficManagerProfiles(); err != nil {
		return err
	}
	return nil
}
