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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/webpubsub/armwebpubsub"
)

// WebPubSubGenerator covers Microsoft.SignalRService/webPubSub services
// and their hubs, custom certificates, custom domains and shared private
// link resources.
//
// Skipped (see results-containers.md):
//   - azurerm_web_pubsub_network_acl: get-only singleton — the network ACL
//     is a property block on the Web PubSub service itself (its Terraform
//     ID is the service ID), armwebpubsub has no NetworkACL client/List,
//     and the block is returned with defaults for every service.
type WebPubSubGenerator struct {
	AzureService
}

// webPubSubTerraformType maps a Web PubSub service to its Terraform type
// by its service kind: SocketIO services are azurerm_web_pubsub_socketio.
func webPubSubTerraformType(service *armwebpubsub.ResourceInfo) string {
	if service != nil && service.Kind != nil && *service.Kind == armwebpubsub.ServiceKindSocketIO {
		return "azurerm_web_pubsub_socketio"
	}
	return "azurerm_web_pubsub"
}

func (g *WebPubSubGenerator) InitResources() error {
	ctx := context.Background()
	client, err := armwebpubsub.NewClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	var services []*armwebpubsub.ResourceInfo
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			services = append(services, page.Value...)
		}
	} else {
		pager := client.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			services = append(services, page.Value...)
		}
	}

	certClient, err := armwebpubsub.NewCustomCertificatesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	domainClient, err := armwebpubsub.NewCustomDomainsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	hubsClient, err := armwebpubsub.NewHubsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	splClient, err := armwebpubsub.NewSharedPrivateLinkResourcesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}

	for _, service := range services {
		g.AppendSimpleResource(*service.ID, *service.Name, webPubSubTerraformType(service))
		parsed, err := ParseAzureResourceID(*service.ID)
		if err != nil {
			return err
		}
		rg, name := parsed.ResourceGroup, *service.Name

		certPager := certClient.NewListPager(rg, name, nil)
		for certPager.More() {
			page, err := certPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_web_pubsub_custom_certificate")
			}
		}

		domainPager := domainClient.NewListPager(rg, name, nil)
		for domainPager.More() {
			page, err := domainPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_web_pubsub_custom_domain")
			}
		}

		hubPager := hubsClient.NewListPager(rg, name, nil)
		for hubPager.More() {
			page, err := hubPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_web_pubsub_hub")
			}
		}

		splPager := splClient.NewListPager(rg, name, nil)
		for splPager.More() {
			page, err := splPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_web_pubsub_shared_private_link_resource")
			}
		}
	}
	return nil
}
