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

	accesscontextmanagerv1 "google.golang.org/api/accesscontextmanager/v1"
	cloudresourcemanagerv1 "google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/option"
)

// TestAccessContextManagerGenerator resolves the organization from
// canned Resource Manager ancestry, then feeds canned Access
// Context Manager responses through the generator and verifies the
// emitted types and IDs: the access policy, its combined IAM
// resource (the canned :getIamPolicy returns one binding), an
// access level, a service perimeter, an authorized orgs desc, and a
// GCP user access binding.
func TestAccessContextManagerGenerator(t *testing.T) {
	policy := "accessPolicies/123"
	transport := &cannedTransport{routes: []cannedRoute{
		{":getAncestry", `{"ancestor":[{"resourceId":{"type":"organization","id":"6789"}}]}`},
		{":getIamPolicy", `{"bindings":[{"role":"roles/accesscontextmanager.policyReader","members":["user:a@example.com"]}]}`},
		{"/organizations/6789/gcpUserAccessBindings", `{"gcpUserAccessBindings":[{"name":"organizations/6789/gcpUserAccessBindings/bind1"}]}`},
		{"/accessPolicies/123/accessLevels", `{"accessLevels":[{"name":"` + policy + `/accessLevels/lev1"}]}`},
		{"/accessPolicies/123/servicePerimeters", `{"servicePerimeters":[{"name":"` + policy + `/servicePerimeters/per1"}]}`},
		{"/accessPolicies/123/authorizedOrgsDescs", `{"authorizedOrgsDescs":[{"name":"` + policy + `/authorizedOrgsDescs/desc1"}]}`},
		// The policies list path is the bare collection; the parent
		// travels as a query parameter in this client.
		{"/v1/accessPolicies", `{"accessPolicies":[{"name":"` + policy + `"}]}`},
	}}

	ctx := context.Background()
	clientOpts := []option.ClientOption{
		option.WithHTTPClient(transport.client()),
		option.WithoutAuthentication(),
	}
	acmService, err := accesscontextmanagerv1.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("accesscontextmanager.NewService: %v", err)
	}
	crmService, err := cloudresourcemanagerv1.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("cloudresourcemanager.NewService: %v", err)
	}

	g := &AccessContextManagerGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
	})
	g.initResources(ctx, acmService, crmService, transport.client())

	expected := map[string][]string{
		"google_access_context_manager_access_policy":           {policy},
		"google_access_context_manager_access_policy_iam":       {policy},
		"google_access_context_manager_access_level":            {policy + "/accessLevels/lev1"},
		"google_access_context_manager_service_perimeter":       {policy + "/servicePerimeters/per1"},
		"google_access_context_manager_authorized_orgs_desc":    {policy + "/authorizedOrgsDescs/desc1"},
		"google_access_context_manager_gcp_user_access_binding": {"organizations/6789/gcpUserAccessBindings/bind1"},
	}
	assertResourceIDs(t, g.Resources, expected)
}
