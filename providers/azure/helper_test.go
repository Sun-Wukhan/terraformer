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
	"testing"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

func TestCanonicalAzureID(t *testing.T) {
	const rg = "/subscriptions/s/resourceGroups/Devtest/providers/"
	tests := []struct {
		in, want string
	}{
		{rg + "Microsoft.EventHub/namespaces/nm/authorizationrules/Root", rg + "Microsoft.EventHub/namespaces/nm/authorizationRules/Root"},
		{rg + "Microsoft.EventHub/namespaces/nm/eventhubs/logs/consumergroups/$Default", rg + "Microsoft.EventHub/namespaces/nm/eventhubs/logs/consumerGroups/$Default"},
		{rg + "Microsoft.Relay/namespaces/r/hybridconnections/hc/authorizationrules/l", rg + "Microsoft.Relay/namespaces/r/hybridConnections/hc/authorizationRules/l"},
		{rg + "Microsoft.ServiceBus/namespaces/sb/topics/t/authorizationrules/p", rg + "Microsoft.ServiceBus/namespaces/sb/topics/t/authorizationRules/p"},
		{rg + "Microsoft.Web/serverfarms/plan", rg + "Microsoft.Web/serverFarms/plan"},
		{rg + "Microsoft.LoadTestService/loadtests/lt", rg + "Microsoft.LoadTestService/loadTests/lt"},
		{"/subscriptions/s/resourcegroups/Devtest/providers/microsoft.insights/metricalerts/a", rg + "Microsoft.Insights/metricAlerts/a"},
		{rg + "microsoft.insights/actiongroups/Smart Detection", rg + "Microsoft.Insights/actionGroups/Smart Detection"},
		// Segments are scoped per namespace: IoT Hub keeps `ConsumerGroups`.
		{rg + "Microsoft.Devices/IotHubs/h/eventHubEndpoints/events/ConsumerGroups/g", rg + "Microsoft.Devices/IotHubs/h/eventHubEndpoints/events/ConsumerGroups/g"},
		// Resource names are never rewritten, even when they look like keys.
		{rg + "Microsoft.EventHub/namespaces/nm/authorizationRules/consumergroups", rg + "Microsoft.EventHub/namespaces/nm/authorizationRules/consumergroups"},
	}
	for _, tt := range tests {
		if got := canonicalAzureID(tt.in); got != tt.want {
			t.Errorf("canonicalAzureID(%q)\n got %q\nwant %q", tt.in, got, tt.want)
		}
	}
}

func TestResourceGroupOfID(t *testing.T) {
	if got := resourceGroupOfID("/subscriptions/s/resourcegroups/Other/providers/X/y/z"); got != "Other" {
		t.Fatalf("got %q, want Other", got)
	}
	if got := resourceGroupOfID("/subscriptions/s/providers/Microsoft.Authorization/policyDefinitions/p"); got != "" {
		t.Fatalf("got %q, want empty for subscription-scoped ID", got)
	}
}

func TestAzureServiceInitialCleanup(t *testing.T) {
	svc := &AzureService{}
	svc.SetArgs(map[string]interface{}{"resource_group": "Devtest"})
	svc.Resources = []terraformutils.Resource{
		terraformutils.NewSimpleResource("/subscriptions/s/resourcegroups/devtest/providers/Microsoft.Web/serverfarms/p", "p", "azurerm_service_plan", "azurerm", nil),
		terraformutils.NewSimpleResource("/subscriptions/s/resourceGroups/Other/providers/Microsoft.LoadTestService/loadTests/lt", "lt", "azurerm_load_test", "azurerm", nil),
		terraformutils.NewSimpleResource("/subscriptions/s/providers/Microsoft.Authorization/policyDefinitions/d", "d", "azurerm_policy_definition", "azurerm", nil),
	}

	svc.InitialCleanup()

	if len(svc.Resources) != 2 {
		t.Fatalf("got %d resources, want the in-group and the subscription-scoped one", len(svc.Resources))
	}
	if got, want := svc.Resources[0].InstanceState.ID, "/subscriptions/s/resourceGroups/devtest/providers/Microsoft.Web/serverFarms/p"; got != want {
		t.Fatalf("ID not canonicalized: got %q, want %q", got, want)
	}
	if svc.Resources[1].InstanceInfo.Type != "azurerm_policy_definition" {
		t.Fatalf("subscription-scoped resource dropped: %+v", svc.Resources[1])
	}
}

func TestIsAPIMServiceScope(t *testing.T) {
	const svc = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.ApiManagement/service/apim"
	for scope, want := range map[string]bool{
		svc + "/":                 true,
		svc:                       true,
		svc + "/apis":             false,
		svc + "/apis/echo":        false,
		svc + "/products/starter": false,
	} {
		if got := isAPIMServiceScope(scope); got != want {
			t.Errorf("isAPIMServiceScope(%q) = %v, want %v", scope, got, want)
		}
	}
}
