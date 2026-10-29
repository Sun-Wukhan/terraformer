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

// AWS Resilience Hub V2 generator (service key "resiliencehubv2").

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/resiliencehubv2"
)

var resilienceHubV2AllowEmptyValues = []string{"tags."}

type ResilienceHubV2Generator struct {
	AWSService
}

func (g *ResilienceHubV2Generator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := resiliencehubv2.NewFromConfig(config)
	if err := g.createPolicies(svc); err != nil {
		return err
	}
	if err := g.createServices(svc); err != nil {
		return err
	}
	return g.createSystems(svc)
}

// aws_resiliencehubv2_policy — importer ID is the policy ARN.
func (g *ResilienceHubV2Generator) createPolicies(svc *resiliencehubv2.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListPolicies(context.TODO(), &resiliencehubv2.ListPoliciesInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, policy := range out.PolicySummaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(policy.PolicyArn), StringValue(policy.Name), "aws_resiliencehubv2_policy", "aws", resilienceHubV2AllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

// aws_resiliencehubv2_service — importer ID is the service ARN.
// Per-service children:
//   - aws_resiliencehubv2_assertion — ID "service_arn,assertion_id"
//   - aws_resiliencehubv2_input_source — ID "service_arn,input_source_id"
//   - aws_resiliencehubv2_service_function — ID
//     "service_arn,service_function_id"
func (g *ResilienceHubV2Generator) createServices(svc *resiliencehubv2.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListServices(context.TODO(), &resiliencehubv2.ListServicesInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, service := range out.ServiceSummaries {
			serviceArn := StringValue(service.ServiceArn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				serviceArn, StringValue(service.Name), "aws_resiliencehubv2_service", "aws", resilienceHubV2AllowEmptyValues))
			if err := g.createAssertions(svc, serviceArn); err != nil {
				return err
			}
			if err := g.createInputSources(svc, serviceArn); err != nil {
				return err
			}
			if err := g.createServiceFunctions(svc, serviceArn); err != nil {
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

func (g *ResilienceHubV2Generator) createAssertions(svc *resiliencehubv2.Client, serviceArn string) error {
	var nextToken *string
	for {
		out, err := svc.ListAssertions(context.TODO(), &resiliencehubv2.ListAssertionsInput{ServiceArn: &serviceArn, NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, assertion := range out.Assertions {
			assertionID := StringValue(assertion.AssertionId)
			id := serviceArn + "," + assertionID
			g.Resources = append(g.Resources, terraformutils.NewResource(
				id, assertionID, "aws_resiliencehubv2_assertion", "aws",
				map[string]string{"service_arn": serviceArn, "assertion_id": assertionID},
				resilienceHubV2AllowEmptyValues, map[string]interface{}{}))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

func (g *ResilienceHubV2Generator) createInputSources(svc *resiliencehubv2.Client, serviceArn string) error {
	var nextToken *string
	for {
		out, err := svc.ListInputSources(context.TODO(), &resiliencehubv2.ListInputSourcesInput{ServiceArn: &serviceArn, NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, inputSource := range out.InputSourceSummaries {
			inputSourceID := StringValue(inputSource.InputSourceId)
			id := serviceArn + "," + inputSourceID
			g.Resources = append(g.Resources, terraformutils.NewResource(
				id, inputSourceID, "aws_resiliencehubv2_input_source", "aws",
				map[string]string{"service_arn": serviceArn, "input_source_id": inputSourceID},
				resilienceHubV2AllowEmptyValues, map[string]interface{}{}))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

func (g *ResilienceHubV2Generator) createServiceFunctions(svc *resiliencehubv2.Client, serviceArn string) error {
	var nextToken *string
	for {
		out, err := svc.ListServiceFunctions(context.TODO(), &resiliencehubv2.ListServiceFunctionsInput{ServiceArn: &serviceArn, NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, function := range out.ServiceFunctions {
			functionID := StringValue(function.ServiceFunctionId)
			id := serviceArn + "," + functionID
			g.Resources = append(g.Resources, terraformutils.NewResource(
				id, StringValue(function.Name), "aws_resiliencehubv2_service_function", "aws",
				map[string]string{"service_arn": serviceArn, "service_function_id": functionID},
				resilienceHubV2AllowEmptyValues, map[string]interface{}{}))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

// aws_resiliencehubv2_system — importer ID is the system ARN.
// aws_resiliencehubv2_user_journey — ID "system_arn,user_journey_id",
// enumerated per system.
func (g *ResilienceHubV2Generator) createSystems(svc *resiliencehubv2.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListSystems(context.TODO(), &resiliencehubv2.ListSystemsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, system := range out.SystemSummaries {
			systemArn := StringValue(system.SystemArn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				systemArn, StringValue(system.Name), "aws_resiliencehubv2_system", "aws", resilienceHubV2AllowEmptyValues))
			if err := g.createUserJourneys(svc, systemArn); err != nil {
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

func (g *ResilienceHubV2Generator) createUserJourneys(svc *resiliencehubv2.Client, systemArn string) error {
	var nextToken *string
	for {
		out, err := svc.ListUserJourneys(context.TODO(), &resiliencehubv2.ListUserJourneysInput{SystemArn: &systemArn, NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, journey := range out.UserJourneySummaries {
			journeyID := StringValue(journey.UserJourneyId)
			id := systemArn + "," + journeyID
			g.Resources = append(g.Resources, terraformutils.NewResource(
				id, StringValue(journey.Name), "aws_resiliencehubv2_user_journey", "aws",
				map[string]string{"system_arn": systemArn, "user_journey_id": journeyID},
				resilienceHubV2AllowEmptyValues, map[string]interface{}{}))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
