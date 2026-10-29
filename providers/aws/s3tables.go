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
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/s3tables"
)

// S3 Tables sweep. Importer IDs verified against provider docs:
// table bucket = its ARN; namespace = "bucket-arn;namespace";
// table / table policy / table replication = "bucket-arn;namespace;name";
// table bucket policy / table bucket replication = the table bucket ARN
// (same composite pattern as the other per-bucket sub-resources).

var s3TablesAllowEmptyValues = []string{"tags."}

type S3TablesGenerator struct {
	AWSService
}

func (g *S3TablesGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := s3tables.NewFromConfig(config)
	ctx := context.TODO()
	add := func(id, name, typ string) {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(id, name, typ, "aws", s3TablesAllowEmptyValues))
	}
	p := s3tables.NewListTableBucketsPaginator(svc, &s3tables.ListTableBucketsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, bucket := range page.TableBuckets {
			arn := StringValue(bucket.Arn)
			add(arn, StringValue(bucket.Name), "aws_s3tables_table_bucket")
			if _, err := svc.GetTableBucketPolicy(ctx, &s3tables.GetTableBucketPolicyInput{TableBucketARN: &arn}); err == nil {
				add(arn, StringValue(bucket.Name), "aws_s3tables_table_bucket_policy")
			}
			if _, err := svc.GetTableBucketReplication(ctx, &s3tables.GetTableBucketReplicationInput{TableBucketARN: &arn}); err == nil {
				add(arn, StringValue(bucket.Name), "aws_s3tables_table_bucket_replication")
			}
			nsP := s3tables.NewListNamespacesPaginator(svc, &s3tables.ListNamespacesInput{TableBucketARN: &arn})
			for nsP.HasMorePages() {
				nsPage, err := nsP.NextPage(ctx)
				if err != nil {
					break
				}
				for _, ns := range nsPage.Namespaces {
					// The SDK models a namespace as a []string (hierarchical
					// name parts); the provider ID uses the dotted name.
					nsName := strings.Join(ns.Namespace, ".")
					nsID := arn + ";" + nsName
					add(nsID, nsID, "aws_s3tables_namespace")
					tP := s3tables.NewListTablesPaginator(svc, &s3tables.ListTablesInput{TableBucketARN: &arn, Namespace: &nsName})
					for tP.HasMorePages() {
						tPage, err := tP.NextPage(ctx)
						if err != nil {
							break
						}
						for _, table := range tPage.Tables {
							tID := nsID + ";" + StringValue(table.Name)
							add(tID, tID, "aws_s3tables_table")
							if _, err := svc.GetTablePolicy(ctx, &s3tables.GetTablePolicyInput{TableBucketARN: &arn, Namespace: &nsName, Name: table.Name}); err == nil {
								add(tID, tID, "aws_s3tables_table_policy")
							}
							if _, err := svc.GetTableReplication(ctx, &s3tables.GetTableReplicationInput{TableArn: table.TableARN}); err == nil {
								add(tID, tID, "aws_s3tables_table_replication")
							}
						}
					}
				}
			}
		}
	}
	return nil
}
