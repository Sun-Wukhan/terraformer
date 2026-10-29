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
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
)

// Depth coverage for API Gateway v1 (aws_api_gateway_* types not covered by
// api_gateway.go). Importer ID formats verified against the Terraform AWS
// provider docs for each resource.
//
// Skipped in this family:
//   - aws_api_gateway_account: singleton per-region account settings
//     (cloudwatch role / throttle defaults); not an enumerable resource.
//   - aws_api_gateway_rest_api_put: write-path-only action; its import ID is
//     the rest_api_id itself and the PUT body/triggers cannot be enumerated
//     from the API, so it would duplicate aws_api_gateway_rest_api.

func (g *APIGatewayGenerator) loadGatewayDepth(svc *apigateway.Client) error {
	if err := g.loadClientCertificates(svc); err != nil {
		return err
	}
	if err := g.loadGatewayDomainNames(svc); err != nil {
		return err
	}
	if err := g.loadGatewayAccessAssociations(svc); err != nil {
		return err
	}
	if err := g.loadGatewayRestAPIDepth(svc); err != nil {
		return err
	}
	if err := g.loadGatewayUsagePlanKeys(svc); err != nil {
		return err
	}
	return nil
}

func (g *APIGatewayGenerator) loadClientCertificates(svc *apigateway.Client) error {
	p := apigateway.NewGetClientCertificatesPaginator(svc, &apigateway.GetClientCertificatesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, cert := range page.Items {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(cert.ClientCertificateId),
				StringValue(cert.ClientCertificateId),
				"aws_api_gateway_client_certificate",
				"aws",
				apiGatewayAllowEmptyValues))
		}
	}
	return nil
}

func (g *APIGatewayGenerator) loadGatewayDomainNames(svc *apigateway.Client) error {
	var position *string
	for {
		out, err := svc.GetDomainNames(context.TODO(), &apigateway.GetDomainNamesInput{Position: position})
		if err != nil {
			return err
		}
		for _, domain := range out.Items {
			domainName := StringValue(domain.DomainName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				domainName,
				domainName,
				"aws_api_gateway_domain_name",
				"aws",
				apiGatewayAllowEmptyValues))
			if err := g.loadBasePathMappings(svc, domainName); err != nil {
				return err
			}
		}
		position = out.Position
		if position == nil {
			break
		}
	}
	return nil
}

func (g *APIGatewayGenerator) loadBasePathMappings(svc *apigateway.Client, domainName string) error {
	var position *string
	for {
		out, err := svc.GetBasePathMappings(context.TODO(), &apigateway.GetBasePathMappingsInput{
			DomainName: &domainName,
			Position:   position,
		})
		if err != nil {
			return err
		}
		for _, mapping := range out.Items {
			basePath := StringValue(mapping.BasePath)
			if basePath == "(none)" {
				basePath = ""
			}
			// Import ID is domain_name/base_path (trailing slash for root).
			id := domainName + "/" + basePath
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				id,
				"aws_api_gateway_base_path_mapping",
				"aws",
				apiGatewayAllowEmptyValues))
		}
		position = out.Position
		if position == nil {
			break
		}
	}
	return nil
}

func (g *APIGatewayGenerator) loadGatewayAccessAssociations(svc *apigateway.Client) error {
	var position *string
	for {
		out, err := svc.GetDomainNameAccessAssociations(context.TODO(), &apigateway.GetDomainNameAccessAssociationsInput{Position: position})
		if err != nil {
			return err
		}
		for _, assoc := range out.Items {
			// Import ID is the association ARN.
			arn := StringValue(assoc.DomainNameAccessAssociationArn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn,
				arn,
				"aws_api_gateway_domain_name_access_association",
				"aws",
				apiGatewayAllowEmptyValues))
		}
		position = out.Position
		if position == nil {
			break
		}
	}
	return nil
}

func (g *APIGatewayGenerator) loadGatewayRestAPIDepth(svc *apigateway.Client) error {
	p := apigateway.NewGetRestApisPaginator(svc, &apigateway.GetRestApisInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, restAPI := range page.Items {
			restAPIID := StringValue(restAPI.Id)
			if StringValue(restAPI.Policy) != "" {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					restAPIID,
					restAPIID,
					"aws_api_gateway_rest_api_policy",
					"aws",
					apiGatewayAllowEmptyValues))
			}
			if err := g.loadDeployments(svc, restAPIID); err != nil {
				return err
			}
			if err := g.loadDocumentationVersions(svc, restAPIID); err != nil {
				return err
			}
			if err := g.loadRequestValidators(svc, restAPIID); err != nil {
				return err
			}
			if err := g.loadMethodSettings(svc, restAPIID); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *APIGatewayGenerator) loadDeployments(svc *apigateway.Client, restAPIID string) error {
	p := apigateway.NewGetDeploymentsPaginator(svc, &apigateway.GetDeploymentsInput{RestApiId: &restAPIID})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, deployment := range page.Items {
			id := restAPIID + "/" + StringValue(deployment.Id)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				id,
				"aws_api_gateway_deployment",
				"aws",
				apiGatewayAllowEmptyValues))
		}
	}
	return nil
}

func (g *APIGatewayGenerator) loadDocumentationVersions(svc *apigateway.Client, restAPIID string) error {
	var position *string
	for {
		page, err := svc.GetDocumentationVersions(context.TODO(), &apigateway.GetDocumentationVersionsInput{RestApiId: &restAPIID, Position: position})
		if err != nil {
			return err
		}
		for _, version := range page.Items {
			id := restAPIID + "/" + StringValue(version.Version)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				id,
				"aws_api_gateway_documentation_version",
				"aws",
				apiGatewayAllowEmptyValues))
		}
		position = page.Position
		if position == nil {
			break
		}
	}
	return nil
}

func (g *APIGatewayGenerator) loadRequestValidators(svc *apigateway.Client, restAPIID string) error {
	var position *string
	for {
		out, err := svc.GetRequestValidators(context.TODO(), &apigateway.GetRequestValidatorsInput{
			RestApiId: &restAPIID,
			Position:  position,
		})
		if err != nil {
			return err
		}
		for _, validator := range out.Items {
			id := restAPIID + "/" + StringValue(validator.Id)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				StringValue(validator.Name),
				"aws_api_gateway_request_validator",
				"aws",
				apiGatewayAllowEmptyValues))
		}
		position = out.Position
		if position == nil {
			break
		}
	}
	return nil
}

func (g *APIGatewayGenerator) loadMethodSettings(svc *apigateway.Client, restAPIID string) error {
	out, err := svc.GetStages(context.TODO(), &apigateway.GetStagesInput{RestApiId: &restAPIID})
	if err != nil {
		return err
	}
	for _, stage := range out.Item {
		for methodPath := range stage.MethodSettings {
			// Import ID is rest_api_id/stage_name/method_path.
			id := restAPIID + "/" + StringValue(stage.StageName) + "/" + methodPath
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				strings.ReplaceAll(id, "/", "_"),
				"aws_api_gateway_method_settings",
				"aws",
				apiGatewayAllowEmptyValues))
		}
	}
	return nil
}

func (g *APIGatewayGenerator) loadGatewayUsagePlanKeys(svc *apigateway.Client) error {
	p := apigateway.NewGetUsagePlansPaginator(svc, &apigateway.GetUsagePlansInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, plan := range page.Items {
			planID := StringValue(plan.Id)
			kp := apigateway.NewGetUsagePlanKeysPaginator(svc, &apigateway.GetUsagePlanKeysInput{UsagePlanId: &planID})
			for kp.HasMorePages() {
				kpage, err := kp.NextPage(context.TODO())
				if err != nil {
					return err
				}
				for _, key := range kpage.Items {
					id := planID + "/" + StringValue(key.Id)
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						id,
						StringValue(key.Name),
						"aws_api_gateway_usage_plan_key",
						"aws",
						apiGatewayAllowEmptyValues))
				}
			}
		}
	}
	return nil
}
