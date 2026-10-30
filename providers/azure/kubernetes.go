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

	armcontainerservice "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v5"
	armdeploymentsafeguards "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armdeploymentsafeguards"
	armkubernetesconfiguration "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/kubernetesconfiguration/armkubernetesconfiguration"
)

// KubernetesGenerator covers Microsoft.ContainerService managed clusters
// and their children (agent pools, extensions, Flux configurations,
// trusted-access role bindings and deployment safeguards).
type KubernetesGenerator struct {
	AzureService
}

// kubernetesClusterTerraformType maps a managed cluster to its Terraform
// type: AKS Automatic clusters carry the "Automatic" SKU name, every other
// cluster is a regular azurerm_kubernetes_cluster.
func kubernetesClusterTerraformType(cluster *armcontainerservice.ManagedCluster) string {
	if cluster != nil && cluster.SKU != nil && cluster.SKU.Name != nil &&
		string(*cluster.SKU.Name) == "Automatic" {
		return "azurerm_kubernetes_automatic_cluster"
	}
	return "azurerm_kubernetes_cluster"
}

func (g *KubernetesGenerator) listClusters() ([]*armcontainerservice.ManagedCluster, error) {
	client, err := armcontainerservice.NewManagedClustersClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	var out []*armcontainerservice.ManagedCluster
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return out, err
			}
			out = append(out, page.Value...)
		}
	} else {
		pager := client.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return out, err
			}
			out = append(out, page.Value...)
		}
	}
	return out, nil
}

func (g *KubernetesGenerator) appendClusterChildren(cluster *armcontainerservice.ManagedCluster) error {
	ctx := context.Background()
	parsed, err := ParseAzureResourceID(*cluster.ID)
	if err != nil {
		return err
	}
	rg, clusterName := parsed.ResourceGroup, *cluster.Name

	poolsClient, err := armcontainerservice.NewAgentPoolsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	poolPager := poolsClient.NewListPager(rg, clusterName, nil)
	for poolPager.More() {
		page, err := poolPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_kubernetes_cluster_node_pool")
		}
	}

	bindingsClient, err := armcontainerservice.NewTrustedAccessRoleBindingsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	bindingPager := bindingsClient.NewListPager(rg, clusterName, nil)
	for bindingPager.More() {
		page, err := bindingPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_kubernetes_cluster_trusted_access_role_binding")
		}
	}

	// Extensions and Flux configurations are scoped to the cluster via
	// Microsoft.KubernetesConfiguration; for AKS the scope segments are
	// Microsoft.ContainerService / managedClusters / <cluster name>.
	extensionsClient, err := armkubernetesconfiguration.NewExtensionsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	extPager := extensionsClient.NewListPager(rg, "Microsoft.ContainerService", "managedClusters", clusterName, nil)
	for extPager.More() {
		page, err := extPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_kubernetes_cluster_extension")
		}
	}

	fluxClient, err := armkubernetesconfiguration.NewFluxConfigurationsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	fluxPager := fluxClient.NewListPager(rg, "Microsoft.ContainerService", "managedClusters", clusterName, nil)
	for fluxPager.More() {
		page, err := fluxPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_kubernetes_flux_configuration")
		}
	}

	// Deployment safeguards are listed by cluster resource URI; the
	// Terraform ID of a safeguard is the cluster ID itself.
	safeguardsClient, err := armdeploymentsafeguards.NewDeploymentSafeguardsClient(g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	safeguardPager := safeguardsClient.NewListPager(*cluster.ID, nil)
	for safeguardPager.More() {
		page, err := safeguardPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for range page.Value {
			g.AppendSimpleResource(*cluster.ID, clusterName, "azurerm_kubernetes_cluster_deployment_safeguard")
		}
	}
	return nil
}

func (g *KubernetesGenerator) InitResources() error {
	clusters, err := g.listClusters()
	if err != nil {
		return err
	}
	for _, cluster := range clusters {
		g.AppendSimpleResource(*cluster.ID, *cluster.Name, kubernetesClusterTerraformType(cluster))
		if err := g.appendClusterChildren(cluster); err != nil {
			return err
		}
	}
	return nil
}
