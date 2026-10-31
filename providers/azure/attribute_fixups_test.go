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
	"encoding/json"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

func convertedResource(typ, name string, item map[string]interface{}) terraformutils.Resource {
	r := terraformutils.NewSimpleResource("/subscriptions/s/resourceGroups/rg/x/"+name, name, typ, "azurerm", nil)
	r.Item = item
	return r
}

func TestAzureServicePostRefreshCleanup(t *testing.T) {
	report := terraformutils.NewImportReport()
	svc := &AzureService{}
	svc.SetName("test")
	svc.SetReport(report)
	svc.Resources = []terraformutils.Resource{
		convertedResource("azurerm_windows_virtual_machine", "vm", map[string]interface{}{
			"name":                   "vm",
			"os_managed_disk_id":     "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Compute/disks/d",
			"platform_fault_domain":  "-1",
			"admin_username":         "adm",
			"admin_password":         importedPasswordSentinel,
			"source_image_reference": []interface{}{map[string]interface{}{"offer": "WindowsServer"}},
		}),
		convertedResource("azurerm_windows_virtual_machine", "vm-in-vmss", map[string]interface{}{
			"platform_fault_domain": "1",
			"source_image_id":       "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Compute/images/i",
			"os_managed_disk_id":    "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Compute/disks/d2",
		}),
		convertedResource("azurerm_windows_virtual_machine", "vm-from-disk", map[string]interface{}{
			"os_managed_disk_id": "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Compute/disks/d3",
			"admin_username":     "adm",
			"admin_password":     importedPasswordSentinel,
			"patch_mode":         "AutomaticByOS",
			"size":               "Standard_B2s",
			"os_disk": []interface{}{map[string]interface{}{
				"name":                 "d3",
				"storage_account_type": "Premium_LRS",
				"caching":              "ReadWrite",
			}},
		}),
		convertedResource("azurerm_eventhub_namespace", "ehns", map[string]interface{}{"minimum_tls_version": "1.0"}),
		convertedResource("azurerm_eventhub_namespace", "ehns12", map[string]interface{}{"minimum_tls_version": "1.2"}),
		convertedResource("azurerm_windows_web_app", "legacy-web", map[string]interface{}{
			"site_config": []interface{}{map[string]interface{}{
				"application_stack": []interface{}{map[string]interface{}{
					"php_version":   "5.6",
					"node_version":  "6.9.1",
					"current_stack": "php",
				}},
			}},
		}),
		convertedResource("azurerm_windows_web_app", "modern-web", map[string]interface{}{
			"site_config": []interface{}{map[string]interface{}{
				"application_stack": []interface{}{map[string]interface{}{
					"php_version":  "7.4",
					"node_version": "~18",
				}},
			}},
		}),
		convertedResource("azurerm_managed_disk", "shared", map[string]interface{}{"max_shares": "2"}),
		convertedResource("azurerm_managed_disk", "single", map[string]interface{}{"max_shares": "1"}),
		convertedResource("azurerm_windows_function_app", "fn", map[string]interface{}{
			"key_vault_reference_identity_id": "SystemAssigned",
			"storage_uses_managed_identity":   "false",
			"storage_account_access_key":      "secret",
		}),
		convertedResource("azurerm_windows_web_app", "web-uai", map[string]interface{}{
			"key_vault_reference_identity_id": "/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/u",
		}),
		convertedResource("azurerm_monitor_metric_alert", "alert", map[string]interface{}{
			"scopes": []interface{}{"/subscriptions/s/resourcegroups/rg/providers/microsoft.insights/components/c"},
			"application_insights_web_test_location_availability_criteria": []interface{}{map[string]interface{}{
				"web_test_id": "/subscriptions/s/resourcegroups/rg/providers/microsoft.insights/webtests/w",
			}},
			"tags": map[string]interface{}{"hidden-link:/subscriptions/s/resourcegroups/rg": "Resource"},
		}),
		convertedResource("azurerm_app_service_certificate", "uploaded", map[string]interface{}{"name": "uploaded"}),
		convertedResource("azurerm_app_service_certificate", "kv", map[string]interface{}{"key_vault_secret_id": "https://kv/secrets/s"}),
		convertedResource("azurerm_api_management_logger", "eh-logger", map[string]interface{}{
			"eventhub": []interface{}{map[string]interface{}{"name": "hub"}},
		}),
		convertedResource("azurerm_api_management_logger", "ai-logger", map[string]interface{}{
			"application_insights": []interface{}{map[string]interface{}{"instrumentation_key": "k"}},
		}),
		convertedResource("azurerm_api_management_named_value", "secret-nv", map[string]interface{}{"secret": "true"}),
		convertedResource("azurerm_api_management_named_value", "plain-nv", map[string]interface{}{"value": "v"}),
		convertedResource("azurerm_api_management_subscription", "unnamed-sub", map[string]interface{}{"display_name": ""}),
		convertedResource("azurerm_api_management_subscription", "named-sub", map[string]interface{}{"display_name": "Starter"}),
	}

	svc.PostRefreshCleanup()

	byName := map[string]terraformutils.Resource{}
	for _, r := range svc.Resources {
		byName[strings.TrimPrefix(r.ResourceName, "tfer--")] = r
	}
	if _, ok := byName["uploaded"]; ok {
		t.Errorf("directly uploaded certificate must be dropped")
	}
	if _, ok := byName["kv"]; !ok {
		t.Errorf("key vault certificate must be kept")
	}
	for _, dropped := range []string{"eh-logger", "secret-nv", "unnamed-sub"} {
		if _, ok := byName[dropped]; ok {
			t.Errorf("%s cannot be represented and must be dropped", dropped)
		}
	}
	for _, kept := range []string{"ai-logger", "plain-nv", "named-sub"} {
		if _, ok := byName[kept]; !ok {
			t.Errorf("%s must be kept", kept)
		}
	}
	if len(report.Failures()) != 4 {
		t.Errorf("every dropped resource must be reported once, got %+v", report.Failures())
	}

	vm := byName["vm"].Item
	if _, ok := vm["os_managed_disk_id"]; ok {
		t.Errorf("os_managed_disk_id must be dropped on an image-based VM: %v", vm)
	}
	if _, ok := vm["platform_fault_domain"]; ok {
		t.Errorf("platform_fault_domain = -1 must be dropped: %v", vm)
	}
	if vm["name"] != "vm" || vm["admin_username"] != "adm" {
		t.Errorf("unrelated attributes must be kept: %v", vm)
	}
	if vm["admin_password"] != importedPasswordPlaceholder {
		t.Errorf("import sentinel must be replaced by a valid placeholder: %v", vm["admin_password"])
	}
	vmss := byName["vm-in-vmss"].Item
	if vmss["platform_fault_domain"] != "1" {
		t.Errorf("real platform_fault_domain must be kept")
	}
	if _, ok := vmss["os_managed_disk_id"]; ok {
		t.Errorf("os_managed_disk_id must be dropped when source_image_id is set")
	}
	disk := byName["vm-from-disk"].Item
	if disk["os_managed_disk_id"] == nil {
		t.Errorf("os_managed_disk_id must be kept on a VM without image: %v", disk)
	}
	for _, gone := range []string{"admin_username", "admin_password", "patch_mode"} {
		if _, ok := disk[gone]; ok {
			t.Errorf("%s conflicts with os_managed_disk_id and must be dropped: %v", gone, disk)
		}
	}
	osDisk := disk["os_disk"].([]interface{})[0].(map[string]interface{})
	if _, ok := osDisk["name"]; ok || osDisk["caching"] != "ReadWrite" {
		t.Errorf("os_disk.name must be dropped, os_disk.caching kept: %v", osDisk)
	}
	if disk["size"] != "Standard_B2s" {
		t.Errorf("size must be kept: %v", disk)
	}

	if _, ok := byName["ehns"].Item["minimum_tls_version"]; ok {
		t.Errorf("minimum_tls_version 1.0 must be dropped")
	}
	if byName["ehns12"].Item["minimum_tls_version"] != "1.2" {
		t.Errorf("minimum_tls_version 1.2 must be kept")
	}
	legacy := byName["legacy-web"].Item["site_config"].([]interface{})[0].(map[string]interface{})["application_stack"].([]interface{})[0].(map[string]interface{})
	if _, ok := legacy["php_version"]; ok {
		t.Errorf("legacy php_version must be dropped: %v", legacy)
	}
	if _, ok := legacy["node_version"]; ok {
		t.Errorf("legacy node_version must be dropped: %v", legacy)
	}
	if legacy["current_stack"] != "php" {
		t.Errorf("sibling attributes must be kept: %v", legacy)
	}
	modern := byName["modern-web"].Item["site_config"].([]interface{})[0].(map[string]interface{})["application_stack"].([]interface{})[0].(map[string]interface{})
	if modern["php_version"] != "7.4" || modern["node_version"] != "~18" {
		t.Errorf("supported runtime versions must be kept: %v", modern)
	}
	if byName["shared"].Item["max_shares"] != "2" {
		t.Errorf("max_shares >= 2 must be kept")
	}
	if _, ok := byName["single"].Item["max_shares"]; ok {
		t.Errorf("max_shares = 1 must be dropped")
	}

	fn := byName["fn"].Item
	if _, ok := fn["key_vault_reference_identity_id"]; ok {
		t.Errorf("SystemAssigned placeholder must be dropped: %v", fn)
	}
	if _, ok := fn["storage_uses_managed_identity"]; ok {
		t.Errorf("storage_uses_managed_identity = false must be dropped: %v", fn)
	}
	if fn["storage_account_access_key"] != "secret" {
		t.Errorf("storage_account_access_key must be kept: %v", fn)
	}
	if byName["web-uai"].Item["key_vault_reference_identity_id"] == nil {
		t.Errorf("real user-assigned identity ID must be kept")
	}

	alert := byName["alert"].Item
	if got := alert["scopes"].([]interface{})[0]; got != "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Insights/components/c" {
		t.Errorf("scopes not canonicalized: %v", got)
	}
	criteria := alert["application_insights_web_test_location_availability_criteria"].([]interface{})[0].(map[string]interface{})
	if criteria["web_test_id"] != "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Insights/webTests/w" {
		t.Errorf("nested ID not canonicalized: %v", criteria)
	}
	if _, ok := alert["tags"].(map[string]interface{})["hidden-link:/subscriptions/s/resourcegroups/rg"]; !ok {
		t.Errorf("tags must not be touched: %v", alert["tags"])
	}
}

func TestSetAttributeKeepsStateAndConfigInSync(t *testing.T) {
	r := convertedResource("azurerm_windows_virtual_machine", "vm", map[string]interface{}{"admin_password": importedPasswordSentinel})
	r.InstanceState.Attributes = map[string]string{"admin_password": importedPasswordSentinel}
	r.StateJSON = []byte(`{"admin_password":"ignored-as-imported","name":"vm"}`)

	if err := setAttribute(&r, "admin_password", importedPasswordPlaceholder); err != nil {
		t.Fatal(err)
	}
	if r.Item["admin_password"] != importedPasswordPlaceholder || r.InstanceState.Attributes["admin_password"] != importedPasswordPlaceholder {
		t.Errorf("item/flatmap not updated: %v %v", r.Item, r.InstanceState.Attributes)
	}
	var state map[string]string
	if err := json.Unmarshal(r.StateJSON, &state); err != nil {
		t.Fatal(err)
	}
	if state["admin_password"] != importedPasswordPlaceholder || state["name"] != "vm" {
		t.Errorf("typed state not updated consistently: %v", state)
	}
}

func TestApplyKeepZero(t *testing.T) {
	ai := convertedResource("azurerm_application_insights", "ai", nil)
	applyKeepZero(&ai)
	if len(ai.AllowEmptyValues) != 1 || ai.AllowEmptyValues[0] != "^sampling_percentage$" {
		t.Errorf("sampling_percentage zero must be kept: %v", ai.AllowEmptyValues)
	}
	other := convertedResource("azurerm_resource_group", "rg", nil)
	applyKeepZero(&other)
	if len(other.AllowEmptyValues) != 0 {
		t.Errorf("no keep-zero patterns expected: %v", other.AllowEmptyValues)
	}
}

func TestIsDefaultConsumerGroup(t *testing.T) {
	str := func(s string) *string { return &s }
	if !isDefaultConsumerGroup(str("$Default")) || !isDefaultConsumerGroup(nil) {
		t.Errorf("$Default and nil must be treated as the built-in group")
	}
	if isDefaultConsumerGroup(str("orders-processor")) {
		t.Errorf("user consumer groups must be kept")
	}
}

func TestIsSystemScopeMap(t *testing.T) {
	str := func(s string) *string { return &s }
	cases := map[string]struct {
		item *armcontainerregistry.ScopeMap
		want bool
	}{
		"nil":    {nil, true},
		"system": {&armcontainerregistry.ScopeMap{ID: str("/id"), Name: str("_repositories_admin")}, true},
		"user":   {&armcontainerregistry.ScopeMap{ID: str("/id"), Name: str("ci-pull")}, false},
	}
	for name, tc := range cases {
		if got := isSystemScopeMap(tc.item); got != tc.want {
			t.Errorf("%s: got %v, want %v", name, got, tc.want)
		}
	}
}
