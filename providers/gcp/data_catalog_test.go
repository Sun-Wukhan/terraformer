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
	datacatalogv1 "google.golang.org/api/datacatalog/v1"
	"google.golang.org/api/option"
)

// TestDataCatalogGenerator feeds canned Data Catalog responses and
// verifies entry groups, taxonomies and policy tags expand their IAM
// trios, and entries and tags are emitted with full names as IDs.
func TestDataCatalogGenerator(t *testing.T) {
	loc := "projects/test-project/locations/us-central1"
	group := loc + "/entryGroups/@bigquery"
	entry := group + "/entries/entry1"
	taxonomy := loc + "/taxonomies/tax1"
	policyTag := taxonomy + "/policyTags/pt1"
	transport := &cannedTransport{routes: []cannedRoute{
		{":getIamPolicy", `{"bindings":[{"role":"roles/datacatalog.viewer","members":["user:a@example.com"]}]}`},
		{"/entryGroups/@bigquery/entries/entry1/tags", `{"tags":[{"name":"` + entry + `/tags/tag2"}]}`},
		{"/entryGroups/@bigquery/entries", `{"entries":[{"name":"` + entry + `"}]}`},
		{"/entryGroups/@bigquery/tags", `{"tags":[{"name":"` + group + `/tags/tag1"}]}`},
		{"/locations/us-central1/entryGroups", `{"entryGroups":[{"name":"` + group + `"}]}`},
		{"/taxonomies/tax1/policyTags", `{"policyTags":[{"name":"` + policyTag + `"}]}`},
		{"/locations/us-central1/taxonomies", `{"taxonomies":[{"name":"` + taxonomy + `"}]}`},
	}}
	ctx := context.Background()
	svc, err := datacatalogv1.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("datacatalog.NewService: %v", err)
	}
	g := &DataCatalogGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	if err := g.initResources(ctx, svc); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_data_catalog_entry_group":             {group},
		"google_data_catalog_entry_group_iam_policy":  {group},
		"google_data_catalog_entry_group_iam_binding": {group},
		"google_data_catalog_entry_group_iam_member":  {group},
		"google_data_catalog_entry":                   {entry},
		"google_data_catalog_tag":                     {group + "/tags/tag1", entry + "/tags/tag2"},
		"google_data_catalog_taxonomy":                {taxonomy},
		"google_data_catalog_taxonomy_iam_policy":     {taxonomy},
		"google_data_catalog_taxonomy_iam_binding":    {taxonomy},
		"google_data_catalog_taxonomy_iam_member":     {taxonomy},
		"google_data_catalog_policy_tag":              {policyTag},
		"google_data_catalog_policy_tag_iam_policy":   {policyTag},
		"google_data_catalog_policy_tag_iam_binding":  {policyTag},
		"google_data_catalog_policy_tag_iam_member":   {policyTag},
	})
}
