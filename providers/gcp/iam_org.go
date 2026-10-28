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
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	accessapprovalv1 "google.golang.org/api/accessapproval/v1"
	crmv1 "google.golang.org/api/cloudresourcemanager/v1"
	crmv3 "google.golang.org/api/cloudresourcemanager/v3"
	essentialcontactsv1 "google.golang.org/api/essentialcontacts/v1"
	iamv1 "google.golang.org/api/iam/v1"
	iamv2 "google.golang.org/api/iam/v2"
	orgpolicyv2 "google.golang.org/api/orgpolicy/v2"
)

var iamOrgAllowEmptyValues = []string{""}

var iamOrgAdditionalFields = map[string]interface{}{}

// IamOrgGenerator enumerates the resource-manager / IAM-policy estate
// reachable from one project import: the project itself, its ancestor
// folders (discovered by folder BFS from the organization and the
// ancestry chain) and its organization, resolved through Cloud
// Resource Manager ancestry (the scc.go / networksecurity.go pattern).
//
// Covered: folders and their IAM trio, the organization IAM trio and
// organization custom roles, the project IAM trio, service account
// IAM trios and service account keys, organization policies at
// project/folder/organization scope plus organization custom
// constraints, resource-manager liens on the project, Access Approval
// settings singletons at all three scopes, Essential Contacts at all
// three scopes, Tags (keys, values, their IAM trios and project /
// folder / organization tag bindings), IAM deny policies at all three
// scopes (iam/v2 ListPolicies — the pinned client does expose a list
// call, superseding the older note in iam_workload_identity.go),
// principal access boundary policies (iam/v3beta, raw REST) and
// policy bindings at project/folder/organization scope (iam/v3beta,
// raw REST).
//
// Deliberately not generated:
//   - google_iam_access_boundary_policy, google_iam_project_access_policy,
//     google_iam_folder_access_policy and
//     google_iam_organization_access_policy: these beta "access
//     policy" attachments have no list call in the iam/v2, v2beta or
//     v3beta surfaces (only the policy definitions —
//     principalAccessBoundaryPolicies — and the policyBindings
//     collections are listable, and both are generated here); an
//     access policy is addressed by a user-chosen name that cannot be
//     discovered from any listable parent.
//   - google_project_service_identity, google_folder_service_identity
//     and google_organization_service_identity: a service identity is
//     provisioned through the Service Usage generateServiceIdentity
//     call; there is no listable or gettable object and reading it
//     would mutate state (same reasoning as
//     google_scc_notification_service_account in scc.go).
//   - google_tags_location_tag_binding: location-scoped tag bindings
//     are listed only per location through regional Resource Manager
//     endpoints; the pinned CRM v3 client exposes only the global
//     TagBindings list (generated here as google_tags_tag_binding)
//     and no location enumeration to drive the regional lists.
//   - google_resource_manager_boundary, google_resource_manager_boundary_config,
//     google_resource_manager_capability and
//     google_resource_manager_capability_config: the CRM v1/v3
//     clients have no Boundaries or Capabilities service at all and
//     no raw-REST list endpoint for either collection appears in the
//     CRM discovery documents, so neither collection is discoverable.
type IamOrgGenerator struct {
	GCPService
}

type iamOrgScope struct {
	kind string // "project", "folder" or "organization"
	name string // "projects/{id}", "folders/{id}" or "organizations/{id}"
	id   string // project ID, folder ID or organization ID
	// numericID is the project number for the project scope (deny
	// policies and policy bindings address projects by number); for
	// folders and organizations id is already numeric.
	numericID string
}

func (g *IamOrgGenerator) InitResources() error {
	ctx := context.Background()
	crm1, err := crmv1.NewService(ctx)
	if err != nil {
		return err
	}
	crm3, err := crmv3.NewService(ctx)
	if err != nil {
		return err
	}
	iamSvc, err := iamv1.NewService(ctx)
	if err != nil {
		return err
	}
	iam2Svc, err := iamv2.NewService(ctx)
	if err != nil {
		return err
	}
	orgSvc, err := orgpolicyv2.NewService(ctx)
	if err != nil {
		return err
	}
	aaSvc, err := accessapprovalv1.NewService(ctx)
	if err != nil {
		return err
	}
	ecSvc, err := essentialcontactsv1.NewService(ctx)
	if err != nil {
		return err
	}
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, crm1, crm3, iamSvc, iam2Svc, orgSvc, aaSvc, ecSvc, restClient)
}

func (g *IamOrgGenerator) newIamOrgResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
	if attrs == nil {
		attrs = map[string]string{}
	}
	parts := strings.Split(id, "/")
	return terraformutils.NewResource(
		id,
		parts[len(parts)-1],
		resourceType,
		g.ProviderName,
		attrs,
		iamOrgAllowEmptyValues,
		iamOrgAdditionalFields,
	)
}

func (g *IamOrgGenerator) initResources(ctx context.Context, crm1 *crmv1.Service, crm3 *crmv3.Service, iamSvc *iamv1.Service, iam2Svc *iamv2.Service, orgSvc *orgpolicyv2.Service, aaSvc *accessapprovalv1.Service, ecSvc *essentialcontactsv1.Service, restClient *http.Client) error {
	project := g.GetArgs()["project"].(string)
	scopes := []iamOrgScope{{kind: "project", name: "projects/" + project, id: project}}
	if proj, err := crm1.Projects.Get(project).Context(ctx).Do(); err == nil && proj != nil {
		scopes[0].numericID = strconv.FormatInt(proj.ProjectNumber, 10)
	} else if err != nil {
		log.Println(err)
	}
	var orgID string
	var ancestorFolders []string
	if crm1 != nil {
		if ancestry, err := crm1.Projects.GetAncestry(project, &crmv1.GetAncestryRequest{}).Context(ctx).Do(); err == nil {
			for _, ancestor := range ancestry.Ancestor {
				if ancestor.ResourceId == nil {
					continue
				}
				switch ancestor.ResourceId.Type {
				case "folder":
					ancestorFolders = append(ancestorFolders, ancestor.ResourceId.Id)
				case "organization":
					orgID = ancestor.ResourceId.Id
				}
			}
		} else {
			log.Println(err)
		}
	}
	// Folder BFS: start from the organization and the ancestry
	// folders, list child folders level by level.
	folderIDs := []string{}
	seenFolder := map[string]bool{}
	queue := []string{}
	if orgID != "" {
		queue = append(queue, "organizations/"+orgID)
	}
	for _, f := range ancestorFolders {
		queue = append(queue, "folders/"+f)
	}
	// Ancestor folders themselves are in scope even if BFS from the
	// organization cannot reach them (permissions).
	for _, f := range ancestorFolders {
		if !seenFolder[f] {
			seenFolder[f] = true
			folderIDs = append(folderIDs, f)
		}
	}
	for len(queue) > 0 {
		parent := queue[0]
		queue = queue[1:]
		if crm3 == nil {
			break
		}
		err := crm3.Folders.List().Parent(parent).Pages(ctx, func(page *crmv3.ListFoldersResponse) error {
			for _, folder := range page.Folders {
				id := strings.TrimPrefix(folder.Name, "folders/")
				if seenFolder[id] {
					continue
				}
				seenFolder[id] = true
				folderIDs = append(folderIDs, id)
				queue = append(queue, folder.Name)
			}
			return nil
		})
		if err != nil {
			log.Println(err)
		}
	}
	for _, f := range folderIDs {
		scopes = append(scopes, iamOrgScope{kind: "folder", name: "folders/" + f, id: f, numericID: f})
	}
	if orgID != "" {
		scopes = append(scopes, iamOrgScope{kind: "organization", name: "organizations/" + orgID, id: orgID, numericID: orgID})
	}

	g.loadFolderResources(ctx, crm3, folderIDs)
	g.loadProjectAndOrgIam(ctx, crm1, crm3, restClient, project, orgID)
	g.loadServiceAccountIam(ctx, iamSvc, project)
	for _, scope := range scopes {
		g.loadOrgPolicy(ctx, orgSvc, scope)
		g.loadLegacyOrgPolicy(ctx, crm1, scope)
		g.loadAccessApproval(ctx, aaSvc, scope)
		g.loadEssentialContacts(ctx, ecSvc, scope)
		g.loadDenyPolicies(ctx, iam2Svc, scope)
		g.loadPolicyBindings(ctx, restClient, scope)
	}
	g.loadOrgCustomConstraints(ctx, orgSvc, orgID)
	g.loadLiens(ctx, crm3, project)
	g.loadTags(ctx, crm3, scopes, project)
	g.loadPrincipalAccessBoundaryPolicies(ctx, restClient, orgID)
	return nil
}

func (g *IamOrgGenerator) loadFolderResources(ctx context.Context, crm3 *crmv3.Service, folderIDs []string) {
	if crm3 == nil {
		return
	}
	for _, id := range folderIDs {
		name := "folders/" + id
		g.Resources = append(g.Resources, g.newIamOrgResource(name, "google_folder", map[string]string{"folder": name}))
		policy, err := crm3.Folders.GetIamPolicy(name, &crmv3.GetIamPolicyRequest{}).Context(ctx).Do()
		if err != nil {
			log.Println(err)
			continue
		}
		bindings := []iamBinding{}
		for _, b := range policy.Bindings {
			bindings = append(bindings, iamBinding{Role: b.Role, Members: b.Members})
		}
		g.Resources = append(g.Resources, expandIamPolicy(bindings, name,
			"google_folder_iam_policy", "google_folder_iam_binding", "google_folder_iam_member",
			map[string]string{"folder": name}, g.newIamOrgResource)...)
	}
}

func (g *IamOrgGenerator) loadProjectAndOrgIam(ctx context.Context, crm1 *crmv1.Service, crm3 *crmv3.Service, restClient *http.Client, project, orgID string) {
	if crm1 != nil {
		if policy, err := crm1.Projects.GetIamPolicy(project, &crmv1.GetIamPolicyRequest{}).Context(ctx).Do(); err == nil {
			bindings := []iamBinding{}
			for _, b := range policy.Bindings {
				bindings = append(bindings, iamBinding{Role: b.Role, Members: b.Members})
			}
			g.Resources = append(g.Resources, expandIamPolicy(bindings, project,
				"google_project_iam_policy", "google_project_iam_binding", "google_project_iam_member",
				map[string]string{"project": project}, g.newIamOrgResource)...)
		} else {
			log.Println(err)
		}
	}
	if orgID == "" || crm3 == nil {
		return
	}
	orgName := "organizations/" + orgID
	if policy, err := crm3.Organizations.GetIamPolicy(orgName, &crmv3.GetIamPolicyRequest{}).Context(ctx).Do(); err == nil {
		bindings := []iamBinding{}
		for _, b := range policy.Bindings {
			bindings = append(bindings, iamBinding{Role: b.Role, Members: b.Members})
		}
		g.Resources = append(g.Resources, expandIamPolicy(bindings, orgID,
			"google_organization_iam_policy", "google_organization_iam_binding", "google_organization_iam_member",
			map[string]string{"org_id": orgID}, g.newIamOrgResource)...)
	} else {
		log.Println(err)
	}
	// Organization custom roles over raw REST (iam/v1 roles list).
	if restClient != nil {
		roles, err := restListResources(ctx, restClient, "https://iam.googleapis.com/v1/"+orgName+"/roles")
		if err != nil {
			log.Println(err)
			return
		}
		for _, role := range roles {
			g.Resources = append(g.Resources, g.newIamOrgResource(role.Name, "google_organization_iam_custom_role", map[string]string{"org_id": orgID}))
		}
	}
}

func (g *IamOrgGenerator) loadServiceAccountIam(ctx context.Context, iamSvc *iamv1.Service, project string) {
	if iamSvc == nil {
		return
	}
	parent := "projects/" + project
	err := iamSvc.Projects.ServiceAccounts.List(parent).Pages(ctx, func(page *iamv1.ListServiceAccountsResponse) error {
		for _, sa := range page.Accounts {
			saResource := parent + "/serviceAccounts/" + sa.Email
			policy, err := iamSvc.Projects.ServiceAccounts.GetIamPolicy(saResource).Context(ctx).Do()
			if err == nil {
				bindings := []iamBinding{}
				for _, b := range policy.Bindings {
					bindings = append(bindings, iamBinding{Role: b.Role, Members: b.Members})
				}
				g.Resources = append(g.Resources, expandIamPolicy(bindings, saResource,
					"google_service_account_iam_policy", "google_service_account_iam_binding", "google_service_account_iam_member",
					map[string]string{"service_account_id": saResource, "project": project}, g.newIamOrgResource)...)
			} else {
				log.Println(err)
			}
			keys, err := iamSvc.Projects.ServiceAccounts.Keys.List(saResource).Context(ctx).Do()
			if err != nil {
				log.Println(err)
				continue
			}
			for _, key := range keys.Keys {
				g.Resources = append(g.Resources, g.newIamOrgResource(key.Name, "google_service_account_key", map[string]string{"service_account_id": saResource, "project": project}))
			}
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (g *IamOrgGenerator) loadOrgPolicy(ctx context.Context, orgSvc *orgpolicyv2.Service, scope iamOrgScope) {
	if orgSvc == nil {
		return
	}
	emit := func(name, resourceType string) {
		g.Resources = append(g.Resources, g.newIamOrgResource(name, resourceType, map[string]string{}))
	}
	switch scope.kind {
	case "project":
		_ = orgSvc.Projects.Policies.List(scope.name).Pages(ctx, func(page *orgpolicyv2.GoogleCloudOrgpolicyV2ListPoliciesResponse) error {
			for _, p := range page.Policies {
				emit(p.Name, "google_org_policy_policy")
			}
			return nil
		})
	case "folder":
		_ = orgSvc.Folders.Policies.List(scope.name).Pages(ctx, func(page *orgpolicyv2.GoogleCloudOrgpolicyV2ListPoliciesResponse) error {
			for _, p := range page.Policies {
				emit(p.Name, "google_org_policy_policy")
			}
			return nil
		})
	case "organization":
		_ = orgSvc.Organizations.Policies.List(scope.name).Pages(ctx, func(page *orgpolicyv2.GoogleCloudOrgpolicyV2ListPoliciesResponse) error {
			for _, p := range page.Policies {
				emit(p.Name, "google_org_policy_policy")
			}
			return nil
		})
	}
}

// loadLegacyOrgPolicy enumerates the legacy (v1) organization
// policies set directly on a scope: google_project_organization_policy,
// google_folder_organization_policy and google_organization_policy,
// whose provider IDs are "{scope}/{constraint}".
func (g *IamOrgGenerator) loadLegacyOrgPolicy(ctx context.Context, crm1 *crmv1.Service, scope iamOrgScope) {
	if crm1 == nil {
		return
	}
	req := &crmv1.ListOrgPoliciesRequest{}
	var resp *crmv1.ListOrgPoliciesResponse
	var err error
	var resourceType, prefix string
	switch scope.kind {
	case "project":
		resp, err = crm1.Projects.ListOrgPolicies(scope.id, req).Context(ctx).Do()
		resourceType, prefix = "google_project_organization_policy", scope.id
	case "folder":
		resp, err = crm1.Folders.ListOrgPolicies(scope.name, req).Context(ctx).Do()
		resourceType, prefix = "google_folder_organization_policy", scope.name
	case "organization":
		resp, err = crm1.Organizations.ListOrgPolicies(scope.name, req).Context(ctx).Do()
		resourceType, prefix = "google_organization_policy", scope.id
	}
	if err != nil {
		log.Println(err)
		return
	}
	if resp == nil {
		return
	}
	for _, policy := range resp.Policies {
		if policy.Constraint == "" {
			continue
		}
		id := prefix + "/" + policy.Constraint
		g.Resources = append(g.Resources, g.newIamOrgResource(id, resourceType, map[string]string{}))
	}
}

func (g *IamOrgGenerator) loadOrgCustomConstraints(ctx context.Context, orgSvc *orgpolicyv2.Service, orgID string) {
	if orgSvc == nil || orgID == "" {
		return
	}
	err := orgSvc.Organizations.CustomConstraints.List("organizations/"+orgID).Pages(ctx, func(page *orgpolicyv2.GoogleCloudOrgpolicyV2ListCustomConstraintsResponse) error {
		for _, c := range page.CustomConstraints {
			g.Resources = append(g.Resources, g.newIamOrgResource(c.Name, "google_org_policy_custom_constraint", map[string]string{"org_id": orgID}))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (g *IamOrgGenerator) loadAccessApproval(ctx context.Context, aaSvc *accessapprovalv1.Service, scope iamOrgScope) {
	if aaSvc == nil {
		return
	}
	var name string
	var resourceType string
	var err error
	switch scope.kind {
	case "project":
		var s *accessapprovalv1.AccessApprovalSettings
		s, err = aaSvc.Projects.GetAccessApprovalSettings(scope.name + "/accessApprovalSettings").Context(ctx).Do()
		if s != nil {
			name = s.Name
		}
		resourceType = "google_project_access_approval_settings"
	case "folder":
		var s *accessapprovalv1.AccessApprovalSettings
		s, err = aaSvc.Folders.GetAccessApprovalSettings(scope.name + "/accessApprovalSettings").Context(ctx).Do()
		if s != nil {
			name = s.Name
		}
		resourceType = "google_folder_access_approval_settings"
	case "organization":
		var s *accessapprovalv1.AccessApprovalSettings
		s, err = aaSvc.Organizations.GetAccessApprovalSettings(scope.name + "/accessApprovalSettings").Context(ctx).Do()
		if s != nil {
			name = s.Name
		}
		resourceType = "google_organization_access_approval_settings"
	}
	if err != nil {
		log.Println(err)
		return
	}
	if name != "" {
		g.Resources = append(g.Resources, g.newIamOrgResource(name, resourceType, map[string]string{}))
	}
}

func (g *IamOrgGenerator) loadEssentialContacts(ctx context.Context, ecSvc *essentialcontactsv1.Service, scope iamOrgScope) {
	if ecSvc == nil {
		return
	}
	emit := func(page *essentialcontactsv1.GoogleCloudEssentialcontactsV1ListContactsResponse) error {
		for _, c := range page.Contacts {
			g.Resources = append(g.Resources, g.newIamOrgResource(c.Name, "google_essential_contacts_contact", map[string]string{"parent": scope.name}))
		}
		return nil
	}
	var err error
	switch scope.kind {
	case "project":
		err = ecSvc.Projects.Contacts.List(scope.name).Pages(ctx, emit)
	case "folder":
		err = ecSvc.Folders.Contacts.List(scope.name).Pages(ctx, emit)
	case "organization":
		err = ecSvc.Organizations.Contacts.List(scope.name).Pages(ctx, emit)
	}
	if err != nil {
		log.Println(err)
	}
}

func (g *IamOrgGenerator) loadDenyPolicies(ctx context.Context, iam2Svc *iamv2.Service, scope iamOrgScope) {
	if iam2Svc == nil {
		return
	}
	attachment := scope.name
	if scope.kind == "project" && scope.numericID != "" {
		attachment = "projects/" + scope.numericID
	}
	parent := "policies/" + url.PathEscape("cloudresourcemanager.googleapis.com/"+attachment) + "/denypolicies"
	// url.PathEscape escapes "/" as %2F, which is the encoding the
	// iam/v2 attachment point requires.
	err := iam2Svc.Policies.ListPolicies(parent).Pages(ctx, func(page *iamv2.GoogleIamV2ListPoliciesResponse) error {
		for _, p := range page.Policies {
			g.Resources = append(g.Resources, g.newIamOrgResource(p.Name, "google_iam_deny_policy", map[string]string{}))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (g *IamOrgGenerator) loadPolicyBindings(ctx context.Context, restClient *http.Client, scope iamOrgScope) {
	if restClient == nil {
		return
	}
	parent := scope.name
	if scope.kind == "project" && scope.numericID != "" {
		parent = "projects/" + scope.numericID
	}
	items, err := restListResources(ctx, restClient, "https://iam.googleapis.com/v3beta/"+parent+"/locations/global/policyBindings")
	if err != nil {
		log.Println(err)
		return
	}
	resourceType := map[string]string{
		"project":      "google_iam_projects_policy_binding",
		"folder":       "google_iam_folders_policy_binding",
		"organization": "google_iam_organizations_policy_binding",
	}[scope.kind]
	for _, item := range items {
		g.Resources = append(g.Resources, g.newIamOrgResource(item.Name, resourceType, map[string]string{}))
	}
}

func (g *IamOrgGenerator) loadPrincipalAccessBoundaryPolicies(ctx context.Context, restClient *http.Client, orgID string) {
	if restClient == nil || orgID == "" {
		return
	}
	items, err := restListResources(ctx, restClient, fmt.Sprintf("https://iam.googleapis.com/v3beta/organizations/%s/locations/global/principalAccessBoundaryPolicies", orgID))
	if err != nil {
		log.Println(err)
		return
	}
	for _, item := range items {
		g.Resources = append(g.Resources, g.newIamOrgResource(item.Name, "google_iam_principal_access_boundary_policy", map[string]string{"organization": orgID}))
	}
}

func (g *IamOrgGenerator) loadLiens(ctx context.Context, crm3 *crmv3.Service, project string) {
	if crm3 == nil {
		return
	}
	err := crm3.Liens.List().Parent("projects/"+project).Pages(ctx, func(page *crmv3.ListLiensResponse) error {
		for _, lien := range page.Liens {
			g.Resources = append(g.Resources, g.newIamOrgResource(lien.Name, "google_resource_manager_lien", map[string]string{"parent": "projects/" + project}))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (g *IamOrgGenerator) loadTags(ctx context.Context, crm3 *crmv3.Service, scopes []iamOrgScope, project string) {
	if crm3 == nil {
		return
	}
	parents := []string{"projects/" + project}
	for _, scope := range scopes {
		if scope.kind == "organization" {
			parents = append(parents, scope.name)
		}
	}
	seenKey := map[string]bool{}
	for _, parent := range parents {
		err := crm3.TagKeys.List().Parent(parent).Pages(ctx, func(page *crmv3.ListTagKeysResponse) error {
			for _, key := range page.TagKeys {
				if seenKey[key.Name] {
					continue
				}
				seenKey[key.Name] = true
				g.Resources = append(g.Resources, g.newIamOrgResource(key.Name, "google_tags_tag_key", map[string]string{"parent": parent}))
				if policy, err := crm3.TagKeys.GetIamPolicy(key.Name, &crmv3.GetIamPolicyRequest{}).Context(ctx).Do(); err == nil {
					bindings := []iamBinding{}
					for _, b := range policy.Bindings {
						bindings = append(bindings, iamBinding{Role: b.Role, Members: b.Members})
					}
					g.Resources = append(g.Resources, expandIamPolicy(bindings, key.Name,
						"google_tags_tag_key_iam_policy", "google_tags_tag_key_iam_binding", "google_tags_tag_key_iam_member",
						map[string]string{}, g.newIamOrgResource)...)
				} else {
					log.Println(err)
				}
				g.loadTagValues(ctx, crm3, key.Name)
			}
			return nil
		})
		if err != nil {
			log.Println(err)
		}
	}
	// Tag bindings attached to the project, folders and organization.
	for _, scope := range scopes {
		parent := "//cloudresourcemanager.googleapis.com/" + scope.name
		if scope.kind == "project" && scope.numericID != "" {
			parent = "//cloudresourcemanager.googleapis.com/projects/" + scope.numericID
		}
		err := crm3.TagBindings.List().Parent(parent).Pages(ctx, func(page *crmv3.ListTagBindingsResponse) error {
			for _, binding := range page.TagBindings {
				g.Resources = append(g.Resources, g.newIamOrgResource(binding.Name, "google_tags_tag_binding", map[string]string{}))
			}
			return nil
		})
		if err != nil {
			log.Println(err)
		}
	}
}

func (g *IamOrgGenerator) loadTagValues(ctx context.Context, crm3 *crmv3.Service, keyName string) {
	err := crm3.TagValues.List().Parent(keyName).Pages(ctx, func(page *crmv3.ListTagValuesResponse) error {
		for _, value := range page.TagValues {
			g.Resources = append(g.Resources, g.newIamOrgResource(value.Name, "google_tags_tag_value", map[string]string{"parent": keyName}))
			if policy, err := crm3.TagValues.GetIamPolicy(value.Name, &crmv3.GetIamPolicyRequest{}).Context(ctx).Do(); err == nil {
				bindings := []iamBinding{}
				for _, b := range policy.Bindings {
					bindings = append(bindings, iamBinding{Role: b.Role, Members: b.Members})
				}
				g.Resources = append(g.Resources, expandIamPolicy(bindings, value.Name,
					"google_tags_tag_value_iam_policy", "google_tags_tag_value_iam_binding", "google_tags_tag_value_iam_member",
					map[string]string{}, g.newIamOrgResource)...)
			} else {
				log.Println(err)
			}
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}
