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
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
)

var bedrockAllowEmptyValues = []string{"tags."}

type BedrockGenerator struct {
	AWSService
}

// InitResources enumerates the account/region-level Amazon Bedrock control
// plane: guardrails, custom models, inference profiles, provisioned
// throughput, evaluation and model invocation jobs, plus the two regional
// singleton configurations (invocation logging and enforced guardrail).
//
// Deliberately not generated, because the Bedrock API offers no way to
// enumerate them:
//   - aws_bedrock_guardrail_version (versions are created imperatively;
//     there is no ListGuardrailVersions API)
//   - aws_bedrock_foundation_model_agreement (offers are listed, accepted
//     agreements are not)
//   - aws_bedrock_use_case_for_model_access (an account form submission,
//     not a listable resource)
func (g *BedrockGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	client := bedrock.NewFromConfig(config)

	if err := g.addGuardrails(client); err != nil {
		return err
	}
	if err := g.addCustomModels(client); err != nil {
		return err
	}
	if err := g.addInferenceProfiles(client); err != nil {
		return err
	}
	if err := g.addProvisionedModelThroughputs(client); err != nil {
		return err
	}
	if err := g.addEvaluationJobs(client); err != nil {
		return err
	}
	if err := g.addModelInvocationJobs(client); err != nil {
		return err
	}
	if err := g.addModelInvocationLoggingConfiguration(client, config.Region); err != nil {
		return err
	}
	return g.addEnforcedGuardrailConfiguration(client)
}

func (g *BedrockGenerator) addGuardrails(client *bedrock.Client) error {
	p := bedrock.NewListGuardrailsPaginator(client, &bedrock.ListGuardrailsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, guardrail := range page.Guardrails {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*guardrail.Id,
				*guardrail.Name,
				"aws_bedrock_guardrail",
				"aws",
				bedrockAllowEmptyValues,
			))
		}
	}
	return nil
}

func (g *BedrockGenerator) addCustomModels(client *bedrock.Client) error {
	p := bedrock.NewListCustomModelsPaginator(client, &bedrock.ListCustomModelsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, model := range page.ModelSummaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*model.ModelArn,
				*model.ModelName,
				"aws_bedrock_custom_model",
				"aws",
				bedrockAllowEmptyValues,
			))
		}
	}
	return nil
}

func (g *BedrockGenerator) addInferenceProfiles(client *bedrock.Client) error {
	p := bedrock.NewListInferenceProfilesPaginator(client, &bedrock.ListInferenceProfilesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, profile := range page.InferenceProfileSummaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*profile.InferenceProfileId,
				*profile.InferenceProfileName,
				"aws_bedrock_inference_profile",
				"aws",
				bedrockAllowEmptyValues,
			))
		}
	}
	return nil
}

func (g *BedrockGenerator) addProvisionedModelThroughputs(client *bedrock.Client) error {
	p := bedrock.NewListProvisionedModelThroughputsPaginator(client, &bedrock.ListProvisionedModelThroughputsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, throughput := range page.ProvisionedModelSummaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*throughput.ProvisionedModelArn,
				*throughput.ProvisionedModelName,
				"aws_bedrock_provisioned_model_throughput",
				"aws",
				bedrockAllowEmptyValues,
			))
		}
	}
	return nil
}

func (g *BedrockGenerator) addEvaluationJobs(client *bedrock.Client) error {
	p := bedrock.NewListEvaluationJobsPaginator(client, &bedrock.ListEvaluationJobsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, job := range page.JobSummaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*job.JobArn,
				*job.JobName,
				"aws_bedrock_evaluation_job",
				"aws",
				bedrockAllowEmptyValues,
			))
		}
	}
	return nil
}

func (g *BedrockGenerator) addModelInvocationJobs(client *bedrock.Client) error {
	p := bedrock.NewListModelInvocationJobsPaginator(client, &bedrock.ListModelInvocationJobsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, job := range page.InvocationJobSummaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*job.JobArn,
				*job.JobName,
				"aws_bedrock_model_invocation_job",
				"aws",
				bedrockAllowEmptyValues,
			))
		}
	}
	return nil
}

// The invocation logging configuration is a per-region singleton; the
// provider keys it by region name.
func (g *BedrockGenerator) addModelInvocationLoggingConfiguration(client *bedrock.Client, region string) error {
	out, err := client.GetModelInvocationLoggingConfiguration(context.TODO(), &bedrock.GetModelInvocationLoggingConfigurationInput{})
	if err != nil {
		return err
	}
	if out.LoggingConfig == nil {
		return nil
	}
	g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
		region,
		"bedrock_model_invocation_logging",
		"aws_bedrock_model_invocation_logging_configuration",
		"aws",
		bedrockAllowEmptyValues,
	))
	return nil
}

// Enforced guardrail configurations are account/region level; the provider
// keys each one by the configuration ID returned by the API.
func (g *BedrockGenerator) addEnforcedGuardrailConfiguration(client *bedrock.Client) error {
	p := bedrock.NewListEnforcedGuardrailsConfigurationPaginator(client, &bedrock.ListEnforcedGuardrailsConfigurationInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, enforced := range page.GuardrailsConfig {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*enforced.ConfigId,
				"bedrock_enforced_guardrail",
				"aws_bedrock_enforced_guardrail_configuration",
				"aws",
				bedrockAllowEmptyValues,
			))
		}
	}
	return nil
}
