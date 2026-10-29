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

// DynamoDB depth coverage (AWS sweep, database cluster).
//
// Skipped in this family, with reasons:
//   - aws_dynamodb_table_item: data-plane item; items are table data, not
//     enumerable infrastructure (a Scan would be unbounded).
//   - aws_dynamodb_table_replica: importer ID is "<table>:<main region>", but
//     the DynamoDB API does not identify which replica region is the main
//     (source) region of a global table, so a verified ID cannot be built.
//   - aws_dynamodb_tag: single-tag association glue; tags are attributes of
//     the table resource itself.

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func (g *DynamoDbGenerator) listTableNames(svc *dynamodb.Client) ([]string, error) {
	var names []string
	p := dynamodb.NewListTablesPaginator(svc, &dynamodb.ListTablesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return names, err
		}
		names = append(names, page.TableNames...)
	}
	return names, nil
}

// Per-table depth: GSIs, Kinesis streaming destinations, resource policies.
func (g *DynamoDbGenerator) loadDynamoTableDepth(svc *dynamodb.Client, tableNames []string) error {
	for _, tableName := range tableNames {
		desc, err := svc.DescribeTable(context.TODO(), &dynamodb.DescribeTableInput{TableName: aws.String(tableName)})
		if err != nil {
			return err
		}
		table := desc.Table

		// aws_dynamodb_global_secondary_index, ID = "<table>,<index>".
		for _, gsi := range table.GlobalSecondaryIndexes {
			indexName := StringValue(gsi.IndexName)
			id := tableName + "," + indexName
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, tableName+"_"+indexName, "aws_dynamodb_global_secondary_index", "aws", dynamodbAllowEmptyValues))
		}

		// aws_dynamodb_kinesis_streaming_destination, ID = "<table>,<stream ARN>".
		kinesis, err := svc.DescribeKinesisStreamingDestination(context.TODO(), &dynamodb.DescribeKinesisStreamingDestinationInput{TableName: aws.String(tableName)})
		if err == nil {
			for _, dest := range kinesis.KinesisDataStreamDestinations {
				if dest.DestinationStatus != dynamodbtypes.DestinationStatusActive && dest.DestinationStatus != dynamodbtypes.DestinationStatusEnabling {
					continue
				}
				streamARN := StringValue(dest.StreamArn)
				id := tableName + "," + streamARN
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, tableName+"_kinesis", "aws_dynamodb_kinesis_streaming_destination", "aws", dynamodbAllowEmptyValues))
			}
		}

		// aws_dynamodb_resource_policy, ID = table ARN; exists only when a
		// policy is attached.
		policy, err := svc.GetResourcePolicy(context.TODO(), &dynamodb.GetResourcePolicyInput{ResourceArn: table.TableArn})
		if err == nil && policy != nil && StringValue(policy.Policy) != "" {
			arn := StringValue(table.TableArn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, tableName, "aws_dynamodb_resource_policy", "aws", dynamodbAllowEmptyValues))
		}
	}
	return nil
}

// aws_dynamodb_contributor_insights: ID = table name, or "<table>/<index>"
// for index-level insights. Only ENABLED summaries are live resources.
func (g *DynamoDbGenerator) loadDynamoContributorInsights(svc *dynamodb.Client) error {
	p := dynamodb.NewListContributorInsightsPaginator(svc, &dynamodb.ListContributorInsightsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, summary := range page.ContributorInsightsSummaries {
			if summary.ContributorInsightsStatus != dynamodbtypes.ContributorInsightsStatusEnabled {
				continue
			}
			tableName := StringValue(summary.TableName)
			indexName := StringValue(summary.IndexName)
			id := tableName
			name := tableName
			if indexName != "" {
				id = tableName + "/" + indexName
				name = tableName + "_" + indexName
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, name, "aws_dynamodb_contributor_insights", "aws", dynamodbAllowEmptyValues))
		}
	}
	return nil
}

// aws_dynamodb_global_table (legacy 2017 global tables), ID = table name.
// ListGlobalTables uses LastEvaluatedGlobalTableName pagination (no paginator helper).
func (g *DynamoDbGenerator) loadDynamoGlobalTables(svc *dynamodb.Client) error {
	var startName *string
	for {
		page, err := svc.ListGlobalTables(context.TODO(), &dynamodb.ListGlobalTablesInput{ExclusiveStartGlobalTableName: startName})
		if err != nil {
			return err
		}
		for _, gt := range page.GlobalTables {
			name := StringValue(gt.GlobalTableName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_dynamodb_global_table", "aws", dynamodbAllowEmptyValues))
		}
		if page.LastEvaluatedGlobalTableName == nil {
			return nil
		}
		startName = page.LastEvaluatedGlobalTableName
	}
}

// aws_dynamodb_table_export, ID = export ARN.
func (g *DynamoDbGenerator) loadDynamoTableExports(svc *dynamodb.Client) error {
	p := dynamodb.NewListExportsPaginator(svc, &dynamodb.ListExportsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, exp := range page.ExportSummaries {
			arn := StringValue(exp.ExportArn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_dynamodb_table_export", "aws", dynamodbAllowEmptyValues))
		}
	}
	return nil
}

func (g *DynamoDbGenerator) loadDynamoDepth(svc *dynamodb.Client) error {
	tableNames, err := g.listTableNames(svc)
	if err != nil {
		return err
	}
	if err := g.loadDynamoTableDepth(svc, tableNames); err != nil {
		return err
	}
	if err := g.loadDynamoContributorInsights(svc); err != nil {
		return err
	}
	if err := g.loadDynamoGlobalTables(svc); err != nil {
		return err
	}
	if err := g.loadDynamoTableExports(svc); err != nil {
		return err
	}
	return nil
}
