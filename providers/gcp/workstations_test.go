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
	workstationsv1 "google.golang.org/api/workstations/v1"
)

// TestWorkstationsGenerator feeds canned Cloud Workstations
// responses and verifies clusters, configs, workstations and both
// IAM trios.
func TestWorkstationsGenerator(t *testing.T) {
	loc := "projects/test-project/locations/us-central1"
	cluster := loc + "/workstationClusters/wc1"
	config := cluster + "/workstationConfigs/cfg1"
	workstation := config + "/workstations/ws1"
	policy := `{"bindings":[{"role":"roles/workstations.user","members":["user:a@example.com"]}]}`
	transport := &cannedTransport{routes: []cannedRoute{
		{":getIamPolicy", policy},
		{"/workstationConfigs/cfg1/workstations", `{"workstations":[{"name":"` + workstation + `"}]}`},
		{"/workstationClusters/wc1/workstationConfigs", `{"workstationConfigs":[{"name":"` + config + `"}]}`},
		{"/workstationClusters", `{"workstationClusters":[{"name":"` + cluster + `"}]}`},
	}}
	ctx := context.Background()
	clientOpts := []option.ClientOption{option.WithHTTPClient(transport.client()), option.WithoutAuthentication()}
	wsService, err := workstationsv1.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("workstations.NewService: %v", err)
	}
	g := &WorkstationsGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project", "region": compute.Region{Name: "us-central1"}})
	if err := g.initResources(ctx, wsService); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_workstations_workstation_cluster":            {cluster},
		"google_workstations_workstation_config":             {config},
		"google_workstations_workstation_config_iam_policy":  {config},
		"google_workstations_workstation_config_iam_binding": {config},
		"google_workstations_workstation_config_iam_member":  {config},
		"google_workstations_workstation":                    {workstation},
		"google_workstations_workstation_iam_policy":         {workstation},
		"google_workstations_workstation_iam_binding":        {workstation},
		"google_workstations_workstation_iam_member":         {workstation},
	})
}
