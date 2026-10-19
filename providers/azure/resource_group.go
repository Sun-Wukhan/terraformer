// Copyright 2019 The Terraformer Authors.
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

type ResourceGroupGenerator struct {
	AzureService
}

func (g ResourceGroupGenerator) createResources(ctx context.Context, pager *runtime.Pager[armresources.ResourceGroupsClientListResponse]) []terraformutils.Resource {
	var resources []terraformutils.Resource
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			break
		}
		for _, group := range page.Value {
			resources = append(resources, terraformutils.NewSimpleResource(
				*group.ID,
				*group.Name,
				"azurerm_resource_group",
				"azurerm",
				[]string{}))
		}
	}
	return resources
}

func (g *ResourceGroupGenerator) InitResources() error {
	ctx := context.Background()
	client, err := armresources.NewResourceGroupsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}

	if rg := g.getResourceGroup(); rg != "" {
		resp, err := client.Get(ctx, rg, nil)
		if err != nil {
			return err
		}
		g.Resources = []terraformutils.Resource{
			terraformutils.NewSimpleResource(
				*resp.ResourceGroup.ID,
				*resp.ResourceGroup.Name,
				"azurerm_resource_group",
				"azurerm",
				[]string{}),
		}
		return nil
	}
	g.Resources = g.createResources(ctx, client.NewListPager(nil))
	return nil
}
