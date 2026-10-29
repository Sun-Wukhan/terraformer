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
	"github.com/aws/aws-sdk-go-v2/service/chatbot"
)

var chatbotAllowEmptyValues = []string{"tags."}

// ChatbotGenerator covers AWS Chatbot. Both channel configuration resources
// import with the chat configuration ARN.
type ChatbotGenerator struct {
	AWSService
}

func (g *ChatbotGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := chatbot.NewFromConfig(config)

	slack, err := svc.DescribeSlackChannelConfigurations(context.TODO(), &chatbot.DescribeSlackChannelConfigurationsInput{})
	if err != nil {
		return err
	}
	for _, configuration := range slack.SlackChannelConfigurations {
		arn := StringValue(configuration.ChatConfigurationArn)
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			arn,
			StringValue(configuration.ConfigurationName),
			"aws_chatbot_slack_channel_configuration",
			"aws",
			chatbotAllowEmptyValues))
	}

	var nextToken *string
	for {
		teams, err := svc.ListMicrosoftTeamsChannelConfigurations(context.TODO(), &chatbot.ListMicrosoftTeamsChannelConfigurationsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, configuration := range teams.TeamChannelConfigurations {
			arn := StringValue(configuration.ChatConfigurationArn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn,
				StringValue(configuration.ConfigurationName),
				"aws_chatbot_teams_channel_configuration",
				"aws",
				chatbotAllowEmptyValues))
		}
		nextToken = teams.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
