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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/datashare/armdatashare"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// DataShareGenerator covers Azure Data Share via armdatashare. Datasets
// are a single ARM resource; the provider splits them per kind.
type DataShareGenerator struct {
	AzureService
}

// dataShareDatasetType maps the ARM dataset kind to the provider resource
// type, or "" for kinds with no resource in this sweep (ADLS Gen1, SQL).
func dataShareDatasetType(kind *armdatashare.DataSetKind) string {
	if kind == nil {
		return ""
	}
	switch *kind {
	case armdatashare.DataSetKindBlob, armdatashare.DataSetKindBlobFolder, armdatashare.DataSetKindContainer:
		return "azurerm_data_share_dataset_blob_storage"
	case armdatashare.DataSetKindAdlsGen2File, armdatashare.DataSetKindAdlsGen2FileSystem, armdatashare.DataSetKindAdlsGen2Folder:
		return "azurerm_data_share_dataset_data_lake_gen2"
	case armdatashare.DataSetKindKustoCluster:
		return "azurerm_data_share_dataset_kusto_cluster"
	case armdatashare.DataSetKindKustoDatabase:
		return "azurerm_data_share_dataset_kusto_database"
	}
	return ""
}

func (g *DataShareGenerator) listAccounts() ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	accountsClient, err := armdatashare.NewAccountsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	sharesClient, err := armdatashare.NewSharesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	dataSetsClient, err := armdatashare.NewDataSetsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	processAccounts := func(accounts []*armdatashare.Account) error {
		for _, account := range accounts {
			resources = append(resources, terraformutils.NewSimpleResource(
				*account.ID, *account.Name, "azurerm_data_share_account", g.ProviderName, []string{}))
			id, err := ParseAzureResourceID(*account.ID)
			if err != nil {
				return err
			}
			sharesPager := sharesClient.NewListByAccountPager(id.ResourceGroup, *account.Name, nil)
			for sharesPager.More() {
				page, err := sharesPager.NextPage(ctx)
				if err != nil {
					log.Println(err)
					return err
				}
				for _, share := range page.Value {
					if share.ID == nil || share.Name == nil {
						continue
					}
					resources = append(resources, terraformutils.NewSimpleResource(
						*share.ID, *share.Name, "azurerm_data_share", g.ProviderName, []string{}))
					dataSetsPager := dataSetsClient.NewListBySharePager(id.ResourceGroup, *account.Name, *share.Name, nil)
					for dataSetsPager.More() {
						dataSetsPage, err := dataSetsPager.NextPage(ctx)
						if err != nil {
							log.Println(err)
							return err
						}
						for _, dataSet := range dataSetsPage.Value {
							dataSetProperties := dataSet.GetDataSet()
							if dataSetProperties == nil || dataSetProperties.ID == nil || dataSetProperties.Name == nil {
								continue
							}
							resourceType := dataShareDatasetType(dataSetProperties.Kind)
							if resourceType == "" {
								continue
							}
							resources = append(resources, terraformutils.NewSimpleResource(
								*dataSetProperties.ID, *dataSetProperties.Name, resourceType, g.ProviderName, []string{}))
						}
					}
				}
			}
		}
		return nil
	}

	if rg := g.getResourceGroup(); rg != "" {
		pager := accountsClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			if err := processAccounts(page.Value); err != nil {
				return nil, err
			}
		}
	} else {
		pager := accountsClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			if err := processAccounts(page.Value); err != nil {
				return nil, err
			}
		}
	}

	return resources, nil
}

func (g *DataShareGenerator) InitResources() error {
	resources, err := g.listAccounts()
	if err != nil {
		return err
	}
	g.Resources = append(g.Resources, resources...)
	return nil
}
