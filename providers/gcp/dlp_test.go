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
	dlp "google.golang.org/api/dlp/v2"
	"google.golang.org/api/option"
)

func TestDlpGeneratorEnumeration(t *testing.T) {
	ctx := context.Background()
	parent := "projects/test-project/locations/us-central1"
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: "/deidentifyTemplates", body: `{"deidentifyTemplates":[{"name":"` + parent + `/deidentifyTemplates/dt1"}]}`},
		{substr: "/discoveryConfigs", body: `{"discoveryConfigs":[{"name":"` + parent + `/discoveryConfigs/dc1"}]}`},
		{substr: "/inspectTemplates", body: `{"inspectTemplates":[{"name":"` + parent + `/inspectTemplates/it1"}]}`},
		{substr: "/jobTriggers", body: `{"jobTriggers":[{"name":"` + parent + `/jobTriggers/jt1"}]}`},
		{substr: "/storedInfoTypes", body: `{"storedInfoTypes":[{"name":"` + parent + `/storedInfoTypes/sit1"}]}`},
	}}
	svc, err := dlp.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	g := &DlpGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project", "region": compute.Region{Name: "us-central1"}})
	if err := g.initResources(ctx, svc); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_data_loss_prevention_deidentify_template": {parent + "/deidentifyTemplates/dt1"},
		"google_data_loss_prevention_discovery_config":    {parent + "/discoveryConfigs/dc1"},
		"google_data_loss_prevention_inspect_template":    {parent + "/inspectTemplates/it1"},
		"google_data_loss_prevention_job_trigger":         {parent + "/jobTriggers/jt1"},
		"google_data_loss_prevention_stored_info_type":    {parent + "/storedInfoTypes/sit1"},
	})
}
