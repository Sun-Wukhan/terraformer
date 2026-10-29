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

// DAX coverage (AWS sweep, database cluster). Importer IDs are names.

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/dax"
)

var daxAllowEmptyValues = []string{"tags."}

type DAXGenerator struct {
	AWSService
}

func (g *DAXGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := dax.NewFromConfig(config)

	// The DAX SDK ships no paginators; page manually via NextToken.
	var clusterToken *string
	for {
		page, err := svc.DescribeClusters(context.TODO(), &dax.DescribeClustersInput{NextToken: clusterToken})
		if err != nil {
			return err
		}
		for _, cluster := range page.Clusters {
			name := StringValue(cluster.ClusterName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_dax_cluster", "aws", daxAllowEmptyValues))
		}
		if page.NextToken == nil {
			break
		}
		clusterToken = page.NextToken
	}

	var parameterGroupToken *string
	for {
		page, err := svc.DescribeParameterGroups(context.TODO(), &dax.DescribeParameterGroupsInput{NextToken: parameterGroupToken})
		if err != nil {
			return err
		}
		for _, pg := range page.ParameterGroups {
			name := StringValue(pg.ParameterGroupName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_dax_parameter_group", "aws", daxAllowEmptyValues))
		}
		if page.NextToken == nil {
			break
		}
		parameterGroupToken = page.NextToken
	}

	var subnetGroupToken *string
	for {
		page, err := svc.DescribeSubnetGroups(context.TODO(), &dax.DescribeSubnetGroupsInput{NextToken: subnetGroupToken})
		if err != nil {
			return err
		}
		for _, sg := range page.SubnetGroups {
			name := StringValue(sg.SubnetGroupName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_dax_subnet_group", "aws", daxAllowEmptyValues))
		}
		if page.NextToken == nil {
			break
		}
		subnetGroupToken = page.NextToken
	}

	return nil
}
