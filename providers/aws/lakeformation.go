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

// Lake Formation coverage (AWS sweep, database cluster).
//
// Skipped in this family, with reasons:
//   - aws_lakeformation_data_lake_settings: singleton settings object for
//     the account/region.
//   - aws_lakeformation_identity_center_configuration: singleton per catalog
//     (DescribeLakeFormationIdentityCenterConfiguration returns the one
//     configuration, if any).
//   - aws_lakeformation_permissions: a permission grant; the provider ID is
//     a hash of principal + resource computed at create time and the
//     resource does not support import.
//   - aws_lakeformation_resource_lf_tag / aws_lakeformation_resource_lf_tags:
//     LF-tag-to-resource association glue; tags on a resource are attributes
//     enumerable via GetResourceLFTags but the association resources have no
//     stable importer identity separate from the resource itself.

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	"github.com/aws/aws-sdk-go-v2/service/lakeformation"
	lftypes "github.com/aws/aws-sdk-go-v2/service/lakeformation/types"
)

var lakeFormationAllowEmptyValues = []string{"tags."}

type LakeFormationGenerator struct {
	AWSService
}

func (g *LakeFormationGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := lakeformation.NewFromConfig(config)
	account, err := g.getAccountNumber(config)
	if err != nil {
		return err
	}
	catalogID := StringValue(account)

	// aws_lakeformation_lf_tag, ID = tag key.
	{
		out, err := svc.ListLFTags(context.TODO(), &lakeformation.ListLFTagsInput{CatalogId: account})
		if err != nil {
			return err
		}
		for _, tag := range out.LFTags {
			key := StringValue(tag.TagKey)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				key, key, "aws_lakeformation_lf_tag", "aws", lakeFormationAllowEmptyValues))
		}
	}

	// aws_lakeformation_lf_tag_expression, ID = expression name.
	{
		out, err := svc.ListLFTagExpressions(context.TODO(), &lakeformation.ListLFTagExpressionsInput{CatalogId: account})
		if err != nil {
			return err
		}
		for _, expr := range out.LFTagExpressions {
			name := StringValue(expr.Name)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_lakeformation_lf_tag_expression", "aws", lakeFormationAllowEmptyValues))
		}
	}

	// aws_lakeformation_resource, ID = resource ARN.
	{
		out, err := svc.ListResources(context.TODO(), &lakeformation.ListResourcesInput{})
		if err != nil {
			return err
		}
		for _, res := range out.ResourceInfoList {
			arn := StringValue(res.ResourceArn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_lakeformation_resource", "aws", lakeFormationAllowEmptyValues))
		}
	}

	// aws_lakeformation_opt_in: only data-location opt-ins are emitted;
	// their provider ID is the data location ARN.
	{
		out, err := svc.ListLakeFormationOptIns(context.TODO(), &lakeformation.ListLakeFormationOptInsInput{})
		if err != nil {
			return err
		}
		for _, optIn := range out.LakeFormationOptInsInfoList {
			if optIn.Resource == nil || optIn.Resource.DataLocation == nil {
				continue
			}
			arn := StringValue(optIn.Resource.DataLocation.ResourceArn)
			if arn == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_lakeformation_opt_in", "aws", lakeFormationAllowEmptyValues))
		}
	}

	// aws_lakeformation_data_cells_filter, enumerated per Glue table.
	// ID = "<database name>,<filter name>,<table catalog ID>,<table name>".
	glueSvc := glue.NewFromConfig(config)
	dbPaginator := glue.NewGetDatabasesPaginator(glueSvc, &glue.GetDatabasesInput{})
	for dbPaginator.HasMorePages() {
		dbPage, err := dbPaginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, db := range dbPage.DatabaseList {
			dbName := StringValue(db.Name)
			tablePaginator := glue.NewGetTablesPaginator(glueSvc, &glue.GetTablesInput{DatabaseName: aws.String(dbName)})
			for tablePaginator.HasMorePages() {
				tablePage, err := tablePaginator.NextPage(context.TODO())
				if err != nil {
					return err
				}
				for _, table := range tablePage.TableList {
					tableName := StringValue(table.Name)
					filters, err := svc.ListDataCellsFilter(context.TODO(), &lakeformation.ListDataCellsFilterInput{
						Table: &lftypes.TableResource{
							CatalogId:    aws.String(catalogID),
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					})
					if err != nil {
						return err
					}
					for _, filter := range filters.DataCellsFilters {
						filterName := StringValue(filter.Name)
						id := dbName + "," + filterName + "," + catalogID + "," + tableName
						g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
							id, dbName+"_"+tableName+"_"+filterName, "aws_lakeformation_data_cells_filter", "aws", lakeFormationAllowEmptyValues))
					}
				}
			}
		}
	}

	return nil
}
