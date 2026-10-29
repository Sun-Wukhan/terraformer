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

// CloudWatch Omni generator (service key "cloudwatchomni").
//
// Skipped types in this family (see results ledger):
//   - aws_cloudwatchomni_organization_domain: organization singleton with
//     no list API; the SDK exposes only GetDomainForOrganization, keyed by a
//     caller-supplied organization ID.

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchomni"
)

var cloudwatchomniAllowEmptyValues = []string{"tags."}

type CloudWatchOmniGenerator struct {
	AWSService
}

func (g *CloudWatchOmniGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := cloudwatchomni.NewFromConfig(config)
	if err := g.createDomains(svc); err != nil {
		return err
	}
	if err := g.createSpaces(svc); err != nil {
		return err
	}
	return g.createViews(svc)
}

// aws_cloudwatchomni_domain — importer ID is the domain ID.
func (g *CloudWatchOmniGenerator) createDomains(svc *cloudwatchomni.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListDomains(context.TODO(), &cloudwatchomni.ListDomainsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, domain := range out.Items {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(domain.DomainId), StringValue(domain.Name),
				"aws_cloudwatchomni_domain", "aws", cloudwatchomniAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

func (g *CloudWatchOmniGenerator) createSpaces(svc *cloudwatchomni.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListSpaces(context.TODO(), &cloudwatchomni.ListSpacesInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, space := range out.Items {
			spaceID := StringValue(space.SpaceId)
			// aws_cloudwatchomni_space — importer ID is the space ID.
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				spaceID, StringValue(space.Name),
				"aws_cloudwatchomni_space", "aws", cloudwatchomniAllowEmptyValues))
			if err := g.createSpaceChildren(svc, spaceID); err != nil {
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

func (g *CloudWatchOmniGenerator) createSpaceChildren(svc *cloudwatchomni.Client, spaceID string) error {
	// aws_cloudwatchomni_dashboard — importer ID "space_id,dashboard_id".
	var nextToken *string
	for {
		out, err := svc.ListOmniDashboards(context.TODO(), &cloudwatchomni.ListOmniDashboardsInput{SpaceId: &spaceID, NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, dashboard := range out.Items {
			id := spaceID + "," + StringValue(dashboard.DashboardId)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				id, StringValue(dashboard.Name), "aws_cloudwatchomni_dashboard", "aws",
				map[string]string{"space_id": spaceID}, cloudwatchomniAllowEmptyValues, map[string]interface{}{}))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	// aws_cloudwatchomni_access_profile — importer ID "space_id,profile_id".
	nextToken = nil
	for {
		out, err := svc.ListAccessProfiles(context.TODO(), &cloudwatchomni.ListAccessProfilesInput{SpaceId: &spaceID, NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, profile := range out.Items {
			id := spaceID + "," + StringValue(profile.ProfileId)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				id, StringValue(profile.Name), "aws_cloudwatchomni_access_profile", "aws",
				map[string]string{"space_id": spaceID}, cloudwatchomniAllowEmptyValues, map[string]interface{}{}))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	// aws_cloudwatchomni_access_grant — importer ID is the grant ID.
	// Service-managed grants (created automatically with each space) are not
	// managed by the provider resource and are excluded.
	nextToken = nil
	for {
		out, err := svc.ListAccessGrants(context.TODO(), &cloudwatchomni.ListAccessGrantsInput{SpaceId: &spaceID, NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, grant := range out.Items {
			if grant.GrantType == "SERVICE_MANAGED" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewResource(
				StringValue(grant.GrantId), StringValue(grant.Name), "aws_cloudwatchomni_access_grant", "aws",
				map[string]string{"space_id": spaceID}, cloudwatchomniAllowEmptyValues, map[string]interface{}{}))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

// aws_cloudwatchomni_view — importer ID is the view name. Only USER views
// are managed by the provider resource.
func (g *CloudWatchOmniGenerator) createViews(svc *cloudwatchomni.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListViews(context.TODO(), &cloudwatchomni.ListViewsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, view := range out.Items {
			if view.Type != "USER" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(view.Name), StringValue(view.Name),
				"aws_cloudwatchomni_view", "aws", cloudwatchomniAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
