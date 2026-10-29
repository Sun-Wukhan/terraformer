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

// DocumentDB depth coverage (AWS sweep, database cluster), including
// DocumentDB Elastic Clusters via the docdbelastic client.

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/docdb"
	"github.com/aws/aws-sdk-go-v2/service/docdbelastic"
)

func (g *DocDBGenerator) loadDocDBDepth(config aws.Config) error {
	svc := docdb.NewFromConfig(config)

	sp := docdb.NewDescribeDBClusterSnapshotsPaginator(svc, &docdb.DescribeDBClusterSnapshotsInput{})
	for sp.HasMorePages() {
		page, err := sp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, snap := range page.DBClusterSnapshots {
			name := StringValue(snap.DBClusterSnapshotIdentifier)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_docdb_cluster_snapshot", "aws", docDBAllowEmptyValues))
		}
	}

	es := docdb.NewDescribeEventSubscriptionsPaginator(svc, &docdb.DescribeEventSubscriptionsInput{})
	for es.HasMorePages() {
		page, err := es.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, sub := range page.EventSubscriptionsList {
			name := StringValue(sub.CustSubscriptionId)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_docdb_event_subscription", "aws", docDBAllowEmptyValues))
		}
	}

	gc := docdb.NewDescribeGlobalClustersPaginator(svc, &docdb.DescribeGlobalClustersInput{})
	for gc.HasMorePages() {
		page, err := gc.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, cluster := range page.GlobalClusters {
			name := StringValue(cluster.GlobalClusterIdentifier)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_docdb_global_cluster", "aws", docDBAllowEmptyValues))
		}
	}

	// aws_docdbelastic_cluster, ID = cluster ARN.
	elastic := docdbelastic.NewFromConfig(config)
	ep := docdbelastic.NewListClustersPaginator(elastic, &docdbelastic.ListClustersInput{})
	for ep.HasMorePages() {
		page, err := ep.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, cluster := range page.Clusters {
			arn := StringValue(cluster.ClusterArn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(cluster.ClusterName), "aws_docdbelastic_cluster", "aws", docDBAllowEmptyValues))
		}
	}

	return nil
}
