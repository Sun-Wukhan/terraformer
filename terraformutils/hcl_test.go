// Copyright 2020 The Terraformer Authors.
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
	"strings"
	"testing"
)

func TestPrintResource(t *testing.T) {
	var resources []Resource
	var nested []map[string]interface{}
	nested = append(nested, mapI("field1", "egg"))
	importResource := prepare("ID1", "type1", map[string]string{
		"type1":                  "ID2",
		"map1.%":                 "1",
		"map1.foo":               "bar",
		"nested.#":               "1",
		"nested.0.map1.#":        "1",
		"nested.0.map1.0.field1": "egg",
		"nested2.#":              "1",
		"nested2.0.field1":       "spam",
		"nested2.0.map2.%":       "1",
		"nested2.0.map2.foo":     "bar",
	}, map[string]interface{}{
		"type1":   "ID2",
		"map1":    mapI("foo", "bar"),
		"nested":  mapI("map1", nested),
		"nested2": map[string]interface{}{"map2": mapI("bar", "foo"), "field1": "egg"},
	})
	resources = append(resources, importResource)
	providerData := map[string]interface{}{}
	output := "hcl"
	data, _ := HclPrintResource(resources, providerData, output, true)

	if strings.Count(string(data), "map1 = ") != 1 {
		t.Errorf("failed to parse data %s", string(data))
	}
	if strings.Count(string(data), "map2 = ") != 1 {
		t.Errorf("failed to parse data %s", string(data))
	}
}

func TestHclPrintResourceKeepsListOrder(t *testing.T) {
	importResource := prepare("ID1", "type1", map[string]string{
		"perms.#":         "3",
		"perms.0":         "Update",
		"perms.1":         "Get",
		"perms.2":         "Backup",
		"rule.#":          "2",
		"rule.0.priority": "200",
		"rule.1.priority": "100",
	}, map[string]interface{}{
		"perms": []interface{}{"Update", "Get", "Backup"},
		"rule": []interface{}{
			map[string]interface{}{"priority": "200"},
			map[string]interface{}{"priority": "100"},
		},
		"zeta":  "last",
		"alpha": "first",
	})
	data, err := HclPrintResource([]Resource{importResource}, map[string]interface{}{}, "hcl", true)
	if err != nil {
		t.Fatal(err)
	}
	out := string(data)
	if !strings.Contains(out, `["Update", "Get", "Backup"]`) {
		t.Errorf("list values were reordered:\n%s", out)
	}
	if strings.Index(out, `"200"`) > strings.Index(out, `"100"`) {
		t.Errorf("repeated blocks were reordered:\n%s", out)
	}
	if strings.Index(out, "alpha") > strings.Index(out, "zeta") {
		t.Errorf("attributes not sorted by key:\n%s", out)
	}
}

func TestStripInterpolationOnly(t *testing.T) {
	in := "value = \"${azurerm_resource_group.tfer--rg1.id}\"\n" +
		"name = \"prefix-${var.foo}-suffix\"\n" +
		"ref = \"${data.terraform_remote_state.vpc.outputs.vpc_id}\"\n"
	got := string(stripInterpolationOnly([]byte(in)))
	want := "value = azurerm_resource_group.tfer--rg1.id\n" +
		"name = \"prefix-${var.foo}-suffix\"\n" +
		"ref = data.terraform_remote_state.vpc.outputs.vpc_id\n"
	if got != want {
		t.Errorf("stripInterpolationOnly() =\n%s\nwant\n%s", got, want)
	}
}
