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

	"google.golang.org/api/compute/v1"
	gkehubv1 "google.golang.org/api/gkehub/v1"
	"google.golang.org/api/option"
)

// TestGkeHubGenerator feeds canned GKE Hub responses — the generated
// v1 client for fleets, features, memberships and scopes, raw REST
// for feature memberships and rollout sequences — and verifies the
// emitted types and IDs, including the IAM trios.
func TestGkeHubGenerator(t *testing.T) {
	glob := "projects/test-project/locations/global"
	feature := glob + "/features/clusterupgrade"
	membership := glob + "/memberships/m1"
	scope := glob + "/scopes/default"
	policy := `{"bindings":[{"role":"roles/gkehub.viewer","members":["user:a@example.com"]}]}`
	transport := &cannedTransport{routes: []cannedRoute{
		{":getIamPolicy", policy},
		{"/features/clusterupgrade/memberships", `{"memberships":[{"name":"` + feature + `/memberships/m1"}]}`},
		{"/memberships/m1/bindings", `{"membershipBindings":[{"name":"` + membership + `/bindings/b1"}]}`},
		{"/memberships/m1/rbacrolebindings", `{"rbacrolebindings":[{"name":"` + membership + `/rbacrolebindings/r1"}]}`},
		{"/scopes/default/namespaces", `{"scopeNamespaces":[{"name":"` + scope + `/namespaces/ns1"}]}`},
		{"/scopes/default/rbacrolebindings", `{"rbacrolebindings":[{"name":"` + scope + `/rbacrolebindings/r1"}]}`},
		{"/fleets", `{"fleets":[{"name":"` + glob + `/fleets/default"}]}`},
		{"/features", `{"resources":[{"name":"` + feature + `"}]}`},
		{"/memberships", `{"resources":[{"name":"` + membership + `"}]}`},
		{"/scopes", `{"scopes":[{"name":"` + scope + `"}]}`},
		{"/rolloutSequences", `{"rolloutSequences":[{"name":"` + glob + `/rolloutSequences/rs1"}]}`},
	}}

	ctx := context.Background()
	clientOpts := []option.ClientOption{
		option.WithHTTPClient(transport.client()),
		option.WithoutAuthentication(),
	}
	hubService, err := gkehubv1.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("gkehub.NewService: %v", err)
	}

	g := &GkeHubGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	if err := g.initResources(ctx, hubService, transport.client()); err != nil {
		t.Fatalf("initResources: %v", err)
	}

	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_gke_hub_fleet":                        {glob + "/fleets/default"},
		"google_gke_hub_feature":                      {feature},
		"google_gke_hub_feature_iam_policy":           {feature},
		"google_gke_hub_feature_iam_binding":          {feature},
		"google_gke_hub_feature_iam_member":           {feature},
		"google_gke_hub_feature_membership":           {feature + "/memberships/m1"},
		"google_gke_hub_membership":                   {membership},
		"google_gke_hub_membership_iam_policy":        {membership},
		"google_gke_hub_membership_iam_binding":       {membership},
		"google_gke_hub_membership_iam_member":        {membership},
		"google_gke_hub_membership_binding":           {membership + "/bindings/b1"},
		"google_gke_hub_membership_rbac_role_binding": {membership + "/rbacrolebindings/r1"},
		"google_gke_hub_scope":                        {scope},
		"google_gke_hub_scope_iam_policy":             {scope},
		"google_gke_hub_scope_iam_binding":            {scope},
		"google_gke_hub_scope_iam_member":             {scope},
		"google_gke_hub_namespace":                    {scope + "/namespaces/ns1"},
		"google_gke_hub_scope_rbac_role_binding":      {scope + "/rbacrolebindings/r1"},
		"google_gke_hub_rollout_sequence":             {glob + "/rolloutSequences/rs1"},
	})
}
