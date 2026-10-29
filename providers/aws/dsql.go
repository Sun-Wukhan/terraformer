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

package aws

// Aurora DSQL coverage (AWS sweep, database cluster).
//
// aws_dsql_cluster_peering is enumerated from each cluster's
// MultiRegionProperties: a cluster with a witness region / linked clusters
// has a peering, whose provider ID is the cluster identifier.

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dsql"
)

var dsqlAllowEmptyValues = []string{"tags."}

type DSQLGenerator struct {
	AWSService
}

func (g *DSQLGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := dsql.NewFromConfig(config)

	p := dsql.NewListClustersPaginator(svc, &dsql.ListClustersInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, cluster := range page.Clusters {
			id := StringValue(cluster.Identifier)
			// aws_dsql_cluster, ID = cluster identifier.
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_dsql_cluster", "aws", dsqlAllowEmptyValues))

			// aws_dsql_cluster_policy, ID = cluster identifier; exists only
			// when a policy is attached.
			policy, err := svc.GetClusterPolicy(context.TODO(), &dsql.GetClusterPolicyInput{Identifier: aws.String(id)})
			if err == nil && policy != nil && StringValue(policy.Policy) != "" {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, id, "aws_dsql_cluster_policy", "aws", dsqlAllowEmptyValues))
			}

			// aws_dsql_cluster_peering, ID = cluster identifier.
			full, err := svc.GetCluster(context.TODO(), &dsql.GetClusterInput{Identifier: aws.String(id)})
			if err == nil && full != nil && full.MultiRegionProperties != nil &&
				(StringValue(full.MultiRegionProperties.WitnessRegion) != "" || len(full.MultiRegionProperties.Clusters) > 0) {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, id, "aws_dsql_cluster_peering", "aws", dsqlAllowEmptyValues))
			}
		}
	}
	return nil
}
