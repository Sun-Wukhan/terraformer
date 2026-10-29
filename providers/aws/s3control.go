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
	"fmt"
	"log"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/s3control"
)

// S3 Control sweep (storage cluster).
//
// Importer IDs verified against the Terraform AWS provider docs:
//   - aws_s3_access_point, aws_s3control_access_point_policy,
//     aws_s3control_multi_region_access_point and its policy: "account:name"
//   - aws_s3control_access_grant / access_grants_location: "account,id" (comma)
//   - aws_s3control_access_grants_instance: account ID (enumerated via
//     ListAccessGrantsInstances, so it is discovered, not assumed)
//   - aws_s3control_access_grants_instance_resource_policy: account ID,
//     emitted per listed instance only when a resource policy exists
//   - aws_s3control_bucket (+ lifecycle / policy): the Outposts bucket ARN
//   - aws_s3control_directory_bucket_access_point_scope: "name,account"
//     (comma, per the provider docs' terraform import example), emitted per
//     directory-bucket access point only when GetAccessPointScope succeeds
//   - aws_s3control_multi_region_access_point_routes: the MRAP ARN
//   - aws_s3control_object_lambda_access_point (+ policy): "account:name"
//   - aws_s3control_storage_lens_configuration: "account:config-id"

var s3ControlAllowEmptyValues = []string{"tags."}

type S3ControlGenerator struct {
	AWSService
}

func (g *S3ControlGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	acctPtr, err := g.getAccountNumber(config)
	if err != nil {
		return err
	}
	acct := StringValue(acctPtr)
	svc := s3control.NewFromConfig(config)
	ctx := context.TODO()
	add := func(id, name, typ string) {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(id, name, typ, "aws", s3ControlAllowEmptyValues))
	}

	// Access points (general purpose / directory / Outposts buckets).
	p := s3control.NewListAccessPointsPaginator(svc, &s3control.ListAccessPointsInput{AccountId: &acct})
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, ap := range page.AccessPointList {
			id := acct + ":" + StringValue(ap.Name)
			add(id, id, "aws_s3_access_point")
			if _, err := svc.GetAccessPointPolicy(ctx, &s3control.GetAccessPointPolicyInput{AccountId: &acct, Name: ap.Name}); err == nil {
				add(id, id, "aws_s3control_access_point_policy")
			}
		}
	}
	// Directory-bucket access points: scope sub-resource.
	if dp := s3control.NewListAccessPointsForDirectoryBucketsPaginator(svc, &s3control.ListAccessPointsForDirectoryBucketsInput{AccountId: &acct}); dp != nil {
		for dp.HasMorePages() {
			page, err := dp.NextPage(ctx)
			if err != nil {
				log.Println("listing directory bucket access points:", err)
				break
			}
			for _, ap := range page.AccessPointList {
				if _, err := svc.GetAccessPointScope(ctx, &s3control.GetAccessPointScopeInput{AccountId: &acct, Name: ap.Name}); err == nil {
					id := StringValue(ap.Name) + "," + acct
					add(id, id, "aws_s3control_directory_bucket_access_point_scope")
				}
			}
		}
	}
	// Object Lambda access points.
	olInput := &s3control.ListAccessPointsForObjectLambdaInput{AccountId: &acct}
	for {
		out, err := svc.ListAccessPointsForObjectLambda(ctx, olInput)
		if err != nil {
			break
		}
		for _, ap := range out.ObjectLambdaAccessPointList {
			id := acct + ":" + StringValue(ap.Name)
			add(id, id, "aws_s3control_object_lambda_access_point")
			if _, err := svc.GetAccessPointPolicyForObjectLambda(ctx, &s3control.GetAccessPointPolicyForObjectLambdaInput{AccountId: &acct, Name: ap.Name}); err == nil {
				add(id, id, "aws_s3control_object_lambda_access_point_policy")
			}
		}
		if out.NextToken == nil {
			break
		}
		olInput.NextToken = out.NextToken
	}
	// S3 on Outposts buckets.
	rbInput := &s3control.ListRegionalBucketsInput{AccountId: &acct}
	for {
		out, err := svc.ListRegionalBuckets(ctx, rbInput)
		if err != nil {
			break
		}
		for _, b := range out.RegionalBucketList {
			arn := StringValue(b.BucketArn)
			if arn == "" {
				continue
			}
			add(arn, StringValue(b.Bucket), "aws_s3control_bucket")
			if _, err := svc.GetBucketLifecycleConfiguration(ctx, &s3control.GetBucketLifecycleConfigurationInput{AccountId: &acct, Bucket: &arn}); err == nil {
				add(arn, StringValue(b.Bucket), "aws_s3control_bucket_lifecycle_configuration")
			}
			if _, err := svc.GetBucketPolicy(ctx, &s3control.GetBucketPolicyInput{AccountId: &acct, Bucket: &arn}); err == nil {
				add(arn, StringValue(b.Bucket), "aws_s3control_bucket_policy")
			}
		}
		if out.NextToken == nil {
			break
		}
		rbInput.NextToken = out.NextToken
	}
	// Multi-Region Access Points.
	mrInput := &s3control.ListMultiRegionAccessPointsInput{AccountId: &acct}
	for {
		out, err := svc.ListMultiRegionAccessPoints(ctx, mrInput)
		if err != nil {
			break
		}
		for _, mr := range out.AccessPoints {
			id := acct + ":" + StringValue(mr.Name)
			add(id, id, "aws_s3control_multi_region_access_point")
			if _, err := svc.GetMultiRegionAccessPointPolicy(ctx, &s3control.GetMultiRegionAccessPointPolicyInput{AccountId: &acct, Name: mr.Name}); err == nil {
				add(id, id, "aws_s3control_multi_region_access_point_policy")
			}
			mrapARN := fmt.Sprintf("arn:aws:s3::%s:accesspoint/%s", acct, StringValue(mr.Alias))
			if routes, err := svc.GetMultiRegionAccessPointRoutes(ctx, &s3control.GetMultiRegionAccessPointRoutesInput{AccountId: &acct, Mrap: &mrapARN}); err == nil && len(routes.Routes) > 0 {
				add(mrapARN, StringValue(mr.Name), "aws_s3control_multi_region_access_point_routes")
			}
		}
		if out.NextToken == nil {
			break
		}
		mrInput.NextToken = out.NextToken
	}
	// Access Grants.
	agInput := &s3control.ListAccessGrantsInput{AccountId: &acct}
	for {
		out, err := svc.ListAccessGrants(ctx, agInput)
		if err != nil {
			break
		}
		for _, grant := range out.AccessGrantsList {
			id := acct + "," + StringValue(grant.AccessGrantId)
			add(id, id, "aws_s3control_access_grant")
		}
		if out.NextToken == nil {
			break
		}
		agInput.NextToken = out.NextToken
	}
	agiInput := &s3control.ListAccessGrantsInstancesInput{AccountId: &acct}
	for {
		out, err := svc.ListAccessGrantsInstances(ctx, agiInput)
		if err != nil {
			break
		}
		for _, inst := range out.AccessGrantsInstancesList {
			_ = inst
			// Provider ID for the instance (and its resource policy) is the
			// account ID; existence is discovered from this list call.
			add(acct, acct, "aws_s3control_access_grants_instance")
			if _, err := svc.GetAccessGrantsInstanceResourcePolicy(ctx, &s3control.GetAccessGrantsInstanceResourcePolicyInput{AccountId: &acct}); err == nil {
				add(acct, acct, "aws_s3control_access_grants_instance_resource_policy")
			}
		}
		if out.NextToken == nil {
			break
		}
		agiInput.NextToken = out.NextToken
	}
	aglInput := &s3control.ListAccessGrantsLocationsInput{AccountId: &acct}
	for {
		out, err := svc.ListAccessGrantsLocations(ctx, aglInput)
		if err != nil {
			break
		}
		for _, loc := range out.AccessGrantsLocationsList {
			id := acct + "," + StringValue(loc.AccessGrantsLocationId)
			add(id, id, "aws_s3control_access_grants_location")
		}
		if out.NextToken == nil {
			break
		}
		aglInput.NextToken = out.NextToken
	}
	// Storage Lens configurations.
	slInput := &s3control.ListStorageLensConfigurationsInput{AccountId: &acct}
	for {
		out, err := svc.ListStorageLensConfigurations(ctx, slInput)
		if err != nil {
			break
		}
		for _, cfg := range out.StorageLensConfigurationList {
			id := acct + ":" + StringValue(cfg.Id)
			add(id, id, "aws_s3control_storage_lens_configuration")
		}
		if out.NextToken == nil {
			break
		}
		slInput.NextToken = out.NextToken
	}
	return nil
}
