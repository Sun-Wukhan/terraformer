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
	"github.com/aws/aws-sdk-go-v2/service/ivs"
	"github.com/aws/aws-sdk-go-v2/service/ivschat"
)

var ivsAllowEmptyValues = []string{"tags."}

type IvsGenerator struct {
	AWSService
}

// InitResources enumerates Amazon IVS channels, playback key pairs and
// recording configurations (all keyed by ARN, the provider importer ID)
// and IVS Chat rooms (by ARN) and logging configurations (by logging
// configuration ID, from ListLoggingConfigurations summaries).
func (g *IvsGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := ivs.NewFromConfig(config)

	channelInput := &ivs.ListChannelsInput{}
	for {
		out, err := svc.ListChannels(context.TODO(), channelInput)
		if err != nil {
			return err
		}
		for _, c := range out.Channels {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(c.Arn), StringValue(c.Name), "aws_ivs_channel", "aws", ivsAllowEmptyValues))
		}
		if out.NextToken == nil {
			break
		}
		channelInput.NextToken = out.NextToken
	}
	keyInput := &ivs.ListPlaybackKeyPairsInput{}
	for {
		out, err := svc.ListPlaybackKeyPairs(context.TODO(), keyInput)
		if err != nil {
			return err
		}
		for _, k := range out.KeyPairs {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(k.Arn), StringValue(k.Name), "aws_ivs_playback_key_pair", "aws", ivsAllowEmptyValues))
		}
		if out.NextToken == nil {
			break
		}
		keyInput.NextToken = out.NextToken
	}
	recInput := &ivs.ListRecordingConfigurationsInput{}
	for {
		out, err := svc.ListRecordingConfigurations(context.TODO(), recInput)
		if err != nil {
			return err
		}
		for _, r := range out.RecordingConfigurations {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(r.Arn), StringValue(r.Name), "aws_ivs_recording_configuration", "aws", ivsAllowEmptyValues))
		}
		if out.NextToken == nil {
			break
		}
		recInput.NextToken = out.NextToken
	}
	return g.loadChat(ivschat.NewFromConfig(config))
}

func (g *IvsGenerator) loadChat(svc *ivschat.Client) error {
	roomInput := &ivschat.ListRoomsInput{}
	for {
		out, err := svc.ListRooms(context.TODO(), roomInput)
		if err != nil {
			return err
		}
		for _, r := range out.Rooms {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(r.Arn), StringValue(r.Name), "aws_ivschat_room", "aws", ivsAllowEmptyValues))
		}
		if out.NextToken == nil {
			break
		}
		roomInput.NextToken = out.NextToken
	}
	logInput := &ivschat.ListLoggingConfigurationsInput{}
	for {
		out, err := svc.ListLoggingConfigurations(context.TODO(), logInput)
		if err != nil {
			return err
		}
		for _, l := range out.LoggingConfigurations {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(l.Id), StringValue(l.Name), "aws_ivschat_logging_configuration", "aws", ivsAllowEmptyValues))
		}
		if out.NextToken == nil {
			return nil
		}
		logInput.NextToken = out.NextToken
	}
}
