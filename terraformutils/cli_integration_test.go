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

// Live CLI integration tests for the CLI-backed provider client.
//
// These tests exercise the real `terraform`/`tofu` binaries against the
// public registries WITHOUT any cloud credentials:
//   - `init` (provider auto-install) + `providers schema -json` + our
//     JSON-to-configschema conversion, fuzzed over every resource type of
//     several real providers (this is the highest-risk new code);
//   - the `show -json` -> InstanceState pipeline using a hand-written state
//     file (the full back half of bulk import, no `apply` needed).
//
// They only run with TERRAFORMER_INTEGRATION_TEST=1 so plain `go test ./...`
// stays offline.

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils/providerwrapper"

	"github.com/hashicorp/terraform/terraform"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

func integrationEnabled(t *testing.T) bool {
	t.Helper()
	if os.Getenv("TERRAFORMER_INTEGRATION_TEST") != "1" {
		t.Skip("live CLI test; set TERRAFORMER_INTEGRATION_TEST=1 to run")
		return false
	}
	return true
}

func cliBinaries(t *testing.T) map[string]string {
	t.Helper()
	bins := map[string]string{}
	for _, name := range []string{"terraform", "tofu"} {
		if p, err := exec.LookPath(name); err == nil {
			bins[name] = p
		}
	}
	if len(bins) == 0 {
		t.Skip("neither terraform nor tofu found on PATH")
	}
	return bins
}

// TestLiveProviderSchema runs init + schema fetch + conversion for real
// providers and calls ImpliedType() on every resource type, which shakes out
// type-expression parser edge cases across real-world schemas.
func TestLiveProviderSchema(t *testing.T) {
	if !integrationEnabled(t) {
		return
	}
	bins := cliBinaries(t)

	// cli -> providers to exercise. tofu gets one provider to prove the
	// OpenTofu registry path; terraform covers the breadth.
	matrix := map[string]map[string]int{}
	if _, ok := bins["terraform"]; ok {
		matrix["terraform"] = map[string]int{
			"aws": 100, "google": 100, "azurerm": 100, "kubernetes": 10,
		}
	}
	if _, ok := bins["tofu"]; ok {
		matrix["tofu"] = map[string]int{"aws": 100}
	}

	for cliName, providers := range matrix {
		bin := bins[cliName]
		for providerName, minTypes := range providers {
			providerName := providerName
			t.Run(cliName+"/"+providerName, func(t *testing.T) {
				t.Setenv("TERRAFORMER_TF_BIN", bin)
				w, err := providerwrapper.NewProviderWrapper(providerName, cty.EmptyObjectVal, false)
				if err != nil {
					t.Fatalf("NewProviderWrapper(%s): %v", providerName, err)
				}
				defer w.Kill()
				t.Logf("CLI binary in use: %s", w.CLIBin())

				schema := w.GetSchema()
				if schema == nil {
					t.Fatal("GetSchema() returned nil")
				}
				if n := len(schema.ResourceTypes); n < minTypes {
					t.Errorf("only %d resource types for %s, want at least %d", n, providerName, minTypes)
				} else {
					t.Logf("%s: %d resource types", providerName, n)
				}

				// Fuzz ImpliedType over the whole real schema.
				failed := 0
				for typeName, rs := range schema.ResourceTypes {
					func() {
						defer func() {
							if r := recover(); r != nil {
								t.Errorf("ImpliedType() panicked for %s: %v", typeName, r)
								failed++
							}
						}()
						if rs.Block == nil {
							t.Errorf("nil block for %s", typeName)
							failed++
							return
						}
						_ = rs.Block.ImpliedType()
					}()
				}
				if failed > 0 {
					t.Errorf("%d resource types failed ImpliedType()", failed)
				}

				// Spot-check the read-only attribute walker on real schemas.
				var sample []string
				for typeName := range schema.ResourceTypes {
					sample = append(sample, typeName)
					if len(sample) == 5 {
						break
					}
				}
				if _, err := w.GetReadOnlyAttributes(sample); err != nil {
					t.Errorf("GetReadOnlyAttributes: %v", err)
				}
			})
		}
	}
}

// dummyValue builds a non-null cty.Value of exactly the given type, used to
// synthesize realistic state values from a real provider schema.
func dummyValue(t cty.Type) cty.Value {
	switch {
	case t == cty.String:
		return cty.StringVal("test")
	case t == cty.Number:
		return cty.NumberIntVal(42)
	case t == cty.Bool:
		return cty.True
	case t == cty.DynamicPseudoType:
		return cty.StringVal("test")
	case t.IsListType():
		return cty.ListVal([]cty.Value{dummyValue(t.ElementType())})
	case t.IsSetType():
		return cty.SetVal([]cty.Value{dummyValue(t.ElementType())})
	case t.IsMapType():
		return cty.MapVal(map[string]cty.Value{"key": dummyValue(t.ElementType())})
	case t.IsObjectType():
		vals := map[string]cty.Value{}
		for name, at := range t.AttributeTypes() {
			vals[name] = dummyValue(at)
		}
		return cty.ObjectVal(vals)
	case t.IsTupleType():
		vals := make([]cty.Value, len(t.TupleElementTypes()))
		for i, et := range t.TupleElementTypes() {
			vals[i] = dummyValue(et)
		}
		return cty.TupleVal(vals)
	default:
		return cty.NullVal(t)
	}
}

// TestShowStatePipeline validates the whole back half of bulk import without
// cloud credentials: hand-written tfstate -> real `<cli> show -json` ->
// our parser -> cty -> terraform.InstanceState.
func TestShowStatePipeline(t *testing.T) {
	if !integrationEnabled(t) {
		return
	}
	bins := cliBinaries(t)

	for cliName, bin := range bins {
		t.Run(cliName, func(t *testing.T) {
			t.Setenv("TERRAFORMER_TF_BIN", bin)
			w, err := providerwrapper.NewProviderWrapper("aws", cty.EmptyObjectVal, false)
			if err != nil {
				t.Fatalf("NewProviderWrapper(aws): %v", err)
			}
			defer w.Kill()

			schema := w.GetSchema()
			rs, ok := schema.ResourceTypes["aws_vpc"]
			if !ok || rs.Block == nil {
				t.Fatal("aws_vpc not in fetched schema")
			}
			implied := rs.Block.ImpliedType()
			vals, err := ctyjson.Marshal(dummyValue(implied), implied)
			if err != nil {
				t.Fatalf("marshaling dummy state: %v", err)
			}

			host := "registry.terraform.io"
			if strings.Contains(w.CLIBin(), "tofu") {
				host = "registry.opentofu.org"
			}
			stateDoc := map[string]interface{}{
				"version":           4,
				"terraform_version": "1.0.0",
				"serial":            1,
				"lineage":           "integration-test",
				"outputs":           map[string]interface{}{},
				"resources": []interface{}{
					map[string]interface{}{
						"mode":     "managed",
						"type":     "aws_vpc",
						"name":     "tfer__test",
						"provider": `provider["` + host + `/hashicorp/aws"]`,
						"instances": []interface{}{
							map[string]interface{}{
								"schema_version":       rs.Version,
								"attributes":           json.RawMessage(vals),
								"sensitive_attributes": []interface{}{},
							},
						},
					},
				},
			}
			stateBytes, _ := json.Marshal(stateDoc)
			// Write into the wrapper's workdir so `show -json` runs against
			// the real CLI output format.
			if err := os.WriteFile(filepath.Join(w.WorkDir(), "terraform.tfstate"), stateBytes, 0644); err != nil {
				t.Fatalf("writing state: %v", err)
			}

			r := &Resource{
				ResourceName:  "tfer__test",
				InstanceInfo:  &terraform.InstanceInfo{Type: "aws_vpc", Id: "aws_vpc.tfer__test"},
				InstanceState: &terraform.InstanceState{ID: "vpc-0123456789abcdef0"},
			}
			addrToResource := map[string]*Resource{"aws_vpc.tfer__test": r}
			if err := collectShowState(w, addrToResource, map[string]bool{}); err != nil {
				t.Fatalf("collectShowState: %v", err)
			}
			if r.InstanceState == nil {
				t.Fatal("InstanceState not populated")
			}
			// The shim takes the ID from the state's id attribute, exactly
			// like the old gRPC Refresh path did.
			if r.InstanceState.ID != "test" {
				t.Errorf("ID = %q, want id attribute from state", r.InstanceState.ID)
			}
			if got := r.InstanceState.Attributes["cidr_block"]; got != "test" {
				t.Errorf("cidr_block = %q, want %q", got, "test")
			}
			t.Logf("state pipeline OK: %d flatmap attributes", len(r.InstanceState.Attributes))
		})
	}
}
