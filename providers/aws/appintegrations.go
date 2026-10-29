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
	"github.com/aws/aws-sdk-go-v2/service/appintegrations"
)

var appIntegrationsAllowEmptyValues = []string{"tags."}

// AppIntegrationsGenerator covers Amazon AppIntegrations. Import IDs
// verified against the Terraform AWS provider docs: data integrations import
// with their UUID id, event integrations with their name.
type AppIntegrationsGenerator struct {
	AWSService
}

func (g *AppIntegrationsGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := appintegrations.NewFromConfig(config)

	var nextToken *string
	for {
		out, err := svc.ListDataIntegrations(context.TODO(), &appintegrations.ListDataIntegrationsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, integration := range out.DataIntegrations {
			// The list summary carries no Id field; the import ID (the data
			// integration UUID) is the final segment of its ARN.
			id := StringValue(integration.Arn)
			if idx := strings.LastIndex(id, "/"); idx >= 0 {
				id = id[idx+1:]
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				StringValue(integration.Name),
				"aws_appintegrations_data_integration",
				"aws",
				appIntegrationsAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}

	nextToken = nil
	for {
		out, err := svc.ListEventIntegrations(context.TODO(), &appintegrations.ListEventIntegrationsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, integration := range out.EventIntegrations {
			name := StringValue(integration.Name)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name,
				name,
				"aws_appintegrations_event_integration",
				"aws",
				appIntegrationsAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
