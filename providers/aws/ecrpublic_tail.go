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
	"github.com/aws/aws-sdk-go-v2/service/ecrpublic"
)

// loadRepositoryPolicies adds aws_ecrpublic_repository_policy resources for
// repositories that have a policy. Provider import ID (verified): repository name.
func (g *EcrPublicGenerator) loadRepositoryPolicies(svc *ecrpublic.Client) {
	p := ecrpublic.NewDescribeRepositoriesPaginator(svc, &ecrpublic.DescribeRepositoriesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return
		}
		for _, repository := range page.Repositories {
			if _, err := svc.GetRepositoryPolicy(context.TODO(), &ecrpublic.GetRepositoryPolicyInput{
				RepositoryName: repository.RepositoryName,
				RegistryId:     repository.RegistryId,
			}); err == nil {
				name := StringValue(repository.RepositoryName)
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					name, name, "aws_ecrpublic_repository_policy", "aws", ecrPublicAllowEmptyValues))
			}
		}
	}
}
