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
	"github.com/aws/aws-sdk-go-v2/service/pinpoint"
)

var pinpointAllowEmptyValues = []string{"tags."}

// pinpointChannelTypes maps Pinpoint channel type keys (from GetChannels)
// to the provider resource for that channel. Each channel resource imports
// with the application ID.
var pinpointChannelTypes = map[string]string{
	"ADM":               "aws_pinpoint_adm_channel",
	"APNS":              "aws_pinpoint_apns_channel",
	"APNS_SANDBOX":      "aws_pinpoint_apns_sandbox_channel",
	"APNS_VOIP":         "aws_pinpoint_apns_voip_channel",
	"APNS_VOIP_SANDBOX": "aws_pinpoint_apns_voip_sandbox_channel",
	"BAIDU":             "aws_pinpoint_baidu_channel",
	"EMAIL":             "aws_pinpoint_email_channel",
	"GCM":               "aws_pinpoint_gcm_channel",
	"SMS":               "aws_pinpoint_sms_channel",
}

// PinpointGenerator covers Amazon Pinpoint (aws_pinpoint_*).
type PinpointGenerator struct {
	AWSService
}

func (g *PinpointGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := pinpoint.NewFromConfig(config)

	var token *string
	for {
		page, err := svc.GetApps(context.TODO(), &pinpoint.GetAppsInput{Token: token})
		if err != nil {
			return err
		}
		for _, app := range page.ApplicationsResponse.Item {
			appID := StringValue(app.Id)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				appID,
				StringValue(app.Name),
				"aws_pinpoint_app",
				"aws",
				pinpointAllowEmptyValues))
			if err := g.loadPinpointAppChildren(svc, appID); err != nil {
				return err
			}
		}
		token = page.ApplicationsResponse.NextToken
		if token == nil {
			break
		}
	}

	if err := g.loadPinpointEmailTemplates(svc); err != nil {
		return err
	}
	return nil
}

func (g *PinpointGenerator) loadPinpointAppChildren(svc *pinpoint.Client, appID string) error {
	channels, err := svc.GetChannels(context.TODO(), &pinpoint.GetChannelsInput{ApplicationId: &appID})
	if err != nil {
		return err
	}
	for channelType := range channels.ChannelsResponse.Channels {
		if resourceType, ok := pinpointChannelTypes[channelType]; ok {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				appID,
				appID,
				resourceType,
				"aws",
				pinpointAllowEmptyValues))
		}
	}

	stream, err := svc.GetEventStream(context.TODO(), &pinpoint.GetEventStreamInput{ApplicationId: &appID})
	if err != nil {
		return err
	}
	if stream.EventStream != nil && StringValue(stream.EventStream.DestinationStreamArn) != "" {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			appID,
			appID,
			"aws_pinpoint_event_stream",
			"aws",
			pinpointAllowEmptyValues))
	}
	return nil
}

func (g *PinpointGenerator) loadPinpointEmailTemplates(svc *pinpoint.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListTemplates(context.TODO(), &pinpoint.ListTemplatesInput{
			NextToken:    nextToken,
			TemplateType: aws.String("EMAIL"),
		})
		if err != nil {
			return err
		}
		for _, template := range out.TemplatesResponse.Item {
			name := StringValue(template.TemplateName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name,
				name,
				"aws_pinpoint_email_template",
				"aws",
				pinpointAllowEmptyValues))
		}
		nextToken = out.TemplatesResponse.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
