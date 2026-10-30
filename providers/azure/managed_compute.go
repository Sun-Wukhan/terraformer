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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/devopsinfrastructure/armdevopsinfrastructure"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redhatopenshift/armredhatopenshift"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redisenterprise/armredisenterprise"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armmanagedapplications"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// ManagedComputeGenerator covers the managed-platform resources in the
// compute cluster: Microsoft.Solutions managed applications and
// application definitions, Microsoft.DevOpsInfrastructure pools
// (Managed DevOps Pools), Microsoft.Cache/redisEnterprise clusters
// (azurerm_managed_redis) and Microsoft.RedHatOpenShift clusters.
//
// Skipped in this family:
//   - azurerm_managed_disk_sas_token: side-effect-only resource whose
//     importer ID is the managed disk's own ARM ID; it has no ARM object
//     and no List surface (it only mints a SAS token on create).
//   - azurerm_managed_lustre_file_system: armstoragecache v1.0.0 (the
//     only track-2 StorageCache module) has no AmlFilesystems client —
//     only Caches and StorageTargets — so Microsoft.StorageCache/
//     amlFilesystems has no track-2 List surface.
//   - azurerm_managed_redis_access_policy_assignment: armredisenterprise
//     v1.2.0 exposes clusters and databases only; there is no access
//     policy assignment List/Get surface.
//   - azurerm_managed_redis_geo_replication: its importer ID is the
//     redisEnterprise cluster's own ARM ID (the same object emitted as
//     azurerm_managed_redis); geo-replication linking is an action on
//     the cluster, not an independent listable object.
type ManagedComputeGenerator struct {
	AzureService
}

func (g *ManagedComputeGenerator) appendSimple(id, name, resourceType string) {
	if id == "" || name == "" {
		return
	}
	g.Resources = append(g.Resources, terraformutils.NewSimpleResource(id, name, resourceType, "azurerm", []string{}))
}

func (g *ManagedComputeGenerator) InitResources() error {
	ctx := context.Background()
	subscriptionID := g.getSubscriptionID()
	cred := g.getCredential()
	opts := g.armClientOptions()
	rg := g.getResourceGroup()

	// Managed applications + definitions
	appClient, err := armmanagedapplications.NewApplicationsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	if rg != "" {
		pager := appClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_managed_application")
			}
		}
	} else {
		pager := appClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_managed_application")
			}
		}
	}
	defClient, err := armmanagedapplications.NewApplicationDefinitionsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	if rg != "" {
		pager := defClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_managed_application_definition")
			}
		}
	}

	// Managed DevOps Pools
	poolClient, err := armdevopsinfrastructure.NewPoolsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	if rg != "" {
		pager := poolClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_managed_devops_pool")
			}
		}
	} else {
		pager := poolClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_managed_devops_pool")
			}
		}
	}

	// Managed Redis (Redis Enterprise clusters)
	redisClient, err := armredisenterprise.NewClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	if rg != "" {
		pager := redisClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_managed_redis")
			}
		}
	} else {
		pager := redisClient.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_managed_redis")
			}
		}
	}

	// Red Hat OpenShift clusters
	aroClient, err := armredhatopenshift.NewOpenShiftClustersClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	if rg != "" {
		pager := aroClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_redhat_openshift_cluster")
			}
		}
	} else {
		pager := aroClient.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_redhat_openshift_cluster")
			}
		}
	}
	return nil
}
