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
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/operationalinsights/armoperationalinsights"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// LogAnalyticsGenerator covers Log Analytics via armoperationalinsights:
// workspaces, clusters, and per-workspace children (data export rules,
// saved searches, linked services, linked storage accounts, storage
// insights, Windows data sources and custom-log tables).
//
// Skipped in this family (see docs/azure.md "Known limits"):
//   - azurerm_log_analytics_cluster_customer_managed_key imports by the
//     cluster ID itself; it is a singleton association on the cluster.
//   - azurerm_log_analytics_query_pack and
//     azurerm_log_analytics_query_pack_query: armoperationalinsights v1.2.0
//     has no query packs client/List.
//   - azurerm_log_analytics_solution is a Microsoft.OperationsManagement
//     resource; armoperationalinsights v1.2.0 has no Solutions client.
//   - azurerm_log_analytics_workspace_table has no Import section in the
//     provider docs (no provider importer).
//   - azurerm_log_analytics_workspace_table_microsoft: the
//     armoperationalinsights Tables list returns legacy table objects with
//     no plan/schema marker distinguishing built-in Microsoft tables.
type LogAnalyticsGenerator struct {
	AzureService
}

// logAnalyticsDataSourceType maps an armoperationalinsights data source kind
// to the azurerm data source resource type covered by this generator.
func logAnalyticsDataSourceType(kind string) string {
	switch kind {
	case "windowsEvent":
		return "azurerm_log_analytics_datasource_windows_event"
	case "windowsPerformanceCounter":
		return "azurerm_log_analytics_datasource_windows_performance_counter"
	}
	return ""
}

// isCustomLogTable reports whether a workspace table is a custom-log table
// (custom log tables carry the _CL name suffix).
func isCustomLogTable(name string) bool {
	return strings.HasSuffix(name, "_CL")
}

func (g *LogAnalyticsGenerator) InitResources() error {
	ctx := context.Background()
	sub := g.getSubscriptionID()
	cred := g.getCredential()
	opts := g.armClientOptions()
	rgFilter := g.getResourceGroup()
	var resources []terraformutils.Resource
	add := func(id, name, typ string) {
		if id != "" && name != "" && typ != "" {
			resources = append(resources, terraformutils.NewSimpleResource(id, name, typ, "azurerm", []string{}))
		}
	}

	clustersClient, err := armoperationalinsights.NewClustersClient(sub, cred, opts)
	if err != nil {
		return err
	}
	if rgFilter != "" {
		pager := clustersClient.NewListByResourceGroupPager(rgFilter, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_log_analytics_cluster")
				}
			}
		}
	} else {
		pager := clustersClient.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_log_analytics_cluster")
				}
			}
		}
	}

	workspacesClient, err := armoperationalinsights.NewWorkspacesClient(sub, cred, opts)
	if err != nil {
		return err
	}
	type wsRef struct{ rg, name string }
	var workspaces []wsRef
	if rgFilter != "" {
		pager := workspacesClient.NewListByResourceGroupPager(rgFilter, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_log_analytics_workspace")
					workspaces = append(workspaces, wsRef{rgFilter, *item.Name})
				}
			}
		}
	} else {
		pager := workspacesClient.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				if item.ID == nil || item.Name == nil {
					continue
				}
				add(*item.ID, *item.Name, "azurerm_log_analytics_workspace")
				if parsed, err := ParseAzureResourceID(*item.ID); err == nil {
					workspaces = append(workspaces, wsRef{parsed.ResourceGroup, *item.Name})
				}
			}
		}
	}

	exportsClient, err := armoperationalinsights.NewDataExportsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	savedClient, err := armoperationalinsights.NewSavedSearchesClient(sub, cred, opts)
	if err != nil {
		return err
	}
	linkedSvcClient, err := armoperationalinsights.NewLinkedServicesClient(sub, cred, opts)
	if err != nil {
		return err
	}
	linkedStorageClient, err := armoperationalinsights.NewLinkedStorageAccountsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	insightsClient, err := armoperationalinsights.NewStorageInsightConfigsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	dataSourcesClient, err := armoperationalinsights.NewDataSourcesClient(sub, cred, opts)
	if err != nil {
		return err
	}
	tablesClient, err := armoperationalinsights.NewTablesClient(sub, cred, opts)
	if err != nil {
		return err
	}

	for _, ws := range workspaces {
		ePager := exportsClient.NewListByWorkspacePager(ws.rg, ws.name, nil)
		for ePager.More() {
			page, err := ePager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_log_analytics_data_export_rule")
				}
			}
		}
		if resp, err := savedClient.ListByWorkspace(ctx, ws.rg, ws.name, nil); err == nil {
			for _, item := range resp.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_log_analytics_saved_search")
				}
			}
		} else {
			log.Println(err)
		}
		lPager := linkedSvcClient.NewListByWorkspacePager(ws.rg, ws.name, nil)
		for lPager.More() {
			page, err := lPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_log_analytics_linked_service")
				}
			}
		}
		lsPager := linkedStorageClient.NewListByWorkspacePager(ws.rg, ws.name, nil)
		for lsPager.More() {
			page, err := lsPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_log_analytics_linked_storage_account")
				}
			}
		}
		sPager := insightsClient.NewListByWorkspacePager(ws.rg, ws.name, nil)
		for sPager.More() {
			page, err := sPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_log_analytics_storage_insights")
				}
			}
		}
		dPager := dataSourcesClient.NewListByWorkspacePager(ws.rg, ws.name, "", nil)
		for dPager.More() {
			page, err := dPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID == nil || item.Name == nil || item.Kind == nil {
					continue
				}
				add(*item.ID, *item.Name, logAnalyticsDataSourceType(string(*item.Kind)))
			}
		}
		tPager := tablesClient.NewListByWorkspacePager(ws.rg, ws.name, nil)
		for tPager.More() {
			page, err := tPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil && isCustomLogTable(*item.Name) {
					add(*item.ID, *item.Name, "azurerm_log_analytics_workspace_table_custom_log")
				}
			}
		}
	}
	g.Resources = resources
	return nil
}
