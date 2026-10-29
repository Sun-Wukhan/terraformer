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

// Glue depth coverage (AWS sweep, database cluster).
//
// Skipped in this family, with reasons:
//   - aws_glue_catalog: the Glue Data Catalog itself is a singleton per
//     account/region, not an enumerable resource.
//   - aws_glue_data_catalog_encryption_settings: singleton settings object
//     for the account's catalog.
//   - aws_glue_resource_policy: singleton policy object for the account's
//     catalog (GetResourcePolicy with no resource argument).
//   - aws_glue_partition: partitions are data-plane slices of a table; a
//     table can hold millions, so they are not enumerable infrastructure.

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	gluetypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
)

var glueDepthAllowEmptyValues = []string{"tags."}

func (g *GlueGenerator) loadGlueDepth(svc *glue.Client, account *string) error {
	catalogID := StringValue(account)

	// aws_glue_classifier, ID = classifier name.
	{
		p := glue.NewGetClassifiersPaginator(svc, &glue.GetClassifiersInput{})
		for p.HasMorePages() {
			page, err := p.NextPage(context.TODO())
			if err != nil {
				return err
			}
			for _, classifier := range page.Classifiers {
				name := glueClassifierName(classifier)
				if name == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					name, name, "aws_glue_classifier", "aws", glueDepthAllowEmptyValues))
			}
		}
	}

	// aws_glue_connection, ID = "<catalog ID>:<connection name>".
	{
		p := glue.NewGetConnectionsPaginator(svc, &glue.GetConnectionsInput{CatalogId: account})
		for p.HasMorePages() {
			page, err := p.NextPage(context.TODO())
			if err != nil {
				return err
			}
			for _, conn := range page.ConnectionList {
				name := StringValue(conn.Name)
				id := catalogID + ":" + name
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, name, "aws_glue_connection", "aws", glueDepthAllowEmptyValues))
			}
		}
	}

	// aws_glue_data_quality_ruleset, ID = ruleset name.
	{
		p := glue.NewListDataQualityRulesetsPaginator(svc, &glue.ListDataQualityRulesetsInput{})
		for p.HasMorePages() {
			page, err := p.NextPage(context.TODO())
			if err != nil {
				return err
			}
			for _, ruleset := range page.Rulesets {
				name := StringValue(ruleset.Name)
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					name, name, "aws_glue_data_quality_ruleset", "aws", glueDepthAllowEmptyValues))
			}
		}
	}

	// aws_glue_dev_endpoint, ID = endpoint name.
	{
		out, err := svc.ListDevEndpoints(context.TODO(), &glue.ListDevEndpointsInput{})
		if err != nil {
			return err
		}
		for _, name := range out.DevEndpointNames {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_glue_dev_endpoint", "aws", glueDepthAllowEmptyValues))
		}
	}

	// aws_glue_ml_transform, ID = transform ID.
	{
		p := glue.NewListMLTransformsPaginator(svc, &glue.ListMLTransformsInput{})
		for p.HasMorePages() {
			page, err := p.NextPage(context.TODO())
			if err != nil {
				return err
			}
			for _, transformID := range page.TransformIds {
				id := transformID
				name := id
				if out, err := svc.GetMLTransform(context.TODO(), &glue.GetMLTransformInput{TransformId: aws.String(transformID)}); err == nil && out != nil && out.Name != nil {
					name = StringValue(out.Name)
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, name, "aws_glue_ml_transform", "aws", glueDepthAllowEmptyValues))
			}
		}
	}

	// aws_glue_registry, ID = registry ARN; plus schemas per registry.
	{
		p := glue.NewListRegistriesPaginator(svc, &glue.ListRegistriesInput{})
		for p.HasMorePages() {
			page, err := p.NextPage(context.TODO())
			if err != nil {
				return err
			}
			for _, registry := range page.Registries {
				registryARN := StringValue(registry.RegistryArn)
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					registryARN, StringValue(registry.RegistryName), "aws_glue_registry", "aws", glueDepthAllowEmptyValues))

				// aws_glue_schema, ID = schema ARN.
				sp := glue.NewListSchemasPaginator(svc, &glue.ListSchemasInput{
					RegistryId: &gluetypes.RegistryId{RegistryArn: aws.String(registryARN)},
				})
				for sp.HasMorePages() {
					spage, err := sp.NextPage(context.TODO())
					if err != nil {
						return err
					}
					for _, schema := range spage.Schemas {
						schemaARN := StringValue(schema.SchemaArn)
						g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
							schemaARN, StringValue(schema.SchemaName), "aws_glue_schema", "aws", glueDepthAllowEmptyValues))
					}
				}
			}
		}
	}

	// aws_glue_security_configuration, ID = name.
	{
		p := glue.NewGetSecurityConfigurationsPaginator(svc, &glue.GetSecurityConfigurationsInput{})
		for p.HasMorePages() {
			page, err := p.NextPage(context.TODO())
			if err != nil {
				return err
			}
			for _, conf := range page.SecurityConfigurations {
				name := StringValue(conf.Name)
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					name, name, "aws_glue_security_configuration", "aws", glueDepthAllowEmptyValues))
			}
		}
	}

	// aws_glue_workflow, ID = workflow name.
	{
		p := glue.NewListWorkflowsPaginator(svc, &glue.ListWorkflowsInput{})
		for p.HasMorePages() {
			page, err := p.NextPage(context.TODO())
			if err != nil {
				return err
			}
			for _, name := range page.Workflows {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					name, name, "aws_glue_workflow", "aws", glueDepthAllowEmptyValues))
			}
		}
	}

	// Per-database / per-table depth.
	databases, err := g.listGlueDatabaseNames(svc)
	if err != nil {
		return err
	}
	for _, dbName := range databases {
		// aws_glue_user_defined_function, ID = "<catalog ID>:<database>:<function>".
		udf, err := svc.GetUserDefinedFunctions(context.TODO(), &glue.GetUserDefinedFunctionsInput{
			CatalogId:    account,
			DatabaseName: aws.String(dbName),
			Pattern:      aws.String(".*"),
		})
		if err != nil {
			return err
		}
		for _, fn := range udf.UserDefinedFunctions {
			fnName := StringValue(fn.FunctionName)
			id := catalogID + ":" + dbName + ":" + fnName
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, dbName+"_"+fnName, "aws_glue_user_defined_function", "aws", glueDepthAllowEmptyValues))
		}

		tables, err := g.listGlueTableNames(svc, dbName)
		if err != nil {
			return err
		}
		for _, tableName := range tables {
			// aws_glue_partition_index,
			// ID = "<catalog ID>:<database>:<table>:<index>".
			indexes, err := svc.GetPartitionIndexes(context.TODO(), &glue.GetPartitionIndexesInput{
				CatalogId:    account,
				DatabaseName: aws.String(dbName),
				TableName:    aws.String(tableName),
			})
			if err != nil {
				return err
			}
			for _, index := range indexes.PartitionIndexDescriptorList {
				indexName := StringValue(index.IndexName)
				id := catalogID + ":" + dbName + ":" + tableName + ":" + indexName
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, dbName+"_"+tableName+"_"+indexName, "aws_glue_partition_index", "aws", glueDepthAllowEmptyValues))
			}

			// aws_glue_catalog_table_optimizer,
			// ID = "<catalog ID>:<database>:<table>:<optimizer type>".
			// The pinned Glue SDK's TableOptimizerType enum only names
			// "compaction"; retention and orphan-file-deletion are valid
			// API values the enum has not caught up with, so spell them
			// as literals of the same string type.
			for _, optimizerType := range []gluetypes.TableOptimizerType{
				gluetypes.TableOptimizerTypeCompaction,
				gluetypes.TableOptimizerType("retention"),
				gluetypes.TableOptimizerType("orphan_file_deletion"),
			} {
				optimizer, err := svc.GetTableOptimizer(context.TODO(), &glue.GetTableOptimizerInput{
					CatalogId:    account,
					DatabaseName: aws.String(dbName),
					TableName:    aws.String(tableName),
					Type:         optimizerType,
				})
				if err != nil || optimizer == nil || optimizer.TableOptimizer == nil || optimizer.TableOptimizer.Configuration == nil {
					continue // no optimizer of this type configured for the table
				}
				id := catalogID + ":" + dbName + ":" + tableName + ":" + string(optimizerType)
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, dbName+"_"+tableName+"_"+string(optimizerType), "aws_glue_catalog_table_optimizer", "aws", glueDepthAllowEmptyValues))
			}
		}
	}

	return nil
}

func glueClassifierName(classifier gluetypes.Classifier) string {
	switch {
	case classifier.GrokClassifier != nil:
		return StringValue(classifier.GrokClassifier.Name)
	case classifier.JsonClassifier != nil:
		return StringValue(classifier.JsonClassifier.Name)
	case classifier.XMLClassifier != nil:
		return StringValue(classifier.XMLClassifier.Name)
	case classifier.CsvClassifier != nil:
		return StringValue(classifier.CsvClassifier.Name)
	}
	return ""
}

func (g *GlueGenerator) listGlueDatabaseNames(svc *glue.Client) ([]string, error) {
	var names []string
	p := glue.NewGetDatabasesPaginator(svc, &glue.GetDatabasesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return names, err
		}
		for _, db := range page.DatabaseList {
			names = append(names, StringValue(db.Name))
		}
	}
	return names, nil
}

func (g *GlueGenerator) listGlueTableNames(svc *glue.Client, databaseName string) ([]string, error) {
	var names []string
	p := glue.NewGetTablesPaginator(svc, &glue.GetTablesInput{DatabaseName: aws.String(databaseName)})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return names, err
		}
		for _, table := range page.TableList {
			names = append(names, StringValue(table.Name))
		}
	}
	return names, nil
}
