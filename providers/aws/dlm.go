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
	"github.com/aws/aws-sdk-go-v2/service/dlm"
	dlmtypes "github.com/aws/aws-sdk-go-v2/service/dlm/types"
)

var dlmAllowEmptyValues = []string{"tags."}

// DlmGenerator covers Data Lifecycle Manager lifecycle policies.
// Provider import ID (verified): policy ID.
type DlmGenerator struct {
	AWSService
}

func (g *DlmGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := dlm.NewFromConfig(config)
	// GetLifecyclePolicies requires a policy state filter; enumerate every state.
	for _, state := range []dlmtypes.GettablePolicyStateValues{
		dlmtypes.GettablePolicyStateValuesEnabled,
		dlmtypes.GettablePolicyStateValuesDisabled,
		dlmtypes.GettablePolicyStateValuesError,
	} {
		out, err := svc.GetLifecyclePolicies(context.TODO(), &dlm.GetLifecyclePoliciesInput{State: state})
		if err != nil {
			return err
		}
		for _, pol := range out.Policies {
			id := StringValue(pol.PolicyId)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_dlm_lifecycle_policy", "aws", dlmAllowEmptyValues))
		}
	}
	return nil
}
