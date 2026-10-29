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
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// s3_subconfig.go (AWS storage sweep): per-bucket sub-configuration resources.
//
// Importer IDs were verified against the Terraform AWS provider docs for each
// type: the bucket-singleton configurations import with the bucket name as ID;
// the named configurations (analytics, inventory, intelligent-tiering and
// metrics) import with "bucket:config-name".
//
// Every type is emitted only when the corresponding Get/List call succeeds
// (and, where a Get always succeeds, only when it reports a real setting), so
// buckets without the configuration do not produce phantom resources.
//
// Skipped in this family (see results ledger):
//   - aws_s3_account_public_access_block: account-level singleton; the
//     provider imports it by account ID, but it is not enumerable (there is
//     exactly one per account and no list/discovery call) — singleton
//     precedent.
//   - aws_s3_object, aws_s3_bucket_object (deprecated alias) and
//     aws_s3_object_copy: data-plane user objects, not bucket configuration;
//     enumerating them would import application data, not infrastructure.

func s3BucketResource(bucket, resourceType string) terraformutils.Resource {
	return terraformutils.NewSimpleResource(bucket, bucket, resourceType, "aws", S3AllowEmptyValues)
}

func (g *S3Generator) bucketSubConfigs(svc *s3.Client, bucket string) []terraformutils.Resource {
	var out []terraformutils.Resource
	ctx := context.TODO()
	b := &s3.GetBucketAclInput{Bucket: &bucket}

	if _, err := svc.GetBucketAbac(ctx, &s3.GetBucketAbacInput{Bucket: &bucket}); err == nil {
		out = append(out, s3BucketResource(bucket, "aws_s3_bucket_abac"))
	}
	if acc, err := svc.GetBucketAccelerateConfiguration(ctx, &s3.GetBucketAccelerateConfigurationInput{Bucket: &bucket}); err == nil && acc.Status != "" {
		out = append(out, s3BucketResource(bucket, "aws_s3_bucket_accelerate_configuration"))
	}
	if _, err := svc.GetBucketAcl(ctx, b); err == nil {
		out = append(out, s3BucketResource(bucket, "aws_s3_bucket_acl"))
	}
	if lst, err := svc.ListBucketAnalyticsConfigurations(ctx, &s3.ListBucketAnalyticsConfigurationsInput{Bucket: &bucket}); err == nil {
		for _, c := range lst.AnalyticsConfigurationList {
			id := bucket + ":" + StringValue(c.Id)
			out = append(out, terraformutils.NewSimpleResource(id, id, "aws_s3_bucket_analytics_configuration", "aws", S3AllowEmptyValues))
		}
	}
	if _, err := svc.GetBucketCors(ctx, &s3.GetBucketCorsInput{Bucket: &bucket}); err == nil {
		out = append(out, s3BucketResource(bucket, "aws_s3_bucket_cors_configuration"))
	}
	if lst, err := svc.ListBucketIntelligentTieringConfigurations(ctx, &s3.ListBucketIntelligentTieringConfigurationsInput{Bucket: &bucket}); err == nil {
		for _, c := range lst.IntelligentTieringConfigurationList {
			id := bucket + ":" + StringValue(c.Id)
			out = append(out, terraformutils.NewSimpleResource(id, id, "aws_s3_bucket_intelligent_tiering_configuration", "aws", S3AllowEmptyValues))
		}
	}
	if lst, err := svc.ListBucketInventoryConfigurations(ctx, &s3.ListBucketInventoryConfigurationsInput{Bucket: &bucket}); err == nil {
		for _, c := range lst.InventoryConfigurationList {
			id := bucket + ":" + StringValue(c.Id)
			out = append(out, terraformutils.NewSimpleResource(id, id, "aws_s3_bucket_inventory", "aws", S3AllowEmptyValues))
		}
	}
	if lc, err := svc.GetBucketLifecycleConfiguration(ctx, &s3.GetBucketLifecycleConfigurationInput{Bucket: &bucket}); err == nil && len(lc.Rules) > 0 {
		out = append(out, s3BucketResource(bucket, "aws_s3_bucket_lifecycle_configuration"))
	}
	if lg, err := svc.GetBucketLogging(ctx, &s3.GetBucketLoggingInput{Bucket: &bucket}); err == nil && lg.LoggingEnabled != nil {
		out = append(out, s3BucketResource(bucket, "aws_s3_bucket_logging"))
	}
	if _, err := svc.GetBucketMetadataConfiguration(ctx, &s3.GetBucketMetadataConfigurationInput{Bucket: &bucket}); err == nil {
		out = append(out, s3BucketResource(bucket, "aws_s3_bucket_metadata_configuration"))
	}
	if lst, err := svc.ListBucketMetricsConfigurations(ctx, &s3.ListBucketMetricsConfigurationsInput{Bucket: &bucket}); err == nil {
		for _, c := range lst.MetricsConfigurationList {
			id := bucket + ":" + StringValue(c.Id)
			out = append(out, terraformutils.NewSimpleResource(id, id, "aws_s3_bucket_metric", "aws", S3AllowEmptyValues))
		}
	}
	if n, err := svc.GetBucketNotificationConfiguration(ctx, &s3.GetBucketNotificationConfigurationInput{Bucket: &bucket}); err == nil &&
		(len(n.TopicConfigurations) > 0 || len(n.QueueConfigurations) > 0 || len(n.LambdaFunctionConfigurations) > 0 || n.EventBridgeConfiguration != nil) {
		out = append(out, s3BucketResource(bucket, "aws_s3_bucket_notification"))
	}
	if _, err := svc.GetObjectLockConfiguration(ctx, &s3.GetObjectLockConfigurationInput{Bucket: &bucket}); err == nil {
		out = append(out, s3BucketResource(bucket, "aws_s3_bucket_object_lock_configuration"))
	}
	if _, err := svc.GetBucketOwnershipControls(ctx, &s3.GetBucketOwnershipControlsInput{Bucket: &bucket}); err == nil {
		out = append(out, s3BucketResource(bucket, "aws_s3_bucket_ownership_controls"))
	}
	if _, err := svc.GetPublicAccessBlock(ctx, &s3.GetPublicAccessBlockInput{Bucket: &bucket}); err == nil {
		out = append(out, s3BucketResource(bucket, "aws_s3_bucket_public_access_block"))
	}
	if _, err := svc.GetBucketReplication(ctx, &s3.GetBucketReplicationInput{Bucket: &bucket}); err == nil {
		out = append(out, s3BucketResource(bucket, "aws_s3_bucket_replication_configuration"))
	}
	if _, err := svc.GetBucketRequestPayment(ctx, &s3.GetBucketRequestPaymentInput{Bucket: &bucket}); err == nil {
		out = append(out, s3BucketResource(bucket, "aws_s3_bucket_request_payment_configuration"))
	}
	if _, err := svc.GetBucketEncryption(ctx, &s3.GetBucketEncryptionInput{Bucket: &bucket}); err == nil {
		out = append(out, s3BucketResource(bucket, "aws_s3_bucket_server_side_encryption_configuration"))
	}
	if v, err := svc.GetBucketVersioning(ctx, &s3.GetBucketVersioningInput{Bucket: &bucket}); err == nil && v.Status != "" {
		out = append(out, s3BucketResource(bucket, "aws_s3_bucket_versioning"))
	}
	if _, err := svc.GetBucketWebsite(ctx, &s3.GetBucketWebsiteInput{Bucket: &bucket}); err == nil {
		out = append(out, s3BucketResource(bucket, "aws_s3_bucket_website_configuration"))
	}
	return out
}
