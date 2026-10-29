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
	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions"
)

var verifiedPermissionsAllowEmptyValues = []string{"tags."}

type VerifiedPermissionsGenerator struct {
	AWSService
}

// InitResources enumerates Verified Permissions policy stores in the
// configured region (ListPolicyStores; provider ID = policy store
// ID) and, per store, its schema (GetSchema; the schema resource's
// provider ID is the policy store ID), emitted only when a schema is
// actually set, plus the store's identity sources, policies and
// policy templates. Their provider importer IDs are composites
// verified against the provider docs:
//   - aws_verifiedpermissions_identity_source:
//     policy_store_id:identity_source_id
//   - aws_verifiedpermissions_policy: policy_id,policy_store_id
//   - aws_verifiedpermissions_policy_template:
//     policy_store_id:policy_template_id
func (g *VerifiedPermissionsGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := verifiedpermissions.NewFromConfig(config)

	p := verifiedpermissions.NewListPolicyStoresPaginator(svc, &verifiedpermissions.ListPolicyStoresInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, store := range page.PolicyStores {
			storeID := StringValue(store.PolicyStoreId)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				storeID,
				storeID,
				"aws_verifiedpermissions_policy_store",
				"aws",
				verifiedPermissionsAllowEmptyValues))
			g.addSchema(svc, storeID)
			g.addIdentitySources(svc, storeID)
			g.addPolicies(svc, storeID)
			g.addPolicyTemplates(svc, storeID)
		}
	}
	return nil
}

func (g *VerifiedPermissionsGenerator) addIdentitySources(svc *verifiedpermissions.Client, storeID string) {
	p := verifiedpermissions.NewListIdentitySourcesPaginator(svc, &verifiedpermissions.ListIdentitySourcesInput{PolicyStoreId: &storeID})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, source := range page.IdentitySources {
			sourceID := StringValue(source.IdentitySourceId)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				storeID+":"+sourceID,
				sourceID,
				"aws_verifiedpermissions_identity_source",
				"aws",
				map[string]string{"policy_store_id": storeID},
				verifiedPermissionsAllowEmptyValues,
				map[string]interface{}{}))
		}
	}
}

func (g *VerifiedPermissionsGenerator) addPolicies(svc *verifiedpermissions.Client, storeID string) {
	p := verifiedpermissions.NewListPoliciesPaginator(svc, &verifiedpermissions.ListPoliciesInput{PolicyStoreId: &storeID})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, policy := range page.Policies {
			policyID := StringValue(policy.PolicyId)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				policyID+","+storeID,
				policyID,
				"aws_verifiedpermissions_policy",
				"aws",
				map[string]string{"policy_store_id": storeID},
				verifiedPermissionsAllowEmptyValues,
				map[string]interface{}{}))
		}
	}
}

func (g *VerifiedPermissionsGenerator) addPolicyTemplates(svc *verifiedpermissions.Client, storeID string) {
	p := verifiedpermissions.NewListPolicyTemplatesPaginator(svc, &verifiedpermissions.ListPolicyTemplatesInput{PolicyStoreId: &storeID})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, template := range page.PolicyTemplates {
			templateID := StringValue(template.PolicyTemplateId)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				storeID+":"+templateID,
				templateID,
				"aws_verifiedpermissions_policy_template",
				"aws",
				map[string]string{"policy_store_id": storeID},
				verifiedPermissionsAllowEmptyValues,
				map[string]interface{}{}))
		}
	}
}

func (g *VerifiedPermissionsGenerator) addSchema(svc *verifiedpermissions.Client, storeID string) {
	out, err := svc.GetSchema(context.TODO(), &verifiedpermissions.GetSchemaInput{PolicyStoreId: &storeID})
	if err != nil || out.Schema == nil || *out.Schema == "" {
		return
	}
	g.Resources = append(g.Resources, terraformutils.NewResource(
		storeID,
		storeID,
		"aws_verifiedpermissions_schema",
		"aws",
		map[string]string{"policy_store_id": storeID},
		verifiedPermissionsAllowEmptyValues,
		map[string]interface{}{}))
}
