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
	"net/http"
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"

	accesscontextmanagerv1 "google.golang.org/api/accesscontextmanager/v1"
	cloudresourcemanagerv1 "google.golang.org/api/cloudresourcemanager/v1"
)

// AccessContextManagerGenerator enumerates the Access Context
// Manager estate of the organization the project belongs to (found
// through Cloud Resource Manager ancestry): access policies, their
// combined IAM resource, access levels, service perimeters,
// authorized orgs descriptions, and the organization's GCP user
// access bindings.
//
// The access policy IAM policy is read over raw REST: the API has
// accessPolicies.getIamPolicy but the generated client in the pinned
// module does not surface it. The provider models that policy as the
// single combined google_access_context_manager_access_policy_iam
// resource (ID = the policy name), emitted when the policy has
// bindings.
//
// Skipped, with reasons (all verified against the provider's Magic
// Modules definitions):
//   - google_access_context_manager_access_level_condition: the
//     provider documents "This resource does not support import";
//     a condition has no ID of its own (its ID is the parent access
//     level's name) — it is embedded in the access level.
//   - google_access_context_manager_service_perimeter_egress_policy,
//     google_access_context_manager_service_perimeter_ingress_policy,
//     google_access_context_manager_service_perimeter_resource and
//     the dry-run variants
//     google_access_context_manager_service_perimeter_dry_run_egress_policy,
//     google_access_context_manager_service_perimeter_dry_run_ingress_policy,
//     google_access_context_manager_service_perimeter_dry_run_resource,
//     plus the audit's short forms
//     google_access_context_manager_egress_policy and
//     google_access_context_manager_ingress_policy: all are embedded
//     in the perimeter's spec/status/dryRunSpec with the perimeter's
//     own name as ID and do not support import.
//   - google_access_context_manager_access_levels and
//     google_access_context_manager_service_perimeters: plural bulk
//     resources duplicating the singular resources over the same API
//     objects (Batch 3 SCC plural precedent).
type AccessContextManagerGenerator struct {
	GCPService
}

const accessContextManagerBaseURL = "https://accesscontextmanager.googleapis.com/v1/"

func (g *AccessContextManagerGenerator) InitResources() error {
	ctx := context.Background()
	acmService, err := accesscontextmanagerv1.NewService(ctx)
	if err != nil {
		log.Println(err)
		acmService = nil
	}
	crmService, err := cloudresourcemanagerv1.NewService(ctx)
	if err != nil {
		log.Println(err)
		crmService = nil
	}
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		log.Println(err)
		restClient = nil
	}
	g.initResources(ctx, acmService, crmService, restClient)
	return nil
}

func (g *AccessContextManagerGenerator) newACMResource(id, resourceType string) terraformutils.Resource {
	parts := strings.Split(id, "/")
	return terraformutils.NewResource(
		id,
		parts[len(parts)-1],
		resourceType,
		g.ProviderName,
		map[string]string{},
		[]string{},
		map[string]interface{}{},
	)
}

func (g *AccessContextManagerGenerator) initResources(ctx context.Context, acmService *accesscontextmanagerv1.Service, crmService *cloudresourcemanagerv1.Service, restClient *http.Client) {
	if acmService == nil || crmService == nil {
		return
	}
	project := g.GetArgs()["project"].(string)
	ancestry, err := crmService.Projects.GetAncestry(project, &cloudresourcemanagerv1.GetAncestryRequest{}).Context(ctx).Do()
	if err != nil {
		log.Println(err)
		g.RecordFailure("", err)
		return
	}
	orgID := ""
	for _, ancestor := range ancestry.Ancestor {
		if ancestor.ResourceId != nil && ancestor.ResourceId.Type == "organization" {
			orgID = ancestor.ResourceId.Id
		}
	}
	if orgID == "" {
		log.Println("accesscontextmanager: no organization ancestor for project", project)
		return
	}
	orgName := "organizations/" + orgID

	bindingsCall := acmService.Organizations.GcpUserAccessBindings.List(orgName)
	if err := bindingsCall.Pages(ctx, func(page *accesscontextmanagerv1.ListGcpUserAccessBindingsResponse) error {
		for _, binding := range page.GcpUserAccessBindings {
			g.Resources = append(g.Resources, g.newACMResource(binding.Name, "google_access_context_manager_gcp_user_access_binding"))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}

	policiesCall := acmService.AccessPolicies.List().Parent(orgName)
	if err := policiesCall.Pages(ctx, func(page *accesscontextmanagerv1.ListAccessPoliciesResponse) error {
		for _, policy := range page.AccessPolicies {
			g.Resources = append(g.Resources, g.newACMResource(policy.Name, "google_access_context_manager_access_policy"))
			g.accessPolicyIAM(ctx, restClient, policy.Name)
			g.accessPolicyChildren(ctx, acmService, policy.Name)
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
}

// accessPolicyIAM emits the combined access-policy IAM resource when
// the policy carries bindings (raw REST getIamPolicy).
func (g *AccessContextManagerGenerator) accessPolicyIAM(ctx context.Context, restClient *http.Client, policyName string) {
	if restClient == nil {
		return
	}
	bindings, err := restGetIamPolicy(ctx, restClient, accessContextManagerBaseURL+policyName+":getIamPolicy")
	if err != nil {
		log.Println(err)
		return
	}
	if len(bindings) == 0 {
		return
	}
	g.Resources = append(g.Resources, g.newACMResource(policyName, "google_access_context_manager_access_policy_iam"))
}

// accessPolicyChildren lists one policy's access levels, service
// perimeters and authorized orgs descriptions.
func (g *AccessContextManagerGenerator) accessPolicyChildren(ctx context.Context, acmService *accesscontextmanagerv1.Service, policyName string) {
	levelsCall := acmService.AccessPolicies.AccessLevels.List(policyName)
	if err := levelsCall.Pages(ctx, func(page *accesscontextmanagerv1.ListAccessLevelsResponse) error {
		for _, level := range page.AccessLevels {
			g.Resources = append(g.Resources, g.newACMResource(level.Name, "google_access_context_manager_access_level"))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	perimetersCall := acmService.AccessPolicies.ServicePerimeters.List(policyName)
	if err := perimetersCall.Pages(ctx, func(page *accesscontextmanagerv1.ListServicePerimetersResponse) error {
		for _, perimeter := range page.ServicePerimeters {
			g.Resources = append(g.Resources, g.newACMResource(perimeter.Name, "google_access_context_manager_service_perimeter"))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	descsCall := acmService.AccessPolicies.AuthorizedOrgsDescs.List(policyName)
	if err := descsCall.Pages(ctx, func(page *accesscontextmanagerv1.ListAuthorizedOrgsDescsResponse) error {
		for _, desc := range page.AuthorizedOrgsDescs {
			g.Resources = append(g.Resources, g.newACMResource(desc.Name, "google_access_context_manager_authorized_orgs_desc"))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
}
