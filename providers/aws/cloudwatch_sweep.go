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

// Sweep extension for the CloudWatch generator (service key "cloudwatch").
//
// Skipped types in this family (see results ledger):
//   - aws_cloudwatch_contributor_managed_insight_rule: ListManagedInsightRules
//     requires a caller-supplied ResourceARN; no API enumerates the resources
//     that carry managed rules.
//   - aws_cloudwatch_event_endpoint: the provider importer exists (endpoint
//     name), but aws-sdk-go-v2 cloudwatchevents exposes no Endpoint operations
//     at all (no List/Describe), so endpoints cannot be enumerated with the SDK.
//   - aws_cloudwatch_event_permission: no list API; permissions are statements
//     embedded in the event bus policy JSON (Put/RemovePermission are write-only).
//   - aws_cloudwatch_otel_enrichment: GetOTelEnrichment is an account-level
//     singleton status with no list API.
//   - aws_cloudwatch_resource_metrics_configuration: Get-only, keyed by a
//     caller-supplied resource ARN; no list API.

import (
	"context"
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	cloudwatchtypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchevents"
)

func (g *CloudWatchGenerator) createSweepResources(cloudwatchSvc *cloudwatch.Client, eventsSvc *cloudwatchevents.Client) error {
	if err := g.createCompositeAlarms(cloudwatchSvc); err != nil {
		return err
	}
	if err := g.createAlarmMuteRules(cloudwatchSvc); err != nil {
		return err
	}
	if err := g.createContributorInsightRules(cloudwatchSvc); err != nil {
		return err
	}
	if err := g.createMetricStreams(cloudwatchSvc); err != nil {
		return err
	}
	if err := g.createEventBuses(eventsSvc); err != nil {
		return err
	}
	if err := g.createEventArchives(eventsSvc); err != nil {
		return err
	}
	if err := g.createEventAPIDestinations(eventsSvc); err != nil {
		return err
	}
	return g.createEventConnections(eventsSvc)
}

// aws_cloudwatch_composite_alarm — importer ID is the alarm name.
func (g *CloudWatchGenerator) createCompositeAlarms(svc *cloudwatch.Client) error {
	p := cloudwatch.NewDescribeAlarmsPaginator(svc, &cloudwatch.DescribeAlarmsInput{
		AlarmTypes: []cloudwatchtypes.AlarmType{cloudwatchtypes.AlarmTypeCompositeAlarm},
	})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, alarm := range page.CompositeAlarms {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(alarm.AlarmName),
				StringValue(alarm.AlarmName),
				"aws_cloudwatch_composite_alarm",
				"aws",
				cloudwatchAllowEmptyValues))
		}
	}
	return nil
}

// aws_cloudwatch_alarm_mute_rule — importer ID is the rule name. The list API
// returns only the rule ARN, whose final segment is the rule name
// (arn:aws:cloudwatch:<region>:<account>:alarm-mute-rule:<name> or .../<name>).
func (g *CloudWatchGenerator) createAlarmMuteRules(svc *cloudwatch.Client) error {
	p := cloudwatch.NewListAlarmMuteRulesPaginator(svc, &cloudwatch.ListAlarmMuteRulesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, summary := range page.AlarmMuteRuleSummaries {
			arn := StringValue(summary.AlarmMuteRuleArn)
			name := arn
			if idx := strings.LastIndex(name, "alarm-mute-rule"); idx >= 0 {
				name = strings.TrimLeft(name[idx+len("alarm-mute-rule"):], ":/")
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name,
				name,
				"aws_cloudwatch_alarm_mute_rule",
				"aws",
				cloudwatchAllowEmptyValues))
		}
	}
	return nil
}

// aws_cloudwatch_contributor_insight_rule — importer ID is the rule name.
// Rules flagged ManagedRule belong to aws_cloudwatch_contributor_managed_insight_rule
// and are not enumerable globally (see skip note above), so they are excluded here.
func (g *CloudWatchGenerator) createContributorInsightRules(svc *cloudwatch.Client) error {
	var nextToken *string
	for {
		out, err := svc.DescribeInsightRules(context.TODO(), &cloudwatch.DescribeInsightRulesInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, rule := range out.InsightRules {
			if rule.ManagedRule != nil && *rule.ManagedRule {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(rule.Name),
				StringValue(rule.Name),
				"aws_cloudwatch_contributor_insight_rule",
				"aws",
				cloudwatchAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

// aws_cloudwatch_metric_stream — importer ID is the stream name.
func (g *CloudWatchGenerator) createMetricStreams(svc *cloudwatch.Client) error {
	p := cloudwatch.NewListMetricStreamsPaginator(svc, &cloudwatch.ListMetricStreamsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, entry := range page.Entries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(entry.Name),
				StringValue(entry.Name),
				"aws_cloudwatch_metric_stream",
				"aws",
				cloudwatchAllowEmptyValues))
		}
	}
	return nil
}

// aws_cloudwatch_event_bus — importer ID is the bus name.
// aws_cloudwatch_event_bus_policy — importer ID is the bus name; emitted for
// every bus whose DescribeEventBus response carries a non-empty policy.
func (g *CloudWatchGenerator) createEventBuses(svc *cloudwatchevents.Client) error {
	// The cloudwatchevents SDK ships no paginators; paginate manually.
	var nextToken *string
	for {
		page, err := svc.ListEventBuses(context.TODO(), &cloudwatchevents.ListEventBusesInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, bus := range page.EventBuses {
			name := StringValue(bus.Name)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name,
				name,
				"aws_cloudwatch_event_bus",
				"aws",
				cloudwatchAllowEmptyValues))
			described, err := svc.DescribeEventBus(context.TODO(), &cloudwatchevents.DescribeEventBusInput{Name: bus.Name})
			if err != nil {
				return err
			}
			if described.Policy != nil && *described.Policy != "" {
				g.Resources = append(g.Resources, terraformutils.NewResource(
					name,
					name,
					"aws_cloudwatch_event_bus_policy",
					"aws",
					map[string]string{"event_bus_name": name},
					cloudwatchAllowEmptyValues,
					map[string]interface{}{}))
			}
		}
		nextToken = page.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

// aws_cloudwatch_event_archive — importer ID is the archive name.
func (g *CloudWatchGenerator) createEventArchives(svc *cloudwatchevents.Client) error {
	var nextToken *string
	for {
		page, err := svc.ListArchives(context.TODO(), &cloudwatchevents.ListArchivesInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, archive := range page.Archives {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(archive.ArchiveName),
				StringValue(archive.ArchiveName),
				"aws_cloudwatch_event_archive",
				"aws",
				cloudwatchAllowEmptyValues))
		}
		nextToken = page.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

// aws_cloudwatch_event_api_destination — importer ID is the destination name.
func (g *CloudWatchGenerator) createEventAPIDestinations(svc *cloudwatchevents.Client) error {
	var nextToken *string
	for {
		page, err := svc.ListApiDestinations(context.TODO(), &cloudwatchevents.ListApiDestinationsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, dest := range page.ApiDestinations {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(dest.Name),
				StringValue(dest.Name),
				"aws_cloudwatch_event_api_destination",
				"aws",
				cloudwatchAllowEmptyValues))
		}
		nextToken = page.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

// aws_cloudwatch_event_connection — importer ID is the connection name.
func (g *CloudWatchGenerator) createEventConnections(svc *cloudwatchevents.Client) error {
	var nextToken *string
	for {
		page, err := svc.ListConnections(context.TODO(), &cloudwatchevents.ListConnectionsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, conn := range page.Connections {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(conn.Name),
				StringValue(conn.Name),
				"aws_cloudwatch_event_connection",
				"aws",
				cloudwatchAllowEmptyValues))
		}
		nextToken = page.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
