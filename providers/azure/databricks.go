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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/databricks/armdatabricks"
)

type DatabricksGenerator struct {
	AzureService
}

func (az *DatabricksGenerator) listWorkspaces() ([]*armdatabricks.Workspace, error) {
	client, err := armdatabricks.NewWorkspacesClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	var resources []*armdatabricks.Workspace
	if rg := az.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
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

func (az *DatabricksGenerator) AppendWorkspace(workspace *armdatabricks.Workspace) {
	az.AppendSimpleResource(*workspace.ID, *workspace.Name, "azurerm_databricks_workspace")
}

// Skipped for Databricks (see results-containers.md):
//   - azurerm_databricks_workspace_root_dbfs_customer_managed_key: the root
//     DBFS customer-managed key is an encryption configuration block on the
//     workspace itself (its Terraform ID is the workspace ID); armdatabricks
//     WorkspacesClient has no List for it — Get on the workspace only.

func (az *DatabricksGenerator) InitResources() error {
	ctx := context.Background()

	workspaces, err := az.listWorkspaces()
	if err != nil {
		return err
	}

	peeringClient, err := armdatabricks.NewVNetPeeringClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
	if err != nil {
		return err
	}
	for _, workspace := range workspaces {
		az.AppendWorkspace(workspace)
		parsed, err := ParseAzureResourceID(*workspace.ID)
		if err != nil {
			return err
		}
		pager := peeringClient.NewListByWorkspacePager(parsed.ResourceGroup, *workspace.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				az.AppendSimpleResource(*item.ID, *item.Name, "azurerm_databricks_virtual_network_peering")
			}
		}
	}

	connectorsClient, err := armdatabricks.NewAccessConnectorsClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
	if err != nil {
		return err
	}
	if rg := az.getResourceGroup(); rg != "" {
		pager := connectorsClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				az.AppendSimpleResource(*item.ID, *item.Name, "azurerm_databricks_access_connector")
			}
		}
	} else {
		pager := connectorsClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				az.AppendSimpleResource(*item.ID, *item.Name, "azurerm_databricks_access_connector")
			}
		}
	}
	return nil
}
