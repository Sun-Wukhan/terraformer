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
	"github.com/aws/aws-sdk-go-v2/service/notifications"
	"github.com/aws/aws-sdk-go-v2/service/notificationscontacts"
)

var notificationsAllowEmptyValues = []string{"tags."}

// NotificationsGenerator covers AWS User Notifications configurations, event rules,
// channel associations, organizational-unit associations, managed-notification
// additional channel associations and notification hubs.
// Provider import IDs (verified): configuration = ARN; event rule = ARN;
// channel association = "<configuration ARN>,<channel ARN>";
// OU association = "<configuration ARN>,<OU ID>"; managed additional channel
// association = "<managed configuration ARN>,<channel ARN>"; hub = region name.
// Skipped in this family:
//   - aws_notifications_managed_notification_account_contact_association: association
//     glue with no list API in the notifications SDK (only Associate/Disassociate
//     operations exist), so existing associations are not enumerable.
//   - aws_notifications_organizations_access: singleton account-level Organizations
//     access toggle; provider ID is the account ID, with no list API.
type NotificationsGenerator struct {
	AWSService
}

func (g *NotificationsGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := notifications.NewFromConfig(config)

	hp := notifications.NewListNotificationHubsPaginator(svc, &notifications.ListNotificationHubsInput{})
	for hp.HasMorePages() {
		page, err := hp.NextPage(context.TODO())
		if err != nil {
			break
		}
		for _, hub := range page.NotificationHubs {
			region := StringValue(hub.NotificationHubRegion)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				region, region, "aws_notifications_notification_hub", "aws", notificationsAllowEmptyValues))
		}
	}

	p := notifications.NewListNotificationConfigurationsPaginator(svc, &notifications.ListNotificationConfigurationsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, cfg := range page.NotificationConfigurations {
			arn := StringValue(cfg.Arn)
			name := StringValue(cfg.Name)
			if name == "" {
				name = arn
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, name, "aws_notifications_notification_configuration", "aws", notificationsAllowEmptyValues))

			ep := notifications.NewListEventRulesPaginator(svc, &notifications.ListEventRulesInput{NotificationConfigurationArn: cfg.Arn})
			for ep.HasMorePages() {
				epage, err := ep.NextPage(context.TODO())
				if err != nil {
					break
				}
				for _, rule := range epage.EventRules {
					ruleARN := StringValue(rule.Arn)
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						ruleARN, ruleARN, "aws_notifications_event_rule", "aws", notificationsAllowEmptyValues))
				}
			}

			channels, err := svc.ListChannels(context.TODO(), &notifications.ListChannelsInput{NotificationConfigurationArn: cfg.Arn})
			if err == nil {
				for _, channelARN := range channels.Channels {
					id := arn + "," + channelARN
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						id, id, "aws_notifications_channel_association", "aws", notificationsAllowEmptyValues))
				}
			}

			ous, err := svc.ListOrganizationalUnits(context.TODO(), &notifications.ListOrganizationalUnitsInput{NotificationConfigurationArn: cfg.Arn})
			if err == nil {
				for _, ouID := range ous.OrganizationalUnits {
					id := arn + "," + ouID
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						id, id, "aws_notifications_organizational_unit_association", "aws", notificationsAllowEmptyValues))
				}
			}
		}
	}

	mp := notifications.NewListManagedNotificationConfigurationsPaginator(svc, &notifications.ListManagedNotificationConfigurationsInput{})
	for mp.HasMorePages() {
		page, err := mp.NextPage(context.TODO())
		if err != nil {
			break
		}
		for _, mcfg := range page.ManagedNotificationConfigurations {
			mARN := StringValue(mcfg.Arn)
			assocs, err := svc.ListManagedNotificationChannelAssociations(context.TODO(),
				&notifications.ListManagedNotificationChannelAssociationsInput{ManagedNotificationConfigurationArn: mcfg.Arn})
			if err != nil {
				continue
			}
			for _, assoc := range assocs.ChannelAssociations {
				id := mARN + "," + StringValue(assoc.ChannelIdentifier)
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, id, "aws_notifications_managed_notification_additional_channel_association", "aws", notificationsAllowEmptyValues))
			}
		}
	}
	return nil
}

// NotificationsContactsGenerator covers User Notifications Contacts email contacts.
// Provider import ID (verified): email contact ARN.
type NotificationsContactsGenerator struct {
	AWSService
}

func (g *NotificationsContactsGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := notificationscontacts.NewFromConfig(config)
	p := notificationscontacts.NewListEmailContactsPaginator(svc, &notificationscontacts.ListEmailContactsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, contact := range page.EmailContacts {
			arn := StringValue(contact.Arn)
			name := StringValue(contact.Name)
			if name == "" {
				name = arn
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, name, "aws_notificationscontacts_email_contact", "aws", notificationsAllowEmptyValues))
		}
	}
	return nil
}
