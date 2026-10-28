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
	"log"
	"net/http"
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	analyticshubv1 "google.golang.org/api/analyticshub/v1"
	bigqueryv2 "google.golang.org/api/bigquery/v2"
	bigqueryconnectionv1 "google.golang.org/api/bigqueryconnection/v1"
	bigquerydatapolicyv1 "google.golang.org/api/bigquerydatapolicy/v1"
	bigquerydatatransferv1 "google.golang.org/api/bigquerydatatransfer/v1"
	bigqueryreservationv1 "google.golang.org/api/bigqueryreservation/v1"
	"google.golang.org/api/compute/v1"
)

// This file extends BigQueryGenerator (bigquery.go, which enumerates
// datasets and tables) with the rest of the BigQuery surface, in the
// same log-and-continue style: a failure in any one collection is
// logged and never aborts the dataset/table enumeration.
//
//   - Per dataset: dataset IAM derived from the dataset's access list
//     (exactly how the google provider implements
//     google_bigquery_dataset_iam_*: it translates access[] entries,
//     there is no datasets.getIamPolicy method in the BigQuery API),
//     table IAM and routine IAM from their getIamPolicy endpoints,
//     routines, and row access policies per table.
//   - Per configured location: capacity commitments, reservations and
//     their assignments, reservation groups (raw REST; the pinned
//     client has no reservation group service), the BI reservation
//     singleton (emitted only when it has capacity, mirroring the
//     provider's read), connections with IAM, data transfer configs,
//     enrolled data transfer data sources, Analytics Hub data
//     exchanges with their listings and IAM, subscriptions, query
//     templates (raw REST; the pinned client has no query template
//     service), and data policies with IAM.
//
// Deliberately not generated:
//   - google_bigquery_dataset_access: access entries have no unique
//     ID (every entry's provider ID is the dataset's own ID) and the
//     provider resource does not support import, so entries cannot be
//     enumerated as separate resources. They remain part of the
//     google_bigquery_dataset resource itself.
//   - google_bigquery_job: jobs are transient execution records, not
//     durable configuration; jobs.list returns history.
//   - google_bigquery_analytics_hub_listing_subscription: the same
//     API object as google_bigquery_analytics_hub_data_exchange_subscription
//     (projects/{project}/locations/{location}/subscriptions/{id});
//     the two provider resources are alternate creation paths for one
//     subscription and are indistinguishable when enumerated.
//     Subscriptions are emitted once, as data exchange subscriptions.
//   - google_bigquery_datapolicyv2_data_policy and
//     google_bigquery_datapolicyv2_data_policy_iam: the v2 provider resources address the same data policy objects
//     (same resource names and IDs) as the v1
//     google_bigquery_datapolicy_data_policy resources emitted here,
//     through the v2 API surface; emitting both would double-manage
//     every data policy.

const (
	bigQueryReservationBaseURL = "https://bigqueryreservation.googleapis.com/v1/"
	analyticsHubBaseURL        = "https://analyticshub.googleapis.com/v1/"
)

func (g *BigQueryGenerator) newDepthResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
	if attrs == nil {
		attrs = map[string]string{}
	}
	parts := strings.Split(id, "/")
	return terraformutils.NewResource(
		id,
		parts[len(parts)-1],
		resourceType,
		g.ProviderName,
		attrs,
		bigQueryAllowEmptyValues,
		map[string]interface{}{},
	)
}

// initDepthResources enumerates the BigQuery depth collections. Any
// service may be nil when its construction failed in InitResources;
// that section is then skipped.
func (g *BigQueryGenerator) initDepthResources(ctx context.Context, bqService *bigqueryv2.Service, reservationService *bigqueryreservationv1.Service, connectionService *bigqueryconnectionv1.Service, transferService *bigquerydatatransferv1.Service, hubService *analyticshubv1.Service, datapolicyService *bigquerydatapolicyv1.Service, restClient *http.Client) {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region)
	locationParent := "projects/" + project + "/locations/" + region.Name
	locationAttrs := map[string]string{"project": project, "location": region.Name}

	if bqService != nil {
		g.loadDatasetDepth(ctx, bqService, project)
	}
	if reservationService != nil {
		g.loadReservations(ctx, reservationService, restClient, locationParent, locationAttrs)
	}
	if connectionService != nil {
		g.loadConnections(ctx, connectionService, locationParent, locationAttrs)
	}
	if transferService != nil {
		g.loadTransferConfigs(ctx, transferService, project, locationParent, locationAttrs)
	}
	if hubService != nil {
		g.loadAnalyticsHub(ctx, hubService, restClient, locationParent, locationAttrs)
	}
	if datapolicyService != nil {
		g.loadDataPolicies(ctx, datapolicyService, locationParent, locationAttrs)
	}
}

func (g *BigQueryGenerator) loadDatasetDepth(ctx context.Context, bqService *bigqueryv2.Service, project string) {
	err := bqService.Datasets.List(project).Pages(ctx, func(page *bigqueryv2.DatasetList) error {
		for _, listed := range page.Datasets {
			datasetID := listed.DatasetReference.DatasetId
			datasetName := "projects/" + project + "/datasets/" + datasetID
			dataset, err := bqService.Datasets.Get(project, datasetID).Context(ctx).Do()
			if err != nil {
				log.Println(err)
				continue
			}
			if bindings := datasetAccessBindings(dataset.Access); len(bindings) > 0 {
				g.Resources = append(g.Resources, expandIamPolicy(
					bindings,
					datasetName,
					"google_bigquery_dataset_iam_policy",
					"google_bigquery_dataset_iam_binding",
					"google_bigquery_dataset_iam_member",
					map[string]string{"project": project, "dataset_id": datasetID},
					g.newDepthResource,
				)...)
			}
			g.loadTableDepth(ctx, bqService, project, datasetID)
			g.loadRoutines(ctx, bqService, project, datasetID)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

// datasetAccessBindings translates a dataset's access list into IAM
// bindings, mirroring the google provider's accessToPolicy translation
// (including the legacy role mapping). Entries without a role —
// authorized views, datasets and routines — are not IAM grants and
// are skipped; an entry whose member cannot be identified is skipped
// as well (the provider errors on it, but enumeration must not abort
// a whole import over one entry).
func datasetAccessBindings(access []*bigqueryv2.DatasetAccess) []iamBinding {
	primitiveRoles := map[string]string{
		"OWNER":  "roles/bigquery.dataOwner",
		"WRITER": "roles/bigquery.dataEditor",
		"READER": "roles/bigquery.dataViewer",
	}
	membersByRole := map[string][]string{}
	var roles []string
	for _, entry := range access {
		if entry.Role == "" {
			continue
		}
		var member string
		switch {
		case entry.UserByEmail != "":
			if strings.Contains(entry.UserByEmail, "gserviceaccount") {
				member = "serviceAccount:" + entry.UserByEmail
			} else {
				member = "user:" + entry.UserByEmail
			}
		case entry.GroupByEmail != "":
			member = "group:" + entry.GroupByEmail
		case entry.Domain != "":
			member = "domain:" + entry.Domain
		case entry.SpecialGroup != "":
			member = entry.SpecialGroup
		case entry.IamMember != "":
			member = "iamMember:" + entry.IamMember
		default:
			continue
		}
		role := entry.Role
		if mapped, ok := primitiveRoles[role]; ok {
			role = mapped
		}
		if _, ok := membersByRole[role]; !ok {
			roles = append(roles, role)
		}
		membersByRole[role] = append(membersByRole[role], member)
	}
	bindings := []iamBinding{}
	for _, role := range roles {
		bindings = append(bindings, iamBinding{Role: role, Members: membersByRole[role]})
	}
	return bindings
}

func (g *BigQueryGenerator) loadTableDepth(ctx context.Context, bqService *bigqueryv2.Service, project, datasetID string) {
	err := bqService.Tables.List(project, datasetID).Pages(ctx, func(page *bigqueryv2.TableList) error {
		for _, table := range page.Tables {
			tableID := table.TableReference.TableId
			tableName := "projects/" + project + "/datasets/" + datasetID + "/tables/" + tableID
			tableAttrs := map[string]string{"project": project, "dataset_id": datasetID, "table_id": tableID}
			policy, err := bqService.Tables.GetIamPolicy(tableName, &bigqueryv2.GetIamPolicyRequest{}).Context(ctx).Do()
			if err != nil {
				log.Println(err)
			} else {
				bindings := []iamBinding{}
				for _, binding := range policy.Bindings {
					bindings = append(bindings, iamBinding{Role: binding.Role, Members: binding.Members})
				}
				g.Resources = append(g.Resources, expandIamPolicy(
					bindings,
					tableName,
					"google_bigquery_table_iam_policy",
					"google_bigquery_table_iam_binding",
					"google_bigquery_table_iam_member",
					tableAttrs,
					g.newDepthResource,
				)...)
			}
			g.loadRowAccessPolicies(ctx, bqService, project, datasetID, tableID, tableAttrs)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (g *BigQueryGenerator) loadRowAccessPolicies(ctx context.Context, bqService *bigqueryv2.Service, project, datasetID, tableID string, tableAttrs map[string]string) {
	err := bqService.RowAccessPolicies.List(project, datasetID, tableID).Pages(ctx, func(page *bigqueryv2.ListRowAccessPoliciesResponse) error {
		for _, policy := range page.RowAccessPolicies {
			ref := policy.RowAccessPolicyReference
			id := "projects/" + ref.ProjectId + "/datasets/" + ref.DatasetId + "/tables/" + ref.TableId + "/rowAccessPolicies/" + ref.PolicyId
			attrs := map[string]string{"policy_id": ref.PolicyId}
			for k, v := range tableAttrs {
				attrs[k] = v
			}
			g.Resources = append(g.Resources, g.newDepthResource(id, "google_bigquery_row_access_policy", attrs))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (g *BigQueryGenerator) loadRoutines(ctx context.Context, bqService *bigqueryv2.Service, project, datasetID string) {
	err := bqService.Routines.List(project, datasetID).Pages(ctx, func(page *bigqueryv2.ListRoutinesResponse) error {
		for _, routine := range page.Routines {
			ref := routine.RoutineReference
			routineName := "projects/" + ref.ProjectId + "/datasets/" + ref.DatasetId + "/routines/" + ref.RoutineId
			routineAttrs := map[string]string{"project": ref.ProjectId, "dataset_id": ref.DatasetId, "routine_id": ref.RoutineId}
			g.Resources = append(g.Resources, g.newDepthResource(routineName, "google_bigquery_routine", routineAttrs))
			policy, err := bqService.Routines.GetIamPolicy(routineName, &bigqueryv2.GetIamPolicyRequest{}).Context(ctx).Do()
			if err != nil {
				log.Println(err)
				continue
			}
			bindings := []iamBinding{}
			for _, binding := range policy.Bindings {
				bindings = append(bindings, iamBinding{Role: binding.Role, Members: binding.Members})
			}
			g.Resources = append(g.Resources, expandIamPolicy(
				bindings,
				routineName,
				"google_bigquery_routine_iam_policy",
				"google_bigquery_routine_iam_binding",
				"google_bigquery_routine_iam_member",
				routineAttrs,
				g.newDepthResource,
			)...)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (g *BigQueryGenerator) loadReservations(ctx context.Context, reservationService *bigqueryreservationv1.Service, restClient *http.Client, locationParent string, locationAttrs map[string]string) {
	err := reservationService.Projects.Locations.CapacityCommitments.List(locationParent).Pages(ctx, func(page *bigqueryreservationv1.ListCapacityCommitmentsResponse) error {
		for _, commitment := range page.CapacityCommitments {
			g.Resources = append(g.Resources, g.newDepthResource(commitment.Name, "google_bigquery_capacity_commitment", locationAttrs))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = reservationService.Projects.Locations.Reservations.List(locationParent).Pages(ctx, func(page *bigqueryreservationv1.ListReservationsResponse) error {
		for _, reservation := range page.Reservations {
			g.Resources = append(g.Resources, g.newDepthResource(reservation.Name, "google_bigquery_reservation", locationAttrs))
			assignmentErr := reservationService.Projects.Locations.Reservations.Assignments.List(reservation.Name).Pages(ctx, func(page *bigqueryreservationv1.ListAssignmentsResponse) error {
				for _, assignment := range page.Assignments {
					g.Resources = append(g.Resources, g.newDepthResource(assignment.Name, "google_bigquery_reservation_assignment", locationAttrs))
				}
				return nil
			})
			if assignmentErr != nil {
				log.Println(assignmentErr)
			}
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	if restClient != nil {
		groups, err := restListResources(ctx, restClient, bigQueryReservationBaseURL+locationParent+"/reservationGroups")
		if err != nil {
			log.Println(err)
		} else {
			for _, group := range groups {
				g.Resources = append(g.Resources, g.newDepthResource(group.Name, "google_bigquery_reservation_group", locationAttrs))
			}
		}
	}
	// The BI reservation is a per-location singleton that conceptually
	// always exists; the provider treats a zero size as absent.
	biReservation, err := reservationService.Projects.Locations.GetBiReservation(locationParent + "/biReservation").Context(ctx).Do()
	if err != nil {
		log.Println(err)
	} else if biReservation.Size > 0 {
		g.Resources = append(g.Resources, g.newDepthResource(biReservation.Name, "google_bigquery_bi_reservation", locationAttrs))
	}
}

func (g *BigQueryGenerator) loadConnections(ctx context.Context, connectionService *bigqueryconnectionv1.Service, locationParent string, locationAttrs map[string]string) {
	err := connectionService.Projects.Locations.Connections.List(locationParent).Pages(ctx, func(page *bigqueryconnectionv1.ListConnectionsResponse) error {
		for _, connection := range page.Connections {
			g.Resources = append(g.Resources, g.newDepthResource(connection.Name, "google_bigquery_connection", locationAttrs))
			policy, err := connectionService.Projects.Locations.Connections.GetIamPolicy(connection.Name, &bigqueryconnectionv1.GetIamPolicyRequest{}).Context(ctx).Do()
			if err != nil {
				log.Println(err)
				continue
			}
			bindings := []iamBinding{}
			for _, binding := range policy.Bindings {
				bindings = append(bindings, iamBinding{Role: binding.Role, Members: binding.Members})
			}
			parts := strings.Split(connection.Name, "/")
			parentAttrs := map[string]string{"connection_id": parts[len(parts)-1]}
			for k, v := range locationAttrs {
				parentAttrs[k] = v
			}
			g.Resources = append(g.Resources, expandIamPolicy(
				bindings,
				connection.Name,
				"google_bigquery_connection_iam_policy",
				"google_bigquery_connection_iam_binding",
				"google_bigquery_connection_iam_member",
				parentAttrs,
				g.newDepthResource,
			)...)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (g *BigQueryGenerator) loadTransferConfigs(ctx context.Context, transferService *bigquerydatatransferv1.Service, project, locationParent string, locationAttrs map[string]string) {
	err := transferService.Projects.Locations.TransferConfigs.List(locationParent).Pages(ctx, func(page *bigquerydatatransferv1.ListTransferConfigsResponse) error {
		for _, config := range page.TransferConfigs {
			g.Resources = append(g.Resources, g.newDepthResource(config.Name, "google_bigquery_data_transfer_config", locationAttrs))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	// Per the provider's own definition, projects.dataSources.list
	// returns the data sources currently enrolled in the project.
	err = transferService.Projects.DataSources.List("projects/"+project).Pages(ctx, func(page *bigquerydatatransferv1.ListDataSourcesResponse) error {
		for _, dataSource := range page.DataSources {
			id := "projects/" + project + "/dataSources/" + dataSource.DataSourceId
			g.Resources = append(g.Resources, g.newDepthResource(id, "google_bigquery_data_transfer_data_source_enrollment", map[string]string{"project": project}))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (g *BigQueryGenerator) loadAnalyticsHub(ctx context.Context, hubService *analyticshubv1.Service, restClient *http.Client, locationParent string, locationAttrs map[string]string) {
	err := hubService.Projects.Locations.DataExchanges.List(locationParent).Pages(ctx, func(page *analyticshubv1.ListDataExchangesResponse) error {
		for _, exchange := range page.DataExchanges {
			g.Resources = append(g.Resources, g.newDepthResource(exchange.Name, "google_bigquery_analytics_hub_data_exchange", locationAttrs))
			g.loadAnalyticsHubIam(func() (*analyticshubv1.Policy, error) {
				return hubService.Projects.Locations.DataExchanges.GetIamPolicy(exchange.Name, &analyticshubv1.GetIamPolicyRequest{}).Context(ctx).Do()
			}, exchange.Name, "google_bigquery_analytics_hub_data_exchange", locationAttrs)
			g.loadListings(ctx, hubService, exchange.Name, locationAttrs)
			g.loadQueryTemplates(ctx, restClient, exchange.Name, locationAttrs)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	// Subscriptions live in the subscriber's project at location
	// level; see the header comment for why they are emitted only as
	// data exchange subscriptions.
	err = hubService.Projects.Locations.Subscriptions.List(locationParent).Pages(ctx, func(page *analyticshubv1.ListSubscriptionsResponse) error {
		for _, subscription := range page.Subscriptions {
			g.Resources = append(g.Resources, g.newDepthResource(subscription.Name, "google_bigquery_analytics_hub_data_exchange_subscription", locationAttrs))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (g *BigQueryGenerator) loadListings(ctx context.Context, hubService *analyticshubv1.Service, exchangeName string, locationAttrs map[string]string) {
	err := hubService.Projects.Locations.DataExchanges.Listings.List(exchangeName).Pages(ctx, func(page *analyticshubv1.ListListingsResponse) error {
		for _, listing := range page.Listings {
			g.Resources = append(g.Resources, g.newDepthResource(listing.Name, "google_bigquery_analytics_hub_listing", locationAttrs))
			g.loadAnalyticsHubIam(func() (*analyticshubv1.Policy, error) {
				return hubService.Projects.Locations.DataExchanges.Listings.GetIamPolicy(listing.Name, &analyticshubv1.GetIamPolicyRequest{}).Context(ctx).Do()
			}, listing.Name, "google_bigquery_analytics_hub_listing", locationAttrs)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (g *BigQueryGenerator) loadAnalyticsHubIam(getPolicy func() (*analyticshubv1.Policy, error), name, resourcePrefix string, parentAttrs map[string]string) {
	// Data exchange and listing IAM calls have distinct generated
	// types; callers pass a closure issuing either one.
	policy, err := getPolicy()
	if err != nil {
		log.Println(err)
		return
	}
	bindings := []iamBinding{}
	for _, binding := range policy.Bindings {
		bindings = append(bindings, iamBinding{Role: binding.Role, Members: binding.Members})
	}
	g.Resources = append(g.Resources, expandIamPolicy(
		bindings,
		name,
		resourcePrefix+"_iam_policy",
		resourcePrefix+"_iam_binding",
		resourcePrefix+"_iam_member",
		parentAttrs,
		g.newDepthResource,
	)...)
}

func (g *BigQueryGenerator) loadQueryTemplates(ctx context.Context, restClient *http.Client, exchangeName string, locationAttrs map[string]string) {
	if restClient == nil {
		return
	}
	templates, err := restListResources(ctx, restClient, analyticsHubBaseURL+exchangeName+"/queryTemplates")
	if err != nil {
		log.Println(err)
		return
	}
	for _, template := range templates {
		g.Resources = append(g.Resources, g.newDepthResource(template.Name, "google_bigquery_analytics_hub_query_template", locationAttrs))
	}
}

func (g *BigQueryGenerator) loadDataPolicies(ctx context.Context, datapolicyService *bigquerydatapolicyv1.Service, locationParent string, locationAttrs map[string]string) {
	err := datapolicyService.Projects.Locations.DataPolicies.List(locationParent).Pages(ctx, func(page *bigquerydatapolicyv1.ListDataPoliciesResponse) error {
		for _, dataPolicy := range page.DataPolicies {
			g.Resources = append(g.Resources, g.newDepthResource(dataPolicy.Name, "google_bigquery_datapolicy_data_policy", locationAttrs))
			policy, err := datapolicyService.Projects.Locations.DataPolicies.GetIamPolicy(dataPolicy.Name, &bigquerydatapolicyv1.GetIamPolicyRequest{}).Context(ctx).Do()
			if err != nil {
				log.Println(err)
				continue
			}
			bindings := []iamBinding{}
			for _, binding := range policy.Bindings {
				bindings = append(bindings, iamBinding{Role: binding.Role, Members: binding.Members})
			}
			g.Resources = append(g.Resources, expandIamPolicy(
				bindings,
				dataPolicy.Name,
				"google_bigquery_datapolicy_data_policy_iam_policy",
				"google_bigquery_datapolicy_data_policy_iam_binding",
				"google_bigquery_datapolicy_data_policy_iam_member",
				locationAttrs,
				g.newDepthResource,
			)...)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}
