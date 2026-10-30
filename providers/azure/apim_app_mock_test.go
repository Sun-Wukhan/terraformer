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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

const mockAPIMServicesJSON = `{"value":[
	{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.ApiManagement/service/apim-mock-1","name":"apim-mock-1","location":"eastus","type":"Microsoft.ApiManagement/service","properties":{"publisherEmail":"a@b.c","publisherName":"Mock"}}
]}`

func TestMockAPIManagementServices(t *testing.T) {
	tr := &cannedTransport{body: mockAPIMServicesJSON}
	client, err := armapimanagement.NewServiceClient(mockSubscriptionID, stubCredential{}, mockClientOptions(tr))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	g := &APIManagementGenerator{}
	resources, err := drainPager(context.Background(), client.NewListPager(nil),
		func(page armapimanagement.ServiceClientListResponse) []*armapimanagement.ServiceResource {
			return page.Value
		},
		func(item *armapimanagement.ServiceResource) (terraformutils.Resource, bool) {
			if item.ID == nil || item.Name == nil {
				return terraformutils.Resource{}, false
			}
			return g.simple(*item.ID, *item.Name, "azurerm_api_management")
		})
	if err != nil {
		t.Fatalf("drain: %v", err)
	}
	if len(resources) != 1 {
		t.Fatalf("got %d resources, want 1", len(resources))
	}
	if resources[0].ResourceName != "tfer--apim-mock-1" {
		t.Errorf("name = %q, want tfer--apim-mock-1", resources[0].ResourceName)
	}
	if resources[0].InstanceInfo.Type != "azurerm_api_management" {
		t.Errorf("type = %q, want azurerm_api_management", resources[0].InstanceInfo.Type)
	}
	if len(tr.urls) == 0 || !strings.Contains(tr.urls[0], "Microsoft.ApiManagement") {
		t.Errorf("expected a Microsoft.ApiManagement call, got %v", tr.urls)
	}
}

const mockSitesJSON = `{"value":[
	{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.Web/sites/web-mock-1","name":"web-mock-1","location":"eastus","type":"Microsoft.Web/sites","kind":"app,linux","properties":{}},
	{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.Web/sites/func-mock-1","name":"func-mock-1","location":"eastus","type":"Microsoft.Web/sites","kind":"functionapp","properties":{}},
	{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.Web/sites/logic-mock-1","name":"logic-mock-1","location":"eastus","type":"Microsoft.Web/sites","kind":"functionapp,workflowapp,linux","properties":{}}
]}`

func TestMockWebAppsSiteClassification(t *testing.T) {
	tr := &cannedTransport{body: mockSitesJSON}
	client, err := armappservice.NewWebAppsClient(mockSubscriptionID, stubCredential{}, mockClientOptions(tr))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	g := &WebAppsGenerator{}
	resources, err := drainPager(context.Background(), client.NewListPager(nil),
		func(page armappservice.WebAppsClientListResponse) []*armappservice.Site { return page.Value },
		func(item *armappservice.Site) (terraformutils.Resource, bool) {
			if item.ID == nil || item.Name == nil {
				return terraformutils.Resource{}, false
			}
			kind := ""
			if item.Kind != nil {
				kind = *item.Kind
			}
			return g.simple(*item.ID, *item.Name, classifySiteKind(kind).resourceType(false))
		})
	if err != nil {
		t.Fatalf("drain: %v", err)
	}
	want := []string{"azurerm_linux_web_app", "azurerm_windows_function_app", "azurerm_logic_app_standard"}
	if len(resources) != len(want) {
		t.Fatalf("got %d resources, want %d", len(resources), len(want))
	}
	for i, w := range want {
		if resources[i].InstanceInfo.Type != w {
			t.Errorf("resource %d type = %q, want %q", i, resources[i].InstanceInfo.Type, w)
		}
	}
	if len(tr.urls) == 0 || !strings.Contains(tr.urls[0], "Microsoft.Web") {
		t.Errorf("expected a Microsoft.Web call, got %v", tr.urls)
	}
}

func TestClassifySiteKindSlots(t *testing.T) {
	cases := map[string]string{
		"app":                         "azurerm_windows_web_app",
		"app,linux":                   "azurerm_linux_web_app",
		"functionapp,linux":           "azurerm_linux_function_app",
		"functionapp,workflowapp":     "azurerm_logic_app_standard",
		"functionapp,linux,container": "azurerm_linux_function_app",
	}
	for kind, want := range cases {
		if got := classifySiteKind(kind).resourceType(false); got != want {
			t.Errorf("classifySiteKind(%q) = %q, want %q", kind, got, want)
		}
	}
	if got := classifySiteKind("app,linux").resourceType(true); got != "azurerm_linux_web_app_slot" {
		t.Errorf("slot type = %q, want azurerm_linux_web_app_slot", got)
	}
	if got := classifySiteKind("functionapp,workflowapp").resourceType(true); got != "" {
		t.Errorf("logic slot type = %q, want empty (no slot resource for Logic Apps Standard)", got)
	}
}

func TestLogicDefinitionResources(t *testing.T) {
	definition := map[string]interface{}{
		"actions": map[string]interface{}{
			"Call_API":  map[string]interface{}{"type": "Http"},
			"Transform": map[string]interface{}{"type": "Xslt"},
		},
		"triggers": map[string]interface{}{
			"manual":   map[string]interface{}{"type": "Request", "kind": "Http"},
			"daily":    map[string]interface{}{"type": "Recurrence"},
			"on_event": map[string]interface{}{"type": "ApiConnection"},
		},
	}
	workflowID := "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Logic/workflows/wf1"
	resources := logicDefinitionResources("azurerm", workflowID, definition)
	byID := map[string]string{}
	for _, r := range resources {
		byID[r.InstanceState.ID] = r.InstanceInfo.Type
	}
	want := map[string]string{
		workflowID + "/actions/Call_API":  "azurerm_logic_app_action_http",
		workflowID + "/actions/Transform": "azurerm_logic_app_action_custom",
		workflowID + "/triggers/manual":   "azurerm_logic_app_trigger_http_request",
		workflowID + "/triggers/daily":    "azurerm_logic_app_trigger_recurrence",
		workflowID + "/triggers/on_event": "azurerm_logic_app_trigger_custom",
	}
	if len(byID) != len(want) {
		t.Fatalf("got %d resources, want %d: %v", len(byID), len(want), byID)
	}
	for id, typ := range want {
		if byID[id] != typ {
			t.Errorf("id %s type = %q, want %q", id, byID[id], typ)
		}
	}
}
