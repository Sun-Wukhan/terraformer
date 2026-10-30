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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/elasticsan/armelasticsan"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// ElasticSANGenerator covers Azure Elastic SAN via armelasticsan: the SAN
// itself, its volume groups and their volumes, all importing by ARM ID.
type ElasticSANGenerator struct {
	AzureService
}

func (g *ElasticSANGenerator) InitResources() error {
	ctx := context.Background()
	sub := g.getSubscriptionID()
	opts := g.armClientOptions()
	var resources []terraformutils.Resource
	add := func(id, name, typ string) {
		if id != "" && name != "" {
			resources = append(resources, terraformutils.NewSimpleResource(id, name, typ, "azurerm", []string{}))
		}
	}
	sansClient, err := armelasticsan.NewElasticSansClient(sub, g.getCredential(), opts)
	if err != nil {
		return err
	}
	groupsClient, err := armelasticsan.NewVolumeGroupsClient(sub, g.getCredential(), opts)
	if err != nil {
		return err
	}
	volumesClient, err := armelasticsan.NewVolumesClient(sub, g.getCredential(), opts)
	if err != nil {
		return err
	}

	type sanRef struct{ rg, name string }
	var sans []sanRef
	if rg := g.getResourceGroup(); rg != "" {
		pager := sansClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, s := range page.Value {
				if s.ID != nil && s.Name != nil {
					add(*s.ID, *s.Name, "azurerm_elastic_san")
					sans = append(sans, sanRef{rg, *s.Name})
				}
			}
		}
	} else {
		pager := sansClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, s := range page.Value {
				if s.ID == nil || s.Name == nil {
					continue
				}
				add(*s.ID, *s.Name, "azurerm_elastic_san")
				if parsed, err := ParseAzureResourceID(*s.ID); err == nil {
					sans = append(sans, sanRef{parsed.ResourceGroup, *s.Name})
				}
			}
		}
	}
	for _, san := range sans {
		gPager := groupsClient.NewListByElasticSanPager(san.rg, san.name, nil)
		for gPager.More() {
			page, err := gPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, grp := range page.Value {
				if grp.ID == nil || grp.Name == nil {
					continue
				}
				add(*grp.ID, *grp.Name, "azurerm_elastic_san_volume_group")
				vPager := volumesClient.NewListByVolumeGroupPager(san.rg, san.name, *grp.Name, nil)
				for vPager.More() {
					vPage, err := vPager.NextPage(ctx)
					if err != nil {
						log.Println(err)
						break
					}
					for _, vol := range vPage.Value {
						if vol.ID != nil && vol.Name != nil {
							add(*vol.ID, *vol.Name, "azurerm_elastic_san_volume")
						}
					}
				}
			}
		}
	}
	g.Resources = resources
	return nil
}
