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
	"google.golang.org/api/option"
	osconfigv1 "google.golang.org/api/osconfig/v1"
	osconfigv2 "google.golang.org/api/osconfig/v2"
)

// TestOsConfigEnumeration asserts patch deployments, zonal OS policy
// assignments and project v2 policy orchestrators (no CRM service is
// supplied, so folder/org orchestrators are not walked).
func TestOsConfigEnumeration(t *testing.T) {
	ctx := context.Background()
	proj := "projects/test-project"
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: "/patchDeployments", body: `{"patchDeployments":[{"name":"` + proj + `/patchDeployments/pd1"}]}`},
		{substr: "/osPolicyAssignments", body: `{"osPolicyAssignments":[{"name":"` + proj + `/locations/us-central1-a/osPolicyAssignments/opa1"}]}`},
		{substr: "/policyOrchestrators", body: `{"policyOrchestrators":[{"name":"` + proj + `/locations/global/policyOrchestrators/po1"}]}`},
	}}
	v1Service, err := osconfigv1.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("v1 NewService: %v", err)
	}
	v2Service, err := osconfigv2.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("v2 NewService: %v", err)
	}
	g := &OsConfigGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region": compute.Region{Name: "us-central1", Zones: []string{
			"https://www.googleapis.com/compute/v1/projects/test-project/zones/us-central1-a",
		}},
	})
	g.initResources(ctx, v1Service, v2Service, nil)
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_os_config_patch_deployment":       {proj + "/patchDeployments/pd1"},
		"google_os_config_os_policy_assignment":   {proj + "/locations/us-central1-a/osPolicyAssignments/opa1"},
		"google_os_config_v2_policy_orchestrator": {proj + "/locations/global/policyOrchestrators/po1"},
	})
}
