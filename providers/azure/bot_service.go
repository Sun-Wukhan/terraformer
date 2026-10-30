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

package azure

import (
	"context"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/botservice/armbotservice"
)

// BotGenerator covers Microsoft.BotService bots, their channels and their
// OAuth connections.
//
// Skipped (see results-containers.md):
//   - azurerm_bot_channels_registration: legacy resource for bots of kind
//     "registration"; the track-2 armbotservice Kind enum (azurebot, bot,
//     designer, function, sdk) has no registration kind and BotsClient has
//     no separate List for registrations, so they cannot be distinguished
//     from Azure bots.
//   - azurerm_bot_web_app: legacy resource for bots of kind "webapp"; same
//     reason — no webapp kind in the track-2 Kind enum and no separate List.
type BotGenerator struct {
	AzureService
}

// botChannelTerraformType maps the channel name reported by the Bot Service
// API (e.g. "FacebookChannel") to its Terraform type. Channels without a
// dedicated Terraform type in this sweep return "".
func botChannelTerraformType(channelName string) string {
	switch channelName {
	case "AlexaChannel":
		return "azurerm_bot_channel_alexa"
	case "DirectLineSpeechChannel":
		return "azurerm_bot_channel_direct_line_speech"
	case "DirectLineChannel":
		return "azurerm_bot_channel_directline"
	case "EmailChannel":
		return "azurerm_bot_channel_email"
	case "FacebookChannel":
		return "azurerm_bot_channel_facebook"
	case "LineChannel":
		return "azurerm_bot_channel_line"
	case "MsTeamsChannel":
		return "azurerm_bot_channel_ms_teams"
	case "SlackChannel":
		return "azurerm_bot_channel_slack"
	case "SmsChannel":
		return "azurerm_bot_channel_sms"
	case "WebChatChannel":
		return "azurerm_bot_channel_web_chat"
	}
	return ""
}

func (g *BotGenerator) InitResources() error {
	ctx := context.Background()
	botsClient, err := armbotservice.NewBotsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	var bots []*armbotservice.Bot
	if rg := g.getResourceGroup(); rg != "" {
		pager := botsClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			bots = append(bots, page.Value...)
		}
	} else {
		pager := botsClient.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			bots = append(bots, page.Value...)
		}
	}

	channelsClient, err := armbotservice.NewChannelsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	connectionsClient, err := armbotservice.NewBotConnectionClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}

	for _, bot := range bots {
		g.AppendSimpleResource(*bot.ID, *bot.Name, "azurerm_bot_service_azure_bot")
		parsed, err := ParseAzureResourceID(*bot.ID)
		if err != nil {
			return err
		}
		rg, botName := parsed.ResourceGroup, *bot.Name

		channelPager := channelsClient.NewListByResourceGroupPager(rg, botName, nil)
		for channelPager.More() {
			page, err := channelPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, channel := range page.Value {
				if channel.Properties == nil || channel.Properties.GetChannel() == nil ||
					channel.Properties.GetChannel().ChannelName == nil {
					continue
				}
				resourceType := botChannelTerraformType(*channel.Properties.GetChannel().ChannelName)
				if resourceType == "" {
					continue
				}
				g.AppendSimpleResource(*channel.ID, *channel.Name, resourceType)
			}
		}

		connectionPager := connectionsClient.NewListByBotServicePager(rg, botName, nil)
		for connectionPager.More() {
			page, err := connectionPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_bot_connection")
			}
		}
	}
	return nil
}
