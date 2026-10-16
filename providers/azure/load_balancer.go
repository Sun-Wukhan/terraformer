// Copyright 2020 The Terraformer Authors.
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
	"regexp"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type LoadBalancerGenerator struct {
	AzureService
}

func (g *LoadBalancerGenerator) listLoadBalancerProbes(resourceGroupName string, loadBalancerName string) ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()

	client, err := armnetwork.NewLoadBalancerProbesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	pager := client.NewListPager(resourceGroupName, loadBalancerName, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			break
		}
		for _, loadBalancerProbe := range page.Value {
			// NOTE:
			// This works out the loadBalancer resource id from current probe
			// /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/group1/providers/Microsoft.Network/loadBalancers/lb1
			// /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/group1/providers/Microsoft.Network/loadBalancers/lb1/probes/probe1
			//
			// As the related data_source in azurerm provider works by starting to look up with loadbalancer_id
			// https://github.com/terraform-providers/terraform-provider-azurerm/blob/v2.18.0/azurerm/internal/services/network/lb_probe_resource.go#L186
			re := regexp.MustCompile(`/probes/.*$`)
			loadBalancerID := re.ReplaceAllLiteralString(*loadBalancerProbe.ID, "")
			resources = append(resources, terraformutils.NewResource(
				*loadBalancerProbe.ID,
				*loadBalancerProbe.Name,
				"azurerm_lb_probe",
				g.ProviderName,
				map[string]string{
					"loadbalancer_id": loadBalancerID,
				},
				[]string{},
				map[string]interface{}{},
			))
		}
	}

	return resources, nil
}

func (g *LoadBalancerGenerator) listInboundNatRules(resourceGroupName string, loadBalancerName string) ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()

	client, err := armnetwork.NewInboundNatRulesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	pager := client.NewListPager(resourceGroupName, loadBalancerName, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			break
		}
		for _, inboundNatRule := range page.Value {
			// NOTE:
			// Similar to above explanation, work out loadbalancer_id for azurerm datasource impl
			re := regexp.MustCompile(`/inboundNatRules/.*$`)
			loadBalancerID := re.ReplaceAllLiteralString(*inboundNatRule.ID, "")
			resources = append(resources, terraformutils.NewResource(
				*inboundNatRule.ID,
				*inboundNatRule.Name,
				"azurerm_lb_nat_rule",
				g.ProviderName,
				map[string]string{
					"loadbalancer_id": loadBalancerID,
				},
				[]string{},
				map[string]interface{}{},
			))
		}
	}

	return resources, nil
}

func (g *LoadBalancerGenerator) listLoadBalancerBackendAddressPools(resourceGroupName string, loadBalancerName string) ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()

	client, err := armnetwork.NewLoadBalancerBackendAddressPoolsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	pager := client.NewListPager(resourceGroupName, loadBalancerName, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			break
		}
		for _, pool := range page.Value {
			// NOTE:
			// Similar to above explanation, work out loadbalancer_id for azurerm datasource impl
			re := regexp.MustCompile(`/backendAddressPools/.*$`)
			loadBalancerID := re.ReplaceAllLiteralString(*pool.ID, "")
			resources = append(resources, terraformutils.NewResource(
				*pool.ID,
				*pool.Name,
				"azurerm_lb_backend_address_pool",
				g.ProviderName,
				map[string]string{
					"loadbalancer_id": loadBalancerID,
				},
				[]string{},
				map[string]interface{}{},
			))
		}
	}

	return resources, nil
}

func (g *LoadBalancerGenerator) listAndAddForLoadBalancers() ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()

	client, err := armnetwork.NewLoadBalancersClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	appendLB := func(loadBalancer *armnetwork.LoadBalancer) error {
		resources = append(resources, terraformutils.NewSimpleResource(
			*loadBalancer.ID,
			*loadBalancer.Name,
			"azurerm_lb",
			g.ProviderName,
			[]string{}))

		id, err := ParseAzureResourceID(*loadBalancer.ID)
		if err != nil {
			return err
		}

		probes, err := g.listLoadBalancerProbes(id.ResourceGroup, *loadBalancer.Name)
		if err != nil {
			return err
		}
		resources = append(resources, probes...)

		inboundNatRules, err := g.listInboundNatRules(id.ResourceGroup, *loadBalancer.Name)
		if err != nil {
			return err
		}
		resources = append(resources, inboundNatRules...)

		backendAddressPools, err := g.listLoadBalancerBackendAddressPools(id.ResourceGroup, *loadBalancer.Name)
		if err != nil {
			return err
		}
		resources = append(resources, backendAddressPools...)
		return nil
	}

	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return resources, err
			}
			for _, loadBalancer := range page.Value {
				if err := appendLB(loadBalancer); err != nil {
					return nil, err
				}
			}
		}
	} else {
		pager := client.NewListAllPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return resources, err
			}
			for _, loadBalancer := range page.Value {
				if err := appendLB(loadBalancer); err != nil {
					return nil, err
				}
			}
		}
	}

	return resources, nil
}

func (g *LoadBalancerGenerator) InitResources() error {
	functions := []func() ([]terraformutils.Resource, error){
		g.listAndAddForLoadBalancers,
	}

	for _, f := range functions {
		resources, err := f()
		if err != nil {
			return err
		}
		g.Resources = append(g.Resources, resources...)
	}

	return nil
}
