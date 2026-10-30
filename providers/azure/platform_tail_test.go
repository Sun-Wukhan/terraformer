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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/automation/armautomation"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/datadog/armdatadog"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/machinelearning/armmachinelearning"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/oracledatabase/armoracledatabase"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/paloaltonetworksngfw/armpanngfw"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

func TestClassifyAutomationConnection(t *testing.T) {
	cases := map[string]string{
		"Azure":                   "azurerm_automation_connection_certificate",
		"AzureClassicCertificate": "azurerm_automation_connection_classic_certificate",
		"AzureServicePrincipal":   "azurerm_automation_connection_service_principal",
		"AzureRunAsConnection":    "azurerm_automation_connection",
		"":                        "azurerm_automation_connection",
	}
	for in, want := range cases {
		if got := classifyAutomationConnection(in); got != want {
			t.Errorf("classifyAutomationConnection(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestClassifyMLCompute(t *testing.T) {
	cases := map[armmachinelearning.ComputeType]string{
		armmachinelearning.ComputeTypeAmlCompute:      "azurerm_machine_learning_compute_cluster",
		armmachinelearning.ComputeTypeComputeInstance: "azurerm_machine_learning_compute_instance",
		armmachinelearning.ComputeTypeAKS:             "azurerm_machine_learning_inference_cluster",
		armmachinelearning.ComputeTypeSynapseSpark:    "azurerm_machine_learning_synapse_spark",
		armmachinelearning.ComputeTypeVirtualMachine:  "",
		armmachinelearning.ComputeTypeDatabricks:      "",
	}
	for in, want := range cases {
		if got := classifyMLCompute(in); got != want {
			t.Errorf("classifyMLCompute(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestClassifyOracleAutonomousDatabase(t *testing.T) {
	cases := []struct {
		props armoracledatabase.AutonomousDatabaseBasePropertiesClassification
		want  string
	}{
		{&armoracledatabase.AutonomousDatabaseProperties{}, "azurerm_oracle_autonomous_database"},
		{&armoracledatabase.AutonomousDatabaseCloneProperties{}, "azurerm_oracle_autonomous_database_clone_from_database"},
		{&armoracledatabase.AutonomousDatabaseFromBackupTimestampProperties{}, "azurerm_oracle_autonomous_database_clone_from_backup"},
		{&armoracledatabase.AutonomousDatabaseCrossRegionDisasterRecoveryProperties{}, "azurerm_oracle_autonomous_database_cross_region_disaster_recovery"},
		{nil, "azurerm_oracle_autonomous_database"},
	}
	for _, c := range cases {
		if got := classifyOracleAutonomousDatabase(c.props); got != c.want {
			t.Errorf("classifyOracleAutonomousDatabase(%T) = %q, want %q", c.props, got, c.want)
		}
	}
}

func TestClassifyPaloFirewall(t *testing.T) {
	vnet := armpanngfw.NetworkTypeVNET
	vwan := armpanngfw.NetworkTypeVWAN
	trueEnum := armpanngfw.BooleanEnumTRUE
	cases := []struct {
		name  string
		props *armpanngfw.FirewallDeploymentProperties
		want  string
	}{
		{"vnet local", &armpanngfw.FirewallDeploymentProperties{
			NetworkProfile:      &armpanngfw.NetworkProfile{NetworkType: &vnet},
			AssociatedRulestack: &armpanngfw.RulestackDetails{},
		}, "azurerm_palo_alto_next_generation_firewall_virtual_network_local_rulestack"},
		{"vhub panorama", &armpanngfw.FirewallDeploymentProperties{
			NetworkProfile:    &armpanngfw.NetworkProfile{NetworkType: &vwan},
			IsPanoramaManaged: &trueEnum,
		}, "azurerm_palo_alto_next_generation_firewall_virtual_hub_panorama"},
		{"vnet strata", &armpanngfw.FirewallDeploymentProperties{
			NetworkProfile: &armpanngfw.NetworkProfile{NetworkType: &vnet},
		}, "azurerm_palo_alto_next_generation_firewall_virtual_network_strata_cloud_manager"},
		{"nil props", nil, "azurerm_palo_alto_next_generation_firewall_virtual_network_strata_cloud_manager"},
	}
	for _, c := range cases {
		if got := classifyPaloFirewall(c.props); got != c.want {
			t.Errorf("%s: classifyPaloFirewall = %q, want %q", c.name, got, c.want)
		}
	}
}

const mockAutomationAccountsJSON = `{"value":[
	{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.Automation/automationAccounts/aa-mock-1","name":"aa-mock-1","location":"eastus","type":"Microsoft.Automation/automationAccounts","properties":{}}
]}`

func TestMockTrack2AutomationAccounts(t *testing.T) {
	tr := &cannedTransport{body: mockAutomationAccountsJSON}
	client, err := armautomation.NewAccountClient(mockSubscriptionID, stubCredential{}, mockClientOptions(tr))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	pager := client.NewListPager(nil)
	ctx := context.Background()
	var resources []terraformutils.Resource
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			t.Fatalf("next page: %v", err)
		}
		for _, item := range page.Value {
			resources = appendAutomationResource(resources, item.ID, item.Name, "azurerm_automation_account", "azurerm")
		}
	}
	if len(resources) != 1 {
		t.Fatalf("got %d resources, want 1", len(resources))
	}
	if resources[0].InstanceInfo.Type != "azurerm_automation_account" {
		t.Errorf("type = %q, want azurerm_automation_account", resources[0].InstanceInfo.Type)
	}
	if len(tr.urls) == 0 || !strings.Contains(tr.urls[0], "automationAccounts") {
		t.Errorf("expected an ARM automationAccounts call, got %v", tr.urls)
	}
}

const mockDatadogMonitorsJSON = `{"value":[
	{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.Datadog/monitors/dd-mock-1","name":"dd-mock-1","location":"eastus","type":"Microsoft.Datadog/monitors","properties":{}}
]}`

func TestMockTrack2DatadogMonitors(t *testing.T) {
	tr := &cannedTransport{body: mockDatadogMonitorsJSON}
	client, err := armdatadog.NewMonitorsClient(mockSubscriptionID, stubCredential{}, mockClientOptions(tr))
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
		for _, item := range page.Value {
			if *item.Name != "dd-mock-1" {
				t.Errorf("name = %q, want dd-mock-1", *item.Name)
			}
			count++
		}
	}
	if count != 1 {
		t.Fatalf("got %d monitors, want 1", count)
	}
	if len(tr.urls) == 0 || !strings.Contains(tr.urls[0], "Microsoft.Datadog") {
		t.Errorf("expected a Microsoft.Datadog call, got %v", tr.urls)
	}
}
