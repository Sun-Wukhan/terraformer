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

package terraformutils

import (
	"encoding/json"
	"reflect"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/terraform"
	"github.com/zclconf/go-cty/cty"
)

func TestPrintTfStateWritesVersion4(t *testing.T) {
	typed := NewSimpleResource("/vnet", "vnet", "azurerm_virtual_network", "azurerm", nil)
	typed.InstanceState.Meta = map[string]interface{}{"schema_version": "2"}
	typed.StateJSON = json.RawMessage(`{"id":"/vnet","name":"Devtest-vnet","flow_timeout_in_minutes":0}`)
	typed.Outputs = map[string]*terraform.OutputState{
		"azurerm_virtual_network_tfer--vnet_id": {Type: "string", Value: "/vnet"},
	}
	legacy := NewSimpleResource("/nsg", "nsg", "azurerm_network_security_group", "azurerm", nil)
	legacy.InstanceState.Attributes = map[string]string{"name": "nsg"}

	out, err := PrintTfState([]Resource{legacy, typed})
	if err != nil {
		t.Fatalf("PrintTfState: %v", err)
	}
	var state struct {
		Version          int    `json:"version"`
		TerraformVersion string `json:"terraform_version"`
		Lineage          string `json:"lineage"`
		Outputs          map[string]struct {
			Value string          `json:"value"`
			Type  json.RawMessage `json:"type"`
		} `json:"outputs"`
		Resources []struct {
			Mode      string `json:"mode"`
			Type      string `json:"type"`
			Name      string `json:"name"`
			Provider  string `json:"provider"`
			Instances []struct {
				SchemaVersion  uint64            `json:"schema_version"`
				Attributes     json.RawMessage   `json:"attributes"`
				AttributesFlat map[string]string `json:"attributes_flat"`
			} `json:"instances"`
		} `json:"resources"`
	}
	if err := json.Unmarshal(out, &state); err != nil {
		t.Fatalf("state is not valid JSON: %v\n%s", err, out)
	}
	if state.Version != 4 {
		t.Fatalf("version = %d, want 4", state.Version)
	}
	if !regexp.MustCompile(`^\d+\.\d+\.\d+`).MatchString(state.TerraformVersion) {
		t.Fatalf("terraform_version %q is not a version", state.TerraformVersion)
	}
	if !regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`).MatchString(state.Lineage) {
		t.Fatalf("lineage %q is not a UUID", state.Lineage)
	}
	if o := state.Outputs["azurerm_virtual_network_tfer--vnet_id"]; o.Value != "/vnet" || string(o.Type) != `"string"` {
		t.Fatalf("output not written as v4 {value,type}: %+v", state.Outputs)
	}
	if len(state.Resources) != 2 {
		t.Fatalf("got %d resources, want 2", len(state.Resources))
	}
	// Sorted by type: network_security_group before virtual_network.
	nsg, vnet := state.Resources[0], state.Resources[1]
	for _, r := range state.Resources {
		if r.Mode != "managed" || r.Provider != `provider["registry.terraform.io/hashicorp/azurerm"]` {
			t.Fatalf("bad mode/provider address: %+v", r)
		}
	}
	if vnet.Type != "azurerm_virtual_network" || vnet.Name != "tfer--vnet" {
		t.Fatalf("unexpected resource order/name: %+v", vnet)
	}
	var gotAttrs, wantAttrs map[string]interface{}
	_ = json.Unmarshal(vnet.Instances[0].Attributes, &gotAttrs)
	_ = json.Unmarshal(typed.StateJSON, &wantAttrs)
	if vnet.Instances[0].SchemaVersion != 2 || !reflect.DeepEqual(gotAttrs, wantAttrs) {
		t.Fatalf("typed attributes not written: got %v, want %v", gotAttrs, wantAttrs)
	}
	if nsg.Instances[0].Attributes != nil || nsg.Instances[0].AttributesFlat["name"] != "nsg" || nsg.Instances[0].AttributesFlat["id"] != "/nsg" {
		t.Fatalf("legacy resource must fall back to attributes_flat with id: %+v", nsg.Instances[0])
	}
}

func TestFlatmapParserDropsZeroNumbersUnlessRequired(t *testing.T) {
	attributes := map[string]string{
		"flow_timeout_in_minutes": "0",
		"priority":                "0",
		"count":                   "3",
		"code":                    "0",
	}
	parser := NewFlatmapParser(attributes, nil, []*regexp.Regexp{regexp.MustCompile(`^priority$`)})
	result, err := parser.Parse(cty.Object(map[string]cty.Type{
		"flow_timeout_in_minutes": cty.Number,
		"priority":                cty.Number,
		"count":                   cty.Number,
		"code":                    cty.String,
	}))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if _, ok := result["flow_timeout_in_minutes"]; ok {
		t.Errorf("optional zero number must be dropped: %v", result)
	}
	if result["priority"] != "0" {
		t.Errorf("required zero number must be kept: %v", result)
	}
	if result["count"] != "3" {
		t.Errorf("non-zero number must be kept: %v", result)
	}
	if result["code"] != "0" {
		t.Errorf("string \"0\" is not a zero number and must be kept: %v", result)
	}
}

func TestFlatmapParserEscapesTemplateSequences(t *testing.T) {
	attributes := map[string]string{
		"body":  "Dear ${DevFirstName}, %{ if x } ok",
		"plain": "no templates here",
	}
	result, err := NewFlatmapParser(attributes, nil, nil).Parse(cty.Object(map[string]cty.Type{
		"body":  cty.String,
		"plain": cty.String,
	}))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if result["body"] != "Dear $${DevFirstName}, %%{ if x } ok" {
		t.Errorf("template sequences not escaped: %q", result["body"])
	}
	if result["plain"] != "no templates here" {
		t.Errorf("plain value altered: %q", result["plain"])
	}
}

func TestFlatmapParserDropsZeroOptionalComputed(t *testing.T) {
	attributes := map[string]string{
		"use_custom_runtime":     "false", // optional+computed: dropped
		"enabled":                "false", // plain optional: kept
		"https_only":             "true",  // optional+computed but non-zero: kept
		"required_flag":          "false", // optional+computed but required-pattern wins: kept
		"stack.#":                "1",
		"stack.0.use_dotnet":     "false", // nested optional+computed: dropped
		"stack.0.dotnet_version": "v8.0",
	}
	dropZero := []*regexp.Regexp{
		regexp.MustCompile(`^use_custom_runtime$`),
		regexp.MustCompile(`^https_only$`),
		regexp.MustCompile(`^required_flag$`),
		regexp.MustCompile(`^stack\.[0-9]+\.use_dotnet$`),
	}
	parser := NewFlatmapParser(attributes, nil, []*regexp.Regexp{regexp.MustCompile(`^required_flag$`)}).WithDropZeroKeys(dropZero)
	result, err := parser.Parse(cty.Object(map[string]cty.Type{
		"use_custom_runtime": cty.Bool,
		"enabled":            cty.Bool,
		"https_only":         cty.Bool,
		"required_flag":      cty.Bool,
		"stack": cty.List(cty.Object(map[string]cty.Type{
			"use_dotnet":     cty.Bool,
			"dotnet_version": cty.String,
		})),
	}))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if _, ok := result["use_custom_runtime"]; ok {
		t.Errorf("optional+computed false must be dropped: %v", result)
	}
	if result["enabled"] != "false" {
		t.Errorf("plain optional false must be kept: %v", result)
	}
	if result["https_only"] != "true" {
		t.Errorf("non-zero optional+computed must be kept: %v", result)
	}
	if result["required_flag"] != "false" {
		t.Errorf("required attribute must be kept even when false: %v", result)
	}
	stack := result["stack"].([]interface{})[0].(map[string]interface{})
	if _, ok := stack["use_dotnet"]; ok || stack["dotnet_version"] != "v8.0" {
		t.Errorf("nested optional+computed false must be dropped, others kept: %v", stack)
	}
}
