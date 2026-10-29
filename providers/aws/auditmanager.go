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

// AWS Audit Manager generator (service key "auditmanager").
//
// Skipped types in this family (see results ledger):
//   - aws_auditmanager_account_registration: account/Region singleton;
//     the provider ID is the Region name and the SDK exposes only
//     GetAccountStatus/GetSettings (no list API).
//   - aws_auditmanager_organization_admin_account_registration: at most
//     one delegated admin account per organization; the SDK exposes only
//     GetOrganizationAdminAccount (no list API).

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/auditmanager"
	"github.com/aws/aws-sdk-go-v2/service/auditmanager/types"
)

var auditManagerAllowEmptyValues = []string{"tags."}

type AuditManagerGenerator struct {
	AWSService
}

func (g *AuditManagerGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := auditmanager.NewFromConfig(config)
	if err := g.createAssessments(svc); err != nil {
		return err
	}
	if err := g.createAssessmentReports(svc); err != nil {
		return err
	}
	if err := g.createControls(svc); err != nil {
		return err
	}
	if err := g.createFrameworks(svc); err != nil {
		return err
	}
	return g.createFrameworkShares(svc)
}

// aws_auditmanager_assessment — importer ID is the assessment ID.
// aws_auditmanager_assessment_delegation — importer ID is
// "assessment_id,role_arn,control_set_id"; delegations are embedded in
// each assessment's metadata returned by ListAssessments.
func (g *AuditManagerGenerator) createAssessments(svc *auditmanager.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListAssessments(context.TODO(), &auditmanager.ListAssessmentsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, assessment := range out.AssessmentMetadata {
			assessmentID := StringValue(assessment.Id)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				assessmentID, StringValue(assessment.Name), "aws_auditmanager_assessment", "aws", auditManagerAllowEmptyValues))
			for _, delegation := range assessment.Delegations {
				id := assessmentID + "," + StringValue(delegation.RoleArn) + "," + StringValue(delegation.ControlSetId)
				g.Resources = append(g.Resources, terraformutils.NewResource(
					id, StringValue(delegation.RoleArn), "aws_auditmanager_assessment_delegation", "aws",
					map[string]string{
						"assessment_id":  assessmentID,
						"role_arn":       StringValue(delegation.RoleArn),
						"control_set_id": StringValue(delegation.ControlSetId),
					},
					auditManagerAllowEmptyValues, map[string]interface{}{}))
			}
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

// aws_auditmanager_assessment_report — importer ID is the report ID.
func (g *AuditManagerGenerator) createAssessmentReports(svc *auditmanager.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListAssessmentReports(context.TODO(), &auditmanager.ListAssessmentReportsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, report := range out.AssessmentReports {
			name := StringValue(report.Name)
			if name == "" {
				name = StringValue(report.Id)
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(report.Id), name, "aws_auditmanager_assessment_report", "aws", auditManagerAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

// aws_auditmanager_control — importer ID is the control ID. ListControls
// requires a control type filter, so both enumerable types are probed.
func (g *AuditManagerGenerator) createControls(svc *auditmanager.Client) error {
	for _, controlType := range []types.ControlType{types.ControlTypeStandard, types.ControlTypeCustom} {
		var nextToken *string
		for {
			out, err := svc.ListControls(context.TODO(), &auditmanager.ListControlsInput{ControlType: controlType, NextToken: nextToken})
			if err != nil {
				return err
			}
			for _, control := range out.ControlMetadataList {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					StringValue(control.Id), StringValue(control.Name), "aws_auditmanager_control", "aws", auditManagerAllowEmptyValues))
			}
			nextToken = out.NextToken
			if nextToken == nil {
				break
			}
		}
	}
	return nil
}

// aws_auditmanager_framework — importer ID is the framework ID.
// ListAssessmentFrameworks requires a framework type filter, so both
// enumerable types are probed.
func (g *AuditManagerGenerator) createFrameworks(svc *auditmanager.Client) error {
	for _, frameworkType := range []types.FrameworkType{types.FrameworkTypeStandard, types.FrameworkTypeCustom} {
		var nextToken *string
		for {
			out, err := svc.ListAssessmentFrameworks(context.TODO(), &auditmanager.ListAssessmentFrameworksInput{FrameworkType: frameworkType, NextToken: nextToken})
			if err != nil {
				return err
			}
			for _, framework := range out.FrameworkMetadataList {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					StringValue(framework.Id), StringValue(framework.Name), "aws_auditmanager_framework", "aws", auditManagerAllowEmptyValues))
			}
			nextToken = out.NextToken
			if nextToken == nil {
				break
			}
		}
	}
	return nil
}

// aws_auditmanager_framework_share — importer ID is the share request
// ID. Only SENT requests are resources of this account.
func (g *AuditManagerGenerator) createFrameworkShares(svc *auditmanager.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListAssessmentFrameworkShareRequests(context.TODO(), &auditmanager.ListAssessmentFrameworkShareRequestsInput{RequestType: types.ShareRequestTypeSent, NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, share := range out.AssessmentFrameworkShareRequests {
			name := StringValue(share.FrameworkName)
			if name == "" {
				name = StringValue(share.Id)
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(share.Id), name, "aws_auditmanager_framework_share", "aws", auditManagerAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
