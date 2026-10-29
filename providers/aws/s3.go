// Copyright 2018 The Terraformer Authors.
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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var S3AllowEmptyValues = []string{"tags."}

var S3AdditionalFields = map[string]interface{}{}

type S3Generator struct {
	AWSService
}

// createResources iterate on all buckets
// for each bucket we check region and choose only bucket from set region
// for each bucket try get bucket policy, if policy exist create additional NewTerraformResource for policy
func (g *S3Generator) createResources(config aws.Config, buckets *s3.ListBucketsOutput, region string) []terraformutils.Resource {
	var resources []terraformutils.Resource
	svc := s3.NewFromConfig(config)
	for _, bucket := range buckets.Buckets {
		resourceName := StringValue(bucket.Name)
		location, err := svc.GetBucketLocation(context.TODO(), &s3.GetBucketLocationInput{Bucket: bucket.Name})
		if err != nil {
			log.Println(err)
			continue
		}
		// check if bucket in region
		constraintString := string(location.LocationConstraint)
		if constraintString == region || (constraintString == "" && region == "us-east-1") {
			attributes := map[string]string{
				"force_destroy": "false",
				"acl":           "private",
			}
			// try get policy
			var policy *s3.GetBucketPolicyOutput
			policy, err = svc.GetBucketPolicy(context.TODO(), &s3.GetBucketPolicyInput{
				Bucket: bucket.Name,
			})

			if err == nil && policy.Policy != nil {
				attributes["policy"] = *policy.Policy
				resources = append(resources, terraformutils.NewResource(
					resourceName,
					resourceName,
					"aws_s3_bucket_policy",
					"aws",
					nil,
					S3AllowEmptyValues,
					S3AdditionalFields))
			}
			resources = append(resources, terraformutils.NewResource(
				resourceName,
				resourceName,
				"aws_s3_bucket",
				"aws",
				attributes,
				S3AllowEmptyValues,
				S3AdditionalFields))
			// AWS storage sweep: per-bucket sub-configuration resources
			// (see s3_subconfig.go).
			resources = append(resources, g.bucketSubConfigs(svc, resourceName)...)
		}
	}
	return resources
}

// Generate TerraformResources from AWS API,
// Need bucket name as ID for terraform resource
func (g *S3Generator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := s3.NewFromConfig(config)

	buckets, err := svc.ListBuckets(context.TODO(), nil)
	if err != nil {
		return err
	}
	g.Resources = g.createResources(config, buckets, g.GetArgs()["region"].(string))
	g.addDirectoryBuckets(svc)
	return nil
}

// addDirectoryBuckets enumerates S3 Express One Zone directory buckets in the
// selected region. Directory buckets are an optional add-on to the core bucket
// listing, so a failure here is logged rather than failing the whole S3 import.
func (g *S3Generator) addDirectoryBuckets(svc *s3.Client) {
	input := &s3.ListDirectoryBucketsInput{}
	for {
		out, err := svc.ListDirectoryBuckets(context.TODO(), input)
		if err != nil {
			log.Println("listing S3 directory buckets:", err)
			return
		}
		for _, bucket := range out.Buckets {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*bucket.Name,
				*bucket.Name,
				"aws_s3_directory_bucket",
				"aws",
				S3AllowEmptyValues,
			))
		}
		if out.ContinuationToken == nil {
			return
		}
		input.ContinuationToken = out.ContinuationToken
	}
}

// PostGenerateHook for add bucket policy json as heredoc
// support only bucket with policy
func (g *S3Generator) PostConvertHook() error {
	for i, resource := range g.Resources {
		if resource.InstanceInfo.Type == "aws_s3_bucket" {
			if val, ok := g.Resources[i].Item["acl"]; ok && val == "private" {
				delete(g.Resources[i].Item, "acl")
			}
			if val, ok := g.Resources[i].Item["policy"]; ok {
				g.Resources[i].Item["policy"] = fmt.Sprintf(`<<POLICY
%s
POLICY`, g.escapeAwsInterpolation(val.(string)))
			}
		}
	}
	return nil
}
