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

type PrivateEndpointGenerator struct {
	AzureService
}

func (az *PrivateEndpointGenerator) listServices() ([]*armnetwork.PrivateLinkService, error) {
	ctx := context.Background()
	client, err := armnetwork.NewPrivateLinkServicesClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
	if err != nil {
		return nil, err
	}
	var resources []*armnetwork.PrivateLinkService
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
		pager := client.NewListBySubscriptionPager(nil)
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

func (az *PrivateEndpointGenerator) AppendServices(link *armnetwork.PrivateLinkService) {
	az.AppendSimpleResource(*link.ID, *link.Name, "azurerm_private_link_service")
}

func (az *PrivateEndpointGenerator) listEndpoints() ([]*armnetwork.PrivateEndpoint, error) {
	ctx := context.Background()
	client, err := armnetwork.NewPrivateEndpointsClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
	if err != nil {
		return nil, err
	}
	var resources []*armnetwork.PrivateEndpoint
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
		pager := client.NewListBySubscriptionPager(nil)
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

func (az *PrivateEndpointGenerator) AppendEndpoint(link *armnetwork.PrivateEndpoint) {
	az.AppendSimpleResource(*link.ID, *link.Name, "azurerm_private_endpoint")
	az.appendASGAssociations(link)
}

// The azurerm association uses a Terraform-specific composite ID of the
// form {privateEndpointId}|{applicationSecurityGroupId} (verified in the
// provider's Import documentation); the associations are the
// applicationSecurityGroups references on the private endpoint object.
func (az *PrivateEndpointGenerator) appendASGAssociations(endpoint *armnetwork.PrivateEndpoint) {
	if endpoint.ID == nil || endpoint.Name == nil || endpoint.Properties == nil {
		return
	}
	for _, asg := range endpoint.Properties.ApplicationSecurityGroups {
		if asg.ID == nil {
			continue
		}
		name := *endpoint.Name + "-asg"
		if asg.Name != nil {
			name = *endpoint.Name + "-" + *asg.Name
		}
		az.appendSimpleAssociation(
			*endpoint.ID+"|"+*asg.ID, *endpoint.Name, &name,
			"azurerm_private_endpoint_application_security_group_association",
			map[string]string{
				"private_endpoint_id":           *endpoint.ID,
				"application_security_group_id": *asg.ID,
			})
	}
}

func (az *PrivateEndpointGenerator) InitResources() error {

	services, err := az.listServices()
	if err != nil {
		return err
	}
	for _, link := range services {
		az.AppendServices(link)
	}
	endpoints, err := az.listEndpoints()
	if err != nil {
		return err
	}
	for _, endpoint := range endpoints {
		az.AppendEndpoint(endpoint)
	}
	return nil
}
