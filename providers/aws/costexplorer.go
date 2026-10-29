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
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
)

var costexplorerAllowEmptyValues = []string{"tags."}

// CostExplorerGenerator covers Cost Explorer anomaly monitors, anomaly
// subscriptions, cost categories and cost allocation tags.
// Provider import IDs (verified): monitor = ARN; subscription = ARN;
// cost category = ARN; cost allocation tag = tag key.
type CostExplorerGenerator struct {
	AWSService
}

func (g *CostExplorerGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := costexplorer.NewFromConfig(config)

	monitors, err := svc.GetAnomalyMonitors(context.TODO(), &costexplorer.GetAnomalyMonitorsInput{})
	if err != nil {
		return err
	}
	for _, m := range monitors.AnomalyMonitors {
		arn := StringValue(m.MonitorArn)
		name := StringValue(m.MonitorName)
		if name == "" {
			name = arn
		}
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			arn, name, "aws_ce_anomaly_monitor", "aws", costexplorerAllowEmptyValues))
	}

	subs, err := svc.GetAnomalySubscriptions(context.TODO(), &costexplorer.GetAnomalySubscriptionsInput{})
	if err != nil {
		return err
	}
	for _, s := range subs.AnomalySubscriptions {
		arn := StringValue(s.SubscriptionArn)
		name := StringValue(s.SubscriptionName)
		if name == "" {
			name = arn
		}
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			arn, name, "aws_ce_anomaly_subscription", "aws", costexplorerAllowEmptyValues))
	}

	cats, err := svc.ListCostCategoryDefinitions(context.TODO(), &costexplorer.ListCostCategoryDefinitionsInput{})
	if err != nil {
		return err
	}
	for _, c := range cats.CostCategoryReferences {
		arn := StringValue(c.CostCategoryArn)
		name := StringValue(c.Name)
		if name == "" {
			name = arn
		}
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			arn, name, "aws_ce_cost_category", "aws", costexplorerAllowEmptyValues))
	}

	tags, err := svc.ListCostAllocationTags(context.TODO(), &costexplorer.ListCostAllocationTagsInput{})
	if err != nil {
		return err
	}
	for _, t := range tags.CostAllocationTags {
		key := StringValue(t.TagKey)
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			key, key, "aws_ce_cost_allocation_tag", "aws", costexplorerAllowEmptyValues))
	}
	return nil
}
