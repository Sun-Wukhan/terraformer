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

	"github.com/hashicorp/terraform/providers"
	"github.com/hashicorp/terraform/terraform"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

// Timeouts for the bulk import CLI invocations.
const (
	bulkPlanTimeout     = 10 * time.Minute
	bulkApplyTimeout    = 10 * time.Minute
	bulkShowTimeout     = 2 * time.Minute
	singleImportTimeout = 5 * time.Minute
)

// importPlanFile is the saved plan the bulk import applies, and only after
// verifying it cannot change any remote object.
const importPlanFile = "import.tfplan"

// importGeneratedFile receives the resource configuration Terraform generates
// for the import blocks (`plan -generate-config-out`). Empty resource blocks
// would fail validation on every required argument, so the config has to come
// from Terraform itself.
const importGeneratedFile = "import_generated.tf"

// This file implements the bulk state-refresh that replaced the old
// per-resource go-plugin gRPC Refresh. It lives in the terraformutils package
// (rather than as a ProviderWrapper method) because providerwrapper cannot
// import terraformutils without creating an import cycle.
//
// Flow:
//  1. Generate import.tf in the wrapper's isolated working directory: one
//     `import` block per resource. Terraform generates the matching resource
//     blocks itself (`-generate-config-out`), so the config always validates.
//  2. `<cli> plan -out` plans the imports (with retries). The saved plan is
//     applied only if `show -json` proves every change is a pure import with
//     a no-op action: any other action would rewrite the real object.
//  3. `<cli> show -json` reads the resulting state; each state value is
//     converted back into the 0.12 terraform.InstanceState model via the
//     provider schema, exactly as the old code did.
//  4. If the whole-set plan fails, the set is bisected and each half is
//     retried, recursively, so one resource whose generated block cannot
//     plan (e.g. a required sensitive argument Terraform cannot generate)
//     degrades alone instead of failing the whole service. Singleton
//     failures are the poisoned resources; they are remembered with their
//     plan error for the report.
//  5. Resources still missing from state get one single
//     `<cli> import <addr> <id>` attempt each, followed by a final
//     `show -json` pass. The config for that fallback is rebuilt from the
//     generated blocks of every leaf attempt, so a poisoned resource's
//     neighbours keep the blocks that let the single import work.
//  6. Resources that never made it into state are recorded in the
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
// per target. Resource blocks are deliberately absent so Terraform generates
// them during plan. IDs are quoted with strconv.Quote semantics so quotes and
// other special characters survive.
func renderImportConfig(targets []importTarget) string {
	var sb strings.Builder
	for _, t := range targets {
		fmt.Fprintf(&sb, "import {\n  to = %s\n  id = %s\n}\n\n", t.address, strconv.Quote(t.resource.InstanceState.ID))
	}
	return sb.String()
}

// bulkImportResources refreshes every resource's InstanceState via the CLI and
// returns the subset that was successfully refreshed. Every resource that is
// skipped or dropped along the way is recorded in report (nil-safe), with
// serviceOf attributing resources to their service for the report.
func bulkImportResources(p *providerwrapper.ProviderWrapper, resources []*Resource, report *ImportReport, serviceOf map[*Resource]string) ([]*Resource, error) {
	resources = dropUnknownTypes(resources, p.GetSchema(), report, serviceOf)
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

	refreshed := map[string]bool{}
	poisoned := map[string]error{}
	leafParts := map[string][]byte{}
	var fatalErr error

	// attempt runs one bulk plan+apply for a subset. Each attempt starts
	// from an empty workdir state: the wrapper's state file is internal to
	// the refresh (the real output state is rebuilt from the returned
	// resources), and a shared accumulating state would make later subset
	// plans try to delete earlier subsets' resources — tripping the
	// import-only guard and defeating the isolation. Refreshed resources
	// are collected into memory after every successful attempt instead.
	// After a leaf attempt (a success, or a single resource) it captures
	// the generated config so the single-import fallback can be rebuilt
	// from every leaf's blocks.
	attempt := func(subset []importTarget) error {
		resetBulkWorkState(p)
		if err := os.WriteFile(filepath.Join(p.WorkDir(), "import.tf"), []byte(renderImportConfig(subset)), 0644); err != nil {
			return fmt.Errorf("writing import.tf: %w", err)
		}
		err := bulkApplyImportPlan(p)
		if err == nil {
			// A show failure after a successful apply is global, not a
			// per-resource problem: don't bisect on it, fail the refresh.
			if cerr := collectShowState(p, addrToResource, refreshed, report, serviceOf); cerr != nil && fatalErr == nil {
				fatalErr = cerr
			}
		}
		if err == nil || len(subset) == 1 {
			if gen, rerr := os.ReadFile(filepath.Join(p.WorkDir(), importGeneratedFile)); rerr == nil {
				for _, t := range subset {
					leafParts[t.address] = gen
				}
			}
		}
		return err
	}

	bisectBulkImport(targets, attempt, poisoned)
	if fatalErr != nil {
		return nil, fatalErr
	}
	for addr, err := range poisoned {
		log.Printf("WARN: bulk import isolated %s as un-plannable, it gets a single-import attempt below: %v", addr, err)
	}
	// Fallback: one single `import` per resource still missing from state.
	needFallback := false
	for addr := range addrToResource {
		if !refreshed[addr] {
			needFallback = true
			break
		}
	}
	singleErrs := map[string]error{}
	if needFallback {
		// Rebuild the workdir config for the single imports: import blocks
		// for everything not yet refreshed, and the generated resource
		// blocks captured from the leaf attempts (with no bisection that is
		// simply the one full attempt's generated file, as before). The
		// fallback starts from an empty state; its imports are collected
		// by the collectShowState call after the loop.
		resetBulkWorkState(p)
		var missing []importTarget
		for _, t := range targets {
			if !refreshed[t.address] {
				missing = append(missing, t)
			}
		}
		if err := os.WriteFile(filepath.Join(p.WorkDir(), "import.tf"), []byte(renderImportConfig(missing)), 0644); err != nil {
			return nil, fmt.Errorf("writing import.tf: %w", err)
		}
		var combined strings.Builder
		seenPart := map[*byte]bool{}
		for _, t := range targets {
			part, ok := leafParts[t.address]
			if !ok || len(part) == 0 {
				continue
			}
			// Several targets of one leaf share the same backing array;
			// write each distinct part once.
			key := &part[0]
			if seenPart[key] {
				continue
			}
			seenPart[key] = true
			combined.Write(part)
			combined.WriteString("\n")
		}
		if combined.Len() > 0 {
			if err := os.WriteFile(filepath.Join(p.WorkDir(), importGeneratedFile), []byte(combined.String()), 0644); err != nil {
				return nil, fmt.Errorf("writing %s: %w", importGeneratedFile, err)
			}
		}
		for addr, r := range addrToResource {
			if refreshed[addr] {
				continue
			}
			log.Printf("WARN: resource %s not in state after bulk apply, trying single import", addr)
			if _, err := p.RunCLI(singleImportTimeout, "import", "-input=false", "-no-color", addr, r.InstanceState.ID); err != nil {
				// Not recorded here: the final loop below is the single
				// recording point for every resource that never refreshed,
				// so one resource can never be counted as two failures.
				log.Printf("ERROR: single import of %s failed: %v", addr, err)
				singleErrs[addr] = err
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
			// A resource that refreshed may still carry a failure recorded
			// mid-pipeline (e.g. a state-conversion error during an earlier
			// bulk pass, before a fallback rescued it). The final outcome
			// wins: it imported, so it is not a failure.
			if r.InstanceInfo != nil {
				report.ClearResourceFailures(serviceOf[r], r.InstanceInfo.Type, r.ResourceName)
			}
		} else {
			log.Printf("ERROR: Unable to refresh resource %s", r.ResourceName)
			err := fmt.Errorf("unable to refresh resource: not present in state after import")
			if perr, ok := poisoned[addr]; ok {
				err = fmt.Errorf("bulk plan failed for this resource alone: %w", perr)
			} else if serr, ok := singleErrs[addr]; ok {
				err = fmt.Errorf("single import failed: %w", serr)
			}
			report.RecordFailure(failureForResource(r, serviceOf[r], StageRefresh, err))
		}
	}
	return out, nil
}

// resetBulkWorkState removes the wrapper workdir's terraform state and
// saved plan so the next bulk attempt or fallback pass starts empty. The
// workdir state is internal to the refresh — output state files are
// rebuilt from the returned resources — so discarding it between attempts
// loses nothing and keeps subset attempts independent of each other.
func resetBulkWorkState(p *providerwrapper.ProviderWrapper) {
	for _, name := range []string{"terraform.tfstate", "terraform.tfstate.backup", importPlanFile} {
		if err := os.Remove(filepath.Join(p.WorkDir(), name)); err != nil && !os.IsNotExist(err) {
			log.Printf("WARN: removing stale %s: %v", name, err)
		}
	}
}

// bisectBulkImport applies the bulk import to targets, splitting the set
// and retrying each half whenever an attempt fails, until every resource
// has either been applied as part of a successful subset or isolated as a
// singleton failure. Singleton failures are recorded in poisoned with the
// attempt's error. attempt is called exactly once for a fully healthy set.
func bisectBulkImport(targets []importTarget, attempt func(subset []importTarget) error, poisoned map[string]error) {
	if len(targets) == 0 {
		return
	}
	if err := attempt(targets); err == nil {
		return
	} else if len(targets) == 1 {
		poisoned[targets[0].address] = err
		return
	}
	mid := len(targets) / 2
	bisectBulkImport(targets[:mid], attempt, poisoned)
	bisectBulkImport(targets[mid:], attempt, poisoned)
}

// dropUnknownTypes removes resources whose type the installed provider does
// not define (e.g. azurerm_app_service, removed in azurerm v4). They all share
// one import.tf, so a single unknown type would otherwise fail every import.
func dropUnknownTypes(resources []*Resource, schema *providers.GetSchemaResponse, report *ImportReport, serviceOf map[*Resource]string) []*Resource {
	if schema == nil {
		return resources
	}
	kept := resources[:0:0]
	for _, r := range resources {
		if r != nil && r.InstanceInfo != nil {
			if _, ok := schema.ResourceTypes[r.InstanceInfo.Type]; !ok {
				report.RecordFailure(failureForResource(r, serviceOf[r], StageRefresh,
					fmt.Errorf("resource type %q is not defined by the installed provider version", r.InstanceInfo.Type)))
				continue
			}
		}
		kept = append(kept, r)
	}
	return kept
}

// bulkApplyImportPlan plans the generated import config and applies the saved
// plan only when it is provably import-only. Any error means nothing was
// applied.
func bulkApplyImportPlan(p *providerwrapper.ProviderWrapper) error {
	var planErr error
	generated := filepath.Join(p.WorkDir(), importGeneratedFile)
	for i := 0; i <= p.GetRetryCount(); i++ {
		if i > 0 {
			log.Printf("WARN: bulk import plan failed, retry %d/%d after %dms", i, p.GetRetryCount(), p.GetRetrySleepMs())
			time.Sleep(time.Duration(p.GetRetrySleepMs()) * time.Millisecond)
		}
		// Terraform refuses to overwrite an existing generated-config file.
		if err := os.Remove(generated); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("removing stale %s: %w", importGeneratedFile, err)
		}
		_, planErr = p.RunCLI(bulkPlanTimeout, "plan", "-input=false", "-no-color",
			"-generate-config-out="+importGeneratedFile, "-out="+importPlanFile)
		if planErr == nil {
			break
		}
	}
	if planErr != nil {
		return fmt.Errorf("plan failed: %w", planErr)
	}
	planJSON, err := p.RunCLI(bulkShowTimeout, "show", "-json", "-no-color", importPlanFile)
	if err != nil {
		return fmt.Errorf("reading plan: %w", err)
	}
	if err := verifyImportOnlyPlan(planJSON); err != nil {
		return err
	}
	if _, err := p.RunCLI(bulkApplyTimeout, "apply", "-input=false", "-no-color", importPlanFile); err != nil {
		return fmt.Errorf("applying import-only plan: %w", err)
	}
	return nil
}

type planResourceChange struct {
	Address string `json:"address"`
	Change  struct {
		Actions   []string        `json:"actions"`
		Importing json.RawMessage `json:"importing"`
	} `json:"change"`
}

// verifyImportOnlyPlan returns an error unless every planned change in a
// `show -json` plan document is an import whose only action is no-op, i.e.
// applying the plan can write local state but never touch a remote object.
func verifyImportOnlyPlan(planJSON []byte) error {
	var doc struct {
		ResourceChanges []planResourceChange `json:"resource_changes"`
	}
	if err := json.Unmarshal(planJSON, &doc); err != nil {
		return fmt.Errorf("parsing plan: %w", err)
	}
	var unsafe []string
	for _, rc := range doc.ResourceChanges {
		noop := len(rc.Change.Actions) == 1 && rc.Change.Actions[0] == "no-op"
		importing := len(rc.Change.Importing) > 0 && string(rc.Change.Importing) != "null"
		if !noop || !importing {
			unsafe = append(unsafe, fmt.Sprintf("%s %v", rc.Address, rc.Change.Actions))
		}
	}
	if len(unsafe) > 0 {
		return fmt.Errorf("refusing to apply: plan would modify %d remote object(s), e.g. %s", len(unsafe), unsafe[0])
	}
	return nil
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
	// Keep the typed attributes for the generated state file.
	r.StateJSON = append(json.RawMessage(nil), values...)
	return nil
}
