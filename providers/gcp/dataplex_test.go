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
	"sort"
	"testing"

	"google.golang.org/api/compute/v1"
	dataplexv1 "google.golang.org/api/dataplex/v1"
	"google.golang.org/api/option"
)

// TestDataplexGenerator feeds canned Dataplex responses — the
// generated client for the lake tree, data scans and catalog types,
// raw REST for glossaries, data products and metadata feeds — through
// the real enumeration code and verifies the emitted resource types
// and IDs, including the IAM expansions.
func TestDataplexGenerator(t *testing.T) {
	loc := "projects/test-project/locations/us-central1"
	lake := loc + "/lakes/lake1"
	zone := lake + "/zones/zone1"
	glossary := loc + "/glossaries/g1"
	product := loc + "/dataProducts/dp1"
	transport := &cannedTransport{routes: []cannedRoute{
		{":getIamPolicy", `{"bindings":[{"role":"roles/dataplex.viewer","members":["user:a@example.com","user:b@example.com"]}]}`},
		// Child collections before their parents: their URLs contain
		// the parent list substrings.
		{"/lakes/lake1/zones/zone1/assets", `{"assets":[{"name":"` + zone + `/assets/asset1"}]}`},
		{"/lakes/lake1/zones", `{"zones":[{"name":"` + zone + `"}]}`},
		{"/lakes/lake1/tasks", `{"tasks":[{"name":"` + lake + `/tasks/task1"}]}`},
		{"/lakes", `{"lakes":[{"name":"` + lake + `"}]}`},
		{"/dataScans", `{"dataScans":[{"name":"` + loc + `/dataScans/scan1"}]}`},
		{"/aspectTypes", `{"aspectTypes":[{"name":"` + loc + `/aspectTypes/at1"}]}`},
		{"/entryTypes", `{"entryTypes":[{"name":"` + loc + `/entryTypes/et1"}]}`},
		{"/entryGroups/eg1/entries", `{"entries":[{"name":"` + loc + `/entryGroups/eg1/entries/entry1"}]}`},
		{"/entryGroups", `{"entryGroups":[{"name":"` + loc + `/entryGroups/eg1"}]}`},
		// Raw REST sections.
		{"/glossaries/g1/categories", `{"categories":[{"name":"` + glossary + `/categories/cat1"}]}`},
		{"/glossaries/g1/terms", `{"terms":[{"name":"` + glossary + `/terms/term1"}]}`},
		{"/glossaries", `{"glossaries":[{"name":"` + glossary + `"}]}`},
		{"/dataProducts/dp1/dataAssets", `{"dataAssets":[{"name":"` + product + `/dataAssets/dpa1"}]}`},
		{"/dataProducts", `{"dataProducts":[{"name":"` + product + `"}]}`},
		{"/metadataFeeds", `{"metadataFeeds":[{"name":"` + loc + `/metadataFeeds/feed1"}]}`},
	}}

	ctx := context.Background()
	clientOpts := []option.ClientOption{
		option.WithHTTPClient(transport.client()),
		option.WithoutAuthentication(),
	}
	dataplexService, err := dataplexv1.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("dataplexv1.NewService: %v", err)
	}

	g := &DataplexGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	if err := g.initResources(ctx, dataplexService, transport.client()); err != nil {
		t.Fatalf("initResources: %v", err)
	}

	expected := map[string][]string{
		"google_dataplex_lake":                     {lake},
		"google_dataplex_lake_iam_policy":          {lake},
		"google_dataplex_lake_iam_binding":         {lake},
		"google_dataplex_lake_iam_member":          {lake},
		"google_dataplex_zone":                     {zone},
		"google_dataplex_zone_iam_policy":          {zone},
		"google_dataplex_zone_iam_binding":         {zone},
		"google_dataplex_zone_iam_member":          {zone},
		"google_dataplex_asset":                    {zone + "/assets/asset1"},
		"google_dataplex_asset_iam_policy":         {zone + "/assets/asset1"},
		"google_dataplex_asset_iam_binding":        {zone + "/assets/asset1"},
		"google_dataplex_asset_iam_member":         {zone + "/assets/asset1"},
		"google_dataplex_task":                     {lake + "/tasks/task1"},
		"google_dataplex_task_iam_policy":          {lake + "/tasks/task1"},
		"google_dataplex_task_iam_binding":         {lake + "/tasks/task1"},
		"google_dataplex_task_iam_member":          {lake + "/tasks/task1"},
		"google_dataplex_datascan":                 {loc + "/dataScans/scan1"},
		"google_dataplex_datascan_iam_policy":      {loc + "/dataScans/scan1"},
		"google_dataplex_datascan_iam_binding":     {loc + "/dataScans/scan1"},
		"google_dataplex_datascan_iam_member":      {loc + "/dataScans/scan1"},
		"google_dataplex_aspect_type":              {loc + "/aspectTypes/at1"},
		"google_dataplex_aspect_type_iam_policy":   {loc + "/aspectTypes/at1"},
		"google_dataplex_aspect_type_iam_binding":  {loc + "/aspectTypes/at1"},
		"google_dataplex_aspect_type_iam_member":   {loc + "/aspectTypes/at1"},
		"google_dataplex_entry_type":               {loc + "/entryTypes/et1"},
		"google_dataplex_entry_type_iam_policy":    {loc + "/entryTypes/et1"},
		"google_dataplex_entry_type_iam_binding":   {loc + "/entryTypes/et1"},
		"google_dataplex_entry_type_iam_member":    {loc + "/entryTypes/et1"},
		"google_dataplex_entry_group":              {loc + "/entryGroups/eg1"},
		"google_dataplex_entry_group_iam_policy":   {loc + "/entryGroups/eg1"},
		"google_dataplex_entry_group_iam_binding":  {loc + "/entryGroups/eg1"},
		"google_dataplex_entry_group_iam_member":   {loc + "/entryGroups/eg1"},
		"google_dataplex_entry":                    {loc + "/entryGroups/eg1/entries/entry1"},
		"google_dataplex_glossary":                 {glossary},
		"google_dataplex_glossary_iam_policy":      {glossary},
		"google_dataplex_glossary_iam_binding":     {glossary},
		"google_dataplex_glossary_iam_member":      {glossary},
		"google_dataplex_glossary_category":        {glossary + "/categories/cat1"},
		"google_dataplex_glossary_term":            {glossary + "/terms/term1"},
		"google_dataplex_data_product":             {product},
		"google_dataplex_data_product_iam_policy":  {product},
		"google_dataplex_data_product_iam_binding": {product},
		"google_dataplex_data_product_iam_member":  {product},
		"google_dataplex_data_product_data_asset":  {product + "/dataAssets/dpa1"},
		"google_dataplex_metadata_feed":            {loc + "/metadataFeeds/feed1"},
	}

	got := map[string]map[string]bool{}
	for _, resource := range g.Resources {
		if got[resource.InstanceInfo.Type] == nil {
			got[resource.InstanceInfo.Type] = map[string]bool{}
		}
		got[resource.InstanceInfo.Type][resource.InstanceState.ID] = true
	}
	if len(got) != len(expected) {
		t.Errorf("resource types = %d, want %d", len(got), len(expected))
	}
	for resourceType, ids := range expected {
		gotIDs := []string{}
		for id := range got[resourceType] {
			gotIDs = append(gotIDs, id)
		}
		sort.Strings(gotIDs)
		wantIDs := append([]string{}, ids...)
		sort.Strings(wantIDs)
		if len(gotIDs) != len(wantIDs) {
			t.Errorf("%s IDs = %v, want %v", resourceType, gotIDs, wantIDs)
			continue
		}
		for i := range wantIDs {
			if gotIDs[i] != wantIDs[i] {
				t.Errorf("%s IDs = %v, want %v", resourceType, gotIDs, wantIDs)
				break
			}
		}
	}
}
