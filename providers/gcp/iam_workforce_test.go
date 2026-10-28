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

	crmv1 "google.golang.org/api/cloudresourcemanager/v1"
	iamv1 "google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
)

// TestWorkforcePoolEnumeration feeds canned ancestry + workforce pool
// responses and asserts pools, the IAM trio, providers and keys.
func TestWorkforcePoolEnumeration(t *testing.T) {
	ctx := context.Background()
	poolName := "organizations/123/locations/global/workforcePools/pool1"
	providerName := poolName + "/providers/oidc1"
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: ":getAncestry", body: `{"ancestor":[{"resourceId":{"type":"organization","id":"123"}}]}`},
		{substr: ":getIamPolicy", body: `{"bindings":[{"role":"roles/iam.workforcePoolAdmin","members":["user:jane@example.com"]}]}`},
		{substr: "/providers/oidc1/keys", body: `{"workforcePoolProviderKeys":[{"name":"` + providerName + `/keys/key1"}]}`},
		{substr: "/providers", body: `{"workforcePoolProviders":[{"name":"` + providerName + `"}]}`},
		{substr: "/workforcePools", body: `{"workforcePools":[{"name":"` + poolName + `"}]}`},
	}}
	opts := []option.ClientOption{option.WithHTTPClient(transport.client()), option.WithoutAuthentication()}
	iamSvc, err := iamv1.NewService(ctx, opts...)
	if err != nil {
		t.Fatal(err)
	}
	crmSvc, err := crmv1.NewService(ctx, opts...)
	if err != nil {
		t.Fatal(err)
	}
	g := &IamGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project"})
	if err := g.initWorkforceResources(ctx, iamSvc, crmSvc); err != nil {
		t.Fatalf("initWorkforceResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_iam_workforce_pool":              {poolName},
		"google_iam_workforce_pool_iam_policy":   {poolName},
		"google_iam_workforce_pool_iam_binding":  {poolName},
		"google_iam_workforce_pool_iam_member":   {poolName},
		"google_iam_workforce_pool_provider":     {providerName},
		"google_iam_workforce_pool_provider_key": {providerName + "/keys/key1"},
	})
}
