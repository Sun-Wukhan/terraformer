// Copyright 2026 The Terraformer Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package terraformutils

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform/terraform"
)

func TestImportReportCounts(t *testing.T) {
	r := NewImportReport()
	r.RecordImported("aws_instance", 3)
	r.RecordImported("aws_instance", 2)
	r.RecordImported("aws_vpc", 1)
	r.RecordImported("aws_vpc", 0) // no-op
	if got := r.ImportedCount(); got != 6 {
		t.Fatalf("ImportedCount = %d, want 6", got)
	}
	if got := r.ImportedByType()["aws_instance"]; got != 5 {
		t.Fatalf("ImportedByType[aws_instance] = %d, want 5", got)
	}
	if r.HasFailures() {
		t.Fatal("HasFailures = true before any failure recorded")
	}
	r.RecordFailure(ImportFailure{Stage: StageRefresh, ResourceType: "aws_instance", Error: "boom"})
	r.RecordFailure(ImportFailure{Stage: StageEnumerate, Service: "iam", Error: "access denied"})
	if !r.HasFailures() || r.FailureCount() != 2 {
		t.Fatalf("FailureCount = %d, want 2", r.FailureCount())
	}
}

func TestImportReportNilSafe(t *testing.T) {
	var r *ImportReport
	r.RecordImported("aws_instance", 1)
	r.RecordFailure(ImportFailure{Error: "x"})
	if r.FailureCount() != 0 || r.ImportedCount() != 0 || r.HasFailures() {
		t.Fatal("nil report should report zeroes")
	}
	if err := r.WriteFailuresJSON(filepath.Join(t.TempDir(), "f.json"), "aws"); err != nil {
		t.Fatalf("nil WriteFailuresJSON: %v", err)
	}
	r.PrintSummary("aws") // must not panic
}

func TestImportReportGrouping(t *testing.T) {
	r := NewImportReport()
	for i := 0; i < 3; i++ {
		r.RecordFailure(ImportFailure{Stage: StageRefresh, Service: "ec2", ResourceType: "aws_instance", Error: "throttled"})
	}
	r.RecordFailure(ImportFailure{Stage: StageRefresh, Service: "ec2", ResourceType: "aws_vpc", Error: "gone"})
	groups := r.groupedFailures()
	if len(groups) != 2 {
		t.Fatalf("groupedFailures = %d groups, want 2", len(groups))
	}
	found := false
	for _, g := range groups {
		if g.count == 3 {
			found = true
		}
	}
	if !found {
		t.Fatal("expected one group with count 3")
	}
}

func TestImportReportWriteJSON(t *testing.T) {
	r := NewImportReport()
	r.RecordImported("aws_instance", 4)
	r.RecordFailure(ImportFailure{
		Stage: StageRefresh, Service: "ec2", ResourceType: "aws_instance",
		ResourceID: "i-123", ResourceName: "web", Error: "not found",
	})
	path := filepath.Join(t.TempDir(), "sub", "terraformer-failures.json")
	if err := r.WriteFailuresJSON(path, "aws"); err != nil {
		t.Fatalf("WriteFailuresJSON: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	var doc reportFile
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if doc.Provider != "aws" || doc.ImportedTotal != 4 || doc.FailureCount != 1 {
		t.Fatalf("unexpected doc: %+v", doc)
	}
	if len(doc.Failures) != 1 || doc.Failures[0].ResourceID != "i-123" || doc.Failures[0].Stage != StageRefresh {
		t.Fatalf("unexpected failures: %+v", doc.Failures)
	}
}

func TestFailureForResource(t *testing.T) {
	res := &Resource{
		ResourceName:  "web",
		InstanceInfo:  &terraform.InstanceInfo{Id: "i-9", Type: "aws_instance"},
		InstanceState: &terraform.InstanceState{ID: "i-9"},
	}
	f := failureForResource(res, "ec2", StageRefresh, errors.New("nope"))
	if f.ResourceType != "aws_instance" || f.ResourceID != "i-9" || f.ResourceName != "web" ||
		f.Service != "ec2" || f.Stage != StageRefresh || f.Error != "nope" {
		t.Fatalf("unexpected failure: %+v", f)
	}
	f = failureForResource(nil, "", StageEnumerate, errors.New("svc down"))
	if f.Error != "svc down" || f.ResourceType != "" {
		t.Fatalf("unexpected nil-resource failure: %+v", f)
	}
}
