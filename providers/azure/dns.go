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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dns/armdns"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type DNSGenerator struct {
	AzureService
}

func (g *DNSGenerator) listRecordSets(resourceGroupName string, zoneName string, top *int32) ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	client, err := armdns.NewRecordSetsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	pager := client.NewListAllByDNSZonePager(resourceGroupName, zoneName, &armdns.RecordSetsClientListAllByDNSZoneOptions{Top: top})
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return resources, err
		}
		for _, recordSet := range page.Value {
			// NOTE:
			// Format example: "Microsoft.Network/dnszones/AAAA"
			recordTypeSplitted := strings.Split(*recordSet.Type, "/")
			recordType := recordTypeSplitted[len(recordTypeSplitted)-1]
			typeResourceNameMap := map[string]string{
				"A":     "azurerm_dns_a_record",
				"AAAA":  "azurerm_dns_aaaa_record",
				"CAA":   "azurerm_dns_caa_record",
				"CNAME": "azurerm_dns_cname_record",
				"MX":    "azurerm_dns_mx_record",
				"NS":    "azurerm_dns_ns_record",
				"PTR":   "azurerm_dns_ptr_record",
				"SRV":   "azurerm_dns_srv_record",
				"TXT":   "azurerm_dns_txt_record",
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

func (g *DNSGenerator) listAndAddForDNSZone() ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	client, err := armdns.NewZonesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	var pageSize int32 = 50

	appendZone := func(zone *armdns.Zone) error {
		resources = append(resources, terraformutils.NewSimpleResource(
			*zone.ID,
			*zone.Name,
			"azurerm_dns_zone",
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
		return nil
	}

	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, &armdns.ZonesClientListByResourceGroupOptions{Top: &pageSize})
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
		pager := client.NewListPager(&armdns.ZonesClientListOptions{Top: &pageSize})
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

func (g *DNSGenerator) InitResources() error {
	functions := []func() ([]terraformutils.Resource, error){
		g.listAndAddForDNSZone,
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
