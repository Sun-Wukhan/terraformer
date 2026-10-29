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
	"github.com/aws/aws-sdk-go-v2/service/s3outposts"
)

// S3 on Outposts sweep. Importer ID verified against provider docs:
// "endpoint-arn,security-group-id,subnet-id" (comma-separated).

var s3OutpostsAllowEmptyValues = []string{"tags."}

type S3OutpostsGenerator struct {
	AWSService
}

func (g *S3OutpostsGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := s3outposts.NewFromConfig(config)
	ctx := context.TODO()
	input := &s3outposts.ListEndpointsInput{}
	for {
		out, err := svc.ListEndpoints(ctx, input)
		if err != nil {
			return err
		}
		for _, ep := range out.Endpoints {
			id := StringValue(ep.EndpointArn) + "," + StringValue(ep.SecurityGroupId) + "," + StringValue(ep.SubnetId)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(id, StringValue(ep.EndpointArn), "aws_s3outposts_endpoint", "aws", s3OutpostsAllowEmptyValues))
		}
		if out.NextToken == nil {
			break
		}
		input.NextToken = out.NextToken
	}
	return nil
}
