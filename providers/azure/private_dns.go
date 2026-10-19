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
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type PrivateDNSGenerator struct {
	AzureService
}

func (g *PrivateDNSGenerator) listRecordSets(resourceGroupName string, privateZoneName string, top *int32) ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	client, err := armprivatedns.NewRecordSetsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	pager := client.NewListPager(resourceGroupName, privateZoneName, &armprivatedns.RecordSetsClientListOptions{Top: top})
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			break
		}
		for _, recordSet := range page.Value {
			// NOTE:
			// Format example: "Microsoft.Network/privateDnsZones/CNAME"
			recordTypeSplitted := strings.Split(*recordSet.Type, "/")
			recordType := recordTypeSplitted[len(recordTypeSplitted)-1]
			typeResourceNameMap := map[string]string{
				"A":     "azurerm_private_dns_a_record",
				"AAAA":  "azurerm_private_dns_aaaa_record",
				"CNAME": "azurerm_private_dns_cname_record",
				"MX":    "azurerm_private_dns_mx_record",
				"PTR":   "azurerm_private_dns_ptr_record",
				"SRV":   "azurerm_private_dns_srv_record",
				"TXT":   "azurerm_private_dns_txt_record",
			}
			if resName, exist := typeResourceNameMap[recordType]; exist {
				resources = append(resources, terraformutils.NewSimpleResource(
					*recordSet.ID,
					*recordSet.Name,
					resName,
					g.ProviderName,
					[]string{}))
			}
		}
	}
	return resources, nil
}

func (g *PrivateDNSGenerator) listVirtualNetworkLinks(resourceGroupName string, privateZoneName string, pageSize *int32) ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	client, err := armprivatedns.NewVirtualNetworkLinksClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	pager := client.NewListPager(resourceGroupName, privateZoneName, &armprivatedns.VirtualNetworkLinksClientListOptions{Top: pageSize})
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			break
		}
		for _, virtualNetworkLink := range page.Value {
			resources = append(resources, terraformutils.NewSimpleResource(
				*virtualNetworkLink.ID,
				*virtualNetworkLink.Name,
				"azurerm_private_dns_zone_virtual_network_link",
				g.ProviderName,
				[]string{}))
		}
	}

	return resources, nil
}

func (g *PrivateDNSGenerator) listAndAddForPrivateDNSZone() ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	client, err := armprivatedns.NewPrivateZonesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	var pageSize int32 = 50

	appendZone := func(zone *armprivatedns.PrivateZone) error {
		resources = append(resources, terraformutils.NewSimpleResource(
			*zone.ID,
			*zone.Name,
			"azurerm_private_dns_zone",
			g.ProviderName,
			[]string{}))

		id, err := ParseAzureResourceID(*zone.ID)
		if err != nil {
			return err
		}

		records, err := g.listRecordSets(id.ResourceGroup, *zone.Name, &pageSize)
		if err != nil {
			return err
		}
		resources = append(resources, records...)

		networkLinks, err := g.listVirtualNetworkLinks(id.ResourceGroup, *zone.Name, &pageSize)
		if err != nil {
			return err
		}
		resources = append(resources, networkLinks...)
		return nil
	}

	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, &armprivatedns.PrivateZonesClientListByResourceGroupOptions{Top: &pageSize})
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return resources, err
			}
			for _, zone := range page.Value {
				if err := appendZone(zone); err != nil {
					return nil, err
				}
			}
		}
	} else {
		pager := client.NewListPager(&armprivatedns.PrivateZonesClientListOptions{Top: &pageSize})
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return resources, err
			}
			for _, zone := range page.Value {
				if err := appendZone(zone); err != nil {
					return nil, err
				}
			}
		}
	}

	return resources, nil
}

func (g *PrivateDNSGenerator) InitResources() error {
	functions := []func() ([]terraformutils.Resource, error){
		g.listAndAddForPrivateDNSZone,
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
