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
	"github.com/aws/aws-sdk-go-v2/service/licensemanager"
)

var licenseManagerAllowEmptyValues = []string{"tags."}

type LicenseManagerGenerator struct {
	AWSService
}

// InitResources enumerates License Manager license configurations
// (ListLicenseConfigurations; provider ID = configuration ARN) and
// distributed grants (ListDistributedGrants; provider ID = grant
// ARN) in the configured region.
//
// Intentionally not generated:
//   - aws_licensemanager_association is association glue between a
//     license configuration and a resource; associations are created
//     implicitly when a configuration is attached to a resource and
//     the enumerated form carries no association ID of its own.
//   - aws_licensemanager_grant_accepter is the accepter side of a
//     grant in the receiving account.
func (g *LicenseManagerGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := licensemanager.NewFromConfig(config)

	input := &licensemanager.ListLicenseConfigurationsInput{}
	for {
		page, err := svc.ListLicenseConfigurations(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, configuration := range page.LicenseConfigurations {
			arn := StringValue(configuration.LicenseConfigurationArn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn,
				StringValue(configuration.Name),
				"aws_licensemanager_license_configuration",
				"aws",
				licenseManagerAllowEmptyValues))
		}
		if page.NextToken == nil {
			break
		}
		input.NextToken = page.NextToken
	}

	grants, err := svc.ListDistributedGrants(context.TODO(), &licensemanager.ListDistributedGrantsInput{})
	if err != nil {
		log.Println(err)
		return nil
	}
	for _, grant := range grants.Grants {
		arn := StringValue(grant.GrantArn)
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			arn,
			StringValue(grant.GrantName),
			"aws_licensemanager_grant",
			"aws",
			licenseManagerAllowEmptyValues))
	}
	return nil
}
