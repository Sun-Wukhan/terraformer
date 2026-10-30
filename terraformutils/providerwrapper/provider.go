// Copyright 2018 The Terraformer Authors.
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

package providerwrapper //nolint

// Architecture.
//
// This package used to drive provider binaries over go-plugin gRPC using the
// internals of Terraform 0.12 (github.com/hashicorp/terraform v0.12.31). That
// client only speaks provider protocol 4/5, so it cannot communicate with
// modern provider releases (protocol 6), and it forced users to hand-install
// provider binaries.
//
// It now shells out to a modern Terraform (>= 1.5) or OpenTofu CLI instead:
//   - NewProviderWrapper creates an isolated temp working directory, writes a
//     minimal versions.tf / provider.tf, and runs `<cli> init`, which
//     auto-installs the provider from the public registry.
//   - Provider schemas come from `<cli> providers schema -json`, converted by
//     schema_json.go into the 0.12 providers.GetSchemaResponse type that the
//     rest of terraformer consumes.
//   - Per-resource state refresh is done in bulk by the terraformutils package
//     via generated `import` blocks + `<cli> plan`, a verified import-only
//     `<cli> apply` of the saved plan, and `<cli> show -json`.
//
// The github.com/hashicorp/terraform v0.12.31 module is therefore kept in
// go.mod as a TYPES-ONLY dependency (configs/configschema, providers). None
// of its runtime/plugin code is used anymore.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils/terraformerstring"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/terraform/configs/configschema"
	"github.com/hashicorp/terraform/providers"
)

// DefaultDataDir is the default directory for storing local data.
const DefaultDataDir = ".terraform"

// DefaultPluginVendorDir is the location in the config directory to look for
// user-added plugin binaries. Terraform only reads from this path if it
// exists, it is never created by terraform.
const DefaultPluginVendorDirV12 = "terraform.d/plugins/" + pluginMachineName

// pluginMachineName is the directory name used in new plugin paths.
const pluginMachineName = runtime.GOOS + "_" + runtime.GOARCH

// cliTimeouts for the various CLI invocations.
const (
	cliInitTimeout    = 10 * time.Minute
	cliSchemaTimeout  = 5 * time.Minute
	cliVersionTimeout = 1 * time.Minute
)

// providerVersionCache remembers the installed provider version per
// terraformer provider name, populated by NewProviderWrapper after a
// successful init. GetProviderVersion serves it without re-running the CLI.
var providerVersionCache sync.Map // map[string]string

// cliInfo records the driven CLI's version and provider registry host, set by
// detectCLIVersion, so the generated state can declare a compatible
// terraform_version and the right provider address.
var cliInfo struct {
	sync.Mutex
	version      string
	registryHost string
}

// defaultCLIVersion is declared in generated state when the CLI version could
// not be detected: the oldest release whose import blocks terraformer needs.
const defaultCLIVersion = "1.5.0"

// defaultRegistryHost is the provider registry assumed for generated state
// when the CLI is not OpenTofu.
const defaultRegistryHost = "registry.terraform.io"

// GetCLIVersion returns the detected terraform/tofu version, or
// defaultCLIVersion when none was detected.
func GetCLIVersion() string {
	cliInfo.Lock()
	defer cliInfo.Unlock()
	if cliInfo.version == "" {
		return defaultCLIVersion
	}
	return cliInfo.version
}

// GetRegistryHost returns the provider registry host of the detected CLI
// (registry.opentofu.org for tofu, registry.terraform.io otherwise).
func GetRegistryHost() string {
	cliInfo.Lock()
	defer cliInfo.Unlock()
	if cliInfo.registryHost == "" {
		return defaultRegistryHost
	}
	return cliInfo.registryHost
}

// GetProviderSource returns the registry source address ("hashicorp/aws")
// for a terraformer provider name, or "" when the name is unknown.
func GetProviderSource(providerName string) string {
	source, _, err := lookupProviderSource(providerName)
	if err != nil {
		return ""
	}
	return source
}

type ProviderWrapper struct {
	providerName string
	tfType       string // terraform provider type, e.g. "aws"
	source       string // registry source address, e.g. "hashicorp/aws"
	cliBin       string // resolved terraform/tofu binary
	workDir      string // isolated temp working directory
	schema       *providers.GetSchemaResponse
	// deprecatedAttrs maps resource type -> flatmap-key patterns of its
	// deprecated attributes (not representable in configschema).
	deprecatedAttrs map[string][]string
	retryCount      int
	retrySleepMs    int
	verbose         bool
}

// resolveCLIBin finds the terraform/tofu CLI to drive: the TERRAFORMER_TF_BIN
// environment override wins, then `terraform` on PATH, then `tofu` on PATH.
func resolveCLIBin() (string, error) {
	if bin := os.Getenv("TERRAFORMER_TF_BIN"); bin != "" {
		resolved, err := exec.LookPath(bin)
		if err != nil {
			return "", fmt.Errorf("TERRAFORMER_TF_BIN=%q: %w", bin, err)
		}
		return resolved, nil
	}
	if bin, err := exec.LookPath("terraform"); err == nil {
		return bin, nil
	}
	if bin, err := exec.LookPath("tofu"); err == nil {
		return bin, nil
	}
	return "", errors.New("no terraform (>= 1.5) or tofu binary found on PATH; install Terraform >= 1.5 or OpenTofu, or set TERRAFORMER_TF_BIN to the CLI binary path")
}

func NewProviderWrapper(providerName string, providerConfig cty.Value, verbose bool, options ...map[string]int) (*ProviderWrapper, error) {
	p := &ProviderWrapper{retryCount: 5, retrySleepMs: 300, verbose: verbose}
	p.providerName = providerName

	if len(options) > 0 {
		if v, ok := options[0]["retryCount"]; ok {
			p.retryCount = v
		}
		if v, ok := options[0]["retrySleepMs"]; ok {
			p.retrySleepMs = v
		}
	}

	cliBin, err := resolveCLIBin()
	if err != nil {
		return nil, err
	}
	p.cliBin = cliBin
	p.detectCLIVersion()

	source, tfType, err := lookupProviderSource(providerName)
	if err != nil {
		return nil, err
	}
	p.source = source
	p.tfType = tfType

	workDir, err := os.MkdirTemp("", "terraformer-")
	if err != nil {
		return nil, fmt.Errorf("creating temp working directory: %w", err)
	}
	p.workDir = workDir
	// Clean up the temp dir if any of the setup steps below fail.
	failed := true
	defer func() {
		if failed {
			os.RemoveAll(workDir)
		}
	}()

	if err := p.writeVersionsTF(); err != nil {
		return nil, err
	}
	if err := p.writeProviderTF(providerConfig); err != nil {
		return nil, err
	}

	// Auto-install the provider from the registry. No more hand-placed
	// provider binaries.
	if _, err := p.runCLI(cliInitTimeout, "init", "-input=false", "-no-color"); err != nil {
		return nil, fmt.Errorf("provider auto-install failed: %w", err)
	}

	schemaOut, err := p.runCLI(cliSchemaTimeout, "providers", "schema", "-json", "-no-color")
	if err != nil {
		return nil, fmt.Errorf("fetching provider schema: %w", err)
	}
	schema, err := parseProvidersSchemaJSON(schemaOut, tfType)
	if err != nil {
		return nil, err
	}
	p.schema = schema
	if p.deprecatedAttrs, err = parseDeprecatedAttributes(schemaOut, tfType); err != nil {
		return nil, err
	}

	p.detectProviderVersion()

	failed = false
	return p, nil
}

// Kill removes the wrapper's isolated working directory.
func (p *ProviderWrapper) Kill() {
	if p.workDir != "" {
		os.RemoveAll(p.workDir)
	}
}

func (p *ProviderWrapper) GetSchema() *providers.GetSchemaResponse {
	return p.schema
}

func (p *ProviderWrapper) GetReadOnlyAttributes(resourceTypes []string) (map[string][]string, error) {
	r := p.GetSchema()
	if r == nil {
		return nil, errors.New("provider schema not initialized")
	}

	if r.Diagnostics.HasErrors() {
		return nil, r.Diagnostics.Err()
	}
	readOnlyAttributes := map[string][]string{}
	for resourceName, obj := range r.ResourceTypes {
		if terraformerstring.ContainsString(resourceTypes, resourceName) {
			readOnlyAttributes[resourceName] = append(readOnlyAttributes[resourceName], "^id$")
			for k, v := range obj.Block.Attributes {
				if !v.Optional && !v.Required {
					if v.Type.IsListType() || v.Type.IsSetType() {
						readOnlyAttributes[resourceName] = append(readOnlyAttributes[resourceName], "^"+k+"\\.(.*)")
					} else {
						readOnlyAttributes[resourceName] = append(readOnlyAttributes[resourceName], "^"+k+"$")
					}
				}
			}
			readOnlyAttributes[resourceName] = p.readObjBlocks(obj.Block.BlockTypes, readOnlyAttributes[resourceName], "-1")
		}
	}
	return readOnlyAttributes, nil
}

// GetRequiredAttributes returns flatmap-key patterns (e.g. `^name$`,
// `^subnet\.[0-9]+\.address_prefix$`) for every attribute the schema marks
// Required in resourceType. Required attributes must be emitted even when
// their value is empty or zero.
func (p *ProviderWrapper) GetRequiredAttributes(resourceType string) []string {
	r := p.GetSchema()
	if r == nil {
		return nil
	}
	rs, ok := r.ResourceTypes[resourceType]
	if !ok || rs.Block == nil {
		return nil
	}
	return attributePatterns(rs.Block, "^", func(a *configschema.Attribute) bool { return a.Required })
}

// GetOptionalComputedAttributes returns flatmap-key patterns for every
// attribute the schema marks both Optional and Computed in resourceType.
// Leaving such an attribute out of the config makes the provider keep the
// live value, so a zero value (false, 0, "") can be dropped without causing a
// diff; emitting it frequently trips ConflictsWith/ExactlyOneOf validation.
func (p *ProviderWrapper) GetOptionalComputedAttributes(resourceType string) []string {
	r := p.GetSchema()
	if r == nil {
		return nil
	}
	rs, ok := r.ResourceTypes[resourceType]
	if !ok || rs.Block == nil {
		return nil
	}
	return attributePatterns(rs.Block, "^", func(a *configschema.Attribute) bool { return a.Optional && a.Computed })
}

// GetDeprecatedAttributes returns flatmap-key patterns for every attribute
// the provider marks deprecated in resourceType. Deprecated attributes are
// aliases of newer ones and conflict with them, so they must not be emitted.
func (p *ProviderWrapper) GetDeprecatedAttributes(resourceType string) []string {
	return p.deprecatedAttrs[resourceType]
}

// attributePatterns walks block collecting anchored flatmap-key patterns for
// the attributes matching pred; prefix is the already-escaped path of block.
func attributePatterns(block *configschema.Block, prefix string, pred func(*configschema.Attribute) bool) []string {
	var patterns []string
	for name, attr := range block.Attributes {
		if pred(attr) {
			patterns = append(patterns, prefix+regexp.QuoteMeta(name)+"$")
		}
	}
	for name, nested := range block.BlockTypes {
		childPrefix := prefix + regexp.QuoteMeta(name) + "\\."
		switch nested.Nesting {
		case configschema.NestingList, configschema.NestingSet:
			childPrefix += "[0-9]+\\."
		case configschema.NestingMap:
			childPrefix += "[^.]+\\."
		}
		patterns = append(patterns, attributePatterns(&nested.Block, childPrefix, pred)...)
	}
	sort.Strings(patterns)
	return patterns
}

func (p *ProviderWrapper) readObjBlocks(block map[string]*configschema.NestedBlock, readOnlyAttributes []string, parent string) []string {
	for k, v := range block {
		if len(v.BlockTypes) > 0 {
			if parent == "-1" {
				readOnlyAttributes = p.readObjBlocks(v.BlockTypes, readOnlyAttributes, k)
			} else {
				readOnlyAttributes = p.readObjBlocks(v.BlockTypes, readOnlyAttributes, parent+"\\.[0-9]+\\."+k)
			}
		}
		fieldCount := 0
		for key, l := range v.Attributes {
			if !l.Optional && !l.Required {
				fieldCount++
				switch v.Nesting {
				case configschema.NestingList:
					if parent == "-1" {
						readOnlyAttributes = append(readOnlyAttributes, "^"+k+"\\.[0-9]+\\."+key+"($|\\.[0-9]+|\\.#)")
					} else {
						readOnlyAttributes = append(readOnlyAttributes, "^"+parent+"\\.(.*)\\."+key+"$")
					}
				case configschema.NestingSet:
					if parent == "-1" {
						readOnlyAttributes = append(readOnlyAttributes, "^"+k+"\\.[0-9]+\\."+key+"$")
					} else {
						readOnlyAttributes = append(readOnlyAttributes, "^"+parent+"\\.(.*)\\."+key+"($|\\.(.*))")
					}
				case configschema.NestingMap:
					readOnlyAttributes = append(readOnlyAttributes, parent+"\\."+key)
				default:
					readOnlyAttributes = append(readOnlyAttributes, parent+"\\."+key+"$")
				}
			}
		}
		if fieldCount == len(v.Block.Attributes) && fieldCount > 0 && len(v.BlockTypes) == 0 {
			readOnlyAttributes = append(readOnlyAttributes, "^"+k)
		}
	}
	return readOnlyAttributes
}

// --- CLI execution helpers ---

// runCLI executes the configured terraform/tofu binary with args in the
// wrapper's working directory, inheriting the caller's environment (cloud
// credentials live there) plus CHECKPOINT_DISABLE=1. Stdout is returned;
// on failure the error carries stdout and stderr for diagnosis.
func (p *ProviderWrapper) runCLI(timeout time.Duration, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, p.cliBin, args...)
	cmd.Dir = p.workDir
	cmd.Env = append(os.Environ(), "CHECKPOINT_DISABLE=1")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if p.verbose {
		log.Printf("[DEBUG] exec: %s %s (dir=%s)", p.cliBin, strings.Join(args, " "), p.workDir)
	}
	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("%s %s timed out after %s\nstdout: %s\nstderr: %s",
			p.cliBin, strings.Join(args, " "), timeout, stdout.String(), stderr.String())
	}
	if err != nil {
		return nil, fmt.Errorf("%s %s failed: %w\nstdout: %s\nstderr: %s",
			p.cliBin, strings.Join(args, " "), err, stdout.String(), stderr.String())
	}
	return stdout.Bytes(), nil
}

// RunCLI is the exported form of runCLI for the bulk import orchestration in
// the terraformutils package (which cannot be imported here without creating
// an import cycle).
func (p *ProviderWrapper) RunCLI(timeout time.Duration, args ...string) ([]byte, error) {
	return p.runCLI(timeout, args...)
}

// WorkDir returns the wrapper's isolated working directory.
func (p *ProviderWrapper) WorkDir() string {
	return p.workDir
}

// CLIBin returns the resolved terraform/tofu binary path.
func (p *ProviderWrapper) CLIBin() string {
	return p.cliBin
}

// GetRetryCount returns the configured number of retries for CLI operations.
func (p *ProviderWrapper) GetRetryCount() int {
	return p.retryCount
}

// GetRetrySleepMs returns the configured sleep between CLI retries in milliseconds.
func (p *ProviderWrapper) GetRetrySleepMs() int {
	return p.retrySleepMs
}

// --- working directory setup ---

func (p *ProviderWrapper) writeVersionsTF() error {
	content := fmt.Sprintf(`terraform {
  required_providers {
    %s = {
      source = "%s"
    }
  }
}
`, p.tfType, p.source)
	if err := os.WriteFile(filepath.Join(p.workDir, "versions.tf"), []byte(content), 0644); err != nil {
		return fmt.Errorf("writing versions.tf: %w", err)
	}
	return nil
}

// writeProviderTF renders the provider configuration (a flat cty object of
// primitives produced by each provider's GetConfig) as a provider block. A
// null or non-object config means "no explicit provider configuration" and no
// file is written; the CLI then falls back to environment variables and
// default credential chains, exactly as before.
func (p *ProviderWrapper) writeProviderTF(config cty.Value) error {
	if config.IsNull() {
		return nil
	}
	if !config.Type().IsObjectType() {
		log.Printf("WARN: provider config is not an object (type %s), skipping provider.tf generation", config.Type().FriendlyName())
		return nil
	}
	f := hclwrite.NewEmptyFile()
	body := f.Body()
	providerBlock := body.AppendNewBlock("provider", []string{p.tfType})
	if !writeConfigBody(providerBlock.Body(), config) {
		return nil
	}
	if err := os.WriteFile(filepath.Join(p.workDir, "provider.tf"), f.Bytes(), 0644); err != nil {
		return fmt.Errorf("writing provider.tf: %w", err)
	}
	return nil
}

// writeConfigBody writes the non-null values of an object into body and
// reports whether anything was written. Object-typed values become nested
// blocks (e.g. azurerm's required `features {}`), everything else an
// attribute; an empty object still yields an empty block.
func writeConfigBody(body *hclwrite.Body, config cty.Value) bool {
	values := config.AsValueMap()
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)
	wrote := false
	for _, name := range names {
		val := values[name]
		if val.IsNull() {
			continue
		}
		if val.Type().IsObjectType() {
			writeConfigBody(body.AppendNewBlock(name, nil).Body(), val)
		} else {
			body.SetAttributeValue(name, val)
		}
		wrote = true
	}
	return wrote
}

// detectCLIVersion records the resolved CLI's version and registry host in
// cliInfo, and logs a warning when it is older than 1.5, the minimum version
// supporting the `import` blocks used for bulk refresh.
func (p *ProviderWrapper) detectCLIVersion() {
	ctx, cancel := context.WithTimeout(context.Background(), cliVersionTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, p.cliBin, "version", "-json")
	cmd.Env = append(os.Environ(), "CHECKPOINT_DISABLE=1")
	out, err := cmd.Output()
	if err != nil {
		return
	}
	var v struct {
		TerraformVersion string `json:"terraform_version"`
		TofuVersion      string `json:"tofu_version"`
	}
	if err := json.Unmarshal(out, &v); err != nil {
		return
	}
	verStr := v.TerraformVersion
	host := defaultRegistryHost
	if verStr == "" {
		verStr = v.TofuVersion
		host = "registry.opentofu.org"
	}
	if verStr != "" {
		cliInfo.Lock()
		cliInfo.version = verStr
		cliInfo.registryHost = host
		cliInfo.Unlock()
	}
	parts := strings.SplitN(verStr, ".", 3)
	if len(parts) < 2 {
		return
	}
	var major, minor int
	if _, err := fmt.Sscanf(parts[0], "%d", &major); err != nil {
		return
	}
	if _, err := fmt.Sscanf(parts[1], "%d", &minor); err != nil {
		return
	}
	if major < 1 || (major == 1 && minor < 5) {
		log.Printf("WARN: %s version %s is older than 1.5; import blocks require terraform >= 1.5 or a recent tofu", p.cliBin, verStr)
	}
}

// detectProviderVersion records the installed provider version (from
// `<cli> version -json` provider_selections) so GetProviderVersion can serve
// it without re-running the CLI.
func (p *ProviderWrapper) detectProviderVersion() {
	out, err := p.runCLI(cliVersionTimeout, "version", "-json", "-no-color")
	if err != nil {
		log.Printf("WARN: unable to detect provider version: %v", err)
		return
	}
	var v struct {
		ProviderSelections map[string]json.RawMessage `json:"provider_selections"`
	}
	if err := json.Unmarshal(out, &v); err != nil {
		log.Printf("WARN: unable to parse version -json output: %v", err)
		return
	}
	for key, raw := range v.ProviderSelections {
		if !strings.HasSuffix(key, "/"+p.source) {
			continue
		}
		// Modern CLI format: "registry.terraform.io/hashicorp/aws": "6.68.0".
		var ver string
		if err := json.Unmarshal(raw, &ver); err == nil && ver != "" {
			providerVersionCache.Store(p.providerName, "~> "+ver)
			return
		}
		// Older CLI format: "registry.terraform.io/hashicorp/aws": {"version": "6.68.0"}.
		var obj struct {
			Version string `json:"version"`
		}
		if err := json.Unmarshal(raw, &obj); err == nil && obj.Version != "" {
			providerVersionCache.Store(p.providerName, "~> "+obj.Version)
			return
		}
	}
	log.Printf("WARN: provider %q not found in version -json provider_selections", p.source)
}

func GetProviderVersion(providerName string) string {
	if v, ok := providerVersionCache.Load(providerName); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	// Fallback: legacy filename-based discovery, for code paths where no
	// ProviderWrapper was ever constructed.
	return getProviderVersionFromFileName(providerName)
}

func getProviderVersionFromFileName(providerName string) string {
	providerFilePath, err := getProviderFileName(providerName)
	if err != nil {
		log.Println("Can't find provider file path. Ensure that you are following https://www.terraform.io/docs/configuration/providers.html#third-party-plugins.")
		return ""
	}
	t := strings.Split(providerFilePath, string(os.PathSeparator))
	providerFileName := t[len(t)-1]
	providerFileNameParts := strings.Split(providerFileName, "_")
	if len(providerFileNameParts) < 2 {
		log.Println("Can't find provider version. Ensure that you are following https://www.terraform.io/docs/configuration/providers.html#plugin-names-and-versions.")
		return ""
	}
	providerVersion := providerFileNameParts[1]
	return "~> " + strings.TrimPrefix(providerVersion, "v")
}

// getProviderFileName and friends are the legacy provider-binary discovery
// helpers. They are no longer used to initialize providers (init now
// auto-installs via the CLI); they remain only as the GetProviderVersion
// fallback above.
func getProviderFileName(providerName string) (string, error) {
	defaultDataDir := os.Getenv("TF_DATA_DIR")
	if defaultDataDir == "" {
		defaultDataDir = DefaultDataDir
	}
	providerFilePath, err := getProviderFileNameV13andV14(defaultDataDir, providerName)
	if err != nil || providerFilePath == "" {
		providerFilePath, err = getProviderFileNameV13andV14(os.Getenv("HOME")+string(os.PathSeparator)+
			".terraform.d", providerName)
	}
	if err != nil || providerFilePath == "" {
		return getProviderFileNameV12(providerName)
	}
	return providerFilePath, nil
}

func getProviderFileNameV13andV14(prefix, providerName string) (string, error) {
	// Read terraform v14 file path
	registryDir := prefix + string(os.PathSeparator) + "providers" + string(os.PathSeparator) +
		"registry.terraform.io"
	providerDirs, err := os.ReadDir(registryDir)
	if err != nil {
		// Read terraform v13 file path
		registryDir = prefix + string(os.PathSeparator) + "plugins" + string(os.PathSeparator) +
			"registry.terraform.io"
		providerDirs, err = os.ReadDir(registryDir)
		if err != nil {
			return "", err
		}
	}
	providerFilePath := ""
	for _, providerDir := range providerDirs {
		pluginPath := registryDir + string(os.PathSeparator) + providerDir.Name() +
			string(os.PathSeparator) + providerName
		dirs, err := os.ReadDir(pluginPath)
		if err != nil {
			continue
		}
		for _, dir := range dirs {
			if !dir.IsDir() {
				continue
			}
			for _, dir := range dirs {
				fullPluginPath := pluginPath + string(os.PathSeparator) + dir.Name() +
					string(os.PathSeparator) + runtime.GOOS + "_" + runtime.GOARCH
				files, err := os.ReadDir(fullPluginPath)
				if err == nil {
					for _, file := range files {
						if strings.HasPrefix(file.Name(), "terraform-provider-"+providerName) {
							providerFilePath = fullPluginPath + string(os.PathSeparator) + file.Name()
						}
					}
				}
			}
		}
	}
	return providerFilePath, nil
}

func getProviderFileNameV12(providerName string) (string, error) {
	defaultDataDir := os.Getenv("TF_DATA_DIR")
	if defaultDataDir == "" {
		defaultDataDir = DefaultDataDir
	}
	pluginPath := defaultDataDir + string(os.PathSeparator) + "plugins" + string(os.PathSeparator) + runtime.GOOS + "_" + runtime.GOARCH
	files, err := os.ReadDir(pluginPath)
	if err != nil {
		pluginPath = os.Getenv("HOME") + string(os.PathSeparator) + "." + DefaultPluginVendorDirV12
		files, err = os.ReadDir(pluginPath)
		if err != nil {
			return "", err
		}
	}
	providerFilePath := ""
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.HasPrefix(file.Name(), "terraform-provider-"+providerName) {
			providerFilePath = pluginPath + string(os.PathSeparator) + file.Name()
		}
	}
	return providerFilePath, nil
}
