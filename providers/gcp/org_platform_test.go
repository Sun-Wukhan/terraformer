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

	apikeysv2 "google.golang.org/api/apikeys/v2"
	assuredworkloadsv1 "google.golang.org/api/assuredworkloads/v1"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
	websecurityscannerv1 "google.golang.org/api/websecurityscanner/v1"
)

// TestOrgPlatformGenerator asserts IDs for API keys, scan configs,
// assured workloads (org ancestry), public CA keys, quota preference
// + adjuster settings, license configurations and the Secure Source
// Manager tree with its IAM trios.
func TestOrgPlatformGenerator(t *testing.T) {
	ctx := context.Background()
	p := "projects/test-project"
	inst := p + "/locations/us-central1/instances/inst1"
	repo := inst + "/repositories/repo1"
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: ":getAncestry", body: `{"ancestor":[{"resourceId":{"type":"organization","id":"123"}}]}`},
		{substr: ":getIamPolicy", body: `{"bindings":[{"role":"roles/securesourcemanager.repoAdmin","members":["user:jane@example.com"]}]}`},
		{substr: "/keys", body: `{"keys":[{"name":"` + p + `/locations/global/keys/key1"}]}`},
		{substr: "/scanConfigs", body: `{"scanConfigs":[{"name":"` + p + `/scanConfigs/sc1"}]}`},
		{substr: "/workloads", body: `{"workloads":[{"name":"organizations/123/locations/us-central1/workloads/w1"}]}`},
		{substr: "externalAccountKeys", body: `{"externalAccountKeys":[{"name":"` + p + `/locations/global/externalAccountKeys/eak1"}]}`},
		{substr: "quotaPreferences", body: `{"quotaPreferences":[{"name":"` + p + `/locations/global/quotaPreferences/qp1"}]}`},
		{substr: "quotaAdjusterSettings", body: `{"name":"` + p + `/locations/global/quotaAdjusterSettings","enablement":"ENABLED"}`},
		{substr: "configurations", body: `{"configurations":[{"name":"` + p + `/locations/global/configurations/cfg1"}]}`},
		{substr: "/branchRules", body: `{"branchRules":[{"name":"` + repo + `/branchRules/br1"}]}`},
		{substr: "/hooks", body: `{"hooks":[{"name":"` + repo + `/hooks/h1"}]}`},
		{substr: "/repositories", body: `{"repositories":[{"name":"` + repo + `"}]}`},
		{substr: "/instances", body: `{"instances":[{"name":"` + inst + `"}]}`},
		{substr: "/locations", body: `{"locations":[{"name":"` + p + `/locations/us-central1"}]}`},
	}}
	opts := []option.ClientOption{option.WithHTTPClient(transport.client()), option.WithoutAuthentication()}
	akSvc, err := apikeysv2.NewService(ctx, opts...)
	if err != nil {
		t.Fatal(err)
	}
	scanSvc, err := websecurityscannerv1.NewService(ctx, opts...)
	if err != nil {
		t.Fatal(err)
	}
	awSvc, err := assuredworkloadsv1.NewService(ctx, opts...)
	if err != nil {
		t.Fatal(err)
	}
	crmSvc := crmTestService(t, transport)
	g := &OrgPlatformGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	if err := g.initResources(ctx, akSvc, scanSvc, awSvc, crmSvc, transport.client()); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_apikeys_key":                                  {p + "/locations/global/keys/key1"},
		"google_security_scanner_scan_config":                 {p + "/scanConfigs/sc1"},
		"google_assured_workloads_workload":                   {"organizations/123/locations/us-central1/workloads/w1"},
		"google_public_ca_external_account_key":               {p + "/locations/global/externalAccountKeys/eak1"},
		"google_cloud_quotas_quota_preference":                {p + "/locations/global/quotaPreferences/qp1"},
		"google_cloud_quotas_quota_adjuster_settings":         {p + "/locations/global/quotaAdjusterSettings"},
		"google_license_manager_configuration":                {p + "/locations/global/configurations/cfg1"},
		"google_secure_source_manager_instance":               {inst},
		"google_secure_source_manager_instance_iam_policy":    {inst},
		"google_secure_source_manager_instance_iam_binding":   {inst},
		"google_secure_source_manager_instance_iam_member":    {inst},
		"google_secure_source_manager_repository":             {repo},
		"google_secure_source_manager_repository_iam_policy":  {repo},
		"google_secure_source_manager_repository_iam_binding": {repo},
		"google_secure_source_manager_repository_iam_member":  {repo},
		"google_secure_source_manager_branch_rule":            {repo + "/branchRules/br1"},
		"google_secure_source_manager_hook":                   {repo + "/hooks/h1"},
	})
}
