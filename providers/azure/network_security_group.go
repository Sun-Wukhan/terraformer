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

type NetworkSecurityGroupGenerator struct {
	AzureService
}

func (az *NetworkSecurityGroupGenerator) listResources() ([]*armnetwork.SecurityGroup, error) {
	ctx := context.Background()
	client, err := armnetwork.NewSecurityGroupsClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
	if err != nil {
		return nil, err
	}
	var resources []*armnetwork.SecurityGroup
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

func (az *NetworkSecurityGroupGenerator) appendResource(resource *armnetwork.SecurityGroup) {
	az.AppendSimpleResourceWithDuplicateCheck(*resource.ID, *resource.Name, "azurerm_network_security_group")
}

func (az *NetworkSecurityGroupGenerator) appendRules(parent *armnetwork.SecurityGroup, resourceGroupID *ResourceID) error {
	ctx := context.Background()
	client, err := armnetwork.NewSecurityRulesClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
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
			az.AppendSimpleResourceWithDuplicateCheck(*item.ID, *item.Name, "azurerm_network_security_rule")
		}
	}
	return nil
}

func (az *NetworkSecurityGroupGenerator) InitResources() error {

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
		err = az.appendRules(resource, resourceGroupID)
		if err != nil {
			return err
		}
	}
	return nil
}
