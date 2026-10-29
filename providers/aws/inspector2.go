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
	"log"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/inspector2"
)

var inspector2AllowEmptyValues = []string{"tags."}

type Inspector2Generator struct {
	AWSService
}

// InitResources enumerates Inspector (v2) filters and member
// associations in the configured region.
//
// Provider importer ID formats (terraform-provider-aws docs):
//   - aws_inspector2_filter: the filter ARN.
//   - aws_inspector2_member_association: the member account ID.
//
// Intentionally not generated:
//   - aws_inspector2_delegated_admin_account is the singleton
//     delegated-administrator designation for the organization.
//   - aws_inspector2_enabler is the singleton Inspector enablement
//     for the calling account/region.
//   - aws_inspector2_organization_configuration is the singleton
//     organization configuration of the delegated administrator
//     (DescribeOrganizationConfiguration).
func (g *Inspector2Generator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := inspector2.NewFromConfig(config)
	g.addFilters(svc)
	g.addMemberAssociations(svc)
	return nil
}

func (g *Inspector2Generator) addFilters(svc *inspector2.Client) {
	p := inspector2.NewListFiltersPaginator(svc, &inspector2.ListFiltersInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, filter := range page.Filters {
			arn := StringValue(filter.Arn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn,
				StringValue(filter.Name),
				"aws_inspector2_filter",
				"aws",
				inspector2AllowEmptyValues))
		}
	}
}

func (g *Inspector2Generator) addMemberAssociations(svc *inspector2.Client) {
	p := inspector2.NewListMembersPaginator(svc, &inspector2.ListMembersInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, member := range page.Members {
			accountID := StringValue(member.AccountId)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				accountID,
				accountID,
				"aws_inspector2_member_association",
				"aws",
				inspector2AllowEmptyValues))
		}
	}
}
