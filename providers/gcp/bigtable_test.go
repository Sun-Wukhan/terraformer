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

	bigtableadminv2 "google.golang.org/api/bigtableadmin/v2"
	"google.golang.org/api/option"
)

// TestBigtableGenerator feeds canned Bigtable Admin responses and
// verifies instances and tables expand their IAM trios, and app
// profiles, authorized views, logical views and materialized views
// are emitted with their full resource names as IDs.
func TestBigtableGenerator(t *testing.T) {
	proj := "projects/test-project"
	inst := proj + "/instances/inst1"
	table := inst + "/tables/t1"
	transport := &cannedTransport{routes: []cannedRoute{
		{":getIamPolicy", `{"bindings":[{"role":"roles/bigtable.user","members":["user:a@example.com"]}]}`},
		{"/instances/inst1/tables/t1/authorizedViews", `{"authorizedViews":[{"name":"` + table + `/authorizedViews/av1"}]}`},
		{"/instances/inst1/tables", `{"tables":[{"name":"` + table + `"}]}`},
		{"/instances/inst1/appProfiles", `{"appProfiles":[{"name":"` + inst + `/appProfiles/ap1"}]}`},
		{"/instances/inst1/logicalViews", `{"logicalViews":[{"name":"` + inst + `/logicalViews/lv1"}]}`},
		{"/instances/inst1/materializedViews", `{"materializedViews":[{"name":"` + inst + `/materializedViews/mv1"}]}`},
		{"/v2/projects/test-project/instances", `{"instances":[{"name":"` + inst + `"}]}`},
	}}
	ctx := context.Background()
	svc, err := bigtableadminv2.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("bigtableadmin.NewService: %v", err)
	}
	g := &BigtableGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project"})
	if err := g.initResources(ctx, svc); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_bigtable_instance":             {inst},
		"google_bigtable_instance_iam_policy":  {inst},
		"google_bigtable_instance_iam_binding": {inst},
		"google_bigtable_instance_iam_member":  {inst},
		"google_bigtable_app_profile":          {inst + "/appProfiles/ap1"},
		"google_bigtable_logical_view":         {inst + "/logicalViews/lv1"},
		"google_bigtable_materialized_view":    {inst + "/materializedViews/mv1"},
		"google_bigtable_table":                {table},
		"google_bigtable_table_iam_policy":     {table},
		"google_bigtable_table_iam_binding":    {table},
		"google_bigtable_table_iam_member":     {table},
		"google_bigtable_authorized_view":      {table + "/authorizedViews/av1"},
	})
}
