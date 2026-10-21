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
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils/providerwrapper"

	"github.com/hashicorp/terraform/terraform"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

// Timeouts for the bulk import CLI invocations.
const (
	bulkApplyTimeout    = 10 * time.Minute
	bulkShowTimeout     = 2 * time.Minute
	singleImportTimeout = 5 * time.Minute
)

// This file implements the bulk state-refresh that replaced the old
// per-resource go-plugin gRPC Refresh. It lives in the terraformutils package
// (rather than as a ProviderWrapper method) because providerwrapper cannot
// import terraformutils without creating an import cycle.
//
// Flow:
//  1. Generate import.tf in the wrapper's isolated working directory: one
//     `import` block plus an empty resource block per resource.
//  2. `<cli> apply -auto-approve` imports everything in one run (with
//     retries; partial failures are tolerated).
//  3. `<cli> show -json` reads the resulting state; each state value is
//     converted back into the 0.12 terraform.InstanceState model via the
//     provider schema, exactly as the old code did.
//  4. Resources still missing from state get one single
//     `<cli> import <addr> <id>` attempt each, followed by a final
//     `show -json` pass.
//  5. Resources that never made it into state are recorded in the
//     ImportReport and dropped, the same graceful degradation the old
//     per-resource refresh had — but no longer silently.

// importTarget pairs a resource with its traversal-safe import label and
// full address (e.g. aws_instance.tfer__foo) used in the generated import.tf.
type importTarget struct {
	resource *Resource
	label    string
	address  string
}

// assignImportTargets gives each importable resource (non-empty ID) a
// traversal-safe label that is unique per resource type, returning the targets
// and an address -> resource map. Resources without an ID are skipped with a
// warning.
func assignImportTargets(resources []*Resource) ([]importTarget, map[string]*Resource) {
	var targets []importTarget
	addrToResource := map[string]*Resource{}
	usedLabels := map[string]map[string]bool{} // resource type -> label -> used

	for _, r := range resources {
		if r == nil || r.InstanceInfo == nil {
			continue
		}
		if r.InstanceState == nil || r.InstanceState.ID == "" {
			log.Printf("WARN: skipping resource %s.%s with empty ID", r.InstanceInfo.Type, r.ResourceName)
			continue
		}
		// TfSanitize output ("tfer--...") is not a valid HCL traversal for
		// `import { to = ... }`, so normalize hyphens to underscores for the
		// import labels. The label only needs to be unique and mappable back;
		// downstream code keeps using ResourceName.
		base := strings.ReplaceAll(TfSanitize(r.ResourceName), "-", "_")
		if usedLabels[r.InstanceInfo.Type] == nil {
			usedLabels[r.InstanceInfo.Type] = map[string]bool{}
		}
		label := base
		for i := 2; usedLabels[r.InstanceInfo.Type][label]; i++ {
			label = fmt.Sprintf("%s_%d", base, i)
		}
		usedLabels[r.InstanceInfo.Type][label] = true

		address := r.InstanceInfo.Type + "." + label
		addrToResource[address] = r
		targets = append(targets, importTarget{resource: r, label: label, address: address})
	}
	return targets, addrToResource
}

// renderImportConfig renders the import.tf file content: one `import` block
// plus an empty resource block per target. IDs are quoted with
// strconv.Quote semantics so quotes and other special characters survive.
func renderImportConfig(targets []importTarget) string {
	var sb strings.Builder
	for _, t := range targets {
		fmt.Fprintf(&sb, "import {\n  to = %s\n  id = %s\n}\n\n", t.address, strconv.Quote(t.resource.InstanceState.ID))
		fmt.Fprintf(&sb, "resource %s %s {\n}\n\n", strconv.Quote(t.resource.InstanceInfo.Type), strconv.Quote(t.label))
	}
	return sb.String()
}

// bulkImportResources refreshes every resource's InstanceState via the CLI and
// returns the subset that was successfully refreshed. Every resource that is
// skipped or dropped along the way is recorded in report (nil-safe), with
// serviceOf attributing resources to their service for the report.
func bulkImportResources(p *providerwrapper.ProviderWrapper, resources []*Resource, report *ImportReport, serviceOf map[*Resource]string) ([]*Resource, error) {
	targets, addrToResource := assignImportTargets(resources)
	if len(targets) == 0 {
		return nil, nil
	}

	// Resources that never became targets (nil or empty ID) are dropped
	// silently by assignImportTargets; account for them here.
	targetSet := make(map[*Resource]bool, len(targets))
	for _, t := range targets {
		targetSet[t.resource] = true
	}
	for _, r := range resources {
		if r == nil || r.InstanceInfo == nil || targetSet[r] {
			continue
		}
		report.RecordFailure(failureForResource(r, serviceOf[r], StageRefresh,
			fmt.Errorf("skipped: resource has an empty ID")))
	}

	if err := os.WriteFile(filepath.Join(p.WorkDir(), "import.tf"), []byte(renderImportConfig(targets)), 0644); err != nil {
		return nil, fmt.Errorf("writing import.tf: %w", err)
	}

	var applyErr error
	for i := 0; i <= p.GetRetryCount(); i++ {
		if i > 0 {
			log.Printf("WARN: bulk import apply failed, retry %d/%d after %dms", i, p.GetRetryCount(), p.GetRetrySleepMs())
			time.Sleep(time.Duration(p.GetRetrySleepMs()) * time.Millisecond)
		}
		_, applyErr = p.RunCLI(bulkApplyTimeout, "apply", "-auto-approve", "-input=false", "-no-color")
		if applyErr == nil {
			break
		}
		log.Printf("WARN: bulk import apply attempt %d failed: %v", i+1, applyErr)
	}
	if applyErr != nil {
		// Partial failure: whatever was imported before the error stays in
		// state, so continue and collect it instead of failing hard.
		log.Printf("ERROR: bulk import apply failed after retries, continuing with partial state: %v", applyErr)
		report.RecordFailure(ImportFailure{Stage: StageRefresh,
			Error: fmt.Sprintf("bulk import apply failed after retries, continued with partial state: %v", applyErr)})
	}

	refreshed := map[string]bool{}
	if err := collectShowState(p, addrToResource, refreshed, report, serviceOf); err != nil {
		return nil, err
	}

	// Fallback: one single `import` per resource still missing from state.
	needFallback := false
	for addr := range addrToResource {
		if !refreshed[addr] {
			needFallback = true
			break
		}
	}
	if needFallback {
		for addr, r := range addrToResource {
			if refreshed[addr] {
				continue
			}
			log.Printf("WARN: resource %s not in state after bulk apply, trying single import", addr)
			if _, err := p.RunCLI(singleImportTimeout, "import", "-input=false", "-no-color", addr, r.InstanceState.ID); err != nil {
				log.Printf("ERROR: single import of %s failed: %v", addr, err)
				report.RecordFailure(failureForResource(r, serviceOf[r], StageRefresh,
					fmt.Errorf("single import failed: %w", err)))
				continue
			}
		}
		if err := collectShowState(p, addrToResource, refreshed, report, serviceOf); err != nil {
			return nil, err
		}
	}

	var out []*Resource
	for addr, r := range addrToResource {
		if refreshed[addr] {
			out = append(out, r)
		} else {
			log.Printf("ERROR: Unable to refresh resource %s", r.ResourceName)
			report.RecordFailure(failureForResource(r, serviceOf[r], StageRefresh,
				fmt.Errorf("unable to refresh resource: not present in state after import")))
		}
	}
	return out, nil
}

// --- show -json handling ---

type showStateResource struct {
	Address string          `json:"address"`
	Type    string          `json:"type"`
	Values  json.RawMessage `json:"values"`
}

type showModule struct {
	Resources    []showStateResource `json:"resources"`
	ChildModules []showModule        `json:"child_modules"`
}

type showJSONDoc struct {
	Values *struct {
		RootModule showModule `json:"root_module"`
	} `json:"values"`
}

// collectShowState runs `<cli> show -json`, matches state resources back to
// our import targets by address, and converts each value into the 0.12
// InstanceState model. Resources already marked in refreshed are skipped.
// Per-resource conversion failures are recorded in report (nil-safe).
func collectShowState(p *providerwrapper.ProviderWrapper, addrToResource map[string]*Resource, refreshed map[string]bool, report *ImportReport, serviceOf map[*Resource]string) error {
	out, err := p.RunCLI(bulkShowTimeout, "show", "-json", "-no-color")
	if err != nil {
		return fmt.Errorf("show -json failed: %w", err)
	}
	var doc showJSONDoc
	if err := json.Unmarshal(out, &doc); err != nil {
		return fmt.Errorf("parsing show -json output: %w", err)
	}
	if doc.Values == nil {
		return nil
	}

	var all []showStateResource
	var walk func(m showModule)
	walk = func(m showModule) {
		all = append(all, m.Resources...)
		for _, child := range m.ChildModules {
			walk(child)
		}
	}
	walk(doc.Values.RootModule)

	for _, sr := range all {
		r, ok := addrToResource[sr.Address]
		if !ok || refreshed[sr.Address] {
			continue
		}
		if err := assignRefreshedState(p, r, sr.Values, sr.Type); err != nil {
			log.Printf("WARN: %v", err)
			report.RecordFailure(failureForResource(r, serviceOf[r], StageRefresh, err))
			continue
		}
		refreshed[sr.Address] = true
	}
	return nil
}

// assignRefreshedState converts one `show -json` resource value into a
// terraform.InstanceState using the provider schema, mirroring what the old
// gRPC Refresh did with NewInstanceStateShimmedFromValue.
func assignRefreshedState(p *providerwrapper.ProviderWrapper, r *Resource, values json.RawMessage, resType string) error {
	schema := p.GetSchema()
	if schema == nil {
		return fmt.Errorf("provider schema not initialized")
	}
	rs, ok := schema.ResourceTypes[resType]
	if !ok || rs.Block == nil {
		return fmt.Errorf("no schema for resource type %q", resType)
	}
	val, err := ctyjson.Unmarshal(values, rs.Block.ImpliedType())
	if err != nil {
		return fmt.Errorf("unmarshaling state for %s: %w", resType, err)
	}
	state := terraform.NewInstanceStateShimmedFromValue(val, int(rs.Version))
	if state.ID == "" && r.InstanceState != nil {
		// Preserve the original import ID when the provider reports no id.
		state.ID = r.InstanceState.ID
	}
	r.InstanceState = state
	return nil
}
