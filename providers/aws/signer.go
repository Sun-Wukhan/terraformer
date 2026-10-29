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
	"github.com/aws/aws-sdk-go-v2/service/signer"
)

var signerAllowEmptyValues = []string{"tags."}

type SignerGenerator struct {
	AWSService
}

// InitResources enumerates Signer signing profiles in the configured
// region (ListSigningProfiles); the provider importer ID is the
// profile name. Per profile it also enumerates the profile's
// cross-account permission statements (ListProfilePermissions);
// the aws_signer_signing_profile_permission provider importer ID is
// profile_name/statement_id (verified against the provider docs).
//
// Intentionally not generated:
//   - aws_signer_signing_job is a one-off signing operation, not a
//     durable object: ListSigningJobs enumerates historical job
//     executions, which are not manageable infrastructure.
func (g *SignerGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := signer.NewFromConfig(config)

	p := signer.NewListSigningProfilesPaginator(svc, &signer.ListSigningProfilesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, profile := range page.Profiles {
			name := StringValue(profile.ProfileName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name,
				name,
				"aws_signer_signing_profile",
				"aws",
				signerAllowEmptyValues))
			g.addProfilePermissions(svc, name)
		}
	}
	return nil
}

func (g *SignerGenerator) addProfilePermissions(svc *signer.Client, profileName string) {
	input := &signer.ListProfilePermissionsInput{ProfileName: &profileName}
	for {
		out, err := svc.ListProfilePermissions(context.TODO(), input)
		if err != nil {
			log.Println(err)
			return
		}
		for _, permission := range out.Permissions {
			statementID := StringValue(permission.StatementId)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				profileName+"/"+statementID,
				statementID,
				"aws_signer_signing_profile_permission",
				"aws",
				map[string]string{"profile_name": profileName},
				signerAllowEmptyValues,
				map[string]interface{}{}))
		}
		if out.NextToken == nil {
			return
		}
		input.NextToken = out.NextToken
	}
}
