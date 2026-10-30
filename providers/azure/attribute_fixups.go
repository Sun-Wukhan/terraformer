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
	"fmt"
	"strconv"
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// This file holds the Azure specific clean-up of refreshed resources. Every
// rule documents a value azurerm itself writes to state yet refuses to accept
// back as configuration (validation ranges, ConflictsWith/ExactlyOneOf,
// placeholders). Leaving such an attribute out of the generated HCL lets the
// provider keep the live value, which is what a faithful import needs.

// attributeDropRule reports whether an attribute carrying value must be left
// out of the generated configuration.
type attributeDropRule func(value interface{}) bool

// dropAlways drops the attribute regardless of its value.
func dropAlways(interface{}) bool { return true }

// dropEqual drops the attribute when its string value equals one of values.
func dropEqual(values ...string) attributeDropRule {
	return func(value interface{}) bool {
		v, ok := value.(string)
		if !ok {
			return false
		}
		for _, want := range values {
			if v == want {
				return true
			}
		}
		return false
	}
}

// dropNumberBelow drops a numeric attribute whose value is below min.
func dropNumberBelow(min float64) attributeDropRule {
	return func(value interface{}) bool {
		v, ok := value.(string)
		if !ok {
			return false
		}
		f, err := strconv.ParseFloat(v, 64)
		return err == nil && f < min
	}
}

// dropUnlessPrefix drops a string attribute whose value does not start with
// prefix.
func dropUnlessPrefix(prefix string) attributeDropRule {
	return func(value interface{}) bool {
		v, ok := value.(string)
		return ok && !strings.HasPrefix(v, prefix)
	}
}

// attributePath addresses an attribute inside r.Item: segments are block or
// attribute names and every list block along the way is traversed
// element-wise, e.g. "site_config/application_stack/php_version".
type attributePath = string

// appServiceDrops lists attributes the App Service resources report with a
// value their own validation rejects.
var appServiceDrops = map[attributePath]attributeDropRule{
	// Must be a user-assigned identity ID, but Azure returns the literal
	// "SystemAssigned" for the default.
	"key_vault_reference_identity_id": dropEqual("SystemAssigned"),
	// false conflicts with the storage_account_access_key reported alongside.
	"storage_uses_managed_identity": dropEqual("false"),
	// Legacy runtimes the current provider no longer lists as valid values.
	"site_config/application_stack/php_version":  dropEqual("5.6", "7.0", "7.2", "7.3"),
	"site_config/application_stack/node_version": dropUnlessPrefix("~"),
}

// eventHubNamespaceDrops lists attributes of Event Hubs / Service Bus
// namespaces azurerm cannot accept back: minimum_tls_version only validates
// "1.2" while existing namespaces may still report 1.0/1.1.
var eventHubNamespaceDrops = map[attributePath]attributeDropRule{
	"minimum_tls_version": dropEqual("1.0", "1.1"),
}

// azureAttributeDrops maps a resource type to the attributes that must be
// dropped from its generated configuration.
var azureAttributeDrops = map[string]map[attributePath]attributeDropRule{
	// max_shares is Optional+Computed and validated to 2..10, but Azure
	// reports 1 for every disk that is not shared.
	"azurerm_managed_disk": {"max_shares": dropNumberBelow(2)},
	// message_retention (Optional+Computed) and the retention_description
	// block are mutually exclusive; keep the newer block.
	"azurerm_eventhub":             {"message_retention": dropAlways},
	"azurerm_eventhub_namespace":   eventHubNamespaceDrops,
	"azurerm_servicebus_namespace": eventHubNamespaceDrops,

	"azurerm_linux_function_app":        appServiceDrops,
	"azurerm_linux_function_app_slot":   appServiceDrops,
	"azurerm_linux_web_app":             appServiceDrops,
	"azurerm_linux_web_app_slot":        appServiceDrops,
	"azurerm_windows_function_app":      appServiceDrops,
	"azurerm_windows_function_app_slot": appServiceDrops,
	"azurerm_windows_web_app":           appServiceDrops,
	"azurerm_windows_web_app_slot":      appServiceDrops,
}

// azureKeepZero lists, per resource type, flatmap-key patterns of optional
// numeric attributes for which 0 is a real setting rather than the SDK's
// "unset" marker, so the zero must stay in the configuration (the provider
// default would otherwise be planned, e.g. sampling_percentage 0 -> 100).
var azureKeepZero = map[string][]string{
	"azurerm_application_insights": {"^sampling_percentage$"},
}

// importedPasswordSentinel is what azurerm writes to admin_password when a VM
// is imported (the API never returns passwords). It fails the provider's own
// password validation, and admin_password is ForceNew, so the configuration
// and the state must carry the same, valid placeholder instead.
const importedPasswordSentinel = "ignored-as-imported"

// importedPasswordPlaceholder satisfies the Windows/Linux admin password
// rules (upper, lower, digit, special character, length) while making its
// nature obvious. It is a placeholder for a value the API does not expose,
// not a credential.
const importedPasswordPlaceholder = "Ignored-As-Imported-1"

// virtualMachineTypes lists the VM resources handled by fixVirtualMachine.
var virtualMachineTypes = map[string]bool{
	"azurerm_linux_virtual_machine":   true,
	"azurerm_windows_virtual_machine": true,
}

// osManagedDiskConflicts are the attributes azurerm declares ConflictsWith
// (or ExactlyOneOf) os_managed_disk_id: a VM created from an existing OS disk
// has no OS profile, so none of them may be set alongside it.
var osManagedDiskConflicts = []attributePath{
	"admin_password", "admin_username", "computer_name", "custom_data",
	"automatic_updates_enabled", "bypass_platform_safety_checks_on_user_schedule_enabled",
	"disable_password_authentication", "hotpatching_enabled", "patch_assessment_mode",
	"patch_mode", "provision_vm_agent", "reboot_setting",
	"os_disk/name", "os_disk/storage_account_type", "os_disk/diff_disk_settings",
}

// fixVirtualMachine reconciles the exactly-one-of between os_managed_disk_id
// and source_image_id/source_image_reference, which azurerm reports together:
// a VM created from an image keeps the image and drops os_managed_disk_id
// (Optional+Computed, and in conflict with most other arguments); a VM
// created from an existing disk keeps os_managed_disk_id and drops everything
// that conflicts with it. platform_fault_domain = -1 is the "not in a scale
// set" marker, only valid together with virtual_machine_scale_set_id.
func fixVirtualMachine(r *terraformutils.Resource) {
	hasImage := !isBlank(r.Item["source_image_id"]) || !isBlankCollection(r.Item["source_image_reference"])
	if hasImage {
		delete(r.Item, "os_managed_disk_id")
	} else {
		for _, path := range osManagedDiskConflicts {
			dropAtPath(r.Item, strings.Split(path, "/"), dropAlways)
		}
	}
	dropAtPath(r.Item, []string{"platform_fault_domain"}, dropNumberBelow(0))
	if r.Item["admin_password"] == importedPasswordSentinel {
		if err := setAttribute(r, "admin_password", importedPasswordPlaceholder); err != nil {
			delete(r.Item, "admin_password")
		}
	}
}

// setAttribute sets a top-level string attribute consistently in the HCL
// item, the flatmap state and the typed state JSON, so configuration and
// state agree and no diff (or replacement) is planned.
func setAttribute(r *terraformutils.Resource, key, value string) error {
	r.Item[key] = value
	if r.InstanceState != nil && r.InstanceState.Attributes != nil {
		r.InstanceState.Attributes[key] = value
	}
	if len(r.StateJSON) == 0 {
		return nil
	}
	var attrs map[string]json.RawMessage
	if err := json.Unmarshal(r.StateJSON, &attrs); err != nil {
		return err
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return err
	}
	attrs[key] = encoded
	r.StateJSON, err = json.Marshal(attrs)
	return err
}

// dropAtPath removes the attribute addressed by path from item when rule
// matches its value, descending into every element of list blocks.
func dropAtPath(item map[string]interface{}, path []string, rule attributeDropRule) {
	if len(path) == 0 || item == nil {
		return
	}
	if len(path) == 1 {
		if value, ok := item[path[0]]; ok && rule(value) {
			delete(item, path[0])
		}
		return
	}
	switch child := item[path[0]].(type) {
	case map[string]interface{}:
		dropAtPath(child, path[1:], rule)
	case []interface{}:
		for _, element := range child {
			if block, ok := element.(map[string]interface{}); ok {
				dropAtPath(block, path[1:], rule)
			}
		}
	}
}

// applyAttributeDrops removes from r.Item every attribute an
// azureAttributeDrops rule matches for r's type.
func applyAttributeDrops(r *terraformutils.Resource) {
	if r.InstanceInfo == nil || r.Item == nil {
		return
	}
	for path, rule := range azureAttributeDrops[r.InstanceInfo.Type] {
		dropAtPath(r.Item, strings.Split(path, "/"), rule)
	}
}

// nonEmptyRequired lists, per resource type, Required string attributes that
// azurerm validates as non-empty although Azure accepts (and returns) empty
// values for them. A resource with such an empty value has no valid
// configuration and is dropped with a report.
var nonEmptyRequired = map[string][]attributePath{
	"azurerm_api_management_subscription": {"display_name"},
	"azurerm_api_management_user":         {"first_name", "last_name", "email"},
}

// unrepresentableReason returns a non-empty reason when r cannot be expressed
// as valid azurerm configuration at all, so it must be dropped (and reported)
// rather than emitted as a resource Terraform rejects.
func unrepresentableReason(r *terraformutils.Resource) string {
	if r.InstanceInfo == nil {
		return ""
	}
	switch r.InstanceInfo.Type {
	case "azurerm_app_service_certificate":
		// The resource needs the certificate material: either a Key Vault
		// secret or the PFX blob. Azure never returns the blob, so a
		// certificate uploaded directly cannot be round-tripped.
		if isBlank(r.Item["key_vault_secret_id"]) && isBlank(r.Item["pfx_blob"]) {
			return "certificate was uploaded directly (no key_vault_secret_id) and Azure does not return its pfx_blob, " +
				"which azurerm_app_service_certificate requires"
		}
	case "azurerm_api_management_logger":
		// An Event Hub logger needs its connection string or endpoint URI,
		// neither of which the API returns.
		if hubs, ok := r.Item["eventhub"].([]interface{}); ok && len(hubs) > 0 {
			hub, _ := hubs[0].(map[string]interface{})
			if isBlank(hub["connection_string"]) && isBlank(hub["endpoint_uri"]) {
				return "Azure does not return the Event Hub connection_string/endpoint_uri the logger requires"
			}
		}
	case "azurerm_api_management_named_value":
		// Secret named values never return their value.
		if isBlank(r.Item["value"]) && isBlankCollection(r.Item["value_from_key_vault"]) {
			return "named value is secret: Azure returns neither value nor value_from_key_vault"
		}
	}
	for _, attr := range nonEmptyRequired[r.InstanceInfo.Type] {
		if value, ok := r.Item[attr]; ok && isBlank(value) {
			return fmt.Sprintf("required attribute %q is empty in Azure, which azurerm rejects", attr)
		}
	}
	return ""
}

// isBlank reports whether an Item value is missing or an empty string.
func isBlank(value interface{}) bool {
	s, ok := value.(string)
	return value == nil || (ok && strings.TrimSpace(s) == "")
}

// isBlankCollection reports whether an Item value is missing or an empty
// list/map (a block with no instances).
func isBlankCollection(value interface{}) bool {
	switch v := value.(type) {
	case nil:
		return true
	case []interface{}:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0
	}
	return false
}

// canonicalizeIDValues rewrites every string in item that looks like an ARM
// resource ID to the casing azurerm's ID parsers require. Some Azure APIs
// return lower-cased IDs (e.g. `/resourcegroups/x/providers/microsoft.insights/...`)
// inside attributes such as metric alert scopes, and azurerm refuses to even
// parse them. Tags are skipped: they are free-form and `hidden-link:` keys
// must stay byte-identical.
func canonicalizeIDValues(item map[string]interface{}) {
	for key, value := range item {
		if key == "tags" {
			continue
		}
		item[key] = canonicalizeIDValue(value)
	}
}

// canonicalizeIDValue applies canonicalAzureID to value and, recursively, to
// its nested maps and lists.
func canonicalizeIDValue(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		if strings.HasPrefix(v, "/subscriptions/") {
			return canonicalAzureID(v)
		}
	case map[string]interface{}:
		for k, inner := range v {
			v[k] = canonicalizeIDValue(inner)
		}
	case []interface{}:
		for i, inner := range v {
			v[i] = canonicalizeIDValue(inner)
		}
	}
	return value
}

// applyKeepZero appends the azureKeepZero patterns of r's type to its
// AllowEmptyValues so the flatmap parser keeps those zeros. It runs before
// the refresh/convert step.
func applyKeepZero(r *terraformutils.Resource) {
	if r.InstanceInfo == nil {
		return
	}
	r.AllowEmptyValues = append(r.AllowEmptyValues, azureKeepZero[r.InstanceInfo.Type]...)
}

// PostRefreshCleanup runs the generic filter cleanup and then the Azure
// specific fix-ups on the refreshed, converted resources: attribute drops,
// VM reconciliation, ID casing inside attribute values, and removal (with a
// reported reason) of resources azurerm cannot represent.
func (az *AzureService) PostRefreshCleanup() {
	az.Service.PostRefreshCleanup()
	kept := az.Resources[:0]
	for i := range az.Resources {
		r := &az.Resources[i]
		if reason := unrepresentableReason(r); reason != "" {
			az.RecordFailure(r.InstanceInfo.Type, fmt.Errorf("%s %s skipped: %s", r.InstanceInfo.Type, r.ResourceName, reason))
			continue
		}
		if r.InstanceInfo != nil && r.Item != nil && virtualMachineTypes[r.InstanceInfo.Type] {
			fixVirtualMachine(r)
		}
		applyAttributeDrops(r)
		canonicalizeIDValues(r.Item)
		kept = append(kept, *r)
	}
	az.Resources = kept
}
