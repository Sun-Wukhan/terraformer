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
	"github.com/aws/aws-sdk-go-v2/service/accountaccess"
)

var accountaccessAllowEmptyValues = []string{"tags."}

// AccountAccessGenerator covers AWS Account Access applications and entitlements.
// Provider import IDs (verified from provider docs): application = application ARN;
// entitlement = "<application ARN>,<entitlement ID>".
type AccountAccessGenerator struct {
	AWSService
}

func (g *AccountAccessGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := accountaccess.NewFromConfig(config)
	p := accountaccess.NewListApplicationsPaginator(svc, &accountaccess.ListApplicationsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, app := range page.Applications {
			arn := StringValue(app.ApplicationArn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_accountaccess_application", "aws", accountaccessAllowEmptyValues))
			ep := accountaccess.NewListEntitlementsPaginator(svc, &accountaccess.ListEntitlementsInput{ApplicationArn: app.ApplicationArn})
			for ep.HasMorePages() {
				epage, err := ep.NextPage(context.TODO())
				if err != nil {
					break
				}
				for _, ent := range epage.Entitlements {
					entID := StringValue(ent.EntitlementId)
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						arn+","+entID, entID, "aws_accountaccess_entitlement", "aws", accountaccessAllowEmptyValues))
				}
			}
		}
	}
	return nil
}
