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

// Sweep extension for the CloudTrail generator (service key "cloudtrail").

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
)

func (g *CloudTrailGenerator) createSweepResources(svc *cloudtrail.Client, config aws.Config) error {
	if err := g.createEventDataStores(svc); err != nil {
		return err
	}
	return g.createDelegatedAdminAccounts(config)
}

// aws_cloudtrail_event_data_store — importer ID is the event data store ARN.
func (g *CloudTrailGenerator) createEventDataStores(svc *cloudtrail.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListEventDataStores(context.TODO(), &cloudtrail.ListEventDataStoresInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, store := range out.EventDataStores {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(store.EventDataStoreArn), StringValue(store.Name),
				"aws_cloudtrail_event_data_store", "aws", cloudtrailAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

// aws_cloudtrail_organization_delegated_admin_account — importer ID is the
// delegated administrator account ID. CloudTrail exposes no list API for
// delegated administrators; they are enumerated through AWS Organizations
// (ListDelegatedAdministratorsForService, service principal
// cloudtrail.amazonaws.com), the same API the provider uses on read.
func (g *CloudTrailGenerator) createDelegatedAdminAccounts(config aws.Config) error {
	orgSvc := organizations.NewFromConfig(config)
	var nextToken *string
	for {
		out, err := orgSvc.ListDelegatedAdministrators(context.TODO(), &organizations.ListDelegatedAdministratorsInput{
			ServicePrincipal: aws.String("cloudtrail.amazonaws.com"),
			NextToken:        nextToken,
		})
		if err != nil {
			return err
		}
		for _, account := range out.DelegatedAdministrators {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				StringValue(account.Id), StringValue(account.Name),
				"aws_cloudtrail_organization_delegated_admin_account", "aws",
				map[string]string{"account_id": StringValue(account.Id)},
				cloudtrailAllowEmptyValues, map[string]interface{}{}))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
