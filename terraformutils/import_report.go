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
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// Stages at which an import failure can be recorded.
const (
	StageEnumerate = "enumerate" // a service failed to list its resources
	StageRefresh   = "refresh"   // a resource could not be refreshed into state
	StageConvert   = "convert"   // a resource's state failed HCL conversion
)

// ImportFailure describes one resource (or one whole service) that did not
// make it into the generated output, with the reason why.
type ImportFailure struct {
	Stage        string `json:"stage"`
	Service      string `json:"service,omitempty"`
	ResourceType string `json:"resource_type,omitempty"`
	ResourceID   string `json:"resource_id,omitempty"`
	ResourceName string `json:"resource_name,omitempty"`
	Error        string `json:"error"`
}

// ImportReport collects the outcome of an import run: how many resources of
// each type were imported, and every failure that was previously only
// visible as a log line (or not visible at all). It is the data behind the
// end-of-run summary, the terraformer-failures.json file, and --strict.
//
// All methods are safe to call on a nil *ImportReport (they no-op), so
// callers may pass nil to disable reporting.
type ImportReport struct {
	mu       sync.Mutex
	imported map[string]int
	failures []ImportFailure
}

func NewImportReport() *ImportReport {
	return &ImportReport{imported: map[string]int{}}
}

// failureForResource builds an ImportFailure describing resource r at the
// given stage. Nil-safe: a nil resource yields a failure with only the
// stage and error set.
func failureForResource(r *Resource, service, stage string, err error) ImportFailure {
	f := ImportFailure{Stage: stage, Service: service, Error: err.Error()}
	if r != nil {
		f.ResourceName = r.ResourceName
		if r.InstanceInfo != nil {
			f.ResourceType = r.InstanceInfo.Type
			f.ResourceID = r.InstanceInfo.Id
		}
	}
	return f
}

// RecordImported adds count imported resources of the given type.
func (r *ImportReport) RecordImported(resourceType string, count int) {
	if r == nil || count <= 0 {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.imported == nil {
		r.imported = map[string]int{}
	}
	r.imported[resourceType] += count
}

// RecordFailure notes one failed resource or service.
func (r *ImportReport) RecordFailure(f ImportFailure) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.failures = append(r.failures, f)
}

// Failures returns a copy of the recorded failures, in recording order.
func (r *ImportReport) Failures() []ImportFailure {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]ImportFailure, len(r.failures))
	copy(out, r.failures)
	return out
}

// FailureCount is the number of recorded failures.
func (r *ImportReport) FailureCount() int {
	if r == nil {
		return 0
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.failures)
}

// HasFailures reports whether anything failed during the import.
func (r *ImportReport) HasFailures() bool {
	return r.FailureCount() > 0
}

// ImportedCount is the total number of imported resources.
func (r *ImportReport) ImportedCount() int {
	if r == nil {
		return 0
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	total := 0
	for _, n := range r.imported {
		total += n
	}
	return total
}

// ImportedByType returns a copy of the per-type imported counts.
func (r *ImportReport) ImportedByType() map[string]int {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make(map[string]int, len(r.imported))
	for k, v := range r.imported {
		out[k] = v
	}
	return out
}

type failureGroup struct {
	key   string
	count int
}

// groupedFailures collapses identical failures (same stage, service, type
// and error) into counted groups, sorted for stable output.
func (r *ImportReport) groupedFailures() []failureGroup {
	groups := map[string]int{}
	for _, f := range r.Failures() {
		key := fmt.Sprintf("[%s] service=%s type=%s: %s", f.Stage, f.Service, f.ResourceType, f.Error)
		groups[key]++
	}
	out := make([]failureGroup, 0, len(groups))
	for k, n := range groups {
		out = append(out, failureGroup{key: k, count: n})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].key < out[j].key })
	return out
}

// PrintSummary logs the end-of-run summary: totals plus one line per
// distinct failure (with a repeat count), so a partial import can no
// longer pass unnoticed in the output.
func (r *ImportReport) PrintSummary(providerName string) {
	if r == nil {
		return
	}
	log.Printf("%s import summary: %d resource(s) imported, %d failure(s)",
		providerName, r.ImportedCount(), r.FailureCount())
	for _, g := range r.groupedFailures() {
		if g.count > 1 {
			log.Printf("  FAILED x%d %s", g.count, g.key)
		} else {
			log.Printf("  FAILED %s", g.key)
		}
	}
}

// reportFile is the on-disk shape of terraformer-failures.json.
type reportFile struct {
	GeneratedAt    time.Time       `json:"generated_at"`
	Provider       string          `json:"provider"`
	ImportedTotal  int             `json:"imported_total"`
	ImportedByType map[string]int  `json:"imported_by_type"`
	FailureCount   int             `json:"failure_count"`
	Failures       []ImportFailure `json:"failures"`
}

// WriteFailuresJSON writes the machine-readable failure report to path
// (creating parent directories), so a partial import can be audited or
// gated on by other tooling.
func (r *ImportReport) WriteFailuresJSON(path, providerName string) error {
	if r == nil {
		return nil
	}
	doc := reportFile{
		GeneratedAt:    time.Now().UTC(),
		Provider:       providerName,
		ImportedTotal:  r.ImportedCount(),
		ImportedByType: r.ImportedByType(),
		FailureCount:   r.FailureCount(),
		Failures:       r.Failures(),
	}
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(path, data, 0o644)
}
