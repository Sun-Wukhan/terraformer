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
	"github.com/aws/aws-sdk-go-v2/service/chimesdkmediapipelines"
)

var chimeMediaPipelinesAllowEmptyValues = []string{"tags."}

// ChimeSdkMediaPipelinesGenerator covers Chime SDK Media Pipelines.
// Media insights pipeline configurations import with their identifier
// (verified against the Terraform AWS provider docs).
type ChimeSdkMediaPipelinesGenerator struct {
	AWSService
}

func (g *ChimeSdkMediaPipelinesGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := chimesdkmediapipelines.NewFromConfig(config)

	var nextToken *string
	for {
		out, err := svc.ListMediaInsightsPipelineConfigurations(context.TODO(), &chimesdkmediapipelines.ListMediaInsightsPipelineConfigurationsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, configuration := range out.MediaInsightsPipelineConfigurations {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(configuration.MediaInsightsPipelineConfigurationId),
				StringValue(configuration.MediaInsightsPipelineConfigurationName),
				"aws_chimesdkmediapipelines_media_insights_pipeline_configuration",
				"aws",
				chimeMediaPipelinesAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
