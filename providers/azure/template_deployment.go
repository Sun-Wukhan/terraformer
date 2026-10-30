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

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// TemplateDeploymentGenerator covers ARM template deployments at
// subscription, resource-group, tenant and management-group scopes. Each
// scope has its own azurerm type but all share the armresources
// DeploymentsClient list pagers.
type TemplateDeploymentGenerator struct {
	AzureService
}

func appendDeploymentPager[T any](ctx context.Context, g *TemplateDeploymentGenerator, pager *runtime.Pager[T], value func(T) []*armresources.DeploymentExtended, resourceType string) {
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return
		}
		for _, item := range value(page) {
			if item.ID == nil || item.Name == nil {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*item.ID,
				*item.Name,
				resourceType,
				g.ProviderName,
				[]string{}))
		}
	}
}

func (g *TemplateDeploymentGenerator) InitResources() error {
	ctx := context.Background()
	client, err := armresources.NewDeploymentsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}

	appendDeploymentPager(ctx, g, client.NewListAtSubscriptionScopePager(nil),
		func(page armresources.DeploymentsClientListAtSubscriptionScopeResponse) []*armresources.DeploymentExtended {
			return page.Value
		},
		"azurerm_subscription_template_deployment")
	appendDeploymentPager(ctx, g, client.NewListAtTenantScopePager(nil),
		func(page armresources.DeploymentsClientListAtTenantScopeResponse) []*armresources.DeploymentExtended {
			return page.Value
		},
		"azurerm_tenant_template_deployment")

	resourceGroups, err := governanceResourceGroupNames(ctx, &g.AzureService)
	if err != nil {
		return err
	}
	for _, rg := range resourceGroups {
		appendDeploymentPager(ctx, g, client.NewListByResourceGroupPager(rg, nil),
			func(page armresources.DeploymentsClientListByResourceGroupResponse) []*armresources.DeploymentExtended {
				return page.Value
			},
			"azurerm_resource_group_template_deployment")
	}

	groups, err := governanceManagementGroups(ctx, &g.AzureService)
	if err != nil {
		return err
	}
	for _, group := range groups {
		appendDeploymentPager(ctx, g, client.NewListAtManagementGroupScopePager(group.Name, nil),
			func(page armresources.DeploymentsClientListAtManagementGroupScopeResponse) []*armresources.DeploymentExtended {
				return page.Value
			},
			"azurerm_management_group_template_deployment")
	}
	return nil
}
