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
	"github.com/aws/aws-sdk-go-v2/service/macie2"
)

var macie2AllowEmptyValues = []string{"tags."}

type Macie2Generator struct {
	AWSService
}

// InitResources enumerates the Macie objects in the configured
// region: classification jobs, custom data identifiers, findings
// filters and member accounts.
//
// Provider importer ID formats (terraform-provider-aws docs):
//   - aws_macie2_classification_job: the job ID.
//   - aws_macie2_custom_data_identifier: the identifier ID.
//   - aws_macie2_findings_filter: the filter ID.
//   - aws_macie2_member: the member account ID.
//
// Intentionally not generated:
//   - aws_macie2_account is the singleton Macie enablement for the
//     account/region.
//   - aws_macie2_classification_export_configuration is the singleton
//     account-level export configuration
//     (GetClassificationExportConfiguration).
//   - aws_macie2_invitation_accepter is the accepter side of a member
//     invitation in the invited account.
//   - aws_macie2_organization_admin_account is the singleton
//     delegated-administrator designation for the organization.
//   - aws_macie2_organization_configuration is the singleton
//     organization configuration of the administrator account
//     (DescribeOrganizationConfiguration).
func (g *Macie2Generator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := macie2.NewFromConfig(config)
	g.addClassificationJobs(svc)
	g.addCustomDataIdentifiers(svc)
	g.addFindingsFilters(svc)
	g.addMembers(svc)
	return nil
}

func (g *Macie2Generator) addClassificationJobs(svc *macie2.Client) {
	p := macie2.NewListClassificationJobsPaginator(svc, &macie2.ListClassificationJobsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, job := range page.Items {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(job.JobId),
				StringValue(job.Name),
				"aws_macie2_classification_job",
				"aws",
				macie2AllowEmptyValues))
		}
	}
}

func (g *Macie2Generator) addCustomDataIdentifiers(svc *macie2.Client) {
	p := macie2.NewListCustomDataIdentifiersPaginator(svc, &macie2.ListCustomDataIdentifiersInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, item := range page.Items {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(item.Id),
				StringValue(item.Name),
				"aws_macie2_custom_data_identifier",
				"aws",
				macie2AllowEmptyValues))
		}
	}
}

func (g *Macie2Generator) addFindingsFilters(svc *macie2.Client) {
	p := macie2.NewListFindingsFiltersPaginator(svc, &macie2.ListFindingsFiltersInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, filter := range page.FindingsFilterListItems {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(filter.Id),
				StringValue(filter.Name),
				"aws_macie2_findings_filter",
				"aws",
				macie2AllowEmptyValues))
		}
	}
}

func (g *Macie2Generator) addMembers(svc *macie2.Client) {
	p := macie2.NewListMembersPaginator(svc, &macie2.ListMembersInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, member := range page.Members {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(member.AccountId),
				StringValue(member.AccountId),
				"aws_macie2_member",
				"aws",
				macie2AllowEmptyValues))
		}
	}
}
