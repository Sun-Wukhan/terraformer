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
	"github.com/aws/aws-sdk-go-v2/service/detective"
)

var detectiveAllowEmptyValues = []string{"tags."}

type DetectiveGenerator struct {
	AWSService
}

// InitResources enumerates Detective behavior graphs in the
// configured region (ListGraphs) and, per graph, its member accounts
// (ListMembers).
//
// Provider importer ID formats (terraform-provider-aws docs):
//   - aws_detective_graph: the graph ARN.
//   - aws_detective_member: graph_arn:account_id.
//
// Intentionally not generated:
//   - aws_detective_invitation_accepter is the accepter side of a
//     member invitation in the invited account.
//   - aws_detective_organization_admin_account is the singleton
//     delegated-administrator designation for the organization.
//   - aws_detective_organization_configuration is the singleton
//     per-graph organization configuration of the administrator
//     account (DescribeOrganizationConfiguration).
func (g *DetectiveGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := detective.NewFromConfig(config)

	graphs, err := svc.ListGraphs(context.TODO(), &detective.ListGraphsInput{})
	if err != nil {
		return err
	}
	for _, graph := range graphs.GraphList {
		graphArn := StringValue(graph.Arn)
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			graphArn,
			arnResourceName(graphArn),
			"aws_detective_graph",
			"aws",
			detectiveAllowEmptyValues))
		g.addMembers(svc, graphArn)
	}
	return nil
}

func (g *DetectiveGenerator) addMembers(svc *detective.Client, graphArn string) {
	p := detective.NewListMembersPaginator(svc, &detective.ListMembersInput{GraphArn: &graphArn})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, member := range page.MemberDetails {
			accountID := StringValue(member.AccountId)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				graphArn+":"+accountID,
				accountID,
				"aws_detective_member",
				"aws",
				map[string]string{"graph_arn": graphArn},
				detectiveAllowEmptyValues,
				map[string]interface{}{}))
		}
	}
}
