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

	"google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
)

func TestWorkloadIdentityEnumeration(t *testing.T) {
	ctx := context.Background()
	poolName := "projects/test-project/locations/global/workloadIdentityPools/pool1"
	namespaceName := poolName + "/namespaces/ns1"
	clientName := "projects/test-project/locations/global/oauthClients/client1"
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: ":getIamPolicy", body: `{"bindings":[{"role":"roles/iam.workloadIdentityPoolAdmin","members":["user:jane@example.com"]}]}`},
		{substr: "/namespaces/ns1/managedIdentities", body: `{"workloadIdentityPoolManagedIdentities":[{"name":"` + namespaceName + `/managedIdentities/mi1"}]}`},
		{substr: "/namespaces", body: `{"workloadIdentityPoolNamespaces":[{"name":"` + namespaceName + `"}]}`},
		{substr: "/providers", body: `{"workloadIdentityPoolProviders":[{"name":"` + poolName + `/providers/aws1"}]}`},
		{substr: "/workloadIdentityPools", body: `{"workloadIdentityPools":[{"name":"` + poolName + `"}]}`},
		{substr: "/oauthClients/client1/credentials", body: `{"oauthClientCredentials":[{"name":"` + clientName + `/credentials/cred1"}]}`},
		{substr: "/oauthClients", body: `{"oauthClients":[{"name":"` + clientName + `"}]}`},
	}}
	service, err := iam.NewService(ctx,
		option.WithHTTPClient(transport.client()),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	g := &IamGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project"})
	if err := g.initWorkloadIdentityResources(ctx, service); err != nil {
		t.Fatalf("initWorkloadIdentityResources: %v", err)
	}

	got := map[string][]string{}
	for _, r := range g.Resources {
		got[r.InstanceInfo.Type] = append(got[r.InstanceInfo.Type], r.InstanceState.ID)
	}

	wantIDs := map[string]string{
		"google_iam_workload_identity_pool":                  poolName,
		"google_iam_workload_identity_pool_iam_policy":       poolName,
		"google_iam_workload_identity_pool_iam_binding":      poolName,
		"google_iam_workload_identity_pool_iam_member":       poolName,
		"google_iam_workload_identity_pool_namespace":        namespaceName,
		"google_iam_workload_identity_pool_managed_identity": namespaceName + "/managedIdentities/mi1",
		"google_iam_workload_identity_pool_provider":         poolName + "/providers/aws1",
		"google_iam_oauth_client":                            clientName,
		"google_iam_oauth_client_credential":                 clientName + "/credentials/cred1",
	}
	for resourceType, wantID := range wantIDs {
		ids := got[resourceType]
		if len(ids) != 1 || ids[0] != wantID {
			t.Errorf("%s: got IDs %v, want [%s]", resourceType, ids, wantID)
		}
	}
}
