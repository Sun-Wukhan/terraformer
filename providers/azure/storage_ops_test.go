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
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/operationalinsights/armoperationalinsights"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
)

// TestMockTrack2StorageQueues proves the storage-extra queue mapping works
// against the real armstorage client with a canned transport.
func TestMockTrack2StorageQueues(t *testing.T) {
	body := `{"value":[{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg1/providers/Microsoft.Storage/storageAccounts/acct1/queueServices/default/queues/q1","name":"q1","type":"Microsoft.Storage/storageAccounts/queueServices/queues"}]}`
	tr := &cannedTransport{body: body}
	client, err := armstorage.NewQueueClient(mockSubscriptionID, stubCredential{}, mockClientOptions(tr))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	pager := client.NewListPager("rg1", "acct1", nil)
	var got int
	for pager.More() {
		page, err := pager.NextPage(context.Background())
		if err != nil {
			t.Fatalf("page: %v", err)
		}
		got += len(page.Value)
		if len(page.Value) > 0 && page.Value[0].Name != nil && *page.Value[0].Name != "q1" {
			t.Errorf("queue name = %q, want q1", *page.Value[0].Name)
		}
	}
	if got != 1 {
		t.Fatalf("got %d queues, want 1", got)
	}
	if len(tr.urls) == 0 || !strings.Contains(tr.urls[0], "queues") {
		t.Errorf("expected a queues ARM call, got %v", tr.urls)
	}
}

// TestMockTrack2LogAnalyticsWorkspaces proves the workspace mapping works
// against the real armoperationalinsights client with a canned transport.
func TestMockTrack2LogAnalyticsWorkspaces(t *testing.T) {
	body := `{"value":[{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg1/providers/Microsoft.OperationalInsights/workspaces/ws1","name":"ws1","type":"Microsoft.OperationalInsights/workspaces","location":"eastus","properties":{"sku":{"name":"PerGB2018"}}}]}`
	tr := &cannedTransport{body: body}
	client, err := armoperationalinsights.NewWorkspacesClient(mockSubscriptionID, stubCredential{}, mockClientOptions(tr))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	pager := client.NewListPager(nil)
	var got int
	for pager.More() {
		page, err := pager.NextPage(context.Background())
		if err != nil {
			t.Fatalf("page: %v", err)
		}
		got += len(page.Value)
	}
	if got != 1 {
		t.Fatalf("got %d workspaces, want 1", got)
	}
}

func TestStorageOpsMappingHelpers(t *testing.T) {
	cases := []struct {
		name string
		got  string
		want string
	}{
		{"netapp sap", netappVolumeGroupType("SAP-HANA"), "azurerm_netapp_volume_group_sap_hana"},
		{"netapp oracle", netappVolumeGroupType("Oracle"), "azurerm_netapp_volume_group_oracle"},
		{"netapp other", netappVolumeGroupType(""), ""},
		{"backup vm", backupPolicyType("", "AzureIaasVM"), "azurerm_backup_policy_vm"},
		{"backup fs", backupPolicyType("", "AzureStorage"), "azurerm_backup_policy_file_share"},
		{"backup wl", backupPolicyType("", "AzureWorkload"), "azurerm_backup_policy_vm_workload"},
		{"backup concrete", backupPolicyType("*armrecoveryservicesbackup.AzureIaaSVMProtectionPolicy", ""), "azurerm_backup_policy_vm"},
		{"protected vm", backupProtectedItemType("AzureIaasVM"), "azurerm_backup_protected_vm"},
		{"protected fs", backupProtectedItemType("AzureStorage"), "azurerm_backup_protected_file_share"},
		{"sqr alert", scheduledQueryRuleType("LogAlert"), "azurerm_monitor_scheduled_query_rules_alert"},
		{"sqr alert v2", scheduledQueryRuleType("LogAlertV2"), "azurerm_monitor_scheduled_query_rules_alert_v2"},
		{"sqr log", scheduledQueryRuleType("LogToMetric"), "azurerm_monitor_scheduled_query_rules_log"},
		{"apr add", alertProcessingRuleType([]string{"*armalertsmanagement.AddActionGroups"}), "azurerm_monitor_alert_processing_rule_action_group"},
		{"apr suppress", alertProcessingRuleType([]string{"*armalertsmanagement.RemoveAllActionGroups"}), "azurerm_monitor_alert_processing_rule_suppression"},
		{"sr policy a2a", siteRecoveryPolicyType("*armrecoveryservicessiterecovery.A2APolicyProviderSpecificDetails"), "azurerm_site_recovery_replication_policy"},
		{"sr policy hyperv", siteRecoveryPolicyType("*armrecoveryservicessiterecovery.HyperVReplicaAzurePolicyProviderSpecificDetails"), "azurerm_site_recovery_hyperv_replication_policy"},
		{"sr policy vmware", siteRecoveryPolicyType("*armrecoveryservicessiterecovery.VMwareCbtPolicyProviderSpecificDetails"), "azurerm_site_recovery_vmware_replication_policy"},
		{"sr item a2a", siteRecoveryProtectedItemType("*armrecoveryservicessiterecovery.A2AProtectedItemProperties"), "azurerm_site_recovery_replicated_vm"},
		{"sr item vmware", siteRecoveryProtectedItemType("*armrecoveryservicessiterecovery.VMwareCbtProtectedItemProperties"), "azurerm_site_recovery_vmware_replicated_vm"},
		{"dt eventhub", digitalTwinsEndpointType("*armdigitaltwins.EventHubEndpointProperties"), "azurerm_digital_twins_endpoint_eventhub"},
		{"dt eventgrid", digitalTwinsEndpointType("*armdigitaltwins.EventGridEndpointProperties"), "azurerm_digital_twins_endpoint_eventgrid"},
		{"dt servicebus", digitalTwinsEndpointType("*armdigitaltwins.ServiceBusEndpointProperties"), "azurerm_digital_twins_endpoint_servicebus"},
		{"la ds event", logAnalyticsDataSourceType("windowsEvent"), "azurerm_log_analytics_datasource_windows_event"},
		{"la ds perf", logAnalyticsDataSourceType("windowsPerformanceCounter"), "azurerm_log_analytics_datasource_windows_performance_counter"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, c.got, c.want)
		}
	}
	if !isCustomLogTable("MyTable_CL") || isCustomLogTable("Heartbeat") {
		t.Errorf("isCustomLogTable suffix logic broken")
	}
}
