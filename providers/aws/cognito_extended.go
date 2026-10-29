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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentity"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
)

// Cognito depth for the AWS sweep. Identity pool role attachments are
// keyed by identity pool ID; user-pool sub-resources that are
// singletons per pool (log delivery configuration, pool-level risk
// configuration) are keyed by user pool ID; identity providers,
// resource servers, users and user groups are keyed by
// "user_pool_id/<name>"; user pool domains by the domain string (the
// domain is read back from DescribeUserPool, since the SDK has no
// ListUserPoolDomains API). Subresources that only exist conditionally
// (roles attached, log delivery configured, a risk configuration set)
// are emitted only when the corresponding Get/Describe call finds
// them.
//
// Deliberately not generated:
//   - aws_cognito_identity_pool_provider_principal_tag: a per-(pool,
//     identity provider) principal-tag attribute map; the SDK exposes
//     it only via GetPrincipalTagAttributeMap for a provider you
//     already know — there is no list API for these mappings.
//   - aws_cognito_user_in_group: user/group membership association
//     glue.
//   - aws_cognito_user_pool_ui_customization: per-(pool, client) UI
//     customization settings; only GetUICustomization for a known
//     client exists, no list API.
//   - aws_cognito_managed_user_pool_client: template-managed clients
//     are indistinguishable from ordinary clients in
//     ListUserPoolClients output, so the managed-client resources
//     cannot be reconstructed from enumeration.
func (g *CognitoGenerator) loadDepth(svcIdp *cognitoidentityprovider.Client, svcID *cognitoidentity.Client) error {
	if err := g.loadIdentityPoolRoles(svcID); err != nil {
		return err
	}
	p := cognitoidentityprovider.NewListUserPoolsPaginator(svcIdp, &cognitoidentityprovider.ListUserPoolsInput{
		MaxResults: aws.Int32(CognitoMaxResults),
	})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, pool := range page.UserPools {
			if err := g.loadUserPoolDepth(svcIdp, StringValue(pool.Id)); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *CognitoGenerator) loadIdentityPoolRoles(svc *cognitoidentity.Client) error {
	p := cognitoidentity.NewListIdentityPoolsPaginator(svc, &cognitoidentity.ListIdentityPoolsInput{
		MaxResults: aws.Int32(CognitoMaxResults),
	})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, pool := range page.IdentityPools {
			poolID := StringValue(pool.IdentityPoolId)
			roles, err := svc.GetIdentityPoolRoles(context.TODO(), &cognitoidentity.GetIdentityPoolRolesInput{IdentityPoolId: aws.String(poolID)})
			if err == nil && len(roles.Roles) > 0 {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					poolID, StringValue(pool.IdentityPoolName), "aws_cognito_identity_pool_roles_attachment", "aws", CognitoAllowEmptyValues))
			}
		}
	}
	return nil
}

func (g *CognitoGenerator) loadUserPoolDepth(svc *cognitoidentityprovider.Client, poolID string) error {
	// Identity providers.
	idpInput := &cognitoidentityprovider.ListIdentityProvidersInput{UserPoolId: aws.String(poolID), MaxResults: aws.Int32(CognitoMaxResults)}
	for {
		out, err := svc.ListIdentityProviders(context.TODO(), idpInput)
		if err != nil {
			return err
		}
		for _, provider := range out.Providers {
			name := StringValue(provider.ProviderName)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				poolID+":"+name, name, "aws_cognito_identity_provider", "aws",
				map[string]string{"user_pool_id": poolID}, CognitoAllowEmptyValues, CognitoAdditionalFields))
		}
		if out.NextToken == nil {
			break
		}
		idpInput.NextToken = out.NextToken
	}
	// Resource servers.
	rsInput := &cognitoidentityprovider.ListResourceServersInput{UserPoolId: aws.String(poolID), MaxResults: aws.Int32(CognitoMaxResults)}
	for {
		out, err := svc.ListResourceServers(context.TODO(), rsInput)
		if err != nil {
			return err
		}
		for _, rs := range out.ResourceServers {
			identifier := StringValue(rs.Identifier)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				poolID+"/"+identifier, StringValue(rs.Name), "aws_cognito_resource_server", "aws",
				map[string]string{"user_pool_id": poolID}, CognitoAllowEmptyValues, CognitoAdditionalFields))
		}
		if out.NextToken == nil {
			break
		}
		rsInput.NextToken = out.NextToken
	}
	// User groups.
	groupInput := &cognitoidentityprovider.ListGroupsInput{UserPoolId: aws.String(poolID), Limit: aws.Int32(CognitoMaxResults)}
	for {
		out, err := svc.ListGroups(context.TODO(), groupInput)
		if err != nil {
			return err
		}
		for _, group := range out.Groups {
			name := StringValue(group.GroupName)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				poolID+":"+name, name, "aws_cognito_user_group", "aws",
				map[string]string{"user_pool_id": poolID}, CognitoAllowEmptyValues, CognitoAdditionalFields))
		}
		if out.NextToken == nil {
			break
		}
		groupInput.NextToken = out.NextToken
	}
	// Users.
	userInput := &cognitoidentityprovider.ListUsersInput{UserPoolId: aws.String(poolID), Limit: aws.Int32(CognitoMaxResults)}
	for {
		out, err := svc.ListUsers(context.TODO(), userInput)
		if err != nil {
			return err
		}
		for _, user := range out.Users {
			username := StringValue(user.Username)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				poolID+"/"+username, username, "aws_cognito_user", "aws",
				map[string]string{"user_pool_id": poolID}, CognitoAllowEmptyValues, CognitoAdditionalFields))
		}
		if out.PaginationToken == nil {
			break
		}
		userInput.PaginationToken = out.PaginationToken
	}
	// Log delivery configuration (per-pool singleton, conditional).
	if out, err := svc.GetLogDeliveryConfiguration(context.TODO(), &cognitoidentityprovider.GetLogDeliveryConfigurationInput{UserPoolId: aws.String(poolID)}); err == nil &&
		out.LogDeliveryConfiguration != nil && len(out.LogDeliveryConfiguration.LogConfigurations) > 0 {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			poolID, poolID, "aws_cognito_log_delivery_configuration", "aws", CognitoAllowEmptyValues))
	}
	// Risk configuration (pool-level only; client-specific
	// configurations share the same resource type but are keyed by
	// client and are not emitted here).
	if out, err := svc.DescribeRiskConfiguration(context.TODO(), &cognitoidentityprovider.DescribeRiskConfigurationInput{UserPoolId: aws.String(poolID)}); err == nil &&
		out.RiskConfiguration != nil && out.RiskConfiguration.ClientId == nil {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			poolID, poolID, "aws_cognito_risk_configuration", "aws", CognitoAllowEmptyValues))
	}
	// Managed login branding: no list API exists, but branding is
	// describable per app client; the provider importer ID is the
	// verified comma-separated "user_pool_id,managed_login_branding_id".
	clientInput := &cognitoidentityprovider.ListUserPoolClientsInput{UserPoolId: aws.String(poolID), MaxResults: aws.Int32(CognitoMaxResults)}
	for {
		out, err := svc.ListUserPoolClients(context.TODO(), clientInput)
		if err != nil {
			return err
		}
		for _, client := range out.UserPoolClients {
			branding, err := svc.DescribeManagedLoginBrandingByClient(context.TODO(), &cognitoidentityprovider.DescribeManagedLoginBrandingByClientInput{
				UserPoolId: aws.String(poolID), ClientId: client.ClientId,
			})
			if err == nil && branding.ManagedLoginBranding != nil && branding.ManagedLoginBranding.ManagedLoginBrandingId != nil {
				brandingID := StringValue(branding.ManagedLoginBranding.ManagedLoginBrandingId)
				g.Resources = append(g.Resources, terraformutils.NewResource(
					poolID+","+brandingID, brandingID, "aws_cognito_managed_login_branding", "aws",
					map[string]string{"user_pool_id": poolID}, CognitoAllowEmptyValues, CognitoAdditionalFields))
			}
		}
		if out.NextToken == nil {
			break
		}
		clientInput.NextToken = out.NextToken
	}
	// User pool domain (read back from the pool itself).
	if out, err := svc.DescribeUserPool(context.TODO(), &cognitoidentityprovider.DescribeUserPoolInput{UserPoolId: aws.String(poolID)}); err == nil && out.UserPool != nil {
		domain := StringValue(out.UserPool.CustomDomain)
		if domain == "" {
			domain = StringValue(out.UserPool.Domain)
		}
		if domain != "" {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				domain, domain, "aws_cognito_user_pool_domain", "aws",
				map[string]string{"user_pool_id": poolID}, CognitoAllowEmptyValues, CognitoAdditionalFields))
		}
	}
	return nil
}
