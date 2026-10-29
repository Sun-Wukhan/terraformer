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

// Neptune coverage (AWS sweep, database cluster). All importer IDs are the
// resource identifiers returned by the Neptune Describe APIs.

import (
	"context"
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/neptune"
)

var neptuneAllowEmptyValues = []string{"tags."}

type NeptuneGenerator struct {
	AWSService
}

func (g *NeptuneGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := neptune.NewFromConfig(config)

	p := neptune.NewDescribeDBClustersPaginator(svc, &neptune.DescribeDBClustersInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, c := range page.DBClusters {
			name := StringValue(c.DBClusterIdentifier)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_neptune_cluster", "aws", neptuneAllowEmptyValues))
		}
	}

	ip := neptune.NewDescribeDBInstancesPaginator(svc, &neptune.DescribeDBInstancesInput{})
	for ip.HasMorePages() {
		page, err := ip.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, inst := range page.DBInstances {
			name := StringValue(inst.DBInstanceIdentifier)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_neptune_cluster_instance", "aws", neptuneAllowEmptyValues))
		}
	}

	ep := neptune.NewDescribeDBClusterEndpointsPaginator(svc, &neptune.DescribeDBClusterEndpointsInput{})
	for ep.HasMorePages() {
		page, err := ep.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, endpoint := range page.DBClusterEndpoints {
			name := StringValue(endpoint.DBClusterEndpointIdentifier)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_neptune_cluster_endpoint", "aws", neptuneAllowEmptyValues))
		}
	}

	cpg := neptune.NewDescribeDBClusterParameterGroupsPaginator(svc, &neptune.DescribeDBClusterParameterGroupsInput{})
	for cpg.HasMorePages() {
		page, err := cpg.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, pg := range page.DBClusterParameterGroups {
			name := StringValue(pg.DBClusterParameterGroupName)
			if strings.Contains(name, ".") {
				continue // default groups
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_neptune_cluster_parameter_group", "aws", neptuneAllowEmptyValues))
		}
	}

	pg := neptune.NewDescribeDBParameterGroupsPaginator(svc, &neptune.DescribeDBParameterGroupsInput{})
	for pg.HasMorePages() {
		page, err := pg.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, group := range page.DBParameterGroups {
			name := StringValue(group.DBParameterGroupName)
			if strings.Contains(name, ".") {
				continue // default groups
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_neptune_parameter_group", "aws", neptuneAllowEmptyValues))
		}
	}

	sp := neptune.NewDescribeDBClusterSnapshotsPaginator(svc, &neptune.DescribeDBClusterSnapshotsInput{})
	for sp.HasMorePages() {
		page, err := sp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, snap := range page.DBClusterSnapshots {
			name := StringValue(snap.DBClusterSnapshotIdentifier)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_neptune_cluster_snapshot", "aws", neptuneAllowEmptyValues))
		}
	}

	sg := neptune.NewDescribeDBSubnetGroupsPaginator(svc, &neptune.DescribeDBSubnetGroupsInput{})
	for sg.HasMorePages() {
		page, err := sg.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, group := range page.DBSubnetGroups {
			name := StringValue(group.DBSubnetGroupName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_neptune_subnet_group", "aws", neptuneAllowEmptyValues))
		}
	}

	es := neptune.NewDescribeEventSubscriptionsPaginator(svc, &neptune.DescribeEventSubscriptionsInput{})
	for es.HasMorePages() {
		page, err := es.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, sub := range page.EventSubscriptionsList {
			name := StringValue(sub.CustSubscriptionId)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_neptune_event_subscription", "aws", neptuneAllowEmptyValues))
		}
	}

	gc := neptune.NewDescribeGlobalClustersPaginator(svc, &neptune.DescribeGlobalClustersInput{})
	for gc.HasMorePages() {
		page, err := gc.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, cluster := range page.GlobalClusters {
			name := StringValue(cluster.GlobalClusterIdentifier)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_neptune_global_cluster", "aws", neptuneAllowEmptyValues))
		}
	}

	return nil
}
