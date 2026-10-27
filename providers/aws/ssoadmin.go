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
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
)

var ssoAdminAllowEmptyValues = []string{"tags."}

type SSOAdminGenerator struct {
	AWSService
}

// InitResources enumerates the IAM Identity Center objects in the
// configured region, per SSO instance: permission sets with their
// managed and customer managed policy attachments, inline policies,
// permissions boundary attachments and account assignments (discovered
// per provisioned account); applications with their assignments, access
// scopes and assignment configuration; the instance access control
// attributes; the instance's additional regions; and trusted token
// issuers.
//
// aws_ssoadmin_managed_policy_attachments_exclusive and
// aws_ssoadmin_customer_managed_policy_attachments_exclusive are
// intentionally not generated: the provider documents that the
// exclusive resources conflict with the individual attachment resources
// for the same permission set, and the attachments themselves are
// generated above, so emitting both would double-manage them. The
// instance's primary region is likewise skipped for aws_ssoadmin_region
// (that resource manages regions added to the instance; the primary
// region is intrinsic to it).
func (g *SSOAdminGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	client := ssoadmin.NewFromConfig(config)

	paginator := ssoadmin.NewListInstancesPaginator(client, &ssoadmin.ListInstancesInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, instance := range page.Instances {
			instanceArn := *instance.InstanceArn
			instanceName := arnResourceName(instanceArn)
			if instance.Name != nil && *instance.Name != "" {
				instanceName = *instance.Name
			}
			if err := g.addPermissionSets(client, instanceArn); err != nil {
				return err
			}
			if err := g.addApplications(client, instanceArn); err != nil {
				return err
			}
			if err := g.addInstanceAccessControlAttributes(client, instanceArn, instanceName); err != nil {
				return err
			}
			if err := g.addRegions(client, instanceArn); err != nil {
				return err
			}
			if err := g.addTrustedTokenIssuers(client, instanceArn); err != nil {
				return err
			}
		}
	}
	return nil
}

// Provider ID format: permission set ARN, instance ARN (comma
// separated). ListPermissionSets returns ARNs only, so each permission
// set is described to obtain its name.
func (g *SSOAdminGenerator) addPermissionSets(client *ssoadmin.Client, instanceArn string) error {
	paginator := ssoadmin.NewListPermissionSetsPaginator(client, &ssoadmin.ListPermissionSetsInput{
		InstanceArn: aws.String(instanceArn),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, permissionSetArn := range page.PermissionSets {
			described, err := client.DescribePermissionSet(context.TODO(), &ssoadmin.DescribePermissionSetInput{
				InstanceArn:      aws.String(instanceArn),
				PermissionSetArn: aws.String(permissionSetArn),
			})
			if err != nil {
				return err
			}
			name := arnResourceName(permissionSetArn)
			if described.PermissionSet != nil && described.PermissionSet.Name != nil {
				name = *described.PermissionSet.Name
			}
			g.Resources = append(g.Resources, terraformutils.NewResource(
				permissionSetArn+","+instanceArn,
				name,
				"aws_ssoadmin_permission_set",
				"aws",
				map[string]string{"instance_arn": instanceArn},
				ssoAdminAllowEmptyValues,
				map[string]interface{}{},
			))
			if err := g.addManagedPolicyAttachments(client, instanceArn, permissionSetArn); err != nil {
				return err
			}
			if err := g.addCustomerManagedPolicyAttachments(client, instanceArn, permissionSetArn); err != nil {
				return err
			}
			if err := g.addPermissionSetInlinePolicy(client, instanceArn, permissionSetArn, name); err != nil {
				return err
			}
			if err := g.addPermissionsBoundaryAttachment(client, instanceArn, permissionSetArn, name); err != nil {
				return err
			}
			if err := g.addAccountAssignments(client, instanceArn, permissionSetArn); err != nil {
				return err
			}
		}
	}
	return nil
}

// Provider ID format: managed policy ARN, permission set ARN, instance
// ARN (comma separated).
func (g *SSOAdminGenerator) addManagedPolicyAttachments(client *ssoadmin.Client, instanceArn, permissionSetArn string) error {
	paginator := ssoadmin.NewListManagedPoliciesInPermissionSetPaginator(client, &ssoadmin.ListManagedPoliciesInPermissionSetInput{
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(permissionSetArn),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, policy := range page.AttachedManagedPolicies {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				*policy.Arn+","+permissionSetArn+","+instanceArn,
				*policy.Name,
				"aws_ssoadmin_managed_policy_attachment",
				"aws",
				map[string]string{
					"instance_arn":       instanceArn,
					"permission_set_arn": permissionSetArn,
				},
				ssoAdminAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

// Provider ID format: policy name, policy path, permission set ARN,
// instance ARN (comma separated).
func (g *SSOAdminGenerator) addCustomerManagedPolicyAttachments(client *ssoadmin.Client, instanceArn, permissionSetArn string) error {
	paginator := ssoadmin.NewListCustomerManagedPolicyReferencesInPermissionSetPaginator(client, &ssoadmin.ListCustomerManagedPolicyReferencesInPermissionSetInput{
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(permissionSetArn),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, reference := range page.CustomerManagedPolicyReferences {
			path := "/"
			if reference.Path != nil && *reference.Path != "" {
				path = *reference.Path
			}
			g.Resources = append(g.Resources, terraformutils.NewResource(
				*reference.Name+","+path+","+permissionSetArn+","+instanceArn,
				*reference.Name,
				"aws_ssoadmin_customer_managed_policy_attachment",
				"aws",
				map[string]string{
					"instance_arn":       instanceArn,
					"permission_set_arn": permissionSetArn,
				},
				ssoAdminAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

// Provider ID format: permission set ARN, instance ARN (comma
// separated). A permission set has an inline policy resource only when
// a policy document is actually set.
func (g *SSOAdminGenerator) addPermissionSetInlinePolicy(client *ssoadmin.Client, instanceArn, permissionSetArn, permissionSetName string) error {
	out, err := client.GetInlinePolicyForPermissionSet(context.TODO(), &ssoadmin.GetInlinePolicyForPermissionSetInput{
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(permissionSetArn),
	})
	if err != nil {
		return err
	}
	if out.InlinePolicy == nil || *out.InlinePolicy == "" {
		return nil
	}
	g.Resources = append(g.Resources, terraformutils.NewResource(
		permissionSetArn+","+instanceArn,
		permissionSetName,
		"aws_ssoadmin_permission_set_inline_policy",
		"aws",
		map[string]string{
			"instance_arn":       instanceArn,
			"permission_set_arn": permissionSetArn,
		},
		ssoAdminAllowEmptyValues,
		map[string]interface{}{},
	))
	return nil
}

// Provider ID format: permission set ARN, instance ARN (comma
// separated). The boundary attachment exists only when a boundary is
// actually set on the permission set.
func (g *SSOAdminGenerator) addPermissionsBoundaryAttachment(client *ssoadmin.Client, instanceArn, permissionSetArn, permissionSetName string) error {
	out, err := client.GetPermissionsBoundaryForPermissionSet(context.TODO(), &ssoadmin.GetPermissionsBoundaryForPermissionSetInput{
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(permissionSetArn),
	})
	if err != nil {
		return err
	}
	if out.PermissionsBoundary == nil ||
		(out.PermissionsBoundary.ManagedPolicyArn == nil && out.PermissionsBoundary.CustomerManagedPolicyReference == nil) {
		return nil
	}
	g.Resources = append(g.Resources, terraformutils.NewResource(
		permissionSetArn+","+instanceArn,
		permissionSetName,
		"aws_ssoadmin_permissions_boundary_attachment",
		"aws",
		map[string]string{
			"instance_arn":       instanceArn,
			"permission_set_arn": permissionSetArn,
		},
		ssoAdminAllowEmptyValues,
		map[string]interface{}{},
	))
	return nil
}

// Provider ID format: principal ID, principal type, target ID, target
// type, permission set ARN, instance ARN (comma separated). Assignments
// are discovered through the accounts the permission set is provisioned
// to; the target type is always AWS_ACCOUNT.
func (g *SSOAdminGenerator) addAccountAssignments(client *ssoadmin.Client, instanceArn, permissionSetArn string) error {
	paginator := ssoadmin.NewListAccountsForProvisionedPermissionSetPaginator(client, &ssoadmin.ListAccountsForProvisionedPermissionSetInput{
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(permissionSetArn),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, accountID := range page.AccountIds {
			if err := g.addAccountAssignmentsForAccount(client, instanceArn, permissionSetArn, accountID); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *SSOAdminGenerator) addAccountAssignmentsForAccount(client *ssoadmin.Client, instanceArn, permissionSetArn, accountID string) error {
	paginator := ssoadmin.NewListAccountAssignmentsPaginator(client, &ssoadmin.ListAccountAssignmentsInput{
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(permissionSetArn),
		AccountId:        aws.String(accountID),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, assignment := range page.AccountAssignments {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				*assignment.PrincipalId+","+string(assignment.PrincipalType)+","+accountID+",AWS_ACCOUNT,"+permissionSetArn+","+instanceArn,
				*assignment.PrincipalId,
				"aws_ssoadmin_account_assignment",
				"aws",
				map[string]string{
					"instance_arn":       instanceArn,
					"permission_set_arn": permissionSetArn,
				},
				ssoAdminAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

func (g *SSOAdminGenerator) addApplications(client *ssoadmin.Client, instanceArn string) error {
	paginator := ssoadmin.NewListApplicationsPaginator(client, &ssoadmin.ListApplicationsInput{
		InstanceArn: aws.String(instanceArn),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, application := range page.Applications {
			applicationArn := *application.ApplicationArn
			name := arnResourceName(applicationArn)
			if application.Name != nil && *application.Name != "" {
				name = *application.Name
			}
			g.Resources = append(g.Resources, terraformutils.NewResource(
				applicationArn,
				name,
				"aws_ssoadmin_application",
				"aws",
				map[string]string{"instance_arn": instanceArn},
				ssoAdminAllowEmptyValues,
				map[string]interface{}{},
			))
			// The assignment configuration is a per-application
			// singleton keyed by the application ARN.
			g.Resources = append(g.Resources, terraformutils.NewResource(
				applicationArn,
				name,
				"aws_ssoadmin_application_assignment_configuration",
				"aws",
				map[string]string{"application_arn": applicationArn},
				ssoAdminAllowEmptyValues,
				map[string]interface{}{},
			))
			if err := g.addApplicationAssignments(client, applicationArn); err != nil {
				return err
			}
			if err := g.addApplicationAccessScopes(client, applicationArn); err != nil {
				return err
			}
		}
	}
	return nil
}

// Provider ID format: application ARN, principal ID, principal type
// (comma separated).
func (g *SSOAdminGenerator) addApplicationAssignments(client *ssoadmin.Client, applicationArn string) error {
	paginator := ssoadmin.NewListApplicationAssignmentsPaginator(client, &ssoadmin.ListApplicationAssignmentsInput{
		ApplicationArn: aws.String(applicationArn),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, assignment := range page.ApplicationAssignments {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				applicationArn+","+*assignment.PrincipalId+","+string(assignment.PrincipalType),
				*assignment.PrincipalId,
				"aws_ssoadmin_application_assignment",
				"aws",
				map[string]string{"application_arn": applicationArn},
				ssoAdminAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

// Provider ID format: application ARN, scope (comma separated).
func (g *SSOAdminGenerator) addApplicationAccessScopes(client *ssoadmin.Client, applicationArn string) error {
	paginator := ssoadmin.NewListApplicationAccessScopesPaginator(client, &ssoadmin.ListApplicationAccessScopesInput{
		ApplicationArn: aws.String(applicationArn),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, scope := range page.Scopes {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				applicationArn+","+*scope.Scope,
				*scope.Scope,
				"aws_ssoadmin_application_access_scope",
				"aws",
				map[string]string{"application_arn": applicationArn},
				ssoAdminAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

// The access control attributes are a per-instance singleton keyed by
// the instance ARN; the resource exists only once attributes have been
// configured for ABAC.
func (g *SSOAdminGenerator) addInstanceAccessControlAttributes(client *ssoadmin.Client, instanceArn, instanceName string) error {
	out, err := client.DescribeInstanceAccessControlAttributeConfiguration(context.TODO(), &ssoadmin.DescribeInstanceAccessControlAttributeConfigurationInput{
		InstanceArn: aws.String(instanceArn),
	})
	if err != nil {
		return err
	}
	if out.InstanceAccessControlAttributeConfiguration == nil ||
		len(out.InstanceAccessControlAttributeConfiguration.AccessControlAttributes) == 0 {
		return nil
	}
	g.Resources = append(g.Resources, terraformutils.NewResource(
		instanceArn,
		instanceName,
		"aws_ssoadmin_instance_access_control_attributes",
		"aws",
		map[string]string{"instance_arn": instanceArn},
		ssoAdminAllowEmptyValues,
		map[string]interface{}{},
	))
	return nil
}

// Provider ID format: instance ARN, region name (comma separated). The
// primary region is skipped: this resource manages regions added to
// the instance, and the primary region cannot be added or removed.
func (g *SSOAdminGenerator) addRegions(client *ssoadmin.Client, instanceArn string) error {
	paginator := ssoadmin.NewListRegionsPaginator(client, &ssoadmin.ListRegionsInput{
		InstanceArn: aws.String(instanceArn),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, region := range page.Regions {
			if region.IsPrimaryRegion {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewResource(
				instanceArn+","+*region.RegionName,
				*region.RegionName,
				"aws_ssoadmin_region",
				"aws",
				map[string]string{"instance_arn": instanceArn},
				ssoAdminAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

func (g *SSOAdminGenerator) addTrustedTokenIssuers(client *ssoadmin.Client, instanceArn string) error {
	paginator := ssoadmin.NewListTrustedTokenIssuersPaginator(client, &ssoadmin.ListTrustedTokenIssuersInput{
		InstanceArn: aws.String(instanceArn),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, issuer := range page.TrustedTokenIssuers {
			name := arnResourceName(*issuer.TrustedTokenIssuerArn)
			if issuer.Name != nil && *issuer.Name != "" {
				name = *issuer.Name
			}
			g.Resources = append(g.Resources, terraformutils.NewResource(
				*issuer.TrustedTokenIssuerArn,
				name,
				"aws_ssoadmin_trusted_token_issuer",
				"aws",
				map[string]string{"instance_arn": instanceArn},
				ssoAdminAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}
