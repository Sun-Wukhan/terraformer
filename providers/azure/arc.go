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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/hybridcompute/armhybridcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/hybridkubernetes/armhybridkubernetes"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/kubernetesconfiguration/armkubernetesconfiguration"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourceconnector/armresourceconnector"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// ArcGenerator covers Azure Arc: HybridCompute machines (+ extensions),
// Arc private link scopes, Resource Bridge appliances, and Arc-enabled
// Kubernetes connected clusters (+ KubernetesConfiguration extensions and
// flux configurations enumerated per cluster).
//
// Skipped in this family:
//   - azurerm_arc_kubernetes_provisioned_cluster: its importer ID is the
//     same Microsoft.Kubernetes/connectedClusters ARM ID as
//     azurerm_arc_kubernetes_cluster and there is no separate List surface;
//     emitting both would duplicate every connected cluster.
type ArcGenerator struct {
	AzureService
}

func (g *ArcGenerator) appendSimple(id, name, resourceType string) {
	if id == "" || name == "" {
		return
	}
	g.Resources = append(g.Resources, terraformutils.NewSimpleResource(id, name, resourceType, "azurerm", []string{}))
}

func (g *ArcGenerator) InitResources() error {
	ctx := context.Background()
	subscriptionID := g.getSubscriptionID()
	cred := g.getCredential()
	opts := g.armClientOptions()
	rg := g.getResourceGroup()

	// Arc machines + extensions
	machineClient, err := armhybridcompute.NewMachinesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	machineExtClient, err := armhybridcompute.NewMachineExtensionsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	var machines []*armhybridcompute.Machine
	if rg != "" {
		pager := machineClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			machines = append(machines, page.Value...)
		}
	} else {
		pager := machineClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			machines = append(machines, page.Value...)
		}
	}
	for _, machine := range machines {
		g.appendSimple(*machine.ID, *machine.Name, "azurerm_arc_machine")
		machineRG := resourceGroupFromID(*machine.ID)
		pager := machineExtClient.NewListPager(machineRG, *machine.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_arc_machine_extension")
			}
		}
	}

	// Arc private link scopes
	plsClient, err := armhybridcompute.NewPrivateLinkScopesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	if rg != "" {
		pager := plsClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_arc_private_link_scope")
			}
		}
	} else {
		pager := plsClient.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_arc_private_link_scope")
			}
		}
	}

	// Resource Bridge appliances
	applianceClient, err := armresourceconnector.NewAppliancesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	if rg != "" {
		pager := applianceClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_arc_resource_bridge_appliance")
			}
		}
	} else {
		pager := applianceClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_arc_resource_bridge_appliance")
			}
		}
	}

	// Arc-enabled Kubernetes: connected clusters + extensions/flux per cluster
	clusterClient, err := armhybridkubernetes.NewConnectedClusterClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	extClient, err := armkubernetesconfiguration.NewExtensionsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	fluxClient, err := armkubernetesconfiguration.NewFluxConfigurationsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	var clusters []*armhybridkubernetes.ConnectedCluster
	if rg != "" {
		pager := clusterClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			clusters = append(clusters, page.Value...)
		}
	} else {
		pager := clusterClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			clusters = append(clusters, page.Value...)
		}
	}
	for _, cluster := range clusters {
		g.appendSimple(*cluster.ID, *cluster.Name, "azurerm_arc_kubernetes_cluster")
		clusterRG := resourceGroupFromID(*cluster.ID)
		pager := extClient.NewListPager(clusterRG, "Microsoft.Kubernetes", "connectedClusters", *cluster.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_arc_kubernetes_cluster_extension")
			}
		}
		fpager := fluxClient.NewListPager(clusterRG, "Microsoft.Kubernetes", "connectedClusters", *cluster.Name, nil)
		for fpager.More() {
			page, err := fpager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_arc_kubernetes_flux_configuration")
			}
		}
	}
	return nil
}
