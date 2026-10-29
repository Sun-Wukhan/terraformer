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

// CloudWatch Observability Admin generator (service key
// "observabilityadmin"). Importer IDs are resource ARNs for every
// implemented type (verified in the provider docs).
//
// Skipped types in this family (see results ledger):
//   - aws_observabilityadmin_telemetry_enrichment: account/Region singleton
//     (one per account per Region); provider ID is the Region name and only
//     GetTelemetryEnrichmentStatus exists.
//   - aws_observabilityadmin_telemetry_evaluation: account/Region singleton;
//     provider ID is the Region name and only GetTelemetryEvaluationStatus
//     exists.
//   - aws_observabilityadmin_telemetry_evaluation_for_organization:
//     organization singleton; provider ID is the Region name and only
//     GetTelemetryEvaluationStatusForOrganization exists.

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/observabilityadmin"
)

var observabilityadminAllowEmptyValues = []string{"tags."}

type ObservabilityAdminGenerator struct {
	AWSService
}

func (g *ObservabilityAdminGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := observabilityadmin.NewFromConfig(config)
	if err := g.createTelemetryPipelines(svc); err != nil {
		return err
	}
	if err := g.createTelemetryRules(svc); err != nil {
		return err
	}
	if err := g.createTelemetryRulesForOrganization(svc); err != nil {
		return err
	}
	if err := g.createCentralizationRulesForOrganization(svc); err != nil {
		return err
	}
	if err := g.createDatasetIntegrations(svc); err != nil {
		return err
	}
	return g.createS3TableIntegrations(svc)
}

// aws_observabilityadmin_telemetry_pipeline — importer ID is the pipeline ARN.
func (g *ObservabilityAdminGenerator) createTelemetryPipelines(svc *observabilityadmin.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListTelemetryPipelines(context.TODO(), &observabilityadmin.ListTelemetryPipelinesInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, pipeline := range out.PipelineSummaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(pipeline.Arn), StringValue(pipeline.Name),
				"aws_observabilityadmin_telemetry_pipeline", "aws", observabilityadminAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

// aws_observabilityadmin_telemetry_rule — importer ID is the rule ARN.
func (g *ObservabilityAdminGenerator) createTelemetryRules(svc *observabilityadmin.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListTelemetryRules(context.TODO(), &observabilityadmin.ListTelemetryRulesInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, rule := range out.TelemetryRuleSummaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(rule.RuleArn), StringValue(rule.RuleName),
				"aws_observabilityadmin_telemetry_rule", "aws", observabilityadminAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

// aws_observabilityadmin_telemetry_rule_for_organization — importer ID is
// the rule ARN.
func (g *ObservabilityAdminGenerator) createTelemetryRulesForOrganization(svc *observabilityadmin.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListTelemetryRulesForOrganization(context.TODO(), &observabilityadmin.ListTelemetryRulesForOrganizationInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, rule := range out.TelemetryRuleSummaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(rule.RuleArn), StringValue(rule.RuleName),
				"aws_observabilityadmin_telemetry_rule_for_organization", "aws", observabilityadminAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

// aws_observabilityadmin_centralization_rule_for_organization — importer ID
// is the rule ARN.
func (g *ObservabilityAdminGenerator) createCentralizationRulesForOrganization(svc *observabilityadmin.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListCentralizationRulesForOrganization(context.TODO(), &observabilityadmin.ListCentralizationRulesForOrganizationInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, rule := range out.CentralizationRuleSummaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(rule.RuleArn), StringValue(rule.RuleName),
				"aws_observabilityadmin_centralization_rule_for_organization", "aws", observabilityadminAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

// aws_observabilityadmin_dataset_integration — importer ID is the
// integration ARN.
func (g *ObservabilityAdminGenerator) createDatasetIntegrations(svc *observabilityadmin.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListDatasetIntegrations(context.TODO(), &observabilityadmin.ListDatasetIntegrationsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, integration := range out.DatasetIntegrationSummaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(integration.Arn), StringValue(integration.Arn),
				"aws_observabilityadmin_dataset_integration", "aws", observabilityadminAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

// aws_observabilityadmin_s3_table_integration — importer ID is the
// integration ARN.
func (g *ObservabilityAdminGenerator) createS3TableIntegrations(svc *observabilityadmin.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListS3TableIntegrations(context.TODO(), &observabilityadmin.ListS3TableIntegrationsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, integration := range out.IntegrationSummaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(integration.Arn), StringValue(integration.Arn),
				"aws_observabilityadmin_s3_table_integration", "aws", observabilityadminAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
