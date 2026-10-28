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

package gcp

import (
	"context"
	"log"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	crmv1 "google.golang.org/api/cloudresourcemanager/v1"
	iamv1 "google.golang.org/api/iam/v1"
)

// initWorkforceResources enumerates Workforce Identity Federation
// pools of the project's organization (resolved through Cloud
// Resource Manager ancestry, superseding the project-scope-only note
// that used to live in iam_workload_identity.go): pools under
// organizations/{org}/locations/global with their IAM policy,
// bindings and members, pool providers and provider keys. The
// provider ID for a pool and its IAM trio is the pool's full name
// (locations/global/workforcePools/{id} under the organization).
//
// Deliberately not generated:
//   - google_iam_workforce_pool_provider_scim_tenant and
//     google_iam_workforce_pool_provider_scim_token: the pinned
//     iam/v1 client has no SCIM tenant or token service at all (no
//     list/get surface), and SCIM tokens are transient credentials
//     whose value is only returned at creation time.
func (g *IamGenerator) initWorkforceResources(ctx context.Context, iamSvc *iamv1.Service, crmSvc *crmv1.Service) error {
	project := g.GetArgs()["project"].(string)
	if crmSvc == nil || iamSvc == nil {
		return nil
	}
	ancestry, err := crmSvc.Projects.GetAncestry(project, &crmv1.GetAncestryRequest{}).Context(ctx).Do()
	if err != nil {
		log.Println(err)
		return nil
	}
	orgID := ""
	for _, ancestor := range ancestry.Ancestor {
		if ancestor.ResourceId != nil && ancestor.ResourceId.Type == "organization" {
			orgID = ancestor.ResourceId.Id
		}
	}
	if orgID == "" {
		log.Println("iam: no organization in project ancestry; skipping workforce pools")
		return nil
	}
	location := "organizations/" + orgID + "/locations/global"
	return iamSvc.Locations.WorkforcePools.List(location).Pages(ctx, func(page *iamv1.ListWorkforcePoolsResponse) error {
		for _, pool := range page.WorkforcePools {
			poolID := workloadIdentityLastSegment(pool.Name)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				pool.Name, poolID, "google_iam_workforce_pool", g.ProviderName,
				map[string]string{"organization": orgID, "location": "global", "workforce_pool_id": poolID},
				IamAllowEmptyValues, IamAdditionalFields))
			g.loadWorkforcePoolIam(ctx, iamSvc, pool.Name, orgID, poolID)
			g.loadWorkforcePoolProviders(ctx, iamSvc, pool.Name, orgID, poolID)
		}
		return nil
	})
}

func (g *IamGenerator) loadWorkforcePoolIam(ctx context.Context, iamSvc *iamv1.Service, poolName, orgID, poolID string) {
	policy, err := iamSvc.Locations.WorkforcePools.GetIamPolicy(poolName, &iamv1.GetIamPolicyRequest{}).Context(ctx).Do()
	if err != nil {
		log.Println(err)
		return
	}
	bindings := []iamBinding{}
	for _, b := range policy.Bindings {
		bindings = append(bindings, iamBinding{Role: b.Role, Members: b.Members})
	}
	newRes := func(id, resourceType string, attrs map[string]string) terraformutils.Resource {
		if attrs == nil {
			attrs = map[string]string{}
		}
		attrs["organization"] = orgID
		attrs["location"] = "global"
		attrs["workforce_pool_id"] = poolID
		return terraformutils.NewResource(id, poolID, resourceType, g.ProviderName, attrs, IamAllowEmptyValues, IamAdditionalFields)
	}
	g.Resources = append(g.Resources, expandIamPolicy(bindings, poolName,
		"google_iam_workforce_pool_iam_policy", "google_iam_workforce_pool_iam_binding", "google_iam_workforce_pool_iam_member",
		map[string]string{}, newRes)...)
}

func (g *IamGenerator) loadWorkforcePoolProviders(ctx context.Context, iamSvc *iamv1.Service, poolName, orgID, poolID string) {
	err := iamSvc.Locations.WorkforcePools.Providers.List(poolName).Pages(ctx, func(page *iamv1.ListWorkforcePoolProvidersResponse) error {
		for _, provider := range page.WorkforcePoolProviders {
			providerID := workloadIdentityLastSegment(provider.Name)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				provider.Name, providerID, "google_iam_workforce_pool_provider", g.ProviderName,
				map[string]string{"organization": orgID, "location": "global", "workforce_pool_id": poolID, "provider_id": providerID},
				IamAllowEmptyValues, IamAdditionalFields))
			keys, err := iamSvc.Locations.WorkforcePools.Providers.Keys.List(provider.Name).Context(ctx).Do()
			if err != nil {
				log.Println(err)
				continue
			}
			for _, key := range keys.WorkforcePoolProviderKeys {
				keyID := workloadIdentityLastSegment(key.Name)
				g.Resources = append(g.Resources, terraformutils.NewResource(
					key.Name, keyID, "google_iam_workforce_pool_provider_key", g.ProviderName,
					map[string]string{"organization": orgID, "location": "global", "workforce_pool_id": poolID, "provider_id": providerID, "key_id": keyID},
					IamAllowEmptyValues, IamAdditionalFields))
			}
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}
