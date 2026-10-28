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
	"testing"

	accessapprovalv1 "google.golang.org/api/accessapproval/v1"
	crmv1 "google.golang.org/api/cloudresourcemanager/v1"
	crmv3 "google.golang.org/api/cloudresourcemanager/v3"
	essentialcontactsv1 "google.golang.org/api/essentialcontacts/v1"
	iamv1 "google.golang.org/api/iam/v1"
	iamv2 "google.golang.org/api/iam/v2"
	"google.golang.org/api/option"
	orgpolicyv2 "google.golang.org/api/orgpolicy/v2"
)

// TestIamOrgGenerator feeds canned CRM/IAM/org-policy responses and
// asserts the emitted types and IDs: folders + folder IAM trio, org
// and project IAM trios, org custom roles, service account IAM trio
// and keys, org policies at three scopes, custom constraints, liens,
// Access Approval singletons, essential contacts, tags with IAM trios
// and bindings, deny policies and v3beta policy bindings / PAB.
func TestIamOrgGenerator(t *testing.T) {
	ctx := context.Background()
	policy := `{"bindings":[{"role":"roles/viewer","members":["user:jane@example.com"]}]}`
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: ":getAncestry", body: `{"ancestor":[{"resourceId":{"type":"folder","id":"111"}},{"resourceId":{"type":"organization","id":"123"}}]}`},
		{substr: ":getIamPolicy", body: policy},
		{substr: "principalAccessBoundaryPolicies", body: `{"principalAccessBoundaryPolicies":[{"name":"organizations/123/locations/global/principalAccessBoundaryPolicies/p1"}]}`},
		{substr: "policyBindings", body: `{"policyBindings":[{"name":"projects/999/locations/global/policyBindings/b1"}]}`},
		{substr: "denypolicies", body: `{"policies":[{"name":"policies/cloudresourcemanager.googleapis.com%2Fprojects%2F999/denypolicies/d1"}]}`},
		{substr: "/v1/organizations/123/roles", body: `{"roles":[{"name":"organizations/123/roles/r1"}]}`},
		{substr: "/keys", body: `{"keys":[{"name":"projects/test-project/serviceAccounts/sa@test-project.iam.gserviceaccount.com/keys/k1"}]}`},
		{substr: "/serviceAccounts", body: `{"accounts":[{"name":"projects/test-project/serviceAccounts/999","email":"sa@test-project.iam.gserviceaccount.com","uniqueId":"999"}]}`},
		{substr: "accessApprovalSettings", body: `{"name":"projects/test-project/accessApprovalSettings"}`},
		{substr: "/v1/projects/test-project", body: `{"projectNumber":"999","projectId":"test-project"}`},
		{substr: ":listOrgPolicies", body: `{"policies":[{"constraint":"constraints/compute.disableSerialPortAccess"}]}`},
		{substr: "customConstraints", body: `{"customConstraints":[{"name":"organizations/123/customConstraints/custom.c1"}]}`},
		{substr: "/organizations/123/policies", body: `{"policies":[{"name":"organizations/123/policies/compute.disableSerialPortAccess"}]}`},
		{substr: "/folders/111/policies", body: `{"policies":[{"name":"folders/111/policies/compute.disableSerialPortAccess"}]}`},
		{substr: "/projects/test-project/policies", body: `{"policies":[{"name":"projects/test-project/policies/compute.disableSerialPortAccess"}]}`},
		{substr: "/folders/222/policies", body: `{"policies":[{"name":"folders/222/policies/compute.disableSerialPortAccess"}]}`},
		{substr: "/contacts", body: `{"contacts":[{"name":"projects/test-project/contacts/c1"}]}`},
		{substr: "/v3/folders", body: `{"folders":[{"name":"folders/222","displayName":"Child"}]}`},
		{substr: "/v3/liens", body: `{"liens":[{"name":"liens/l1"}]}`},
		{substr: "/v3/tagKeys", body: `{"tagKeys":[{"name":"tagKeys/333"}]}`},
		{substr: "/v3/tagValues", body: `{"tagValues":[{"name":"tagValues/444"}]}`},
		{substr: "/v3/tagBindings", body: `{"tagBindings":[{"name":"tagBindings/tb1"}]}`},
	}}
	opts := []option.ClientOption{option.WithHTTPClient(transport.client()), option.WithoutAuthentication()}
	crm1, err := crmv1.NewService(ctx, opts...)
	if err != nil {
		t.Fatal(err)
	}
	crm3, err := crmv3.NewService(ctx, opts...)
	if err != nil {
		t.Fatal(err)
	}
	iamSvc, err := iamv1.NewService(ctx, opts...)
	if err != nil {
		t.Fatal(err)
	}
	iam2Svc, err := iamv2.NewService(ctx, opts...)
	if err != nil {
		t.Fatal(err)
	}
	orgSvc, err := orgpolicyv2.NewService(ctx, opts...)
	if err != nil {
		t.Fatal(err)
	}
	aaSvc, err := accessapprovalv1.NewService(ctx, opts...)
	if err != nil {
		t.Fatal(err)
	}
	ecSvc, err := essentialcontactsv1.NewService(ctx, opts...)
	if err != nil {
		t.Fatal(err)
	}

	g := &IamOrgGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project"})
	if err := g.initResources(ctx, crm1, crm3, iamSvc, iam2Svc, orgSvc, aaSvc, ecSvc, transport.client()); err != nil {
		t.Fatalf("initResources: %v", err)
	}

	denyID := "policies/cloudresourcemanager.googleapis.com%2Fprojects%2F999/denypolicies/d1"
	saID := "projects/test-project/serviceAccounts/sa@test-project.iam.gserviceaccount.com"
	expected := map[string][]string{
		"google_folder":                                {"folders/111", "folders/222"},
		"google_folder_iam_policy":                     {"folders/111", "folders/222"},
		"google_folder_iam_binding":                    {"folders/111", "folders/222"},
		"google_folder_iam_member":                     {"folders/111", "folders/222"},
		"google_project_iam_policy":                    {"test-project"},
		"google_project_iam_binding":                   {"test-project"},
		"google_project_iam_member":                    {"test-project"},
		"google_organization_iam_policy":               {"123"},
		"google_organization_iam_binding":              {"123"},
		"google_organization_iam_member":               {"123"},
		"google_organization_iam_custom_role":          {"organizations/123/roles/r1"},
		"google_service_account_iam_policy":            {saID},
		"google_service_account_iam_binding":           {saID},
		"google_service_account_iam_member":            {saID},
		"google_service_account_key":                   {saID + "/keys/k1"},
		"google_org_policy_policy":                     {"projects/test-project/policies/compute.disableSerialPortAccess", "folders/111/policies/compute.disableSerialPortAccess", "folders/222/policies/compute.disableSerialPortAccess", "organizations/123/policies/compute.disableSerialPortAccess"},
		"google_project_organization_policy":           {"test-project/constraints/compute.disableSerialPortAccess"},
		"google_folder_organization_policy":            {"folders/111/constraints/compute.disableSerialPortAccess", "folders/222/constraints/compute.disableSerialPortAccess"},
		"google_organization_policy":                   {"123/constraints/compute.disableSerialPortAccess"},
		"google_org_policy_custom_constraint":          {"organizations/123/customConstraints/custom.c1"},
		"google_project_access_approval_settings":      {"projects/test-project/accessApprovalSettings"},
		"google_folder_access_approval_settings":       {"projects/test-project/accessApprovalSettings"},
		"google_organization_access_approval_settings": {"projects/test-project/accessApprovalSettings"},
		"google_essential_contacts_contact":            {"projects/test-project/contacts/c1"},
		"google_iam_deny_policy":                       {denyID},
		"google_iam_projects_policy_binding":           {"projects/999/locations/global/policyBindings/b1"},
		"google_iam_folders_policy_binding":            {"projects/999/locations/global/policyBindings/b1"},
		"google_iam_organizations_policy_binding":      {"projects/999/locations/global/policyBindings/b1"},
		"google_iam_principal_access_boundary_policy":  {"organizations/123/locations/global/principalAccessBoundaryPolicies/p1"},
		"google_resource_manager_lien":                 {"liens/l1"},
		"google_tags_tag_key":                          {"tagKeys/333"},
		"google_tags_tag_key_iam_policy":               {"tagKeys/333"},
		"google_tags_tag_key_iam_binding":              {"tagKeys/333"},
		"google_tags_tag_key_iam_member":               {"tagKeys/333"},
		"google_tags_tag_value":                        {"tagValues/444"},
		"google_tags_tag_value_iam_policy":             {"tagValues/444"},
		"google_tags_tag_value_iam_binding":            {"tagValues/444"},
		"google_tags_tag_value_iam_member":             {"tagValues/444"},
		"google_tags_tag_binding":                      {"tagBindings/tb1"},
	}
	assertResourceIDs(t, g.Resources, expected)
}
