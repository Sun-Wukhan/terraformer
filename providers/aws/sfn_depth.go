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
	"github.com/aws/aws-sdk-go-v2/service/sfn"
)

// Depth coverage for Step Functions: state machine aliases, enumerated per
// state machine via ListStateMachineAliases. Import ID is the alias ARN
// (verified against the Terraform AWS provider docs).

func (g *SfnGenerator) loadSfnAliases(svc *sfn.Client) error {
	p := sfn.NewListStateMachinesPaginator(svc, &sfn.ListStateMachinesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, machine := range page.StateMachines {
			var aliasToken *string
			for {
				aliasPage, err := svc.ListStateMachineAliases(context.TODO(), &sfn.ListStateMachineAliasesInput{
					StateMachineArn: machine.StateMachineArn,
					NextToken:       aliasToken,
				})
				if err != nil {
					return err
				}
				for _, alias := range aliasPage.StateMachineAliases {
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						StringValue(alias.StateMachineAliasArn),
						StringValue(alias.StateMachineAliasArn),
						"aws_sfn_alias",
						"aws",
						sfnAllowEmptyValues))
				}
				aliasToken = aliasPage.NextToken
				if aliasToken == nil {
					break
				}
			}
		}
	}
	return nil
}
