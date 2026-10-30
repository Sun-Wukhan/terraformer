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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

// netParent identifies a parent ARM resource for child enumeration: its own
// ID/name plus the name segments of its ancestors, which the child List
// operations of the network SDKs require as separate arguments.
type netParent struct {
	ID              string
	Name            string
	ResourceGroup   string
	ParentName      string
	GrandparentName string
}

func makeParent(id, name *string) (netParent, bool) {
	if id == nil || name == nil {
		return netParent{}, false
	}
	parsed, err := ParseAzureResourceID(*id)
	if err != nil {
		return netParent{}, false
	}
	return netParent{ID: *id, Name: *name, ResourceGroup: parsed.ResourceGroup}, true
}

// resourceGroupNames returns the resource groups a resource-group-scoped
// List should run against: the configured resource group when terraformer
// was invoked with one, otherwise every resource group in the subscription
// (several network child/top-level APIs expose no subscription-scope List).
func (az *AzureService) resourceGroupNames() ([]string, error) {
	if rg := az.getResourceGroup(); rg != "" {
		return []string{rg}, nil
	}
	ctx := context.Background()
	client, err := armresources.NewResourceGroupsClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
	if err != nil {
		return nil, err
	}
	var names []string
	pager := client.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		for _, group := range page.Value {
			if group.Name != nil {
				names = append(names, *group.Name)
			}
		}
	}
	return names, nil
}
