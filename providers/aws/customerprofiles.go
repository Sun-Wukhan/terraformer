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
	"github.com/aws/aws-sdk-go-v2/service/customerprofiles"
)

var customerprofilesAllowEmptyValues = []string{"tags."}

// CustomerProfilesGenerator covers Customer Profiles domains.
// Provider import ID (verified): domain name (the resource ID set by the provider).
// Skipped in this family:
//   - aws_customerprofiles_profile: the Customer Profiles SDK has no ListProfiles
//     operation; profiles are only reachable via SearchProfiles, which requires a
//     caller-supplied key name and value, so the profile inventory is not enumerable.
type CustomerProfilesGenerator struct {
	AWSService
}

func (g *CustomerProfilesGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := customerprofiles.NewFromConfig(config)
	var nextToken *string
	for {
		page, err := svc.ListDomains(context.TODO(), &customerprofiles.ListDomainsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, d := range page.Items {
			name := StringValue(d.DomainName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_customerprofiles_domain", "aws", customerprofilesAllowEmptyValues))
		}
		if page.NextToken == nil {
			break
		}
		nextToken = page.NextToken
	}
	return nil
}
