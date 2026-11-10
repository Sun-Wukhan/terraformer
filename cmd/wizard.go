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
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

// wizardConfig is everything the wizard collects. The same struct drives
// the interactive prompts, the generated command and the unit tests, so
// what the user sees is exactly what would run.
type wizardConfig struct {
	Provider      string
	Resources     []string
	Projects      []string // google only
	Regions       []string // google / aws
	Profile       string   // aws only
	ResourceGroup string   // azure only
	VaultAddress  string   // vault only
	PathOutput    string
	Connect       bool
	Strict        bool
	Verbose       bool
	Compact       bool
}

// The provider registry (wizardProviders) lives in root.go — see the
// comment there for why it must not move back into this file
// (build/multi-build per-provider release builds).

func wizardProviderNames() []string {
	names := make([]string, 0, len(wizardProviders()))
	for name := range wizardProviders() {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// servicesFor returns the provider's supported services, sorted.
// It returns nil for an unknown provider.
func servicesFor(provider string) []string {
	gen, ok := wizardProviders()[provider]
	if !ok {
		return nil
	}
	return providerServices(gen())
}

// Args is the terraformer argument vector (without the binary name).
func (c wizardConfig) Args() []string {
	args := []string{"import", c.Provider}
	if len(c.Resources) > 0 {
		args = append(args, "--resources="+strings.Join(c.Resources, ","))
	}
	switch c.Provider {
	case "google":
		if len(c.Projects) > 0 {
			args = append(args, "--projects="+strings.Join(c.Projects, ","))
		}
		if len(c.Regions) > 0 {
			args = append(args, "--regions="+strings.Join(c.Regions, ","))
		}
	case "aws":
		if len(c.Regions) > 0 {
			args = append(args, "--regions="+strings.Join(c.Regions, ","))
		}
		if c.Profile != "" {
			args = append(args, "--profile="+c.Profile)
		}
	case "azure":
		if c.ResourceGroup != "" {
			args = append(args, "--resource-group="+c.ResourceGroup)
		}
	case "vault":
		if c.VaultAddress != "" {
			args = append(args, "--address="+c.VaultAddress)
		}
	}
	if c.PathOutput != "" && c.PathOutput != DefaultPathOutput {
		args = append(args, "--path-output="+c.PathOutput)
	}
	if !c.Connect {
		args = append(args, "--connect=false")
	}
	if c.Compact {
		args = append(args, "--compact")
	}
	if c.Verbose {
		args = append(args, "--verbose")
	}
	if c.Strict {
		args = append(args, "--strict")
	}
	return args
}

// shellQuote renders one argument safe for a POSIX shell command line.
func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	if strings.IndexFunc(s, func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' ||
			strings.ContainsRune("-_./:=,@%+", r))
	}) == -1 {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// ShellCommand is the copy-pasteable command line for this config.
func (c wizardConfig) ShellCommand() string {
	parts := make([]string, 0, len(c.Args())+1)
	parts = append(parts, "terraformer")
	for _, a := range c.Args() {
		parts = append(parts, shellQuote(a))
	}
	return strings.Join(parts, " ")
}

type envHint struct {
	Name    string
	Purpose string
}

// RequiredEnv lists credentials/config the provider reads from the
// environment. The wizard never asks for secret values and never puts
// them on the generated command line.
func (c wizardConfig) RequiredEnv() []envHint {
	switch c.Provider {
	case "google":
		return []envHint{{"GOOGLE_APPLICATION_CREDENTIALS", "service-account key file, or run: gcloud auth application-default login"}}
	case "aws":
		return []envHint{{"AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY", "or a named profile / SSO login for --profile"}}
	case "azure":
		return []envHint{{"ARM_SUBSCRIPTION_ID", "subscription to import from"}, {"ARM_TENANT_ID / az login", "run: az login, if not already signed in"}}
	case "vault":
		return []envHint{{"VAULT_TOKEN", "token with read/list on the mounts you import (VAULT_ADDR is set by --address)"}}
	case "grafana":
		return []envHint{{"GRAFANA_URL", "e.g. http://localhost:3000"}, {"GRAFANA_AUTH", "admin user:password or an API key"}}
	case "github":
		return []envHint{{"GITHUB_TOKEN", "personal access token"}}
	case "kubernetes":
		return []envHint{{"KUBECONFIG", "path to your kubeconfig, if not ~/.kube/config"}}
	default:
		return []envHint{{strings.ToUpper(c.Provider) + " credentials", "set the provider's usual environment variables / CLI login before running"}}
	}
}

// Validate checks the config against the provider's real service list.
func (c wizardConfig) Validate() error {
	if _, ok := wizardProviders()[c.Provider]; !ok {
		return fmt.Errorf("unknown provider %q (run: terraformer wizard --help, or terraformer import --help)", c.Provider)
	}
	if len(c.Resources) == 0 {
		return fmt.Errorf("no resources selected: pick at least one service, or * for all of them")
	}
	supported := servicesFor(c.Provider)
	// Some providers (kubernetes) build their service list from the live
	// target — kubernetes discovery queries the cluster — so an empty
	// list here means "cannot enumerate right now", not "no services".
	// In that case names cannot be validated up front; the import itself
	// will reject unknown ones.
	if len(supported) > 0 {
		set := make(map[string]bool, len(supported))
		for _, s := range supported {
			set[s] = true
		}
		for _, r := range c.Resources {
			if r == "*" {
				continue
			}
			if !set[r] {
				return fmt.Errorf("%q is not a %s service. Valid services: %s", r, c.Provider, strings.Join(supported, ", "))
			}
		}
	}
	switch c.Provider {
	case "google":
		if len(c.Projects) == 0 {
			return fmt.Errorf("google imports need at least one --projects value")
		}
	case "azure":
		if c.ResourceGroup == "" {
			return fmt.Errorf("azure imports need --resource-group (use * for all resource groups in the subscription)")
		}
	}
	return nil
}

// NextSteps are the commands to run after a successful import. The plan
// check is the acceptance test: a clean import shows "No changes".
func (c wizardConfig) NextSteps() []string {
	out := c.PathOutput
	if out == "" {
		out = DefaultPathOutput
	}
	return []string{
		fmt.Sprintf("cd %s/%s/<service>", out, c.Provider),
		"terraform init    # or: tofu init",
		"terraform plan    # a clean import shows: No changes.",
	}
}

// --- interactive runner -------------------------------------------------

type wizardRunner struct {
	in  *bufio.Reader
	out io.Writer
	cfg wizardConfig
}

func (w *wizardRunner) printf(format string, a ...interface{}) {
	fmt.Fprintf(w.out, format, a...)
}

func (w *wizardRunner) ask(label, def string) string {
	if def != "" {
		w.printf("%s [%s]: ", label, def)
	} else {
		w.printf("%s: ", label)
	}
	line, _ := w.in.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return def
	}
	return line
}

func (w *wizardRunner) askYesNo(label string, def bool) bool {
	hint := "y/N"
	if def {
		hint = "Y/n"
	}
	ans := strings.ToLower(w.ask(label+" ("+hint+")", ""))
	if ans == "" {
		return def
	}
	return ans == "y" || ans == "yes"
}

func splitCSV(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// askResources shows the real service list and accepts names, numbers,
// a comma list mixing both, "list" to re-print, or * for everything.
func (w *wizardRunner) askResources() []string {
	services := servicesFor(w.cfg.Provider)
	if len(services) == 0 {
		// Dynamic-list providers (kubernetes): the service list comes from
		// the live target, so it cannot be shown or validated here.
		w.printf("\n%s builds its service list from the live target, so I can't show it here.\n", w.cfg.Provider)
		for {
			picked := splitCSV(w.ask("Resources (names, comma-separated, or * for all)", ""))
			if len(picked) > 0 {
				return picked
			}
			w.printf("  Pick at least one service, or * for all.\n")
		}
	}
	for {
		w.printf("\n%s services (%d). Type 'list' to see them all, * for all:\n", w.cfg.Provider, len(services))
		for i, s := range services {
			w.printf("  %3d. %s\n", i+1, s)
		}
		raw := w.ask("Resources (names or numbers, comma-separated)", "")
		if strings.EqualFold(raw, "list") {
			continue
		}
		if raw == "*" {
			return []string{"*"}
		}
		var picked []string
		bad := ""
		for _, tok := range splitCSV(raw) {
			if n, err := strconv.Atoi(tok); err == nil {
				if n >= 1 && n <= len(services) {
					picked = append(picked, services[n-1])
				} else {
					bad = tok
				}
				continue
			}
			found := false
			for _, s := range services {
				if s == tok {
					found = true
					break
				}
			}
			if found {
				picked = append(picked, tok)
			} else {
				bad = tok
			}
		}
		if bad != "" {
			w.printf("  %q is not a %s service — check the spelling against the list above.\n", bad, w.cfg.Provider)
			continue
		}
		if len(picked) == 0 {
			w.printf("  Pick at least one service, or * for all.\n")
			continue
		}
		return picked
	}
}

func (w *wizardRunner) runInteractive() error {
	w.printf("Terraformer import wizard\n")
	w.printf("I'll ask what to import, check it, and give you the exact command — no guessing flags.\n\n")

	// 1. Provider
	names := wizardProviderNames()
	for i, n := range names {
		w.printf("  %2d. %s\n", i+1, n)
	}
	for {
		raw := w.ask("\nProvider (name or number)", w.cfg.Provider)
		if n, err := strconv.Atoi(raw); err == nil && n >= 1 && n <= len(names) {
			w.cfg.Provider = names[n-1]
			break
		}
		if _, ok := wizardProviders()[raw]; ok {
			w.cfg.Provider = raw
			break
		}
		w.printf("  %q isn't a provider — pick one from the list.\n", raw)
	}

	// 2. Resources, validated against the provider's real list
	w.cfg.Resources = w.askResources()

	// 3. Provider-specific details
	switch w.cfg.Provider {
	case "google":
		for len(w.cfg.Projects) == 0 {
			w.cfg.Projects = splitCSV(w.ask("GCP project ID(s), comma-separated", ""))
			if len(w.cfg.Projects) == 0 {
				w.printf("  A project is required for google imports.\n")
			}
		}
		w.cfg.Regions = splitCSV(w.ask("Region(s)", "global"))
	case "aws":
		w.cfg.Regions = splitCSV(w.ask("Region(s), e.g. us-east-1,eu-west-1", ""))
		w.cfg.Profile = w.ask("AWS profile", "default")
	case "azure":
		for w.cfg.ResourceGroup == "" {
			w.cfg.ResourceGroup = w.ask("Resource group (* for all in the subscription)", "")
			if w.cfg.ResourceGroup == "" {
				w.printf("  A resource group is required for azure imports.\n")
			}
		}
	case "vault":
		def := os.Getenv("VAULT_ADDR")
		if def == "" {
			def = "http://127.0.0.1:8200"
		}
		w.cfg.VaultAddress = w.ask("Vault address", def)
		w.printf("  Set VAULT_TOKEN in your environment first — I won't put a token on a command line.\n")
	}

	// 4. Common options
	w.cfg.PathOutput = w.ask("Output folder", DefaultPathOutput)
	w.cfg.Strict = w.askYesNo("Fail loudly if anything fails to import (--strict, recommended)", true)
	w.cfg.Verbose = w.askYesNo("Verbose logging", false)
	return w.cfg.Validate()
}

// checkPrereqs reports (but does not block on) missing local tools and
// unset credential variables. It names variables only, never values.
func (w *wizardRunner) checkPrereqs() {
	w.printf("\nBefore you run:\n")
	if _, err := exec.LookPath("terraform"); err != nil {
		if _, err2 := exec.LookPath("tofu"); err2 != nil {
			w.printf("  ! terraform/tofu not found in PATH — install one to verify the import afterwards.\n")
		}
	}
	for _, h := range w.cfg.RequiredEnv() {
		// Only single, plain variable names can be checked directly.
		if !strings.ContainsAny(h.Name, " /") {
			if os.Getenv(h.Name) == "" {
				w.printf("  ! %s is not set — %s\n", h.Name, h.Purpose)
				continue
			}
			w.printf("  ✓ %s is set\n", h.Name)
			continue
		}
		w.printf("  • %s — %s\n", h.Name, h.Purpose)
	}
}

func (w *wizardRunner) printSummary() {
	w.printf("\nYour import command:\n\n  %s\n\n", w.cfg.ShellCommand())
	w.printf("After it finishes, verify the import is clean:\n")
	for _, s := range w.cfg.NextSteps() {
		w.printf("  %s\n", s)
	}
}

func newWizardCmd() *cobra.Command {
	var (
		flagProvider, flagResources, flagProjects, flagRegions string
		flagProfile, flagResourceGroup, flagVaultAddress       string
		flagOutput, flagScript                                 string
		flagDryRun, flagYes, flagStrict                        bool
	)
	cmd := &cobra.Command{
		Use:   "wizard",
		Short: "Interactive import wizard — answer questions, get the exact import command",
		Long: `Guided import: pick a provider, pick from its real list of services,
answer the few provider-specific questions, and the wizard validates
everything, shows the exact terraformer command, and can run it or
save it as a script. Use --dry-run with the flags below to generate a
command without prompts (handy for scripts and docs).`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := wizardConfig{
				Provider:      flagProvider,
				Profile:       flagProfile,
				ResourceGroup: flagResourceGroup,
				VaultAddress:  flagVaultAddress,
				PathOutput:    flagOutput,
				Strict:        flagStrict,
				Connect:       true,
			}
			if flagResources != "" {
				cfg.Resources = splitCSV(flagResources)
			}
			if flagProjects != "" {
				cfg.Projects = splitCSV(flagProjects)
			}
			if flagRegions != "" {
				cfg.Regions = splitCSV(flagRegions)
			}
			w := &wizardRunner{in: bufio.NewReader(os.Stdin), out: cmd.OutOrStdout(), cfg: cfg}

			nonInteractive := flagDryRun || flagProvider != ""
			if !nonInteractive {
				if err := w.runInteractive(); err != nil {
					return err
				}
			} else {
				// Sensible defaults for flag-driven use.
				if w.cfg.Provider == "google" && len(w.cfg.Regions) == 0 {
					w.cfg.Regions = []string{"global"}
				}
				if w.cfg.Provider == "aws" && w.cfg.Profile == "" {
					w.cfg.Profile = "default"
				}
				if err := w.cfg.Validate(); err != nil {
					return err
				}
			}

			w.checkPrereqs()
			w.printSummary()

			if flagScript != "" {
				script := "#!/bin/sh\n# Generated by: terraformer wizard\nset -e\n" + w.cfg.ShellCommand() + "\n"
				if err := os.WriteFile(flagScript, []byte(script), 0o755); err != nil {
					return err
				}
				w.printf("\nSaved as %s — run it with: sh %s\n", flagScript, flagScript)
			}
			if flagDryRun {
				return nil
			}
			run := flagYes
			if !run && !nonInteractive {
				run = w.askYesNo("\nRun this import now?", false)
			}
			if !run {
				w.printf("\nNothing run. Copy the command above whenever you're ready.\n")
				return nil
			}
			w.printf("\nRunning: %s\n\n", w.cfg.ShellCommand())
			root := NewCmdRoot()
			root.SetArgs(w.cfg.Args())
			root.SetOut(cmd.OutOrStdout())
			root.SetErr(cmd.ErrOrStderr())
			return root.Execute()
		},
	}
	cmd.Flags().StringVar(&flagProvider, "provider", "", "provider name, e.g. google, aws, azure, vault, grafana (skips the prompts)")
	cmd.Flags().StringVar(&flagResources, "resources", "", "comma-separated services, or * for all")
	cmd.Flags().StringVar(&flagProjects, "projects", "", "google: comma-separated project IDs")
	cmd.Flags().StringVar(&flagRegions, "regions", "", "google/aws: comma-separated regions")
	cmd.Flags().StringVar(&flagProfile, "profile", "", "aws: profile name")
	cmd.Flags().StringVar(&flagResourceGroup, "resource-group", "", "azure: resource group, or * for all")
	cmd.Flags().StringVar(&flagVaultAddress, "address", "", "vault: server address")
	cmd.Flags().StringVar(&flagOutput, "path-output", DefaultPathOutput, "output folder")
	cmd.Flags().StringVar(&flagScript, "script", "", "also save the command as an executable shell script at this path")
	cmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "validate and print the command, never run it")
	cmd.Flags().BoolVar(&flagYes, "yes", false, "run the import without asking")
	cmd.Flags().BoolVar(&flagStrict, "strict", true, "fail loudly if anything fails to import")
	return cmd
}
