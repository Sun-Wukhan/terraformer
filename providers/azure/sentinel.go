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
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/operationalinsights/armoperationalinsights"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/securityinsights/armsecurityinsights"
)

// SentinelGenerator covers Microsoft Sentinel resources, which are children
// of Log Analytics workspaces: workspaces are enumerated first, then each
// workspace's Sentinel collections are listed.
// Skipped in this family:
//   - azurerm_sentinel_alert_rule_anomaly_built_in and
//     azurerm_sentinel_alert_rule_anomaly_duplicate: these are
//     securityMLAnalyticsSettings resources, and no client/List for that
//     collection exists in the track-2 armsecurityinsights v1.2.0 module.
//   - azurerm_sentinel_metadata: no Metadata client/List exists in the
//     track-2 armsecurityinsights v1.2.0 module.
type SentinelGenerator struct {
	AzureService
}

// sentinelAlertRuleType maps the ARM alert rule kind discriminator to the
// azurerm resource type the provider splits alert rules into.
func sentinelAlertRuleType(kind string) string {
	switch kind {
	case "Scheduled":
		return "azurerm_sentinel_alert_rule_scheduled"
	case "MicrosoftSecurityIncidentCreation":
		return "azurerm_sentinel_alert_rule_ms_security_incident"
	case "Fusion":
		return "azurerm_sentinel_alert_rule_fusion"
	case "MLBehaviorAnalytics":
		return "azurerm_sentinel_alert_rule_machine_learning_behavior_analytics"
	case "NRT":
		return "azurerm_sentinel_alert_rule_nrt"
	case "ThreatIntelligence":
		return "azurerm_sentinel_alert_rule_threat_intelligence"
	}
	return ""
}

// sentinelDataConnectorType maps the ARM data connector kind discriminator
// to the azurerm resource type the provider splits data connectors into.
func sentinelDataConnectorType(kind string) string {
	switch kind {
	case "AmazonWebServicesCloudTrail":
		return "azurerm_sentinel_data_connector_aws_cloud_trail"
	case "AmazonWebServicesS3":
		return "azurerm_sentinel_data_connector_aws_s3"
	case "AzureActiveDirectory":
		return "azurerm_sentinel_data_connector_azure_active_directory"
	case "AzureAdvancedThreatProtection":
		return "azurerm_sentinel_data_connector_azure_advanced_threat_protection"
	case "AzureSecurityCenter":
		return "azurerm_sentinel_data_connector_azure_security_center"
	case "Dynamics365":
		return "azurerm_sentinel_data_connector_dynamics_365"
	case "IoT":
		return "azurerm_sentinel_data_connector_iot"
	case "MicrosoftCloudAppSecurity":
		return "azurerm_sentinel_data_connector_microsoft_cloud_app_security"
	case "MicrosoftDefenderAdvancedThreatProtection":
		return "azurerm_sentinel_data_connector_microsoft_defender_advanced_threat_protection"
	case "MicrosoftThreatIntelligence":
		return "azurerm_sentinel_data_connector_microsoft_threat_intelligence"
	case "MicrosoftThreatProtection":
		return "azurerm_sentinel_data_connector_microsoft_threat_protection"
	case "Office365":
		return "azurerm_sentinel_data_connector_office_365"
	case "Office365Project":
		return "azurerm_sentinel_data_connector_office_365_project"
	case "OfficeATP":
		return "azurerm_sentinel_data_connector_office_atp"
	case "OfficeIRM":
		return "azurerm_sentinel_data_connector_office_irm"
	case "OfficePowerBI":
		return "azurerm_sentinel_data_connector_office_power_bi"
	case "ThreatIntelligence":
		return "azurerm_sentinel_data_connector_threat_intelligence"
	case "ThreatIntelligenceTaxii":
		return "azurerm_sentinel_data_connector_threat_intelligence_taxii"
	}
	return ""
}

func (g *SentinelGenerator) listWorkspaces(ctx context.Context) ([]*armoperationalinsights.Workspace, error) {
	client, err := armoperationalinsights.NewWorkspacesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	var workspaces []*armoperationalinsights.Workspace
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return workspaces, err
			}
			workspaces = append(workspaces, page.Value...)
		}
		return workspaces, nil
	}
	pager := client.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return workspaces, err
		}
		workspaces = append(workspaces, page.Value...)
	}
	return workspaces, nil
}

func (g *SentinelGenerator) appendWorkspaceResources(ctx context.Context, rg, workspaceName string) error {
	sub := g.getSubscriptionID()
	cred := g.getCredential()
	opts := g.armClientOptions()

	alertClient, err := armsecurityinsights.NewAlertRulesClient(sub, cred, opts)
	if err != nil {
		return err
	}
	alertPager := alertClient.NewListPager(rg, workspaceName, nil)
	for alertPager.More() {
		page, err := alertPager.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, item := range page.Value {
			rule := item.GetAlertRule()
			if rule == nil || rule.ID == nil || rule.Name == nil || rule.Kind == nil {
				continue
			}
			if tfType := sentinelAlertRuleType(string(*rule.Kind)); tfType != "" {
				g.AppendSimpleResource(*rule.ID, *rule.Name, tfType)
			}
		}
	}

	autoClient, err := armsecurityinsights.NewAutomationRulesClient(sub, cred, opts)
	if err != nil {
		return err
	}
	autoPager := autoClient.NewListPager(rg, workspaceName, nil)
	for autoPager.More() {
		page, err := autoPager.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, item := range page.Value {
			g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_sentinel_automation_rule")
		}
	}

	dcClient, err := armsecurityinsights.NewDataConnectorsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	dcPager := dcClient.NewListPager(rg, workspaceName, nil)
	for dcPager.More() {
		page, err := dcPager.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, item := range page.Value {
			connector := item.GetDataConnector()
			if connector == nil || connector.ID == nil || connector.Name == nil || connector.Kind == nil {
				continue
			}
			if tfType := sentinelDataConnectorType(string(*connector.Kind)); tfType != "" {
				g.AppendSimpleResource(*connector.ID, *connector.Name, tfType)
			}
		}
	}

	onboardingClient, err := armsecurityinsights.NewSentinelOnboardingStatesClient(sub, cred, opts)
	if err != nil {
		return err
	}
	onboarding, err := onboardingClient.List(ctx, rg, workspaceName, nil)
	if err != nil {
		return err
	}
	for _, item := range onboarding.Value {
		g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_sentinel_log_analytics_workspace_onboarding")
	}

	tiClient, err := armsecurityinsights.NewThreatIntelligenceIndicatorsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	tiPager := tiClient.NewListPager(rg, workspaceName, nil)
	for tiPager.More() {
		page, err := tiPager.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, item := range page.Value {
			info := item.GetThreatIntelligenceInformation()
			if info == nil || info.ID == nil {
				continue
			}
			name := ""
			if info.Name != nil {
				name = *info.Name
			} else {
				parts := strings.Split(*info.ID, "/")
				name = parts[len(parts)-1]
			}
			g.AppendSimpleResource(*info.ID, name, "azurerm_sentinel_threat_intelligence_indicator")
		}
	}

	watchlistClient, err := armsecurityinsights.NewWatchlistsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	watchlists := []string{}
	wlPager := watchlistClient.NewListPager(rg, workspaceName, nil)
	for wlPager.More() {
		page, err := wlPager.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, item := range page.Value {
			g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_sentinel_watchlist")
			watchlists = append(watchlists, *item.Name)
		}
	}

	itemClient, err := armsecurityinsights.NewWatchlistItemsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	for _, watchlistName := range watchlists {
		itemPager := itemClient.NewListPager(rg, workspaceName, watchlistName, nil)
		for itemPager.More() {
			page, err := itemPager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_sentinel_watchlist_item")
			}
		}
	}
	return nil
}

func (g *SentinelGenerator) InitResources() error {
	ctx := context.Background()
	workspaces, err := g.listWorkspaces(ctx)
	if err != nil {
		return err
	}
	for _, workspace := range workspaces {
		parsed, err := ParseAzureResourceID(*workspace.ID)
		if err != nil {
			return err
		}
		if err := g.appendWorkspaceResources(ctx, parsed.ResourceGroup, *workspace.Name); err != nil {
			return err
		}
	}
	return nil
}
