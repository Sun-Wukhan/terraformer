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

// MemoryDB coverage (AWS sweep, database cluster). Importer IDs are the
// resource names returned by the MemoryDB Describe APIs.

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/memorydb"
)

var memoryDBAllowEmptyValues = []string{"tags."}

type MemoryDBGenerator struct {
	AWSService
}

func (g *MemoryDBGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := memorydb.NewFromConfig(config)

	ap := memorydb.NewDescribeACLsPaginator(svc, &memorydb.DescribeACLsInput{})
	for ap.HasMorePages() {
		page, err := ap.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, acl := range page.ACLs {
			name := StringValue(acl.Name)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_memorydb_acl", "aws", memoryDBAllowEmptyValues))
		}
	}

	cp := memorydb.NewDescribeClustersPaginator(svc, &memorydb.DescribeClustersInput{})
	for cp.HasMorePages() {
		page, err := cp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, cluster := range page.Clusters {
			name := StringValue(cluster.Name)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_memorydb_cluster", "aws", memoryDBAllowEmptyValues))
		}
	}

	mp := memorydb.NewDescribeMultiRegionClustersPaginator(svc, &memorydb.DescribeMultiRegionClustersInput{})
	for mp.HasMorePages() {
		page, err := mp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, cluster := range page.MultiRegionClusters {
			name := StringValue(cluster.MultiRegionClusterName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_memorydb_multi_region_cluster", "aws", memoryDBAllowEmptyValues))
		}
	}

	pp := memorydb.NewDescribeParameterGroupsPaginator(svc, &memorydb.DescribeParameterGroupsInput{})
	for pp.HasMorePages() {
		page, err := pp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, pg := range page.ParameterGroups {
			name := StringValue(pg.Name)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_memorydb_parameter_group", "aws", memoryDBAllowEmptyValues))
		}
	}

	sp := memorydb.NewDescribeSnapshotsPaginator(svc, &memorydb.DescribeSnapshotsInput{})
	for sp.HasMorePages() {
		page, err := sp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, snap := range page.Snapshots {
			name := StringValue(snap.Name)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_memorydb_snapshot", "aws", memoryDBAllowEmptyValues))
		}
	}

	sgp := memorydb.NewDescribeSubnetGroupsPaginator(svc, &memorydb.DescribeSubnetGroupsInput{})
	for sgp.HasMorePages() {
		page, err := sgp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, sg := range page.SubnetGroups {
			name := StringValue(sg.Name)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_memorydb_subnet_group", "aws", memoryDBAllowEmptyValues))
		}
	}

	up := memorydb.NewDescribeUsersPaginator(svc, &memorydb.DescribeUsersInput{})
	for up.HasMorePages() {
		page, err := up.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, user := range page.Users {
			name := StringValue(user.Name)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_memorydb_user", "aws", memoryDBAllowEmptyValues))
		}
	}

	return nil
}
