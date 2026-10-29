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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elastictranscoder"
	"github.com/aws/aws-sdk-go-v2/service/mediaconvert"
	"github.com/aws/aws-sdk-go-v2/service/medialive"
	"github.com/aws/aws-sdk-go-v2/service/mediapackagev2"
	"github.com/aws/aws-sdk-go-v2/service/mediastore"
)

var mediaAllowEmptyValues = []string{"tags."}

// --- MediaLive multiplexes (extends the medialive service) -----------------

// Multiplexes are keyed by multiplex ID; multiplex programs by the
// verified provider importer ID "program_name/multiplex_id".
func (g *MediaLiveGenerator) loadMultiplexes(svc *medialive.Client) error {
	input := &medialive.ListMultiplexesInput{}
	for {
		out, err := svc.ListMultiplexes(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, m := range out.Multiplexes {
			multiplexID := StringValue(m.Id)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				multiplexID, StringValue(m.Name), "aws_medialive_multiplex", "aws", medialiveAllowEmptyValues))
			progInput := &medialive.ListMultiplexProgramsInput{MultiplexId: aws.String(multiplexID)}
			for {
				programs, err := svc.ListMultiplexPrograms(context.TODO(), progInput)
				if err != nil {
					return err
				}
				for _, p := range programs.MultiplexPrograms {
					g.Resources = append(g.Resources, terraformutils.NewResource(
						StringValue(p.ProgramName)+"/"+multiplexID, StringValue(p.ProgramName),
						"aws_medialive_multiplex_program", "aws",
						map[string]string{"multiplex_id": multiplexID},
						medialiveAllowEmptyValues, map[string]interface{}{}))
				}
				if programs.NextToken == nil {
					break
				}
				progInput.NextToken = programs.NextToken
			}
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

// --- MediaConvert queues (new mediaconvert service) ------------------------

type MediaConvertGenerator struct {
	AWSService
}

// MediaConvert is an account-endpoint service: the data-plane client
// must be pointed at the endpoint returned by DescribeEndpoints before
// queues can be listed. Queues are keyed by queue name (the provider
// importer ID).
func (g *MediaConvertGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	discovery := mediaconvert.NewFromConfig(config)
	endpoints, err := discovery.DescribeEndpoints(context.TODO(), &mediaconvert.DescribeEndpointsInput{})
	if err != nil {
		return err
	}
	if len(endpoints.Endpoints) == 0 || endpoints.Endpoints[0].Url == nil {
		return nil
	}
	svc := mediaconvert.New(mediaconvert.Options{
		Credentials:  config.Credentials,
		Region:       config.Region,
		BaseEndpoint: endpoints.Endpoints[0].Url,
	})
	input := &mediaconvert.ListQueuesInput{}
	for {
		out, err := svc.ListQueues(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, q := range out.Queues {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(q.Name), StringValue(q.Name),
				"aws_media_convert_queue", "aws", mediaAllowEmptyValues))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

// --- MediaPackage v2 channel groups (extends the media_package service) ----

// Channel groups are keyed by channel group name (the provider
// importer ID).
func (g *MediaPackageGenerator) loadChannelGroups(config aws.Config) error {
	svc := mediapackagev2.NewFromConfig(config)
	input := &mediapackagev2.ListChannelGroupsInput{}
	for {
		out, err := svc.ListChannelGroups(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, group := range out.Items {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(group.ChannelGroupName), StringValue(group.ChannelGroupName),
				"aws_media_packagev2_channel_group", "aws", mediapackageAllowEmptyValues))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

// --- MediaStore container policies (extends the media_store service) -------

// A container policy is keyed by container name and exists only when a
// policy was attached; GetContainerPolicy fails for containers without
// one, which is not an enumeration failure.
func (g *MediaStoreGenerator) loadContainerPolicies(svc *mediastore.Client) error {
	p := mediastore.NewListContainersPaginator(svc, &mediastore.ListContainersInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, container := range page.Containers {
			name := StringValue(container.Name)
			policy, err := svc.GetContainerPolicy(context.TODO(), &mediastore.GetContainerPolicyInput{ContainerName: aws.String(name)})
			if err == nil && policy.Policy != nil && *policy.Policy != "" {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					name, name, "aws_media_store_container_policy", "aws", mediastoreAllowEmptyValues))
			}
		}
	}
	return nil
}

// --- Elastic Transcoder (new elastictranscoder service) --------------------

type ElasticTranscoderGenerator struct {
	AWSService
}

// Pipelines and presets are keyed by their IDs (the provider importer
// IDs). Only customer presets are generated: ListPresets also returns
// the AWS-owned system presets, which cannot be managed as
// aws_elastictranscoder_preset resources for the account.
func (g *ElasticTranscoderGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := elastictranscoder.NewFromConfig(config)

	pipeInput := &elastictranscoder.ListPipelinesInput{}
	for {
		out, err := svc.ListPipelines(context.TODO(), pipeInput)
		if err != nil {
			return err
		}
		for _, p := range out.Pipelines {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(p.Id), StringValue(p.Name),
				"aws_elastictranscoder_pipeline", "aws", mediaAllowEmptyValues))
		}
		if out.NextPageToken == nil {
			break
		}
		pipeInput.PageToken = out.NextPageToken
	}
	presetInput := &elastictranscoder.ListPresetsInput{}
	for {
		out, err := svc.ListPresets(context.TODO(), presetInput)
		if err != nil {
			return err
		}
		for _, p := range out.Presets {
			if StringValue(p.Type) == "System" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(p.Id), StringValue(p.Name),
				"aws_elastictranscoder_preset", "aws", mediaAllowEmptyValues))
		}
		if out.NextPageToken == nil {
			return nil
		}
		presetInput.PageToken = out.NextPageToken
	}
}
