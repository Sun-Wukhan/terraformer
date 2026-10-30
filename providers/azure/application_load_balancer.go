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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/servicenetworking/armservicenetworking"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// ApplicationLoadBalancerGenerator imports Azure Application Load Balancer
// (Microsoft.ServiceNetworking trafficControllers) resources via the
// track-2 armservicenetworking SDK.
type ApplicationLoadBalancerGenerator struct {
	AzureService
}

func (g *ApplicationLoadBalancerGenerator) simple(id, name, resourceType string) (terraformutils.Resource, bool) {
	if id == "" || name == "" {
		return terraformutils.Resource{}, false
	}
	return terraformutils.NewSimpleResource(id, name, resourceType, g.ProviderName, []string{}), true
}

func (g *ApplicationLoadBalancerGenerator) InitResources() error {
	ctx := context.Background()
	client, err := armservicenetworking.NewTrafficControllerInterfaceClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	var controllers []*armservicenetworking.TrafficController
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			controllers = append(controllers, page.Value...)
		}
	} else {
		pager := client.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			controllers = append(controllers, page.Value...)
		}
	}
	for _, controller := range controllers {
		if controller.ID == nil || controller.Name == nil {
			continue
		}
		if res, ok := g.simple(*controller.ID, *controller.Name, "azurerm_application_load_balancer"); ok {
			g.Resources = append(g.Resources, res)
		}
		parsed, err := ParseAzureResourceID(*controller.ID)
		if err != nil {
			continue
		}
		if err := g.listChildren(ctx, parsed.ResourceGroup, *controller.Name); err != nil {
			return err
		}
	}
	return nil
}

func (g *ApplicationLoadBalancerGenerator) listChildren(ctx context.Context, resourceGroup, controllerName string) error {
	frontendsClient, err := armservicenetworking.NewFrontendsInterfaceClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	resources, err := drainPager(ctx, frontendsClient.NewListByTrafficControllerPager(resourceGroup, controllerName, nil),
		func(p armservicenetworking.FrontendsInterfaceClientListByTrafficControllerResponse) []*armservicenetworking.Frontend {
			return p.Value
		},
		func(item *armservicenetworking.Frontend) (terraformutils.Resource, bool) {
			if item.ID == nil || item.Name == nil {
				return terraformutils.Resource{}, false
			}
			return g.simple(*item.ID, *item.Name, "azurerm_application_load_balancer_frontend")
		})
	if err != nil {
		return err
	}
	g.Resources = append(g.Resources, resources...)

	securityPoliciesClient, err := armservicenetworking.NewSecurityPoliciesInterfaceClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	resources, err = drainPager(ctx, securityPoliciesClient.NewListByTrafficControllerPager(resourceGroup, controllerName, nil),
		func(p armservicenetworking.SecurityPoliciesInterfaceClientListByTrafficControllerResponse) []*armservicenetworking.SecurityPolicy {
			return p.Value
		},
		func(item *armservicenetworking.SecurityPolicy) (terraformutils.Resource, bool) {
			if item.ID == nil || item.Name == nil {
				return terraformutils.Resource{}, false
			}
			return g.simple(*item.ID, *item.Name, "azurerm_application_load_balancer_security_policy")
		})
	if err != nil {
		return err
	}
	g.Resources = append(g.Resources, resources...)

	associationsClient, err := armservicenetworking.NewAssociationsInterfaceClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	resources, err = drainPager(ctx, associationsClient.NewListByTrafficControllerPager(resourceGroup, controllerName, nil),
		func(p armservicenetworking.AssociationsInterfaceClientListByTrafficControllerResponse) []*armservicenetworking.Association {
			return p.Value
		},
		func(item *armservicenetworking.Association) (terraformutils.Resource, bool) {
			if item.ID == nil || item.Name == nil {
				return terraformutils.Resource{}, false
			}
			return g.simple(*item.ID, *item.Name, "azurerm_application_load_balancer_subnet_association")
		})
	if err != nil {
		return err
	}
	g.Resources = append(g.Resources, resources...)
	return nil
}
