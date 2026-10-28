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

	biglakev1 "google.golang.org/api/biglake/v1"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

// TestBiglakeGenerator feeds canned responses for all three
// BigLake catalog flavours and verifies: a classic catalog with
// database and table; a hive catalog with database, table and the
// three IAM trios; and an iceberg catalog (extensions-prefixed ID)
// with namespace, table and the three IAM trios, with the Iceberg
// list replies in their spec shapes.
func TestBiglakeGenerator(t *testing.T) {
	proj := "projects/test-project"
	classic := proj + "/locations/US/catalogs/cat1"
	hiveCatalog := proj + "/catalogs/hive1"
	hiveDatabase := hiveCatalog + "/databases/hdb1"
	hiveTable := hiveDatabase + "/tables/ht1"
	iceCatalog := proj + "/catalogs/ice1"
	iceNamespace := iceCatalog + "/namespaces/ns1"
	iceTable := iceNamespace + "/tables/it1"
	transport := &cannedTransport{routes: []cannedRoute{
		{":getIamPolicy", `{"bindings":[{"role":"roles/biglake.admin","members":["user:jane@example.com"]}]}`},
		{"/locations/US/catalogs/cat1/databases/hdb-classic/tables", `{"tables":[{"name":"` + classic + `/databases/hdb-classic/tables/t1"}]}`},
		{"/locations/US/catalogs/cat1/databases", `{"databases":[{"name":"` + classic + `/databases/hdb-classic"}]}`},
		{"/locations/US/catalogs", `{"catalogs":[{"name":"` + classic + `"}]}`},
		{"/hive/v1/projects/test-project/catalogs/hive1/databases/hdb1/tables", `{"tables":[{"name":"` + hiveTable + `"}]}`},
		{"/hive/v1/projects/test-project/catalogs/hive1/databases", `{"databases":[{"name":"` + hiveDatabase + `"}]}`},
		{"/hive/v1/projects/test-project/catalogs", `{"catalogs":[{"name":"` + hiveCatalog + `"}]}`},
		{"/namespaces/ns1/tables", `{"identifiers":[{"namespace":["ns1"],"name":"it1"}]}`},
		{"/restcatalog/v1/projects/test-project/catalogs/ice1/namespaces", `{"namespaces":[["ns1"]]}`},
		{"/restcatalog/extensions/projects/test-project/catalogs", `{"catalogs":[{"name":"` + iceCatalog + `"}]}`},
	}}
	ctx := context.Background()
	service, err := biglakev1.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("biglake NewService: %v", err)
	}
	g := &BiglakeGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	g.initResources(ctx, service, transport.client())

	expected := map[string][]string{
		"google_biglake_catalog":       {classic},
		"google_biglake_database":      {classic + "/databases/hdb-classic"},
		"google_biglake_table":         {classic + "/databases/hdb-classic/tables/t1"},
		"google_biglake_hive_catalog":  {hiveCatalog},
		"google_biglake_hive_database": {hiveDatabase},
		"google_biglake_hive_table":    {hiveTable},
		"google_biglake_iceberg_catalog": {
			"iceberg/v1/restcatalog/extensions/" + iceCatalog,
		},
		"google_biglake_iceberg_namespace": {iceNamespace},
		"google_biglake_iceberg_table":     {iceTable},
	}
	for _, parent := range []struct {
		id     string
		prefix string
	}{
		{hiveCatalog, "google_biglake_hive_catalog"},
		{hiveDatabase, "google_biglake_hive_database"},
		{hiveTable, "google_biglake_hive_table"},
		{"iceberg/v1/restcatalog/extensions/" + iceCatalog, "google_biglake_iceberg_catalog"},
		{iceNamespace, "google_biglake_iceberg_namespace"},
		{iceTable, "google_biglake_iceberg_table"},
	} {
		expected[parent.prefix+"_iam_policy"] = []string{parent.id}
		expected[parent.prefix+"_iam_binding"] = []string{parent.id}
		expected[parent.prefix+"_iam_member"] = []string{parent.id}
	}
	assertResourceIDs(t, g.Resources, expected)
}
