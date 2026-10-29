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
	"github.com/aws/aws-sdk-go-v2/service/organizations"
)

// loadTailResources adds Organizations trusted service access, delegated
// administrators and the organization resource policy.
// Provider import IDs (verified): aws_service_access = service principal;
// delegated administrator = "<account ID>/<service principal>";
// resource policy = resource policy ID.
// Skipped in this family:
//   - aws_organizations_tag: a single tag key/value pair on an arbitrary
//     Organizations resource; there is no standalone tag inventory (tags are
//     only readable per-resource via ListTagsForResource), so tag resources are
//     not enumerable as independent objects.
func (g *OrganizationGenerator) loadTailResources(svc *organizations.Client) {
	access, err := svc.ListAWSServiceAccessForOrganization(context.TODO(), &organizations.ListAWSServiceAccessForOrganizationInput{})
	if err == nil {
		for _, principal := range access.EnabledServicePrincipals {
			sp := StringValue(principal.ServicePrincipal)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				sp, sp, "aws_organizations_aws_service_access", "aws", organizationAllowEmptyValues))
		}
	}

	admins, err := svc.ListDelegatedAdministrators(context.TODO(), &organizations.ListDelegatedAdministratorsInput{})
	if err == nil {
		for _, admin := range admins.DelegatedAdministrators {
			accountID := StringValue(admin.Id)
			services, err := svc.ListDelegatedServicesForAccount(context.TODO(),
				&organizations.ListDelegatedServicesForAccountInput{AccountId: admin.Id})
			if err != nil {
				continue
			}
			for _, svcPrincipal := range services.DelegatedServices {
				id := accountID + "/" + StringValue(svcPrincipal.ServicePrincipal)
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, id, "aws_organizations_delegated_administrator", "aws", organizationAllowEmptyValues))
			}
		}
	}

	// An organization has at most one resource policy; DescribeResourcePolicy
	// returns it (the pinned SDK's PolicyType enum has no RESOURCE_POLICY filter
	// value for ListPolicies). Emit it when one exists.
	if out, err := svc.DescribeResourcePolicy(context.TODO(), &organizations.DescribeResourcePolicyInput{}); err == nil && out.ResourcePolicy != nil && out.ResourcePolicy.ResourcePolicySummary != nil {
		id := StringValue(out.ResourcePolicy.ResourcePolicySummary.Id)
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			id, id, "aws_organizations_resource_policy", "aws", organizationAllowEmptyValues))
	}
}
