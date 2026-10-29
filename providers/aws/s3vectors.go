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
	"github.com/aws/aws-sdk-go-v2/service/s3vectors"
)

// S3 Vectors sweep. Importer IDs verified against provider docs:
// vector bucket = its ARN; index = its index ARN; vector bucket policy =
// the vector bucket ARN (per the provider docs' import section).

var s3VectorsAllowEmptyValues = []string{"tags."}

type S3VectorsGenerator struct {
	AWSService
}

func (g *S3VectorsGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := s3vectors.NewFromConfig(config)
	ctx := context.TODO()
	add := func(id, name, typ string) {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(id, name, typ, "aws", s3VectorsAllowEmptyValues))
	}
	input := &s3vectors.ListVectorBucketsInput{}
	for {
		out, err := svc.ListVectorBuckets(ctx, input)
		if err != nil {
			return err
		}
		for _, bucket := range out.VectorBuckets {
			add(StringValue(bucket.VectorBucketArn), StringValue(bucket.VectorBucketName), "aws_s3vectors_vector_bucket")
			if _, err := svc.GetVectorBucketPolicy(ctx, &s3vectors.GetVectorBucketPolicyInput{VectorBucketName: bucket.VectorBucketName}); err == nil {
				add(StringValue(bucket.VectorBucketArn), StringValue(bucket.VectorBucketName), "aws_s3vectors_vector_bucket_policy")
			}
			idxInput := &s3vectors.ListIndexesInput{VectorBucketName: bucket.VectorBucketName}
			for {
				idxOut, err := svc.ListIndexes(ctx, idxInput)
				if err != nil {
					break
				}
				for _, idx := range idxOut.Indexes {
					add(StringValue(idx.IndexArn), StringValue(idx.IndexName), "aws_s3vectors_index")
				}
				if idxOut.NextToken == nil {
					break
				}
				idxInput.NextToken = idxOut.NextToken
			}
		}
		if out.NextToken == nil {
			break
		}
		input.NextToken = out.NextToken
	}
	return nil
}
