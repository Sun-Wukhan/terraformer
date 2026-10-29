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
	"github.com/aws/aws-sdk-go-v2/service/workmail"
)

var workMailAllowEmptyValues = []string{"tags."}

// WorkMailGenerator covers Amazon WorkMail. Import IDs verified against the
// Terraform AWS provider docs: organizations import with the organization
// ID, domains with organization_id,domain_name, groups with
// organization_id,group_id; users follow the same family composite pattern
// (organization_id,user_id).
//
// Skipped in this family:
//   - aws_workmail_default_domain: per-organization singleton pointer to the
//     default domain; the domain itself is emitted as aws_workmail_domain.
type WorkMailGenerator struct {
	AWSService
}

func (g *WorkMailGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := workmail.NewFromConfig(config)

	var nextToken *string
	for {
		out, err := svc.ListOrganizations(context.TODO(), &workmail.ListOrganizationsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, org := range out.OrganizationSummaries {
			orgID := StringValue(org.OrganizationId)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				orgID,
				StringValue(org.Alias),
				"aws_workmail_organization",
				"aws",
				workMailAllowEmptyValues))
			if err := g.loadWorkMailDomains(svc, orgID); err != nil {
				return err
			}
			if err := g.loadWorkMailGroups(svc, orgID); err != nil {
				return err
			}
			if err := g.loadWorkMailUsers(svc, orgID); err != nil {
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

func (g *WorkMailGenerator) loadWorkMailDomains(svc *workmail.Client, orgID string) error {
	var nextToken *string
	for {
		out, err := svc.ListMailDomains(context.TODO(), &workmail.ListMailDomainsInput{
			OrganizationId: &orgID,
			NextToken:      nextToken,
		})
		if err != nil {
			return err
		}
		for _, domain := range out.MailDomains {
			// Import ID is organization_id,domain_name.
			id := orgID + "," + StringValue(domain.DomainName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				StringValue(domain.DomainName),
				"aws_workmail_domain",
				"aws",
				workMailAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

func (g *WorkMailGenerator) loadWorkMailGroups(svc *workmail.Client, orgID string) error {
	var nextToken *string
	for {
		out, err := svc.ListGroups(context.TODO(), &workmail.ListGroupsInput{
			OrganizationId: &orgID,
			NextToken:      nextToken,
		})
		if err != nil {
			return err
		}
		for _, group := range out.Groups {
			id := orgID + "," + StringValue(group.Id)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				StringValue(group.Name),
				"aws_workmail_group",
				"aws",
				workMailAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

func (g *WorkMailGenerator) loadWorkMailUsers(svc *workmail.Client, orgID string) error {
	var nextToken *string
	for {
		out, err := svc.ListUsers(context.TODO(), &workmail.ListUsersInput{
			OrganizationId: &orgID,
			NextToken:      nextToken,
		})
		if err != nil {
			return err
		}
		for _, user := range out.Users {
			id := orgID + "," + StringValue(user.Id)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				StringValue(user.Name),
				"aws_workmail_user",
				"aws",
				workMailAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
