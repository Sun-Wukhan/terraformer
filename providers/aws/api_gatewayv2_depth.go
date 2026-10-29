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
	apigatewayv2v2 "github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
)

// Depth coverage for API Gateway V2 (aws_apigatewayv2_* types not covered by
// api_gatewayv2.go, which uses the v1 SDK). This file uses the v2 SDK client.
// Importer ID formats verified against the Terraform AWS provider docs.

func (g *APIGatewayV2Generator) loadGatewayV2Depth() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := apigatewayv2v2.NewFromConfig(config)

	if err := g.loadV2DomainNames(svc); err != nil {
		return err
	}
	if err := g.loadV2APIDepth(svc); err != nil {
		return err
	}
	return nil
}

func (g *APIGatewayV2Generator) loadV2DomainNames(svc *apigatewayv2v2.Client) error {
	var nextToken *string
	for {
		out, err := svc.GetDomainNames(context.TODO(), &apigatewayv2v2.GetDomainNamesInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, domain := range out.Items {
			domainName := StringValue(domain.DomainName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				domainName,
				domainName,
				"aws_apigatewayv2_domain_name",
				"aws",
				apiGatewayV2AllowEmptyValues))
			if err := g.loadV2APIMappings(svc, domainName); err != nil {
				return err
			}
			if err := g.loadV2RoutingRules(svc, domainName); err != nil {
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

func (g *APIGatewayV2Generator) loadV2APIMappings(svc *apigatewayv2v2.Client, domainName string) error {
	var nextToken *string
	for {
		out, err := svc.GetApiMappings(context.TODO(), &apigatewayv2v2.GetApiMappingsInput{
			DomainName: &domainName,
			NextToken:  nextToken,
		})
		if err != nil {
			return err
		}
		for _, mapping := range out.Items {
			// Import ID is api_mapping_id/domain_name.
			id := StringValue(mapping.ApiMappingId) + "/" + domainName
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				id,
				"aws_apigatewayv2_api_mapping",
				"aws",
				apiGatewayV2AllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

func (g *APIGatewayV2Generator) loadV2RoutingRules(svc *apigatewayv2v2.Client, domainName string) error {
	var nextToken *string
	for {
		out, err := svc.ListRoutingRules(context.TODO(), &apigatewayv2v2.ListRoutingRulesInput{
			DomainName: &domainName,
			NextToken:  nextToken,
		})
		if err != nil {
			return err
		}
		for _, rule := range out.RoutingRules {
			// Import ID is the routing rule ARN.
			arn := StringValue(rule.RoutingRuleArn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn,
				arn,
				"aws_apigatewayv2_routing_rule",
				"aws",
				apiGatewayV2AllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

func (g *APIGatewayV2Generator) loadV2APIDepth(svc *apigatewayv2v2.Client) error {
	var nextToken *string
	for {
		out, err := svc.GetApis(context.TODO(), &apigatewayv2v2.GetApisInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, api := range out.Items {
			apiID := StringValue(api.ApiId)
			if err := g.loadV2Stages(svc, apiID); err != nil {
				return err
			}
			if err := g.loadV2Deployments(svc, apiID); err != nil {
				return err
			}
			if err := g.loadV2Integrations(svc, apiID); err != nil {
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

func (g *APIGatewayV2Generator) loadV2Stages(svc *apigatewayv2v2.Client, apiID string) error {
	var nextToken *string
	for {
		out, err := svc.GetStages(context.TODO(), &apigatewayv2v2.GetStagesInput{
			ApiId:     &apiID,
			NextToken: nextToken,
		})
		if err != nil {
			return err
		}
		for _, stage := range out.Items {
			// Import ID is api_id/stage_name.
			id := apiID + "/" + StringValue(stage.StageName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				id,
				"aws_apigatewayv2_stage",
				"aws",
				apiGatewayV2AllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

func (g *APIGatewayV2Generator) loadV2Deployments(svc *apigatewayv2v2.Client, apiID string) error {
	var nextToken *string
	for {
		out, err := svc.GetDeployments(context.TODO(), &apigatewayv2v2.GetDeploymentsInput{
			ApiId:     &apiID,
			NextToken: nextToken,
		})
		if err != nil {
			return err
		}
		for _, deployment := range out.Items {
			// Import ID is api_id/deployment_id.
			id := apiID + "/" + StringValue(deployment.DeploymentId)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				id,
				"aws_apigatewayv2_deployment",
				"aws",
				apiGatewayV2AllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

func (g *APIGatewayV2Generator) loadV2Integrations(svc *apigatewayv2v2.Client, apiID string) error {
	var nextToken *string
	for {
		out, err := svc.GetIntegrations(context.TODO(), &apigatewayv2v2.GetIntegrationsInput{
			ApiId:     &apiID,
			NextToken: nextToken,
		})
		if err != nil {
			return err
		}
		for _, integration := range out.Items {
			integrationID := StringValue(integration.IntegrationId)
			id := apiID + "/" + integrationID
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				id,
				"aws_apigatewayv2_integration",
				"aws",
				apiGatewayV2AllowEmptyValues))
			if err := g.loadV2IntegrationResponses(svc, apiID, integrationID); err != nil {
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

func (g *APIGatewayV2Generator) loadV2IntegrationResponses(svc *apigatewayv2v2.Client, apiID, integrationID string) error {
	var nextToken *string
	for {
		out, err := svc.GetIntegrationResponses(context.TODO(), &apigatewayv2v2.GetIntegrationResponsesInput{
			ApiId:         &apiID,
			IntegrationId: &integrationID,
			NextToken:     nextToken,
		})
		if err != nil {
			return err
		}
		for _, response := range out.Items {
			// Import ID is api_id/integration_id/integration_response_id.
			id := apiID + "/" + integrationID + "/" + StringValue(response.IntegrationResponseId)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				id,
				"aws_apigatewayv2_integration_response",
				"aws",
				apiGatewayV2AllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
