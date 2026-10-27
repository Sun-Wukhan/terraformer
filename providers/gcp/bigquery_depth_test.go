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

	analyticshubv1 "google.golang.org/api/analyticshub/v1"
	bigqueryv2 "google.golang.org/api/bigquery/v2"
	bigqueryconnectionv1 "google.golang.org/api/bigqueryconnection/v1"
	bigquerydatapolicyv1 "google.golang.org/api/bigquerydatapolicy/v1"
	bigquerydatatransferv1 "google.golang.org/api/bigquerydatatransfer/v1"
	bigqueryreservationv1 "google.golang.org/api/bigqueryreservation/v1"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

// TestBigQueryDepthGenerator feeds canned BigQuery and satellite API
// responses through the depth enumeration (initDepthResources) and
// verifies the emitted resource types and IDs: dataset IAM translated
// from the dataset access list, table and routine IAM, routines, row
// access policies, reservations and assignments, reservation groups,
// the BI reservation, connections, transfer configs and enrolled data
// sources, Analytics Hub exchanges/listings/subscriptions/query
// templates, and data policies.
func TestBigQueryDepthGenerator(t *testing.T) {
	proj := "projects/test-project"
	loc := proj + "/locations/us-central1"
	dataset := proj + "/datasets/ds1"
	table := dataset + "/tables/t1"
	exchange := loc + "/dataExchanges/ex1"
	transport := &cannedTransport{routes: []cannedRoute{
		{":getIamPolicy", `{"bindings":[{"role":"roles/bigquery.dataViewer","members":["user:a@example.com"]}]}`},
		// BigQuery core: child collections before their parents,
		// whose URL substrings they contain.
		{"/datasets/ds1/tables/t1/rowAccessPolicies", `{"rowAccessPolicies":[{"rowAccessPolicyReference":{"projectId":"test-project","datasetId":"ds1","tableId":"t1","policyId":"rap1"}}]}`},
		{"/datasets/ds1/tables", `{"tables":[{"id":"test-project:ds1.t1","tableReference":{"projectId":"test-project","datasetId":"ds1","tableId":"t1"}}]}`},
		{"/datasets/ds1/routines", `{"routines":[{"routineReference":{"projectId":"test-project","datasetId":"ds1","routineId":"r1"}}]}`},
		{"/datasets/ds1", `{"datasetReference":{"projectId":"test-project","datasetId":"ds1"},"access":[{"role":"READER","userByEmail":"reader@example.com"},{"role":"OWNER","specialGroup":"projectOwners"},{"view":{"projectId":"test-project","datasetId":"ds1","tableId":"t1"}}]}`},
		{"/bigquery/v2/projects/test-project/datasets", `{"datasets":[{"id":"test-project:ds1","datasetReference":{"projectId":"test-project","datasetId":"ds1"}}]}`},
		// Reservations.
		{"/locations/us-central1/capacityCommitments", `{"capacityCommitments":[{"name":"` + loc + `/capacityCommitments/cc1"}]}`},
		{"/reservations/res1/assignments", `{"assignments":[{"name":"` + loc + `/reservations/res1/assignments/asg1"}]}`},
		{"/locations/us-central1/reservations", `{"reservations":[{"name":"` + loc + `/reservations/res1"}]}`},
		{"/locations/us-central1/reservationGroups", `{"reservationGroups":[{"name":"` + loc + `/reservationGroups/grp1"}]}`},
		{"/locations/us-central1/biReservation", `{"name":"` + loc + `/biReservation","size":"100"}`},
		// Connections, transfer, Analytics Hub, data policies.
		{"/locations/us-central1/connections", `{"connections":[{"name":"` + loc + `/connections/conn1"}]}`},
		{"/locations/us-central1/transferConfigs", `{"transferConfigs":[{"name":"` + loc + `/transferConfigs/tc1"}]}`},
		{"/v1/projects/test-project/dataSources", `{"dataSources":[{"dataSourceId":"carbon-id","name":"projects/test-project/dataSources/carbon-id"}]}`},
		{"/dataExchanges/ex1/listings", `{"listings":[{"name":"` + exchange + `/listings/lst1"}]}`},
		{"/dataExchanges/ex1/queryTemplates", `{"queryTemplates":[{"name":"` + exchange + `/queryTemplates/qt1"}]}`},
		{"/locations/us-central1/dataExchanges", `{"dataExchanges":[{"name":"` + exchange + `"}]}`},
		{"/locations/us-central1/subscriptions", `{"subscriptions":[{"name":"` + loc + `/subscriptions/sub1"}]}`},
		{"/locations/us-central1/dataPolicies", `{"dataPolicies":[{"name":"` + loc + `/dataPolicies/dpol1"}]}`},
	}}

	ctx := context.Background()
	clientOpts := []option.ClientOption{
		option.WithHTTPClient(transport.client()),
		option.WithoutAuthentication(),
	}
	bqService, err := bigqueryv2.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("bigquery.NewService: %v", err)
	}
	reservationService, err := bigqueryreservationv1.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("bigqueryreservation.NewService: %v", err)
	}
	connectionService, err := bigqueryconnectionv1.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("bigqueryconnection.NewService: %v", err)
	}
	transferService, err := bigquerydatatransferv1.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("bigquerydatatransfer.NewService: %v", err)
	}
	hubService, err := analyticshubv1.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("analyticshub.NewService: %v", err)
	}
	datapolicyService, err := bigquerydatapolicyv1.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("bigquerydatapolicy.NewService: %v", err)
	}

	g := &BigQueryGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	g.initDepthResources(ctx, bqService, reservationService, connectionService, transferService, hubService, datapolicyService, transport.client())

	expected := map[string][]string{
		"google_bigquery_dataset_iam_policy":                       {dataset},
		"google_bigquery_dataset_iam_binding":                      {dataset},
		"google_bigquery_dataset_iam_member":                       {dataset},
		"google_bigquery_table_iam_policy":                         {table},
		"google_bigquery_table_iam_binding":                        {table},
		"google_bigquery_table_iam_member":                         {table},
		"google_bigquery_row_access_policy":                        {table + "/rowAccessPolicies/rap1"},
		"google_bigquery_routine":                                  {dataset + "/routines/r1"},
		"google_bigquery_routine_iam_policy":                       {dataset + "/routines/r1"},
		"google_bigquery_routine_iam_binding":                      {dataset + "/routines/r1"},
		"google_bigquery_routine_iam_member":                       {dataset + "/routines/r1"},
		"google_bigquery_capacity_commitment":                      {loc + "/capacityCommitments/cc1"},
		"google_bigquery_reservation":                              {loc + "/reservations/res1"},
		"google_bigquery_reservation_assignment":                   {loc + "/reservations/res1/assignments/asg1"},
		"google_bigquery_reservation_group":                        {loc + "/reservationGroups/grp1"},
		"google_bigquery_bi_reservation":                           {loc + "/biReservation"},
		"google_bigquery_connection":                               {loc + "/connections/conn1"},
		"google_bigquery_connection_iam_policy":                    {loc + "/connections/conn1"},
		"google_bigquery_connection_iam_binding":                   {loc + "/connections/conn1"},
		"google_bigquery_connection_iam_member":                    {loc + "/connections/conn1"},
		"google_bigquery_data_transfer_config":                     {loc + "/transferConfigs/tc1"},
		"google_bigquery_data_transfer_data_source_enrollment":     {proj + "/dataSources/carbon-id"},
		"google_bigquery_analytics_hub_data_exchange":              {exchange},
		"google_bigquery_analytics_hub_data_exchange_iam_policy":   {exchange},
		"google_bigquery_analytics_hub_data_exchange_iam_binding":  {exchange},
		"google_bigquery_analytics_hub_data_exchange_iam_member":   {exchange},
		"google_bigquery_analytics_hub_listing":                    {exchange + "/listings/lst1"},
		"google_bigquery_analytics_hub_listing_iam_policy":         {exchange + "/listings/lst1"},
		"google_bigquery_analytics_hub_listing_iam_binding":        {exchange + "/listings/lst1"},
		"google_bigquery_analytics_hub_listing_iam_member":         {exchange + "/listings/lst1"},
		"google_bigquery_analytics_hub_query_template":             {exchange + "/queryTemplates/qt1"},
		"google_bigquery_analytics_hub_data_exchange_subscription": {loc + "/subscriptions/sub1"},
		"google_bigquery_datapolicy_data_policy":                   {loc + "/dataPolicies/dpol1"},
		"google_bigquery_datapolicy_data_policy_iam_policy":        {loc + "/dataPolicies/dpol1"},
		"google_bigquery_datapolicy_data_policy_iam_binding":       {loc + "/dataPolicies/dpol1"},
		"google_bigquery_datapolicy_data_policy_iam_member":        {loc + "/dataPolicies/dpol1"},
	}

	got := map[string]map[string]bool{}
	counts := map[string]int{}
	for _, resource := range g.Resources {
		if got[resource.InstanceInfo.Type] == nil {
			got[resource.InstanceInfo.Type] = map[string]bool{}
		}
		got[resource.InstanceInfo.Type][resource.InstanceState.ID] = true
		counts[resource.InstanceInfo.Type]++
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
	// The dataset access list carries two IAM grants (READER -> a
	// dataViewer member, OWNER -> a dataOwner member; the authorized
	// view is not IAM), so dataset IAM expands to 1 policy + 2
	// bindings + 2 members even though all share one ID. Every other
	// IAM expansion is 1+1+1 from the canned policy.
	if counts["google_bigquery_dataset_iam_binding"] != 2 {
		t.Errorf("dataset IAM bindings = %d, want 2", counts["google_bigquery_dataset_iam_binding"])
	}
	if counts["google_bigquery_dataset_iam_member"] != 2 {
		t.Errorf("dataset IAM members = %d, want 2", counts["google_bigquery_dataset_iam_member"])
	}
	if len(g.Resources) != 38 {
		t.Errorf("total resources = %d, want 38", len(g.Resources))
	}
}
