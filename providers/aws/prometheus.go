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

// Amazon Managed Service for Prometheus (AMP) generator (service key
// "prometheus"). Workspaces and scrapers are listed globally; the remaining
// types are per-workspace / per-scraper singletons or children enumerated
// through the parent.

import (
	"context"
	"errors"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/amp"
	"github.com/aws/aws-sdk-go-v2/service/amp/types"
)

var prometheusAllowEmptyValues = []string{"tags."}

type PrometheusGenerator struct {
	AWSService
}

func (g *PrometheusGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := amp.NewFromConfig(config)
	workspaceIDs, err := g.createWorkspaces(svc)
	if err != nil {
		return err
	}
	for _, workspaceID := range workspaceIDs {
		if err := g.createWorkspaceChildren(svc, workspaceID); err != nil {
			return err
		}
	}
	return g.createScrapers(svc)
}

// aws_prometheus_workspace — importer ID is the workspace ID.
// aws_prometheus_workspace_configuration — importer ID is the workspace ID;
// DescribeWorkspaceConfiguration succeeds for every workspace.
func (g *PrometheusGenerator) createWorkspaces(svc *amp.Client) ([]string, error) {
	var workspaceIDs []string
	var nextToken *string
	for {
		out, err := svc.ListWorkspaces(context.TODO(), &amp.ListWorkspacesInput{NextToken: nextToken})
		if err != nil {
			return nil, err
		}
		for _, workspace := range out.Workspaces {
			id := StringValue(workspace.WorkspaceId)
			workspaceIDs = append(workspaceIDs, id)
			name := StringValue(workspace.Alias)
			if name == "" {
				name = id
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, name, "aws_prometheus_workspace", "aws", prometheusAllowEmptyValues))
			g.Resources = append(g.Resources, terraformutils.NewResource(
				id, name, "aws_prometheus_workspace_configuration", "aws",
				map[string]string{"workspace_id": id}, prometheusAllowEmptyValues, map[string]interface{}{}))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return workspaceIDs, nil
}

func (g *PrometheusGenerator) createWorkspaceChildren(svc *amp.Client, workspaceID string) error {
	// aws_prometheus_rule_group_namespace — importer ID is the namespace ARN.
	var nextToken *string
	for {
		out, err := svc.ListRuleGroupsNamespaces(context.TODO(), &amp.ListRuleGroupsNamespacesInput{WorkspaceId: &workspaceID, NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, namespace := range out.RuleGroupsNamespaces {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				StringValue(namespace.Arn), StringValue(namespace.Name), "aws_prometheus_rule_group_namespace", "aws",
				map[string]string{"workspace_id": workspaceID}, prometheusAllowEmptyValues, map[string]interface{}{}))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	// aws_prometheus_anomaly_detector — importer ID "detector_id,workspace_id".
	nextToken = nil
	for {
		out, err := svc.ListAnomalyDetectors(context.TODO(), &amp.ListAnomalyDetectorsInput{WorkspaceId: &workspaceID, NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, detector := range out.AnomalyDetectors {
			id := StringValue(detector.AnomalyDetectorId) + "," + workspaceID
			g.Resources = append(g.Resources, terraformutils.NewResource(
				id, StringValue(detector.Alias), "aws_prometheus_anomaly_detector", "aws",
				map[string]string{"workspace_id": workspaceID}, prometheusAllowEmptyValues, map[string]interface{}{}))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	// Per-workspace singleton configurations — importer ID is the workspace
	// ID; emitted only when the configuration actually exists.
	if _, err := svc.DescribeAlertManagerDefinition(context.TODO(), &amp.DescribeAlertManagerDefinitionInput{WorkspaceId: &workspaceID}); err != nil {
		var notFound *types.ResourceNotFoundException
		if !errors.As(err, &notFound) {
			return err
		}
	} else {
		g.Resources = append(g.Resources, terraformutils.NewResource(
			workspaceID, workspaceID, "aws_prometheus_alert_manager_definition", "aws",
			map[string]string{"workspace_id": workspaceID}, prometheusAllowEmptyValues, map[string]interface{}{}))
	}
	if _, err := svc.DescribeQueryLoggingConfiguration(context.TODO(), &amp.DescribeQueryLoggingConfigurationInput{WorkspaceId: &workspaceID}); err != nil {
		var notFound *types.ResourceNotFoundException
		if !errors.As(err, &notFound) {
			return err
		}
	} else {
		g.Resources = append(g.Resources, terraformutils.NewResource(
			workspaceID, workspaceID, "aws_prometheus_query_logging_configuration", "aws",
			map[string]string{"workspace_id": workspaceID}, prometheusAllowEmptyValues, map[string]interface{}{}))
	}
	policy, err := svc.DescribeResourcePolicy(context.TODO(), &amp.DescribeResourcePolicyInput{WorkspaceId: &workspaceID})
	if err != nil {
		var notFound *types.ResourceNotFoundException
		if !errors.As(err, &notFound) {
			return err
		}
	} else if policy.PolicyDocument != nil && *policy.PolicyDocument != "" {
		g.Resources = append(g.Resources, terraformutils.NewResource(
			workspaceID, workspaceID, "aws_prometheus_resource_policy", "aws",
			map[string]string{"workspace_id": workspaceID}, prometheusAllowEmptyValues, map[string]interface{}{}))
	}
	return nil
}

// aws_prometheus_scraper — importer ID is the scraper ID.
// aws_prometheus_scraper_logging_configuration — importer ID is the scraper
// ID; emitted only for scrapers that carry a logging destination.
func (g *PrometheusGenerator) createScrapers(svc *amp.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListScrapers(context.TODO(), &amp.ListScrapersInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, scraper := range out.Scrapers {
			id := StringValue(scraper.ScraperId)
			name := StringValue(scraper.Alias)
			if name == "" {
				name = id
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, name, "aws_prometheus_scraper", "aws", prometheusAllowEmptyValues))
			logging, err := svc.DescribeScraperLoggingConfiguration(context.TODO(), &amp.DescribeScraperLoggingConfigurationInput{ScraperId: scraper.ScraperId})
			if err != nil {
				var notFound *types.ResourceNotFoundException
				if !errors.As(err, &notFound) {
					return err
				}
				continue
			}
			if logging.LoggingDestination != nil {
				g.Resources = append(g.Resources, terraformutils.NewResource(
					id, name, "aws_prometheus_scraper_logging_configuration", "aws",
					map[string]string{"scraper_id": id}, prometheusAllowEmptyValues, map[string]interface{}{}))
			}
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
