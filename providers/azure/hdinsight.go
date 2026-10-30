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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/hdinsight/armhdinsight"
)

// HDInsightGenerator covers Microsoft.HDInsight clusters. All cluster
// kinds share one ARM resource; the Terraform type is decided by the
// cluster definition kind reported by the API.
type HDInsightGenerator struct {
	AzureService
}

// hdinsightTerraformType maps a cluster definition kind to its Terraform
// type ("" for kinds with no type in this sweep, e.g. Trino/MLServices).
func hdinsightTerraformType(cluster *armhdinsight.Cluster) string {
	if cluster == nil || cluster.Properties == nil || cluster.Properties.ClusterDefinition == nil ||
		cluster.Properties.ClusterDefinition.Kind == nil {
		return ""
	}
	switch strings.ToLower(*cluster.Properties.ClusterDefinition.Kind) {
	case "hadoop":
		return "azurerm_hdinsight_hadoop_cluster"
	case "hbase":
		return "azurerm_hdinsight_hbase_cluster"
	case "interactivequery":
		return "azurerm_hdinsight_interactive_query_cluster"
	case "kafka":
		return "azurerm_hdinsight_kafka_cluster"
	case "spark":
		return "azurerm_hdinsight_spark_cluster"
	}
	return ""
}

func (g *HDInsightGenerator) InitResources() error {
	ctx := context.Background()
	client, err := armhdinsight.NewClustersClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	var clusters []*armhdinsight.Cluster
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			clusters = append(clusters, page.Value...)
		}
	} else {
		pager := client.NewListPager(nil)
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
		if resourceType := hdinsightTerraformType(cluster); resourceType != "" {
			g.AppendSimpleResource(*cluster.ID, *cluster.Name, resourceType)
		}
	}
	return nil
}
