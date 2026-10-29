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

// Sweep extension for the CloudWatch Logs generator (service key "logs").
//
// Skipped types in this family (see results ledger):
//   - aws_cloudwatch_log_storage_tier_policy: GetStorageTierPolicy takes no
//     input and returns the account-level storage tier — an account singleton,
//     not an enumerable per-group resource.
//   - aws_cloudwatch_log_s3_table_integration_source: sources exist only as
//     associations under an S3 Tables integration (ListSourcesForS3TableIntegration
//     is keyed by a parent integration ARN); association glue, not a
//     standalone enumerable resource.

import (
	"context"
	"errors"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

type logsGroupRef struct {
	name string
	arn  string
}

func (g *LogsGenerator) createSweepResources(svc *cloudwatchlogs.Client) error {
	groups, err := g.listLogGroups(svc)
	if err != nil {
		return err
	}
	for _, group := range groups {
		if err := g.createLogGroupChildren(svc, group); err != nil {
			return err
		}
	}
	if err := g.createAccountPolicies(svc); err != nil {
		return err
	}
	if err := g.createLogDestinations(svc); err != nil {
		return err
	}
	if err := g.createLogResourcePolicies(svc); err != nil {
		return err
	}
	if err := g.createLogAnomalyDetectors(svc); err != nil {
		return err
	}
	if err := g.createDeliveries(svc); err != nil {
		return err
	}
	if err := g.createDeliveryDestinations(svc); err != nil {
		return err
	}
	if err := g.createDeliverySources(svc); err != nil {
		return err
	}
	return g.createQueryDefinitions(svc)
}

func (g *LogsGenerator) listLogGroups(svc *cloudwatchlogs.Client) ([]logsGroupRef, error) {
	var groups []logsGroupRef
	p := cloudwatchlogs.NewDescribeLogGroupsPaginator(svc, &cloudwatchlogs.DescribeLogGroupsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return nil, err
		}
		for _, lg := range page.LogGroups {
			arn := StringValue(lg.LogGroupArn)
			if arn == "" {
				arn = StringValue(lg.Arn)
			}
			groups = append(groups, logsGroupRef{name: StringValue(lg.LogGroupName), arn: arn})
		}
	}
	return groups, nil
}

func (g *LogsGenerator) createLogGroupChildren(svc *cloudwatchlogs.Client, group logsGroupRef) error {
	// aws_cloudwatch_log_metric_filter — importer ID "log_group_name:name".
	mf := cloudwatchlogs.NewDescribeMetricFiltersPaginator(svc, &cloudwatchlogs.DescribeMetricFiltersInput{LogGroupName: &group.name})
	for mf.HasMorePages() {
		page, err := mf.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, filter := range page.MetricFilters {
			id := group.name + ":" + StringValue(filter.FilterName)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				id, StringValue(filter.FilterName), "aws_cloudwatch_log_metric_filter", "aws",
				map[string]string{"log_group_name": group.name}, logsAllowEmptyValues, map[string]interface{}{}))
		}
	}
	// aws_cloudwatch_log_subscription_filter — importer ID "log_group_name|name".
	sf := cloudwatchlogs.NewDescribeSubscriptionFiltersPaginator(svc, &cloudwatchlogs.DescribeSubscriptionFiltersInput{LogGroupName: &group.name})
	for sf.HasMorePages() {
		page, err := sf.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, filter := range page.SubscriptionFilters {
			id := group.name + "|" + StringValue(filter.FilterName)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				id, StringValue(filter.FilterName), "aws_cloudwatch_log_subscription_filter", "aws",
				map[string]string{"log_group_name": group.name}, logsAllowEmptyValues, map[string]interface{}{}))
		}
	}
	// aws_cloudwatch_log_stream — importer ID "log_group_name:name".
	ls := cloudwatchlogs.NewDescribeLogStreamsPaginator(svc, &cloudwatchlogs.DescribeLogStreamsInput{LogGroupName: &group.name})
	for ls.HasMorePages() {
		page, err := ls.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, stream := range page.LogStreams {
			id := group.name + ":" + StringValue(stream.LogStreamName)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				id, StringValue(stream.LogStreamName), "aws_cloudwatch_log_stream", "aws",
				map[string]string{"log_group_name": group.name}, logsAllowEmptyValues, map[string]interface{}{}))
		}
	}
	// aws_cloudwatch_log_data_protection_policy — importer ID is the log group name;
	// emitted only for groups that actually carry a policy document.
	dpp, err := svc.GetDataProtectionPolicy(context.TODO(), &cloudwatchlogs.GetDataProtectionPolicyInput{LogGroupIdentifier: &group.name})
	if err != nil {
		return err
	}
	if dpp.PolicyDocument != nil && *dpp.PolicyDocument != "" {
		g.Resources = append(g.Resources, terraformutils.NewResource(
			group.name, group.name, "aws_cloudwatch_log_data_protection_policy", "aws",
			map[string]string{"log_group_name": group.name}, logsAllowEmptyValues, map[string]interface{}{}))
	}
	// aws_cloudwatch_log_index_policy — importer ID is the log group name.
	ip, err := svc.DescribeIndexPolicies(context.TODO(), &cloudwatchlogs.DescribeIndexPoliciesInput{LogGroupIdentifiers: []string{group.name}})
	if err != nil {
		return err
	}
	if len(ip.IndexPolicies) > 0 {
		g.Resources = append(g.Resources, terraformutils.NewResource(
			group.name, group.name, "aws_cloudwatch_log_index_policy", "aws",
			map[string]string{"log_group_name": group.name}, logsAllowEmptyValues, map[string]interface{}{}))
	}
	// aws_cloudwatch_log_transformer — importer ID is the log group ARN.
	_, err = svc.GetTransformer(context.TODO(), &cloudwatchlogs.GetTransformerInput{LogGroupIdentifier: &group.name})
	if err != nil {
		var notFound *types.ResourceNotFoundException
		if !errors.As(err, &notFound) {
			return err
		}
	} else if group.arn != "" {
		g.Resources = append(g.Resources, terraformutils.NewResource(
			group.arn, group.name, "aws_cloudwatch_log_transformer", "aws",
			map[string]string{"log_group_arn": group.arn}, logsAllowEmptyValues, map[string]interface{}{}))
	}
	return nil
}

// aws_cloudwatch_log_account_policy — importer ID "policy_name:policy_type".
func (g *LogsGenerator) createAccountPolicies(svc *cloudwatchlogs.Client) error {
	for _, policyType := range []types.PolicyType{
		types.PolicyTypeDataProtectionPolicy,
		types.PolicyTypeSubscriptionFilterPolicy,
		types.PolicyTypeFieldIndexPolicy,
		types.PolicyTypeTransformerPolicy,
	} {
		// DescribeAccountPolicies has no SDK paginator; paginate manually.
		var nextToken *string
		for {
			page, err := svc.DescribeAccountPolicies(context.TODO(), &cloudwatchlogs.DescribeAccountPoliciesInput{PolicyType: policyType, NextToken: nextToken})
			if err != nil {
				return err
			}
			for _, policy := range page.AccountPolicies {
				id := StringValue(policy.PolicyName) + ":" + string(policy.PolicyType)
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, StringValue(policy.PolicyName), "aws_cloudwatch_log_account_policy", "aws", logsAllowEmptyValues))
			}
			nextToken = page.NextToken
			if nextToken == nil {
				break
			}
		}
	}
	return nil
}

// aws_cloudwatch_log_destination — importer ID is the destination name.
// aws_cloudwatch_log_destination_policy — importer ID is the destination name;
// emitted for destinations carrying a non-empty access policy.
func (g *LogsGenerator) createLogDestinations(svc *cloudwatchlogs.Client) error {
	p := cloudwatchlogs.NewDescribeDestinationsPaginator(svc, &cloudwatchlogs.DescribeDestinationsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, dest := range page.Destinations {
			name := StringValue(dest.DestinationName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_cloudwatch_log_destination", "aws", logsAllowEmptyValues))
			if dest.AccessPolicy != nil && *dest.AccessPolicy != "" {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					name, name, "aws_cloudwatch_log_destination_policy", "aws", logsAllowEmptyValues))
			}
		}
	}
	return nil
}

// aws_cloudwatch_log_resource_policy — importer ID is the policy name.
func (g *LogsGenerator) createLogResourcePolicies(svc *cloudwatchlogs.Client) error {
	// DescribeResourcePolicies has no SDK paginator; paginate manually.
	var nextToken *string
	for {
		page, err := svc.DescribeResourcePolicies(context.TODO(), &cloudwatchlogs.DescribeResourcePoliciesInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, policy := range page.ResourcePolicies {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(policy.PolicyName), StringValue(policy.PolicyName),
				"aws_cloudwatch_log_resource_policy", "aws", logsAllowEmptyValues))
		}
		nextToken = page.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

// aws_cloudwatch_log_anomaly_detector — importer ID is the detector ARN.
func (g *LogsGenerator) createLogAnomalyDetectors(svc *cloudwatchlogs.Client) error {
	p := cloudwatchlogs.NewListLogAnomalyDetectorsPaginator(svc, &cloudwatchlogs.ListLogAnomalyDetectorsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, detector := range page.AnomalyDetectors {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(detector.AnomalyDetectorArn), StringValue(detector.DetectorName),
				"aws_cloudwatch_log_anomaly_detector", "aws", logsAllowEmptyValues))
		}
	}
	return nil
}

// aws_cloudwatch_log_delivery — importer ID is the delivery ID.
func (g *LogsGenerator) createDeliveries(svc *cloudwatchlogs.Client) error {
	p := cloudwatchlogs.NewDescribeDeliveriesPaginator(svc, &cloudwatchlogs.DescribeDeliveriesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, delivery := range page.Deliveries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(delivery.Id), StringValue(delivery.Id),
				"aws_cloudwatch_log_delivery", "aws", logsAllowEmptyValues))
		}
	}
	return nil
}

// aws_cloudwatch_log_delivery_destination — importer ID is the destination name.
// aws_cloudwatch_log_delivery_destination_policy — importer ID is the delivery
// destination name; emitted when the destination carries a policy.
func (g *LogsGenerator) createDeliveryDestinations(svc *cloudwatchlogs.Client) error {
	p := cloudwatchlogs.NewDescribeDeliveryDestinationsPaginator(svc, &cloudwatchlogs.DescribeDeliveryDestinationsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, dest := range page.DeliveryDestinations {
			name := StringValue(dest.Name)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_cloudwatch_log_delivery_destination", "aws", logsAllowEmptyValues))
			policy, err := svc.GetDeliveryDestinationPolicy(context.TODO(), &cloudwatchlogs.GetDeliveryDestinationPolicyInput{DeliveryDestinationName: dest.Name})
			if err != nil {
				return err
			}
			if policy.Policy != nil && policy.Policy.DeliveryDestinationPolicy != nil && *policy.Policy.DeliveryDestinationPolicy != "" {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					name, name, "aws_cloudwatch_log_delivery_destination_policy", "aws", logsAllowEmptyValues))
			}
		}
	}
	return nil
}

// aws_cloudwatch_log_delivery_source — importer ID is the source name.
func (g *LogsGenerator) createDeliverySources(svc *cloudwatchlogs.Client) error {
	p := cloudwatchlogs.NewDescribeDeliverySourcesPaginator(svc, &cloudwatchlogs.DescribeDeliverySourcesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, source := range page.DeliverySources {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(source.Name), StringValue(source.Name),
				"aws_cloudwatch_log_delivery_source", "aws", logsAllowEmptyValues))
		}
	}
	return nil
}

// aws_cloudwatch_query_definition — importer ID is the query definition ID.
func (g *LogsGenerator) createQueryDefinitions(svc *cloudwatchlogs.Client) error {
	// DescribeQueryDefinitions has no SDK paginator; paginate manually.
	var nextToken *string
	for {
		page, err := svc.DescribeQueryDefinitions(context.TODO(), &cloudwatchlogs.DescribeQueryDefinitionsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, def := range page.QueryDefinitions {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(def.QueryDefinitionId), StringValue(def.Name),
				"aws_cloudwatch_query_definition", "aws", logsAllowEmptyValues))
		}
		nextToken = page.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
