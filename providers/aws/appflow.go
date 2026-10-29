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
	"github.com/aws/aws-sdk-go-v2/service/appflow"
)

var appflowAllowEmptyValues = []string{"tags."}

// AppFlowGenerator covers Amazon AppFlow. Both resources import with the
// resource name (verified against the Terraform AWS provider docs;
// connector profiles are enumerated via DescribeConnectorProfiles, the
// service's list-style API).
type AppFlowGenerator struct {
	AWSService
}

func (g *AppFlowGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := appflow.NewFromConfig(config)

	var nextToken *string
	for {
		out, err := svc.DescribeConnectorProfiles(context.TODO(), &appflow.DescribeConnectorProfilesInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, profile := range out.ConnectorProfileDetails {
			name := StringValue(profile.ConnectorProfileName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name,
				name,
				"aws_appflow_connector_profile",
				"aws",
				appflowAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}

	nextToken = nil
	for {
		out, err := svc.ListFlows(context.TODO(), &appflow.ListFlowsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, flow := range out.Flows {
			name := StringValue(flow.FlowName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name,
				name,
				"aws_appflow_flow",
				"aws",
				appflowAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
