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
	"strings"
	"testing"

	"github.com/hashicorp/terraform/providers"
	"github.com/hashicorp/terraform/terraform"
)

func testImportResource(id, resourceName, resourceType string) *Resource {
	return &Resource{
		ResourceName: resourceName,
		InstanceInfo: &terraform.InstanceInfo{Type: resourceType},
		InstanceState: &terraform.InstanceState{
			ID:         id,
			Attributes: map[string]string{},
		},
	}
}

func TestAssignImportTargets(t *testing.T) {
	resources := []*Resource{
		testImportResource("i-123", "web server", "aws_instance"),
		testImportResource("i-456", "web server", "aws_instance"), // duplicate name
		testImportResource(`quo"te`, "db", "aws_db_instance"),     // quote in ID
		testImportResource("x", "db", "aws_db_instance"),          // duplicate name, other type
		testImportResource("", "empty", "aws_instance"),           // empty ID: skipped
		nil, // nil resource: skipped
	}

	targets, addrToResource := assignImportTargets(resources)

	if len(targets) != 4 {
		t.Fatalf("got %d targets, want 4", len(targets))
	}

	// Labels must be valid HCL identifiers and unique per type.
	seen := map[string]bool{}
	for _, tg := range targets {
		for _, r := range tg.label {
			if !(r == '_' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9') {
				t.Fatalf("label %q contains invalid identifier char %q", tg.label, r)
			}
		}
		key := tg.address
		if seen[key] {
			t.Fatalf("duplicate address %q", key)
		}
		seen[key] = true
		if addrToResource[key] != tg.resource {
			t.Fatalf("address map mismatch for %q", key)
		}
	}

	// Duplicate names within one type get numeric suffixes; the same name in
	// another type does not collide.
	var instanceLabels []string
	for _, tg := range targets {
		if strings.HasPrefix(tg.address, "aws_instance.") {
			instanceLabels = append(instanceLabels, tg.label)
		}
	}
	if len(instanceLabels) != 2 || instanceLabels[0] == instanceLabels[1] {
		t.Fatalf("duplicate names not deduped: %v", instanceLabels)
	}
	if !strings.HasSuffix(instanceLabels[1], "_2") {
		t.Fatalf("expected _2 suffix for duplicate, got %q", instanceLabels[1])
	}
}

func TestRenderImportConfig(t *testing.T) {
	resources := []*Resource{
		testImportResource(`we"ird\id`, "my bucket", "aws_s3_bucket"),
		testImportResource("i-1", "my bucket", "aws_s3_bucket"),
	}
	targets, _ := assignImportTargets(resources)
	got := renderImportConfig(targets)

	// ID with quote and backslash must be escaped, not break the HCL string.
	if !strings.Contains(got, `id = "we\"ird\\id"`) {
		t.Fatalf("ID not properly quoted/escaped:\n%s", got)
	}
	// Every target gets an import block; resource blocks are left for
	// Terraform's -generate-config-out, since empty ones fail validation.
	for _, tg := range targets {
		if !strings.Contains(got, "to = "+tg.address) {
			t.Fatalf("missing import block for %s:\n%s", tg.address, got)
		}
	}
	if strings.Contains(got, `resource "`) {
		t.Fatalf("resource blocks must not be rendered (they would fail validation):\n%s", got)
	}
	// Duplicate names deduped with suffix.
	if !strings.Contains(got, "_2") {
		t.Fatalf("expected deduped _2 label:\n%s", got)
	}
}

func TestDropUnknownTypes(t *testing.T) {
	known := testImportResource("/vnet", "vnet", "azurerm_virtual_network")
	removed := testImportResource("/site", "site", "azurerm_app_service")
	schema := &providers.GetSchemaResponse{
		ResourceTypes: map[string]providers.Schema{"azurerm_virtual_network": {}},
	}
	report := NewImportReport()

	kept := dropUnknownTypes([]*Resource{known, removed, nil}, schema, report, map[*Resource]string{removed: "app_service"})

	if len(kept) != 2 || kept[0] != known || kept[1] != nil {
		t.Fatalf("got %v, want the known resource and the nil passthrough", kept)
	}
	failures := report.Failures()
	if len(failures) != 1 || failures[0].ResourceType != "azurerm_app_service" || failures[0].Service != "app_service" {
		t.Fatalf("unknown type not recorded: %+v", failures)
	}
	if got := dropUnknownTypes([]*Resource{removed}, nil, nil, nil); len(got) != 1 {
		t.Fatalf("nil schema must keep everything, got %v", got)
	}
}

func TestVerifyImportOnlyPlan(t *testing.T) {
	tests := []struct {
		name    string
		plan    string
		wantErr bool
	}{
		{"empty plan", `{}`, false},
		{"pure imports", `{"resource_changes":[
			{"address":"a.x","change":{"actions":["no-op"],"importing":{"id":"1"}}},
			{"address":"a.y","change":{"actions":["no-op"],"importing":{"id":"2"}}}]}`, false},
		{"import with update", `{"resource_changes":[
			{"address":"a.x","change":{"actions":["no-op"],"importing":{"id":"1"}}},
			{"address":"a.y","change":{"actions":["update"],"importing":{"id":"2"}}}]}`, true},
		{"import with replace", `{"resource_changes":[
			{"address":"a.x","change":{"actions":["delete","create"],"importing":{"id":"1"}}}]}`, true},
		{"no-op without import", `{"resource_changes":[
			{"address":"a.x","change":{"actions":["no-op"]}}]}`, true},
		{"create", `{"resource_changes":[
			{"address":"a.x","change":{"actions":["create"],"importing":null}}]}`, true},
		{"invalid json", `{`, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := verifyImportOnlyPlan([]byte(tt.plan))
			if (err != nil) != tt.wantErr {
				t.Fatalf("verifyImportOnlyPlan() err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
