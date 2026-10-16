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
	// Every target gets an import block and a matching empty resource block.
	for _, tg := range targets {
		if !strings.Contains(got, "to = "+tg.address) {
			t.Fatalf("missing import block for %s:\n%s", tg.address, got)
		}
		if !strings.Contains(got, `resource "`+tg.resource.InstanceInfo.Type+`" "`+tg.label+`"`) {
			t.Fatalf("missing resource block for %s:\n%s", tg.address, got)
		}
	}
	// Duplicate names deduped with suffix.
	if !strings.Contains(got, "_2") {
		t.Fatalf("expected deduped _2 label:\n%s", got)
	}
}
