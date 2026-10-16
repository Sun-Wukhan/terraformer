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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type PublicIPGenerator struct {
	AzureService
}

func (g *PublicIPGenerator) listAndAddForPublicIPAddress() ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	client, err := armnetwork.NewPublicIPAddressesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	appendPage := func(ips []*armnetwork.PublicIPAddress) {
		for _, publicIP := range ips {
			resources = append(resources, terraformutils.NewSimpleResource(
				*publicIP.ID,
				*publicIP.Name,
				"azurerm_public_ip",
				g.ProviderName,
				[]string{}))
		}
	}
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return resources, err
			}
			appendPage(page.Value)
		}
	} else {
		pager := client.NewListAllPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return resources, err
			}
			appendPage(page.Value)
		}
	}

	return resources, nil
}

func (g *PublicIPGenerator) listAndAddForPublicIPPrefix() ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	client, err := armnetwork.NewPublicIPPrefixesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	appendPage := func(prefixes []*armnetwork.PublicIPPrefix) {
		for _, publicIPPrefix := range prefixes {
			resources = append(resources, terraformutils.NewSimpleResource(
				*publicIPPrefix.ID,
				*publicIPPrefix.Name,
				"azurerm_public_ip_prefix",
				g.ProviderName,
				[]string{}))
		}
	}
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return resources, err
			}
			appendPage(page.Value)
		}
	} else {
		pager := client.NewListAllPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return resources, err
			}
			appendPage(page.Value)
		}
	}

	return resources, nil
}

func (g *PublicIPGenerator) InitResources() error {
	functions := []func() ([]terraformutils.Resource, error){
		g.listAndAddForPublicIPAddress,
		g.listAndAddForPublicIPPrefix,
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
