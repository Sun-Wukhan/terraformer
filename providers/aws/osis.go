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
	"github.com/aws/aws-sdk-go-v2/service/osis"
)

var osisAllowEmptyValues = []string{"tags."}

type OsisGenerator struct {
	AWSService
}

// InitResources enumerates OpenSearch Ingestion pipelines (by pipeline
// name, the provider importer ID) and pipeline endpoints (by endpoint
// ID, from ListPipelineEndpoints).
//
// Deliberately not generated: aws_osis_resource_policy — the service
// only exposes GetResourcePolicy for a resource ARN you already know;
// there is no list API to discover which resources carry a policy (the
// same reason aws_bedrockagentcore_resource_policy is not generated).
func (g *OsisGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := osis.NewFromConfig(config)

	pipeInput := &osis.ListPipelinesInput{}
	for {
		out, err := svc.ListPipelines(context.TODO(), pipeInput)
		if err != nil {
			return err
		}
		for _, p := range out.Pipelines {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(p.PipelineName), StringValue(p.PipelineName),
				"aws_osis_pipeline", "aws", osisAllowEmptyValues))
		}
		if out.NextToken == nil {
			break
		}
		pipeInput.NextToken = out.NextToken
	}

	endpointInput := &osis.ListPipelineEndpointsInput{}
	for {
		out, err := svc.ListPipelineEndpoints(context.TODO(), endpointInput)
		if err != nil {
			return err
		}
		for _, ep := range out.PipelineEndpoints {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(ep.EndpointId), StringValue(ep.EndpointId),
				"aws_osis_pipeline_endpoint", "aws", osisAllowEmptyValues))
		}
		if out.NextToken == nil {
			return nil
		}
		endpointInput.NextToken = out.NextToken
	}
}
