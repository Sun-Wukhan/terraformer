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
	"github.com/aws/aws-sdk-go-v2/service/workspacesweb"
	"github.com/aws/aws-sdk-go-v2/service/workspacesweb/types"
)

var workspacesWebAllowEmptyValues = []string{"tags."}

type WorkspacesWebGenerator struct {
	AWSService
}

// InitResources enumerates the Amazon WorkSpaces Web objects in the
// configured region: portals with their identity providers and their
// eight associated settings resources (browser, data protection, IP
// access, network and user settings, session logger, trust store and
// user access logging), plus each settings object on its own. The list
// summaries for several settings types carry only an ARN, so those
// resources are named after the ARN's final segment; the associations
// are read off the portal summaries, which carry every associated
// settings ARN.
func (g *WorkspacesWebGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	client := workspacesweb.NewFromConfig(config)

	for _, add := range []func(*workspacesweb.Client) error{
		g.addPortals,
		g.addBrowserSettings,
		g.addDataProtectionSettings,
		g.addIpAccessSettings,
		g.addNetworkSettings,
		g.addUserSettings,
		g.addUserAccessLoggingSettings,
		g.addSessionLoggers,
		g.addTrustStores,
	} {
		if err := add(client); err != nil {
			return err
		}
	}
	return nil
}

func (g *WorkspacesWebGenerator) addPortals(client *workspacesweb.Client) error {
	paginator := workspacesweb.NewListPortalsPaginator(client, &workspacesweb.ListPortalsInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, portal := range page.Portals {
			portalArn := *portal.PortalArn
			name := arnResourceName(portalArn)
			if portal.DisplayName != nil && *portal.DisplayName != "" {
				name = *portal.DisplayName
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				portalArn,
				name,
				"aws_workspacesweb_portal",
				"aws",
				workspacesWebAllowEmptyValues,
			))
			g.addPortalAssociations(portal)
			if err := g.addIdentityProviders(client, portalArn); err != nil {
				return err
			}
		}
	}
	return nil
}

// Provider ID format: associated settings ARN, portal ARN.
func (g *WorkspacesWebGenerator) addPortalAssociations(portal types.PortalSummary) {
	portalArn := *portal.PortalArn
	for _, association := range []struct {
		settingsArn  *string
		attribute    string
		resourceType string
	}{
		{portal.BrowserSettingsArn, "browser_settings_arn", "aws_workspacesweb_browser_settings_association"},
		{portal.DataProtectionSettingsArn, "data_protection_settings_arn", "aws_workspacesweb_data_protection_settings_association"},
		{portal.IpAccessSettingsArn, "ip_access_settings_arn", "aws_workspacesweb_ip_access_settings_association"},
		{portal.NetworkSettingsArn, "network_settings_arn", "aws_workspacesweb_network_settings_association"},
		{portal.SessionLoggerArn, "session_logger_arn", "aws_workspacesweb_session_logger_association"},
		{portal.TrustStoreArn, "trust_store_arn", "aws_workspacesweb_trust_store_association"},
		{portal.UserAccessLoggingSettingsArn, "user_access_logging_settings_arn", "aws_workspacesweb_user_access_logging_settings_association"},
		{portal.UserSettingsArn, "user_settings_arn", "aws_workspacesweb_user_settings_association"},
	} {
		if association.settingsArn == nil || *association.settingsArn == "" {
			continue
		}
		g.Resources = append(g.Resources, terraformutils.NewResource(
			*association.settingsArn+","+portalArn,
			arnResourceName(*association.settingsArn),
			association.resourceType,
			"aws",
			map[string]string{
				"portal_arn":          portalArn,
				association.attribute: *association.settingsArn,
			},
			workspacesWebAllowEmptyValues,
			map[string]interface{}{},
		))
	}
}

func (g *WorkspacesWebGenerator) addIdentityProviders(client *workspacesweb.Client, portalArn string) error {
	paginator := workspacesweb.NewListIdentityProvidersPaginator(client, &workspacesweb.ListIdentityProvidersInput{
		PortalArn: aws.String(portalArn),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, provider := range page.IdentityProviders {
			name := arnResourceName(*provider.IdentityProviderArn)
			if provider.IdentityProviderName != nil && *provider.IdentityProviderName != "" {
				name = *provider.IdentityProviderName
			}
			g.Resources = append(g.Resources, terraformutils.NewResource(
				*provider.IdentityProviderArn,
				name,
				"aws_workspacesweb_identity_provider",
				"aws",
				map[string]string{"portal_arn": portalArn},
				workspacesWebAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

func (g *WorkspacesWebGenerator) addBrowserSettings(client *workspacesweb.Client) error {
	paginator := workspacesweb.NewListBrowserSettingsPaginator(client, &workspacesweb.ListBrowserSettingsInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, settings := range page.BrowserSettings {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*settings.BrowserSettingsArn,
				arnResourceName(*settings.BrowserSettingsArn),
				"aws_workspacesweb_browser_settings",
				"aws",
				workspacesWebAllowEmptyValues,
			))
		}
	}
	return nil
}

func (g *WorkspacesWebGenerator) addDataProtectionSettings(client *workspacesweb.Client) error {
	paginator := workspacesweb.NewListDataProtectionSettingsPaginator(client, &workspacesweb.ListDataProtectionSettingsInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, settings := range page.DataProtectionSettings {
			name := arnResourceName(*settings.DataProtectionSettingsArn)
			if settings.DisplayName != nil && *settings.DisplayName != "" {
				name = *settings.DisplayName
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*settings.DataProtectionSettingsArn,
				name,
				"aws_workspacesweb_data_protection_settings",
				"aws",
				workspacesWebAllowEmptyValues,
			))
		}
	}
	return nil
}

func (g *WorkspacesWebGenerator) addIpAccessSettings(client *workspacesweb.Client) error {
	paginator := workspacesweb.NewListIpAccessSettingsPaginator(client, &workspacesweb.ListIpAccessSettingsInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, settings := range page.IpAccessSettings {
			name := arnResourceName(*settings.IpAccessSettingsArn)
			if settings.DisplayName != nil && *settings.DisplayName != "" {
				name = *settings.DisplayName
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*settings.IpAccessSettingsArn,
				name,
				"aws_workspacesweb_ip_access_settings",
				"aws",
				workspacesWebAllowEmptyValues,
			))
		}
	}
	return nil
}

func (g *WorkspacesWebGenerator) addNetworkSettings(client *workspacesweb.Client) error {
	paginator := workspacesweb.NewListNetworkSettingsPaginator(client, &workspacesweb.ListNetworkSettingsInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, settings := range page.NetworkSettings {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*settings.NetworkSettingsArn,
				arnResourceName(*settings.NetworkSettingsArn),
				"aws_workspacesweb_network_settings",
				"aws",
				workspacesWebAllowEmptyValues,
			))
		}
	}
	return nil
}

func (g *WorkspacesWebGenerator) addUserSettings(client *workspacesweb.Client) error {
	paginator := workspacesweb.NewListUserSettingsPaginator(client, &workspacesweb.ListUserSettingsInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, settings := range page.UserSettings {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*settings.UserSettingsArn,
				arnResourceName(*settings.UserSettingsArn),
				"aws_workspacesweb_user_settings",
				"aws",
				workspacesWebAllowEmptyValues,
			))
		}
	}
	return nil
}

func (g *WorkspacesWebGenerator) addUserAccessLoggingSettings(client *workspacesweb.Client) error {
	paginator := workspacesweb.NewListUserAccessLoggingSettingsPaginator(client, &workspacesweb.ListUserAccessLoggingSettingsInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, settings := range page.UserAccessLoggingSettings {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*settings.UserAccessLoggingSettingsArn,
				arnResourceName(*settings.UserAccessLoggingSettingsArn),
				"aws_workspacesweb_user_access_logging_settings",
				"aws",
				workspacesWebAllowEmptyValues,
			))
		}
	}
	return nil
}

func (g *WorkspacesWebGenerator) addSessionLoggers(client *workspacesweb.Client) error {
	paginator := workspacesweb.NewListSessionLoggersPaginator(client, &workspacesweb.ListSessionLoggersInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, logger := range page.SessionLoggers {
			name := arnResourceName(*logger.SessionLoggerArn)
			if logger.DisplayName != nil && *logger.DisplayName != "" {
				name = *logger.DisplayName
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*logger.SessionLoggerArn,
				name,
				"aws_workspacesweb_session_logger",
				"aws",
				workspacesWebAllowEmptyValues,
			))
		}
	}
	return nil
}

func (g *WorkspacesWebGenerator) addTrustStores(client *workspacesweb.Client) error {
	paginator := workspacesweb.NewListTrustStoresPaginator(client, &workspacesweb.ListTrustStoresInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, store := range page.TrustStores {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*store.TrustStoreArn,
				arnResourceName(*store.TrustStoreArn),
				"aws_workspacesweb_trust_store",
				"aws",
				workspacesWebAllowEmptyValues,
			))
		}
	}
	return nil
}
