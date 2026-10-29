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
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
)

var sesv2AllowEmptyValues = []string{"tags."}

// SesV2Generator covers SES API v2 (aws_sesv2_*). Import IDs verified
// against the Terraform AWS provider docs.
//
// Skipped in this family:
//   - aws_sesv2_account_suppression_attributes: singleton account-level
//     settings.
//   - aws_sesv2_account_vdm_attributes: singleton account-level settings.
//   - aws_sesv2_multi_region_endpoint: provider docs have no Import section
//     (no provider importer).
type SesV2Generator struct {
	AWSService
}

func (g *SesV2Generator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := sesv2.NewFromConfig(config)

	if err := g.loadSesV2ConfigurationSets(svc); err != nil {
		return err
	}
	if err := g.loadSesV2ContactLists(svc); err != nil {
		return err
	}
	if err := g.loadSesV2DedicatedIPPools(svc); err != nil {
		return err
	}
	if err := g.loadSesV2EmailIdentities(svc); err != nil {
		return err
	}
	if err := g.loadSesV2Tenants(svc); err != nil {
		return err
	}
	return nil
}

func (g *SesV2Generator) loadSesV2ConfigurationSets(svc *sesv2.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListConfigurationSets(context.TODO(), &sesv2.ListConfigurationSetsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, set := range out.ConfigurationSets {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				set,
				set,
				"aws_sesv2_configuration_set",
				"aws",
				sesv2AllowEmptyValues))
			destinations, err := svc.GetConfigurationSetEventDestinations(context.TODO(), &sesv2.GetConfigurationSetEventDestinationsInput{
				ConfigurationSetName: &set,
			})
			if err != nil {
				return err
			}
			for _, destination := range destinations.EventDestinations {
				// Import ID is configuration_set_name|event_destination_name.
				id := set + "|" + StringValue(destination.Name)
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id,
					StringValue(destination.Name),
					"aws_sesv2_configuration_set_event_destination",
					"aws",
					sesv2AllowEmptyValues))
			}
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

func (g *SesV2Generator) loadSesV2ContactLists(svc *sesv2.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListContactLists(context.TODO(), &sesv2.ListContactListsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, list := range out.ContactLists {
			name := StringValue(list.ContactListName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name,
				name,
				"aws_sesv2_contact_list",
				"aws",
				sesv2AllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

func (g *SesV2Generator) loadSesV2DedicatedIPPools(svc *sesv2.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListDedicatedIpPools(context.TODO(), &sesv2.ListDedicatedIpPoolsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, pool := range out.DedicatedIpPools {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				pool,
				pool,
				"aws_sesv2_dedicated_ip_pool",
				"aws",
				sesv2AllowEmptyValues))
			poolIPs, err := svc.GetDedicatedIps(context.TODO(), &sesv2.GetDedicatedIpsInput{PoolName: &pool})
			if err != nil {
				return err
			}
			for _, ip := range poolIPs.DedicatedIps {
				// Import ID is ip,destination_pool_name.
				id := StringValue(ip.Ip) + "," + pool
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id,
					StringValue(ip.Ip),
					"aws_sesv2_dedicated_ip_assignment",
					"aws",
					sesv2AllowEmptyValues))
			}
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

func (g *SesV2Generator) loadSesV2EmailIdentities(svc *sesv2.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListEmailIdentities(context.TODO(), &sesv2.ListEmailIdentitiesInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, identity := range out.EmailIdentities {
			name := StringValue(identity.IdentityName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name,
				name,
				"aws_sesv2_email_identity",
				"aws",
				sesv2AllowEmptyValues))

			details, err := svc.GetEmailIdentity(context.TODO(), &sesv2.GetEmailIdentityInput{EmailIdentity: &name})
			if err != nil {
				return err
			}
			// Feedback attributes exist for every identity (forwarding flag is
			// always returned by GetEmailIdentity).
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name,
				name,
				"aws_sesv2_email_identity_feedback_attributes",
				"aws",
				sesv2AllowEmptyValues))
			if details.MailFromAttributes != nil && StringValue(details.MailFromAttributes.MailFromDomain) != "" {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					name,
					name,
					"aws_sesv2_email_identity_mail_from_attributes",
					"aws",
					sesv2AllowEmptyValues))
			}

			policies, err := svc.GetEmailIdentityPolicies(context.TODO(), &sesv2.GetEmailIdentityPoliciesInput{EmailIdentity: &name})
			if err != nil {
				return err
			}
			for policyName := range policies.Policies {
				// Import ID is identity/policy_name.
				id := name + "/" + policyName
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id,
					policyName,
					"aws_sesv2_email_identity_policy",
					"aws",
					sesv2AllowEmptyValues))
			}
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

func (g *SesV2Generator) loadSesV2Tenants(svc *sesv2.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListTenants(context.TODO(), &sesv2.ListTenantsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, tenant := range out.Tenants {
			tenantName := StringValue(tenant.TenantName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				tenantName,
				tenantName,
				"aws_sesv2_tenant",
				"aws",
				sesv2AllowEmptyValues))
			resources, err := svc.ListTenantResources(context.TODO(), &sesv2.ListTenantResourcesInput{TenantName: &tenantName})
			if err != nil {
				return err
			}
			for _, resource := range resources.TenantResources {
				// Import ID is tenant_name|resource_arn.
				id := tenantName + "|" + StringValue(resource.ResourceArn)
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id,
					id,
					"aws_sesv2_tenant_resource_association",
					"aws",
					sesv2AllowEmptyValues))
			}
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
