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

// CloudWatch Application Insights generator (service key
// "applicationinsights").

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/applicationinsights"
)

var applicationinsightsAllowEmptyValues = []string{"tags."}

type ApplicationInsightsGenerator struct {
	AWSService
}

// aws_applicationinsights_application — importer ID is the resource group name.
func (g *ApplicationInsightsGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := applicationinsights.NewFromConfig(config)
	var nextToken *string
	for {
		out, err := svc.ListApplications(context.TODO(), &applicationinsights.ListApplicationsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, application := range out.ApplicationInfoList {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(application.ResourceGroupName), StringValue(application.ResourceGroupName),
				"aws_applicationinsights_application", "aws", applicationinsightsAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
