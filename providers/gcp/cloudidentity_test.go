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

	cloudidentity "google.golang.org/api/cloudidentity/v1"
	"google.golang.org/api/option"
)

func TestCloudIdentityEnumeration(t *testing.T) {
	ctx := context.Background()
	group := "groups/abc123"
	policy := "customers/C0123/policies/pol1"
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: "/memberships", body: `{"memberships":[{"name":"` + group + `/memberships/mem1"}]}`},
		{substr: "/groups", body: `{"groups":[{"name":"` + group + `"}]}`},
		{substr: "/policies", body: `{"policies":[{"name":"` + policy + `"}]}`},
	}}
	service, err := cloudidentity.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	g := &CloudIdentityGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project"})
	g.initResources(ctx, service)
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_cloud_identity_group":            {group},
		"google_cloud_identity_group_membership": {group + "/memberships/mem1"},
		"google_cloud_identity_policy":           {policy},
	})
}
