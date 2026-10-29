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

package aws

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/athena"
)

var athenaAllowEmptyValues = []string{"tags."}

// AthenaGenerator covers Athena workgroups, data catalogs, databases, named
// queries, prepared statements and capacity reservations.
// Provider import IDs (verified): workgroup = name; data catalog = name;
// database = name; named query = query ID; prepared statement =
// "<workgroup name>/<statement name>"; capacity reservation = name.
type AthenaGenerator struct {
	AWSService
}

func (g *AthenaGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := athena.NewFromConfig(config)

	wp := athena.NewListWorkGroupsPaginator(svc, &athena.ListWorkGroupsInput{})
	var workgroups []string
	for wp.HasMorePages() {
		page, err := wp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, wg := range page.WorkGroups {
			name := StringValue(wg.Name)
			workgroups = append(workgroups, name)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_athena_workgroup", "aws", athenaAllowEmptyValues))
		}
	}

	cp := athena.NewListDataCatalogsPaginator(svc, &athena.ListDataCatalogsInput{})
	var catalogs []string
	for cp.HasMorePages() {
		page, err := cp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, cat := range page.DataCatalogsSummary {
			name := StringValue(cat.CatalogName)
			catalogs = append(catalogs, name)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_athena_data_catalog", "aws", athenaAllowEmptyValues))
		}
	}
	if len(catalogs) == 0 {
		catalogs = []string{"AwsDataCatalog"}
	}
	seenDB := map[string]bool{}
	for _, catalog := range catalogs {
		dp := athena.NewListDatabasesPaginator(svc, &athena.ListDatabasesInput{CatalogName: aws.String(catalog)})
		for dp.HasMorePages() {
			page, err := dp.NextPage(context.TODO())
			if err != nil {
				break
			}
			for _, db := range page.DatabaseList {
				name := StringValue(db.Name)
				if seenDB[name] {
					continue
				}
				seenDB[name] = true
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					name, name, "aws_athena_database", "aws", athenaAllowEmptyValues))
			}
		}
	}

	for _, wg := range workgroups {
		np := athena.NewListNamedQueriesPaginator(svc, &athena.ListNamedQueriesInput{WorkGroup: aws.String(wg)})
		for np.HasMorePages() {
			page, err := np.NextPage(context.TODO())
			if err != nil {
				break
			}
			for _, qid := range page.NamedQueryIds {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					qid, qid, "aws_athena_named_query", "aws", athenaAllowEmptyValues))
			}
		}
		sp := athena.NewListPreparedStatementsPaginator(svc, &athena.ListPreparedStatementsInput{WorkGroup: aws.String(wg)})
		for sp.HasMorePages() {
			page, err := sp.NextPage(context.TODO())
			if err != nil {
				break
			}
			for _, st := range page.PreparedStatements {
				id := wg + "/" + StringValue(st.StatementName)
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, id, "aws_athena_prepared_statement", "aws", athenaAllowEmptyValues))
			}
		}
	}

	rp := athena.NewListCapacityReservationsPaginator(svc, &athena.ListCapacityReservationsInput{})
	for rp.HasMorePages() {
		page, err := rp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, res := range page.CapacityReservations {
			name := StringValue(res.Name)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_athena_capacity_reservation", "aws", athenaAllowEmptyValues))
		}
	}
	return nil
}
