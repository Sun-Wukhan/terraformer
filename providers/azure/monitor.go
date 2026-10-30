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
	"fmt"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/alertsmanagement/armalertsmanagement"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// MonitorGenerator covers Azure Monitor via armmonitor and alert processing
// rules via armalertsmanagement: action groups, activity log alerts,
// autoscale settings, data collection endpoints/rules and their
// associations, metric alerts, private link scopes and their scoped
// services, scheduled query rules (split by kind) and alert processing
// rules (split by action).
//
// Skipped in this family (see docs/azure.md "Known limits"):
//   - azurerm_monitor_diagnostic_setting: armmonitor DiagnosticSettings are
//     listed only per target resource URI and the provider ID is the
//     composite "{targetResourceId}|{name}"; there is no subscription or
//     resource-group List.
//   - azurerm_monitor_aad_diagnostic_setting: tenant-level Microsoft.AADIAM
//     diagnostic settings have no client in armmonitor/armalertsmanagement.
//   - azurerm_monitor_workspace (Microsoft.Monitor/accounts),
//     azurerm_monitor_alert_prometheus_rule_group and
//     azurerm_monitor_smart_detector_alert_rule: no client/List exists in
//     the track-2 armmonitor v0.13.0 or armalertsmanagement modules.
type MonitorGenerator struct {
	AzureService
}

// scheduledQueryRuleType maps an armmonitor scheduled query rule kind to the
// azurerm resource type.
func scheduledQueryRuleType(kind string) string {
	switch kind {
	case "LogAlert":
		return "azurerm_monitor_scheduled_query_rules_alert"
	case "LogAlertV2":
		return "azurerm_monitor_scheduled_query_rules_alert_v2"
	case "LogToMetric":
		return "azurerm_monitor_scheduled_query_rules_log"
	}
	return ""
}

// alertProcessingRuleType maps an alert processing rule's action type names
// to the azurerm resource type: rules adding action groups are
// azurerm_monitor_alert_processing_rule_action_group; rules whose actions
// only remove action groups (suppression rules) are
// azurerm_monitor_alert_processing_rule_suppression.
func alertProcessingRuleType(actionTypes []string) string {
	for _, a := range actionTypes {
		if strings.Contains(a, "AddActionGroups") {
			return "azurerm_monitor_alert_processing_rule_action_group"
		}
	}
	return "azurerm_monitor_alert_processing_rule_suppression"
}

func (g *MonitorGenerator) InitResources() error {
	ctx := context.Background()
	sub := g.getSubscriptionID()
	cred := g.getCredential()
	opts := g.armClientOptions()
	rg := g.getResourceGroup()
	var resources []terraformutils.Resource
	seen := map[string]bool{}
	add := func(id, name, typ string) {
		if id == "" || name == "" || typ == "" || seen[id] {
			return
		}
		seen[id] = true
		resources = append(resources, terraformutils.NewSimpleResource(id, name, typ, "azurerm", []string{}))
	}

	// armmonitor resources with subscription / resource-group pagers. When a
	// resource group filter is set, the ByResourceGroup pager is used,
	// otherwise the subscription pager.
	actionGroups, err := armmonitor.NewActionGroupsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	if rg != "" {
		pager := actionGroups.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_monitor_action_group")
				}
			}
		}
	} else {
		pager := actionGroups.NewListBySubscriptionIDPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_monitor_action_group")
				}
			}
		}
	}

	activityAlerts, err := armmonitor.NewActivityLogAlertsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	if rg != "" {
		pager := activityAlerts.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_monitor_activity_log_alert")
				}
			}
		}
	} else {
		pager := activityAlerts.NewListBySubscriptionIDPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_monitor_activity_log_alert")
				}
			}
		}
	}

	autoscale, err := armmonitor.NewAutoscaleSettingsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	if rg != "" {
		pager := autoscale.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_monitor_autoscale_setting")
				}
			}
		}
	} else {
		pager := autoscale.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_monitor_autoscale_setting")
				}
			}
		}
	}

	endpoints, err := armmonitor.NewDataCollectionEndpointsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	type namedRef struct{ rg, name string }
	var endpointRefs, ruleRefs []namedRef
	if rg != "" {
		pager := endpoints.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_monitor_data_collection_endpoint")
					endpointRefs = append(endpointRefs, namedRef{rg, *item.Name})
				}
			}
		}
	} else {
		pager := endpoints.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				if item.ID == nil || item.Name == nil {
					continue
				}
				add(*item.ID, *item.Name, "azurerm_monitor_data_collection_endpoint")
				if parsed, err := ParseAzureResourceID(*item.ID); err == nil {
					endpointRefs = append(endpointRefs, namedRef{parsed.ResourceGroup, *item.Name})
				}
			}
		}
	}

	rules, err := armmonitor.NewDataCollectionRulesClient(sub, cred, opts)
	if err != nil {
		return err
	}
	if rg != "" {
		pager := rules.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_monitor_data_collection_rule")
					ruleRefs = append(ruleRefs, namedRef{rg, *item.Name})
				}
			}
		}
	} else {
		pager := rules.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				if item.ID == nil || item.Name == nil {
					continue
				}
				add(*item.ID, *item.Name, "azurerm_monitor_data_collection_rule")
				if parsed, err := ParseAzureResourceID(*item.ID); err == nil {
					ruleRefs = append(ruleRefs, namedRef{parsed.ResourceGroup, *item.Name})
				}
			}
		}
	}

	associations, err := armmonitor.NewDataCollectionRuleAssociationsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	for _, ref := range ruleRefs {
		pager := associations.NewListByRulePager(ref.rg, ref.name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_monitor_data_collection_rule_association")
				}
			}
		}
	}
	for _, ref := range endpointRefs {
		pager := associations.NewListByDataCollectionEndpointPager(ref.rg, ref.name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_monitor_data_collection_rule_association")
				}
			}
		}
	}

	metricAlerts, err := armmonitor.NewMetricAlertsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	if rg != "" {
		pager := metricAlerts.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_monitor_metric_alert")
				}
			}
		}
	} else {
		pager := metricAlerts.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_monitor_metric_alert")
				}
			}
		}
	}

	scopes, err := armmonitor.NewPrivateLinkScopesClient(sub, cred, opts)
	if err != nil {
		return err
	}
	var scopeRefs []namedRef
	if rg != "" {
		pager := scopes.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_monitor_private_link_scope")
					scopeRefs = append(scopeRefs, namedRef{rg, *item.Name})
				}
			}
		}
	} else {
		pager := scopes.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				if item.ID == nil || item.Name == nil {
					continue
				}
				add(*item.ID, *item.Name, "azurerm_monitor_private_link_scope")
				if parsed, err := ParseAzureResourceID(*item.ID); err == nil {
					scopeRefs = append(scopeRefs, namedRef{parsed.ResourceGroup, *item.Name})
				}
			}
		}
	}
	scopedResources, err := armmonitor.NewPrivateLinkScopedResourcesClient(sub, cred, opts)
	if err != nil {
		return err
	}
	for _, ref := range scopeRefs {
		pager := scopedResources.NewListByPrivateLinkScopePager(ref.rg, ref.name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_monitor_private_link_scoped_service")
				}
			}
		}
	}

	queryRules, err := armmonitor.NewScheduledQueryRulesClient(sub, cred, opts)
	if err != nil {
		return err
	}
	handleQueryRule := func(item *armmonitor.ScheduledQueryRuleResource) {
		if item.ID == nil || item.Name == nil || item.Kind == nil {
			return
		}
		add(*item.ID, *item.Name, scheduledQueryRuleType(string(*item.Kind)))
	}
	if rg != "" {
		pager := queryRules.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				handleQueryRule(item)
			}
		}
	} else {
		pager := queryRules.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				handleQueryRule(item)
			}
		}
	}

	processingRules, err := armalertsmanagement.NewAlertProcessingRulesClient(sub, cred, opts)
	if err != nil {
		return err
	}
	handleProcessingRule := func(item *armalertsmanagement.AlertProcessingRule) {
		if item.ID == nil || item.Name == nil || item.Properties == nil {
			return
		}
		actionTypes := make([]string, 0, len(item.Properties.Actions))
		for _, a := range item.Properties.Actions {
			actionTypes = append(actionTypes, fmt.Sprintf("%T", a))
		}
		add(*item.ID, *item.Name, alertProcessingRuleType(actionTypes))
	}
	if rg != "" {
		pager := processingRules.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				handleProcessingRule(item)
			}
		}
	} else {
		pager := processingRules.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				handleProcessingRule(item)
			}
		}
	}

	g.Resources = resources
	return nil
}
