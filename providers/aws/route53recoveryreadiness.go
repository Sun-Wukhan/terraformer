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

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/route53recoveryreadiness"
)

var route53RecoveryReadinessAllowEmptyValues = []string{"tags."}

type Route53RecoveryReadinessGenerator struct {
	AWSService
}

// InitResources enumerates the Amazon Route 53 Application Recovery
// Controller readiness objects: cells, readiness checks, recovery
// groups and resource sets. Every resource in this family is keyed
// (and imported) by its name.
func (g *Route53RecoveryReadinessGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	client := route53recoveryreadiness.NewFromConfig(config)
	for _, add := range []func(*route53recoveryreadiness.Client) error{
		g.addCells,
		g.addReadinessChecks,
		g.addRecoveryGroups,
		g.addResourceSets,
	} {
		if err := add(client); err != nil {
			return err
		}
	}
	return nil
}

func (g *Route53RecoveryReadinessGenerator) addCells(client *route53recoveryreadiness.Client) error {
	input := &route53recoveryreadiness.ListCellsInput{}
	for {
		out, err := client.ListCells(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, cell := range out.Cells {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(cell.CellName),
				StringValue(cell.CellName),
				"aws_route53recoveryreadiness_cell",
				"aws",
				route53RecoveryReadinessAllowEmptyValues,
			))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *Route53RecoveryReadinessGenerator) addReadinessChecks(client *route53recoveryreadiness.Client) error {
	input := &route53recoveryreadiness.ListReadinessChecksInput{}
	for {
		out, err := client.ListReadinessChecks(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, check := range out.ReadinessChecks {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(check.ReadinessCheckName),
				StringValue(check.ReadinessCheckName),
				"aws_route53recoveryreadiness_readiness_check",
				"aws",
				route53RecoveryReadinessAllowEmptyValues,
			))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *Route53RecoveryReadinessGenerator) addRecoveryGroups(client *route53recoveryreadiness.Client) error {
	input := &route53recoveryreadiness.ListRecoveryGroupsInput{}
	for {
		out, err := client.ListRecoveryGroups(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, group := range out.RecoveryGroups {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(group.RecoveryGroupName),
				StringValue(group.RecoveryGroupName),
				"aws_route53recoveryreadiness_recovery_group",
				"aws",
				route53RecoveryReadinessAllowEmptyValues,
			))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *Route53RecoveryReadinessGenerator) addResourceSets(client *route53recoveryreadiness.Client) error {
	input := &route53recoveryreadiness.ListResourceSetsInput{}
	for {
		out, err := client.ListResourceSets(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, set := range out.ResourceSets {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(set.ResourceSetName),
				StringValue(set.ResourceSetName),
				"aws_route53recoveryreadiness_resource_set",
				"aws",
				route53RecoveryReadinessAllowEmptyValues,
			))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}
