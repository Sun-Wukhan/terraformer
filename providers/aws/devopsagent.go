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

// AWS DevOps Agent generator (service key "devopsagent").

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/devopsagent"
)

var devopsagentAllowEmptyValues = []string{"tags."}

type DevOpsAgentGenerator struct {
	AWSService
}

func (g *DevOpsAgentGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := devopsagent.NewFromConfig(config)
	var nextToken *string
	for {
		out, err := svc.ListAgentSpaces(context.TODO(), &devopsagent.ListAgentSpacesInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, space := range out.AgentSpaces {
			spaceID := StringValue(space.AgentSpaceId)
			// aws_devopsagent_agent_space — importer ID is the agent space ID.
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				spaceID, StringValue(space.Name),
				"aws_devopsagent_agent_space", "aws", devopsagentAllowEmptyValues))
			if err := g.createAssets(svc, spaceID); err != nil {
				return err
			}
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

// aws_devopsagent_asset — importer ID "agent_space_id,asset_id".
func (g *DevOpsAgentGenerator) createAssets(svc *devopsagent.Client, agentSpaceID string) error {
	var nextToken *string
	for {
		out, err := svc.ListAssets(context.TODO(), &devopsagent.ListAssetsInput{AgentSpaceId: &agentSpaceID, NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, asset := range out.Items {
			id := agentSpaceID + "," + StringValue(asset.AssetId)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				id, StringValue(asset.AssetId), "aws_devopsagent_asset", "aws",
				map[string]string{"agent_space_id": agentSpaceID}, devopsagentAllowEmptyValues, map[string]interface{}{}))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
