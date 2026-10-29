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
	"github.com/aws/aws-sdk-go-v2/service/amplify"
)

var amplifyAllowEmptyValues = []string{"tags."}

// AmplifyGenerator covers AWS Amplify. Import IDs follow the provider's
// documented formats: app_id for apps, app_id/name composites for backend
// environments, branches and domain associations, app_id/webhook_id for
// webhooks.
type AmplifyGenerator struct {
	AWSService
}

func (g *AmplifyGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := amplify.NewFromConfig(config)

	var nextToken *string
	for {
		out, err := svc.ListApps(context.TODO(), &amplify.ListAppsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, app := range out.Apps {
			appID := StringValue(app.AppId)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				appID,
				StringValue(app.Name),
				"aws_amplify_app",
				"aws",
				amplifyAllowEmptyValues))
			if err := g.loadAmplifyChildren(svc, appID); err != nil {
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

func (g *AmplifyGenerator) loadAmplifyChildren(svc *amplify.Client, appID string) error {
	var nextToken *string
	for {
		out, err := svc.ListBackendEnvironments(context.TODO(), &amplify.ListBackendEnvironmentsInput{
			AppId:     &appID,
			NextToken: nextToken,
		})
		if err != nil {
			return err
		}
		for _, env := range out.BackendEnvironments {
			id := appID + "/" + StringValue(env.EnvironmentName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				StringValue(env.EnvironmentName),
				"aws_amplify_backend_environment",
				"aws",
				amplifyAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}

	nextToken = nil
	for {
		out, err := svc.ListBranches(context.TODO(), &amplify.ListBranchesInput{
			AppId:     &appID,
			NextToken: nextToken,
		})
		if err != nil {
			return err
		}
		for _, branch := range out.Branches {
			id := appID + "/" + StringValue(branch.BranchName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				StringValue(branch.BranchName),
				"aws_amplify_branch",
				"aws",
				amplifyAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}

	nextToken = nil
	for {
		out, err := svc.ListDomainAssociations(context.TODO(), &amplify.ListDomainAssociationsInput{
			AppId:     &appID,
			NextToken: nextToken,
		})
		if err != nil {
			return err
		}
		for _, domain := range out.DomainAssociations {
			id := appID + "/" + StringValue(domain.DomainName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				StringValue(domain.DomainName),
				"aws_amplify_domain_association",
				"aws",
				amplifyAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}

	webhooks, err := svc.ListWebhooks(context.TODO(), &amplify.ListWebhooksInput{AppId: &appID})
	if err != nil {
		return err
	}
	for _, webhook := range webhooks.Webhooks {
		id := appID + "/" + StringValue(webhook.WebhookId)
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			id,
			StringValue(webhook.BranchName),
			"aws_amplify_webhook",
			"aws",
			amplifyAllowEmptyValues))
	}
	return nil
}
