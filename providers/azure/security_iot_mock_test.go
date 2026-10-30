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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cognitiveservices/armcognitiveservices"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/securityinsights/armsecurityinsights"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/servicebus/armservicebus"
)

func TestSentinelAlertRuleTypeMapping(t *testing.T) {
	cases := map[string]string{
		"Scheduled":                         "azurerm_sentinel_alert_rule_scheduled",
		"MicrosoftSecurityIncidentCreation": "azurerm_sentinel_alert_rule_ms_security_incident",
		"Fusion":                            "azurerm_sentinel_alert_rule_fusion",
		"MLBehaviorAnalytics":               "azurerm_sentinel_alert_rule_machine_learning_behavior_analytics",
		"NRT":                               "azurerm_sentinel_alert_rule_nrt",
		"ThreatIntelligence":                "azurerm_sentinel_alert_rule_threat_intelligence",
		"Anomalies":                         "",
		"":                                  "",
	}
	for kind, want := range cases {
		if got := sentinelAlertRuleType(kind); got != want {
			t.Errorf("sentinelAlertRuleType(%q) = %q, want %q", kind, got, want)
		}
	}
}

func TestSentinelDataConnectorTypeMapping(t *testing.T) {
	cases := map[string]string{
		"AmazonWebServicesCloudTrail":               "azurerm_sentinel_data_connector_aws_cloud_trail",
		"AmazonWebServicesS3":                       "azurerm_sentinel_data_connector_aws_s3",
		"AzureActiveDirectory":                      "azurerm_sentinel_data_connector_azure_active_directory",
		"AzureAdvancedThreatProtection":             "azurerm_sentinel_data_connector_azure_advanced_threat_protection",
		"AzureSecurityCenter":                       "azurerm_sentinel_data_connector_azure_security_center",
		"Dynamics365":                               "azurerm_sentinel_data_connector_dynamics_365",
		"IoT":                                       "azurerm_sentinel_data_connector_iot",
		"MicrosoftCloudAppSecurity":                 "azurerm_sentinel_data_connector_microsoft_cloud_app_security",
		"MicrosoftDefenderAdvancedThreatProtection": "azurerm_sentinel_data_connector_microsoft_defender_advanced_threat_protection",
		"MicrosoftThreatIntelligence":               "azurerm_sentinel_data_connector_microsoft_threat_intelligence",
		"MicrosoftThreatProtection":                 "azurerm_sentinel_data_connector_microsoft_threat_protection",
		"Office365":                                 "azurerm_sentinel_data_connector_office_365",
		"Office365Project":                          "azurerm_sentinel_data_connector_office_365_project",
		"OfficeATP":                                 "azurerm_sentinel_data_connector_office_atp",
		"OfficeIRM":                                 "azurerm_sentinel_data_connector_office_irm",
		"OfficePowerBI":                             "azurerm_sentinel_data_connector_office_power_bi",
		"ThreatIntelligence":                        "azurerm_sentinel_data_connector_threat_intelligence",
		"ThreatIntelligenceTaxii":                   "azurerm_sentinel_data_connector_threat_intelligence_taxii",
		"GenericUI":                                 "",
	}
	for kind, want := range cases {
		if got := sentinelDataConnectorType(kind); got != want {
			t.Errorf("sentinelDataConnectorType(%q) = %q, want %q", kind, got, want)
		}
	}
}

func TestCognitiveConnectionTypeMapping(t *testing.T) {
	cases := map[string]string{
		"AccountKey":      "azurerm_cognitive_account_connection_account_key",
		"ManagedIdentity": "azurerm_cognitive_account_connection_account_managed_identity",
		"ApiKey":          "azurerm_cognitive_account_connection_api_key",
		"CustomKeys":      "azurerm_cognitive_account_connection_custom_keys",
		"AAD":             "azurerm_cognitive_account_connection_entra_id",
		"OAuth2":          "",
	}
	for authType, want := range cases {
		if got := cognitiveConnectionType(authType); got != want {
			t.Errorf("cognitiveConnectionType(%q) = %q, want %q", authType, got, want)
		}
	}
}

const mockServiceBusNamespacesJSON = `{"value":[
	{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.ServiceBus/namespaces/sb-mock-1","name":"sb-mock-1","location":"eastus","type":"Microsoft.ServiceBus/namespaces","properties":{}}
]}`

func TestMockTrack2ServiceBusNamespaces(t *testing.T) {
	tr := &cannedTransport{body: mockServiceBusNamespacesJSON}
	client, err := armservicebus.NewNamespacesClient(mockSubscriptionID, stubCredential{}, mockClientOptions(tr))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	pager := client.NewListPager(nil)
	ctx := context.Background()
	count := 0
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			t.Fatalf("next page: %v", err)
		}
		for _, ns := range page.Value {
			if *ns.Name != "sb-mock-1" {
				t.Errorf("name = %q, want sb-mock-1", *ns.Name)
			}
			count++
		}
	}
	if count != 1 {
		t.Fatalf("got %d namespaces, want 1", count)
	}
	if len(tr.urls) == 0 || !strings.Contains(tr.urls[0], "Microsoft.ServiceBus") {
		t.Errorf("expected a Microsoft.ServiceBus call, got %v", tr.urls)
	}
}

const mockSentinelAlertRulesJSON = `{"value":[
	{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.OperationalInsights/workspaces/ws-mock-1/providers/Microsoft.SecurityInsights/alertRules/rule-1","name":"rule-1","type":"Microsoft.SecurityInsights/alertRules","kind":"Scheduled","properties":{"displayName":"rule-1","enabled":true,"severity":"High","query":"AzureActivity","queryFrequency":"PT1H","queryPeriod":"PT1H","triggerOperator":"GreaterThan","triggerThreshold":0}},
	{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.OperationalInsights/workspaces/ws-mock-1/providers/Microsoft.SecurityInsights/alertRules/rule-2","name":"rule-2","type":"Microsoft.SecurityInsights/alertRules","kind":"NRT","properties":{"displayName":"rule-2","enabled":true,"severity":"High","query":"AzureActivity"}}
]}`

func TestMockTrack2SentinelAlertRulesKindMapping(t *testing.T) {
	tr := &cannedTransport{body: mockSentinelAlertRulesJSON}
	client, err := armsecurityinsights.NewAlertRulesClient(mockSubscriptionID, stubCredential{}, mockClientOptions(tr))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	pager := client.NewListPager("rg-mock-1", "ws-mock-1", nil)
	ctx := context.Background()
	got := map[string]string{}
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			t.Fatalf("next page: %v", err)
		}
		for _, item := range page.Value {
			rule := item.GetAlertRule()
			got[*rule.Name] = sentinelAlertRuleType(string(*rule.Kind))
		}
	}
	if got["rule-1"] != "azurerm_sentinel_alert_rule_scheduled" {
		t.Errorf("rule-1 type = %q, want azurerm_sentinel_alert_rule_scheduled", got["rule-1"])
	}
	// NRT is not a kind the SDK models concretely; it must survive as the
	// base AlertRule with its kind string intact so the mapping can emit it.
	if got["rule-2"] != "azurerm_sentinel_alert_rule_nrt" {
		t.Errorf("rule-2 type = %q, want azurerm_sentinel_alert_rule_nrt", got["rule-2"])
	}
}

const mockCognitiveAccountsJSON = `{"value":[
	{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.CognitiveServices/accounts/ai-mock-1","name":"ai-mock-1","location":"eastus","type":"Microsoft.CognitiveServices/accounts","kind":"AIServices","sku":{"name":"S0"},"properties":{}}
]}`

func TestMockTrack2CognitiveAccounts(t *testing.T) {
	tr := &cannedTransport{body: mockCognitiveAccountsJSON}
	client, err := armcognitiveservices.NewAccountsClient(mockSubscriptionID, stubCredential{}, mockClientOptions(tr))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	pager := client.NewListPager(nil)
	ctx := context.Background()
	count := 0
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			t.Fatalf("next page: %v", err)
		}
		for _, account := range page.Value {
			if *account.Kind != "AIServices" {
				t.Errorf("kind = %q, want AIServices", *account.Kind)
			}
			count++
		}
	}
	if count != 1 {
		t.Fatalf("got %d accounts, want 1", count)
	}
	if len(tr.urls) == 0 || !strings.Contains(tr.urls[0], "Microsoft.CognitiveServices") {
		t.Errorf("expected a Microsoft.CognitiveServices call, got %v", tr.urls)
	}
}
