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
	"github.com/aws/aws-sdk-go-v2/service/appfabric"
)

var appFabricAllowEmptyValues = []string{"tags."}

// AppFabricGenerator covers AWS AppFabric. Import IDs verified against the
// Terraform AWS provider docs: app bundles import with their ARN,
// ingestions with app_bundle_arn,ingestion_arn.
//
// Skipped in this family (provider docs have no Import section for these
// resources — no provider importer):
//   - aws_appfabric_app_authorization
//   - aws_appfabric_app_authorization_connection
//   - aws_appfabric_ingestion_destination
type AppFabricGenerator struct {
	AWSService
}

func (g *AppFabricGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := appfabric.NewFromConfig(config)

	var nextToken *string
	for {
		out, err := svc.ListAppBundles(context.TODO(), &appfabric.ListAppBundlesInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, bundle := range out.AppBundleSummaryList {
			bundleArn := StringValue(bundle.Arn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				bundleArn,
				bundleArn,
				"aws_appfabric_app_bundle",
				"aws",
				appFabricAllowEmptyValues))
			if err := g.loadAppFabricIngestions(svc, bundleArn); err != nil {
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

func (g *AppFabricGenerator) loadAppFabricIngestions(svc *appfabric.Client, bundleArn string) error {
	var nextToken *string
	for {
		out, err := svc.ListIngestions(context.TODO(), &appfabric.ListIngestionsInput{
			AppBundleIdentifier: &bundleArn,
			NextToken:           nextToken,
		})
		if err != nil {
			return err
		}
		for _, ingestion := range out.Ingestions {
			id := bundleArn + "," + StringValue(ingestion.Arn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				StringValue(ingestion.Arn),
				"aws_appfabric_ingestion",
				"aws",
				appFabricAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
