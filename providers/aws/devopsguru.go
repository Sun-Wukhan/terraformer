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

// DevOps Guru generator (service key "devopsguru").
//
// Skipped types in this family (see results ledger):
//   - aws_devopsguru_event_sources_config: account/Region singleton; the
//     provider ID is the Region name and only DescribeEventSourcesConfig
//     exists.
//   - aws_devopsguru_service_integration: account/Region singleton; the
//     provider ID is the Region name and only DescribeServiceIntegration
//     exists.

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/devopsguru"
	"github.com/aws/aws-sdk-go-v2/service/devopsguru/types"
)

var devopsguruAllowEmptyValues = []string{"tags."}

type DevOpsGuruGenerator struct {
	AWSService
}

func (g *DevOpsGuruGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := devopsguru.NewFromConfig(config)
	if err := g.createNotificationChannels(svc); err != nil {
		return err
	}
	return g.createResourceCollections(svc)
}

// aws_devopsguru_notification_channel — importer ID is the channel ID.
func (g *DevOpsGuruGenerator) createNotificationChannels(svc *devopsguru.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListNotificationChannels(context.TODO(), &devopsguru.ListNotificationChannelsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, channel := range out.Channels {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(channel.Id), StringValue(channel.Id),
				"aws_devopsguru_notification_channel", "aws", devopsguruAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

// aws_devopsguru_resource_collection — importer ID is the collection type.
// Only one collection type can be configured per account; probe the three
// known types and emit the one that is actually configured (mirrors the
// closed-set enumeration used for CloudWatch Logs account policies).
func (g *DevOpsGuruGenerator) createResourceCollections(svc *devopsguru.Client) error {
	for _, collectionType := range []types.ResourceCollectionType{
		types.ResourceCollectionTypeAwsCloudFormation,
		types.ResourceCollectionTypeAwsService,
		types.ResourceCollectionTypeAwsTags,
	} {
		out, err := svc.GetResourceCollection(context.TODO(), &devopsguru.GetResourceCollectionInput{ResourceCollectionType: collectionType})
		if err != nil {
			return err
		}
		if out.ResourceCollection == nil {
			continue
		}
		configured := false
		if out.ResourceCollection.CloudFormation != nil && len(out.ResourceCollection.CloudFormation.StackNames) > 0 {
			configured = true
		}
		if len(out.ResourceCollection.Tags) > 0 {
			configured = true
		}
		if configured {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				string(collectionType), string(collectionType),
				"aws_devopsguru_resource_collection", "aws", devopsguruAllowEmptyValues))
		}
	}
	return nil
}
