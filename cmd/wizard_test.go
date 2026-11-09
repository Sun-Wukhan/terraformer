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

package cmd

import (
	"bufio"
	"strings"
	"testing"
)

func TestWizardGoogleCommand(t *testing.T) {
	cfg := wizardConfig{
		Provider:   "google",
		Resources:  []string{"networks", "firewall"},
		Projects:   []string{"my-project"},
		Regions:    []string{"global"},
		PathOutput: "generated",
		Connect:    true,
		Strict:     true,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	want := "terraformer import google --resources=networks,firewall --projects=my-project --regions=global --strict"
	if got := cfg.ShellCommand(); got != want {
		t.Fatalf("command mismatch:\n got: %s\nwant: %s", got, want)
	}
}

func TestWizardVaultCommandOmitsToken(t *testing.T) {
	cfg := wizardConfig{
		Provider:     "vault",
		Resources:    []string{"generic_secret"},
		VaultAddress: "http://127.0.0.1:8200",
		PathOutput:   "generated",
		Connect:      true,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	cmd := cfg.ShellCommand()
	if !strings.Contains(cmd, "--address=http://127.0.0.1:8200") {
		t.Fatalf("missing address in %q", cmd)
	}
	if strings.Contains(strings.ToLower(cmd), "token") {
		t.Fatalf("command must never carry a token: %q", cmd)
	}
}

func TestWizardRejectsUnknownService(t *testing.T) {
	cfg := wizardConfig{Provider: "google", Resources: []string{"not_a_service"}, Projects: []string{"p"}}
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "not_a_service") {
		t.Fatalf("expected unknown-service error naming the bad value, got: %v", err)
	}
}

func TestWizardRejectsGoogleWithoutProject(t *testing.T) {
	cfg := wizardConfig{Provider: "google", Resources: []string{"networks"}}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected an error when --projects is missing for google")
	}
}

func TestWizardRejectsAzureWithoutResourceGroup(t *testing.T) {
	cfg := wizardConfig{Provider: "azure", Resources: []string{"resource_group"}}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected an error when --resource-group is missing for azure")
	}
}

func TestWizardShellQuote(t *testing.T) {
	if got := shellQuote("a b;c"); got != "'a b;c'" {
		t.Fatalf("expected quoting, got %s", got)
	}
	if got := shellQuote("plain-1.0/x:y"); got != "plain-1.0/x:y" {
		t.Fatalf("safe value should stay unquoted, got %s", got)
	}
}

// The interactive flow, driven end-to-end with scripted answers:
// provider by number is brittle across the sorted list, so pick by name.
func TestWizardInteractiveGrafana(t *testing.T) {
	answers := strings.Join([]string{
		"grafana",                          // provider
		"grafana_dashboard,grafana_folder", // resources by name
		"",                                 // output folder (default)
		"",                                 // strict (default yes)
		"",                                 // verbose (default no)
	}, "\n") + "\n"
	var out strings.Builder
	w := &wizardRunner{in: bufio.NewReader(strings.NewReader(answers)), out: &out, cfg: wizardConfig{Connect: true}}
	if err := w.runInteractive(); err != nil {
		t.Fatalf("interactive run: %v", err)
	}
	if w.cfg.Provider != "grafana" {
		t.Fatalf("provider = %q", w.cfg.Provider)
	}
	if got := w.cfg.ShellCommand(); !strings.Contains(got, "import grafana --resources=grafana_dashboard,grafana_folder") {
		t.Fatalf("unexpected command: %s", got)
	}
	if !w.cfg.Strict {
		t.Fatal("strict should default to on")
	}
}

func TestWizardEveryProviderHasServices(t *testing.T) {
	for _, name := range wizardProviderNames() {
		// kubernetes builds its service list from a live cluster's
		// discovery API, so it legitimately reports none in a test
		// environment with no cluster configured.
		if name == "kubernetes" {
			continue
		}
		if services := servicesFor(name); len(services) == 0 {
			t.Errorf("provider %s reports no services", name)
		}
	}
}

// Dynamic-list providers cannot be validated offline: their names are
// accepted as given, because there is no list to check them against.
func TestWizardDynamicProviderSkipsValidation(t *testing.T) {
	cfg := wizardConfig{Provider: "kubernetes", Resources: []string{"some_resource_kind"}}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected no validation error without a cluster, got: %v", err)
	}
}
