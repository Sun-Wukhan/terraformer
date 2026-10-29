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
	"fmt"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalog"
)

// loadTailResources adds Service Catalog products, constraints, shares,
// associations, provisioned products, provisioning artifacts, service actions
// and tag options.
// Provider import IDs (verified): product = product ID; constraint = constraint ID;
// portfolio share = "<portfolio ID>:<type>:<principal ID>"; principal association =
// "en,<principal ARN>,<portfolio ID>,<principal type>"; product/portfolio association =
// "en:<portfolio ID>:<product ID>"; provisioned product = ID; provisioning artifact =
// "<artifact ID>:<product ID>"; service action = ID; tag option = ID;
// tag option association = "<tag option ID>:<resource ID>".
// Skipped in this family:
//   - aws_servicecatalog_budget_resource_association: Service Catalog exposes only
//     the reverse lookup ListBudgetsForResource for a caller-supplied resource ID;
//     there is no operation that lists budget associations across resources.
//   - aws_servicecatalog_organizations_access: the provider documentation has no
//     Import section (no provider importer); it is a singleton account-level
//     Organizations access toggle read via GetAWSOrganizationsAccessStatus.
func (g *ServiceCatalogGenerator) loadTailResources(svc *servicecatalog.Client) {
	// Portfolios first (needed for per-portfolio children).
	var portfolioIDs []string
	pp := servicecatalog.NewListPortfoliosPaginator(svc, &servicecatalog.ListPortfoliosInput{})
	for pp.HasMorePages() {
		page, err := pp.NextPage(context.TODO())
		if err != nil {
			break
		}
		for _, portfolio := range page.PortfolioDetails {
			portfolioIDs = append(portfolioIDs, StringValue(portfolio.Id))
		}
	}

	products, err := svc.SearchProductsAsAdmin(context.TODO(), &servicecatalog.SearchProductsAsAdminInput{})
	if err == nil {
		for _, detail := range products.ProductViewDetails {
			if detail.ProductViewSummary == nil {
				continue
			}
			productID := StringValue(detail.ProductViewSummary.ProductId)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				productID, StringValue(detail.ProductViewSummary.Name), "aws_servicecatalog_product", "aws", servicecatalogAllowEmptyValues))

			artifacts, err := svc.ListProvisioningArtifacts(context.TODO(),
				&servicecatalog.ListProvisioningArtifactsInput{ProductId: detail.ProductViewSummary.ProductId})
			if err == nil {
				for _, artifact := range artifacts.ProvisioningArtifactDetails {
					id := StringValue(artifact.Id) + ":" + productID
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						id, id, "aws_servicecatalog_provisioning_artifact", "aws", servicecatalogAllowEmptyValues))
				}
			}
			portfolios, err := svc.ListPortfoliosForProduct(context.TODO(),
				&servicecatalog.ListPortfoliosForProductInput{ProductId: detail.ProductViewSummary.ProductId})
			if err == nil {
				for _, portfolio := range portfolios.PortfolioDetails {
					id := "en:" + StringValue(portfolio.Id) + ":" + productID
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						id, id, "aws_servicecatalog_product_portfolio_association", "aws", servicecatalogAllowEmptyValues))
				}
			}
		}
	}

	for _, portfolioID := range portfolioIDs {
		constraints, err := svc.ListConstraintsForPortfolio(context.TODO(),
			&servicecatalog.ListConstraintsForPortfolioInput{PortfolioId: aws.String(portfolioID)})
		if err == nil {
			for _, c := range constraints.ConstraintDetails {
				id := StringValue(c.ConstraintId)
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, id, "aws_servicecatalog_constraint", "aws", servicecatalogAllowEmptyValues))
			}
		}
		principals, err := svc.ListPrincipalsForPortfolio(context.TODO(),
			&servicecatalog.ListPrincipalsForPortfolioInput{PortfolioId: aws.String(portfolioID)})
		if err == nil {
			for _, principal := range principals.Principals {
				id := fmt.Sprintf("en,%s,%s,%s", StringValue(principal.PrincipalARN), portfolioID, string(principal.PrincipalType))
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, id, "aws_servicecatalog_principal_portfolio_association", "aws", servicecatalogAllowEmptyValues))
			}
		}
	}

	for _, portfolioID := range portfolioIDs {
		shares, err := svc.DescribePortfolioShares(context.TODO(),
			&servicecatalog.DescribePortfolioSharesInput{PortfolioId: aws.String(portfolioID)})
		if err != nil {
			continue
		}
		for _, share := range shares.PortfolioShareDetails {
			id := fmt.Sprintf("%s:%s:%s", portfolioID, string(share.Type), StringValue(share.PrincipalId))
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_servicecatalog_portfolio_share", "aws", servicecatalogAllowEmptyValues))
		}
	}

	provisioned, err := svc.SearchProvisionedProducts(context.TODO(), &servicecatalog.SearchProvisionedProductsInput{})
	if err == nil {
		for _, prod := range provisioned.ProvisionedProducts {
			id := StringValue(prod.Id)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(prod.Name), "aws_servicecatalog_provisioned_product", "aws", servicecatalogAllowEmptyValues))
		}
	}

	actions, err := svc.ListServiceActions(context.TODO(), &servicecatalog.ListServiceActionsInput{})
	if err == nil {
		for _, action := range actions.ServiceActionSummaries {
			id := StringValue(action.Id)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(action.Name), "aws_servicecatalog_service_action", "aws", servicecatalogAllowEmptyValues))
		}
	}

	tagOptions, err := svc.ListTagOptions(context.TODO(), &servicecatalog.ListTagOptionsInput{})
	if err == nil {
		for _, tagOption := range tagOptions.TagOptionDetails {
			id := StringValue(tagOption.Id)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_servicecatalog_tag_option", "aws", servicecatalogAllowEmptyValues))
			resources, err := svc.ListResourcesForTagOption(context.TODO(),
				&servicecatalog.ListResourcesForTagOptionInput{TagOptionId: tagOption.Id})
			if err == nil {
				for _, res := range resources.ResourceDetails {
					assocID := id + ":" + StringValue(res.Id)
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						assocID, assocID, "aws_servicecatalog_tag_option_resource_association", "aws", servicecatalogAllowEmptyValues))
				}
			}
		}
	}
}
