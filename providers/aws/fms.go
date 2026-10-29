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
	"github.com/aws/aws-sdk-go-v2/service/fms"
)

var fmsAllowEmptyValues = []string{"tags."}

type FmsGenerator struct {
	AWSService
}

// InitResources enumerates Firewall Manager policies (ListPolicies,
// ID = policy ID) and resource sets (ListResourceSets, ID = resource
// set ID) in the configured region.
//
// Intentionally not generated: aws_fms_admin_account is the
// singleton Firewall Manager administrator designation for the
// organization (GetAdminAccount); there is at most one.
func (g *FmsGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := fms.NewFromConfig(config)

	policies, err := svc.ListPolicies(context.TODO(), &fms.ListPoliciesInput{})
	if err != nil {
		log.Println(err)
	} else {
		for _, policy := range policies.PolicyList {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(policy.PolicyId),
				StringValue(policy.PolicyName),
				"aws_fms_policy",
				"aws",
				fmsAllowEmptyValues))
		}
	}

	sets, err := svc.ListResourceSets(context.TODO(), &fms.ListResourceSetsInput{})
	if err != nil {
		log.Println(err)
		return nil
	}
	for _, set := range sets.ResourceSets {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			StringValue(set.Id),
			StringValue(set.Name),
			"aws_fms_resource_set",
			"aws",
			fmsAllowEmptyValues))
	}
	return nil
}
