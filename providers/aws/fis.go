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

// AWS Fault Injection Service (FIS) generator (service key "fis").
//
// Skipped types in this family (see results ledger):
//   - aws_fis_safety_lever_state: exactly one safety lever exists per
//     account and Region and it always exists; the provider ID is the
//     Region name and the SDK exposes only GetSafetyLever (no list API).

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/fis"
)

var fisAllowEmptyValues = []string{"tags."}

type FisGenerator struct {
	AWSService
}

func (g *FisGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := fis.NewFromConfig(config)
	var nextToken *string
	for {
		out, err := svc.ListExperimentTemplates(context.TODO(), &fis.ListExperimentTemplatesInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, template := range out.ExperimentTemplates {
			templateID := StringValue(template.Id)
			// aws_fis_experiment_template — importer ID is the template ID.
			name := StringValue(template.Description)
			if name == "" {
				name = templateID
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				templateID, name, "aws_fis_experiment_template", "aws", fisAllowEmptyValues))
			if err := g.createTargetAccountConfigurations(svc, templateID); err != nil {
				return err
			}
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

// aws_fis_target_account_configuration — importer ID
// "account_id,experiment_template_id".
func (g *FisGenerator) createTargetAccountConfigurations(svc *fis.Client, experimentTemplateID string) error {
	var nextToken *string
	for {
		out, err := svc.ListTargetAccountConfigurations(context.TODO(), &fis.ListTargetAccountConfigurationsInput{ExperimentTemplateId: &experimentTemplateID, NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, targetAccount := range out.TargetAccountConfigurations {
			id := StringValue(targetAccount.AccountId) + "," + experimentTemplateID
			g.Resources = append(g.Resources, terraformutils.NewResource(
				id, StringValue(targetAccount.AccountId), "aws_fis_target_account_configuration", "aws",
				map[string]string{"experiment_template_id": experimentTemplateID, "account_id": StringValue(targetAccount.AccountId)},
				fisAllowEmptyValues, map[string]interface{}{}))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
