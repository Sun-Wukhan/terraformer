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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/machinelearning/armmachinelearning"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// MachineLearningGenerator covers Microsoft.MachineLearningServices
// workspaces and their computes.
//
// Skipped in this service (see ledger):
//   - azurerm_machine_learning_datastore_blobstorage /
//     _datastore_datalake_gen2 / _datastore_fileshare: armmachinelearning
//     v1.0.0 has no Datastores client / List operation.
//   - azurerm_machine_learning_workspace_network_outbound_rule_fqdn /
//     _private_endpoint / _service_tag: armmachinelearning v1.0.0 has no
//     outbound-rules (networking) client / List operation.
type MachineLearningGenerator struct {
	AzureService
}

// classifyMLCompute maps the ARM ComputeType discriminator
// (ComputeResource.Properties.GetCompute().ComputeType) to the Terraform
// type the azurerm provider uses. Types with no Terraform counterpart in
// this cluster (VirtualMachine, Databricks, HDInsight, DataFactory,
// DataLakeAnalytics, Kubernetes) return "" and are not emitted.
func classifyMLCompute(t armmachinelearning.ComputeType) string {
	switch t {
	case armmachinelearning.ComputeTypeAmlCompute:
		return "azurerm_machine_learning_compute_cluster"
	case armmachinelearning.ComputeTypeComputeInstance:
		return "azurerm_machine_learning_compute_instance"
	case armmachinelearning.ComputeTypeAKS:
		return "azurerm_machine_learning_inference_cluster"
	case armmachinelearning.ComputeTypeSynapseSpark:
		return "azurerm_machine_learning_synapse_spark"
	default:
		return ""
	}
}

func mlComputeType(c *armmachinelearning.ComputeResource) armmachinelearning.ComputeType {
	if c == nil || c.Properties == nil {
		return ""
	}
	compute := c.Properties.GetCompute()
	if compute == nil || compute.ComputeType == nil {
		return ""
	}
	return *compute.ComputeType
}

func (g *MachineLearningGenerator) listWorkspaces() ([]*armmachinelearning.Workspace, error) {
	client, err := armmachinelearning.NewWorkspacesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	var out []*armmachinelearning.Workspace
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
		pager := client.NewListBySubscriptionPager(nil)
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

func (g *MachineLearningGenerator) listComputes(rg, workspace string) []terraformutils.Resource {
	var resources []terraformutils.Resource
	client, err := armmachinelearning.NewComputeClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		log.Println(err)
		return resources
	}
	ctx := context.Background()
	pager := client.NewListPager(rg, workspace, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return resources
		}
		for _, item := range page.Value {
			tfType := classifyMLCompute(mlComputeType(item))
			if tfType == "" || item.ID == nil || item.Name == nil {
				continue
			}
			resources = append(resources, terraformutils.NewSimpleResource(*item.ID, *item.Name, tfType, g.ProviderName, []string{}))
		}
	}
	return resources
}

func (g *MachineLearningGenerator) InitResources() error {
	workspaces, err := g.listWorkspaces()
	if err != nil {
		return err
	}
	for _, ws := range workspaces {
		if ws.ID == nil || ws.Name == nil {
			continue
		}
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(*ws.ID, *ws.Name, "azurerm_machine_learning_workspace", g.ProviderName, []string{}))
		parsed, err := ParseAzureResourceID(*ws.ID)
		if err != nil {
			return err
		}
		g.Resources = append(g.Resources, g.listComputes(parsed.ResourceGroup, *ws.Name)...)
	}
	return nil
}
