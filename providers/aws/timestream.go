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

// Timestream coverage (AWS sweep, database cluster): Timestream for
// LiveAnalytics (timestreamwrite + timestreamquery) and Timestream for
// InfluxDB.

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb"
	"github.com/aws/aws-sdk-go-v2/service/timestreamquery"
	"github.com/aws/aws-sdk-go-v2/service/timestreamwrite"
)

var timestreamAllowEmptyValues = []string{"tags."}

// --- TimestreamWrite ---

type TimestreamWriteGenerator struct {
	AWSService
}

func (g *TimestreamWriteGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := timestreamwrite.NewFromConfig(config)

	dp := timestreamwrite.NewListDatabasesPaginator(svc, &timestreamwrite.ListDatabasesInput{})
	for dp.HasMorePages() {
		page, err := dp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, db := range page.Databases {
			dbName := StringValue(db.DatabaseName)
			// aws_timestreamwrite_database, ID = database name.
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				dbName, dbName, "aws_timestreamwrite_database", "aws", timestreamAllowEmptyValues))

			// aws_timestreamwrite_table, ID = "<table name>:<database name>".
			tp := timestreamwrite.NewListTablesPaginator(svc, &timestreamwrite.ListTablesInput{DatabaseName: aws.String(dbName)})
			for tp.HasMorePages() {
				tpage, err := tp.NextPage(context.TODO())
				if err != nil {
					return err
				}
				for _, table := range tpage.Tables {
					tableName := StringValue(table.TableName)
					id := tableName + ":" + dbName
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						id, dbName+"_"+tableName, "aws_timestreamwrite_table", "aws", timestreamAllowEmptyValues))
				}
			}
		}
	}
	return nil
}

// --- TimestreamQuery ---

type TimestreamQueryGenerator struct {
	AWSService
}

func (g *TimestreamQueryGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := timestreamquery.NewFromConfig(config)

	// aws_timestreamquery_scheduled_query, ID = scheduled query ARN.
	p := timestreamquery.NewListScheduledQueriesPaginator(svc, &timestreamquery.ListScheduledQueriesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, q := range page.ScheduledQueries {
			arn := StringValue(q.Arn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(q.Name), "aws_timestreamquery_scheduled_query", "aws", timestreamAllowEmptyValues))
		}
	}
	return nil
}

// --- TimestreamInfluxDB ---

type TimestreamInfluxDBGenerator struct {
	AWSService
}

func (g *TimestreamInfluxDBGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := timestreaminfluxdb.NewFromConfig(config)

	// aws_timestreaminfluxdb_db_instance, ID = DB instance ID.
	ip := timestreaminfluxdb.NewListDbInstancesPaginator(svc, &timestreaminfluxdb.ListDbInstancesInput{})
	for ip.HasMorePages() {
		page, err := ip.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, inst := range page.Items {
			id := StringValue(inst.Id)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(inst.Name), "aws_timestreaminfluxdb_db_instance", "aws", timestreamAllowEmptyValues))
		}
	}

	// aws_timestreaminfluxdb_db_cluster, ID = DB cluster ID.
	cp := timestreaminfluxdb.NewListDbClustersPaginator(svc, &timestreaminfluxdb.ListDbClustersInput{})
	for cp.HasMorePages() {
		page, err := cp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, cluster := range page.Items {
			id := StringValue(cluster.Id)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(cluster.Name), "aws_timestreaminfluxdb_db_cluster", "aws", timestreamAllowEmptyValues))
		}
	}
	return nil
}
