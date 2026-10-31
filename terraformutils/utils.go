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

package terraformutils

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils/providerwrapper"
)

type BaseResource struct {
	Tags map[string]string `json:"tags,omitempty"`
}

// State file (format version 4) model, the format every Terraform >= 0.12
// and OpenTofu release reads. Terraformer used to write version 3 with the
// 0.12 provider address (`provider.azurerm`), which modern CLIs can only
// load after a manual 0.13 upgrade / `state replace-provider`.
type stateV4 struct {
	Version          int                 `json:"version"`
	TerraformVersion string              `json:"terraform_version"`
	Serial           uint64              `json:"serial"`
	Lineage          string              `json:"lineage"`
	Outputs          map[string]outputV4 `json:"outputs"`
	Resources        []resourceStateV4   `json:"resources"`
	CheckResults     []json.RawMessage   `json:"check_results"`
}

type outputV4 struct {
	Value json.RawMessage `json:"value"`
	Type  json.RawMessage `json:"type"`
}

type resourceStateV4 struct {
	Mode      string            `json:"mode"`
	Type      string            `json:"type"`
	Name      string            `json:"name"`
	Provider  string            `json:"provider"`
	Instances []instanceStateV4 `json:"instances"`
}

type instanceStateV4 struct {
	SchemaVersion       uint64            `json:"schema_version"`
	Attributes          json.RawMessage   `json:"attributes,omitempty"`
	AttributesFlat      map[string]string `json:"attributes_flat,omitempty"`
	SensitiveAttributes []json.RawMessage `json:"sensitive_attributes"`
}

// NewTfState builds a version-4 state for resources. Instances carry the
// typed JSON attributes recorded at refresh (StateJSON); resources without
// it fall back to the legacy flatmap, which the CLI upgrades on first use.
func NewTfState(resources []Resource) *stateV4 {
	state := &stateV4{
		Version:          4,
		TerraformVersion: providerwrapper.GetCLIVersion(),
		Serial:           1,
		Lineage:          newLineage(),
		Outputs:          map[string]outputV4{},
		Resources:        []resourceStateV4{},
		CheckResults:     nil,
	}
	for _, r := range resources {
		for k, v := range r.Outputs {
			if v == nil {
				continue
			}
			value, err := json.Marshal(v.Value)
			if err != nil {
				log.Printf("WARN: skipping output %s: %v", k, err)
				continue
			}
			state.Outputs[k] = outputV4{Value: value, Type: outputTypeJSON(v.Type)}
		}
	}
	for _, r := range resources {
		if r.InstanceInfo == nil || r.InstanceState == nil {
			continue
		}
		instance := instanceStateV4{
			SchemaVersion:       schemaVersionOf(r),
			SensitiveAttributes: []json.RawMessage{},
		}
		if len(r.StateJSON) > 0 {
			instance.Attributes = r.StateJSON
		} else {
			instance.AttributesFlat = flatAttributesWithID(r)
		}
		state.Resources = append(state.Resources, resourceStateV4{
			Mode:      "managed",
			Type:      r.InstanceInfo.Type,
			Name:      r.ResourceName,
			Provider:  providerAddress(r.Provider),
			Instances: []instanceStateV4{instance},
		})
	}
	sort.Slice(state.Resources, func(i, j int) bool {
		if state.Resources[i].Type != state.Resources[j].Type {
			return state.Resources[i].Type < state.Resources[j].Type
		}
		return state.Resources[i].Name < state.Resources[j].Name
	})
	return state
}

// PrintTfState renders the version-4 state file for resources.
func PrintTfState(resources []Resource) ([]byte, error) {
	out, err := json.MarshalIndent(NewTfState(resources), "", "  ")
	if err != nil {
		return nil, err
	}
	return append(out, '\n'), nil
}

// providerAddress returns the fully-qualified provider address used in state,
// e.g. provider["registry.terraform.io/hashicorp/azurerm"].
func providerAddress(providerName string) string {
	source := providerwrapper.GetProviderSource(providerName)
	if source == "" {
		source = "hashicorp/" + providerName
	}
	return fmt.Sprintf("provider[%q]", providerwrapper.GetRegistryHost()+"/"+source)
}

// schemaVersionOf reads the provider schema version recorded on the
// resource's state at refresh time.
func schemaVersionOf(r Resource) uint64 {
	if r.InstanceState.Meta == nil {
		return 0
	}
	switch v := r.InstanceState.Meta["schema_version"].(type) {
	case string:
		n, _ := strconv.ParseUint(v, 10, 64)
		return n
	case int:
		return uint64(v)
	case float64:
		return uint64(v)
	}
	return 0
}

// flatAttributesWithID returns the legacy flatmap attributes, making sure the
// id the CLI needs is present.
func flatAttributesWithID(r Resource) map[string]string {
	flat := make(map[string]string, len(r.InstanceState.Attributes)+1)
	for k, v := range r.InstanceState.Attributes {
		flat[k] = v
	}
	if _, ok := flat["id"]; !ok && r.InstanceState.ID != "" {
		flat["id"] = r.InstanceState.ID
	}
	return flat
}

// outputTypeJSON maps the legacy output type name to a state v4 type
// expression; terraformer only ever emits string outputs.
func outputTypeJSON(legacyType string) json.RawMessage {
	switch legacyType {
	case "string", "":
		return json.RawMessage(`"string"`)
	case "list":
		return json.RawMessage(`["list","string"]`)
	case "map":
		return json.RawMessage(`["map","string"]`)
	}
	return json.RawMessage(`"dynamic"`)
}

// newLineage returns a random UUID for the state lineage.
func newLineage() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "00000000-0000-4000-8000-000000000000"
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func RefreshResourcesByProvider(providersMapping *ProvidersMapping, providerWrapper *providerwrapper.ProviderWrapper, report *ImportReport) error {
	allResources := providersMapping.ShuffleResources()

	// Attribute each resource to its service for failure reporting.
	serviceOf := make(map[*Resource]string, len(allResources))
	for resource, provider := range providersMapping.resourceToProvider {
		serviceOf[resource] = providersMapping.providerToService[provider]
	}

	// State refresh is done in bulk via generated `import` blocks + a single
	// CLI apply + `show -json` (see bulk_import.go), replacing the old
	// per-resource go-plugin gRPC refresh.
	refreshedResources, err := bulkImportResources(providerWrapper, allResources, report, serviceOf)
	if err != nil {
		return err
	}

	providersMapping.SetResources(refreshedResources)
	return nil
}

func IgnoreKeys(resourcesTypes []string, p *providerwrapper.ProviderWrapper) map[string][]string {
	readOnlyAttributes, err := p.GetReadOnlyAttributes(resourcesTypes)
	if err != nil {
		log.Println("plugin error 2:", err)
		return map[string][]string{}
	}
	return readOnlyAttributes
}

func ParseFilterValues(value string) []string {
	var values []string

	valueBuffering := true
	wrapped := false
	var valueBuffer []byte
	for i := 0; i < len(value); i++ {
		if value[i] == '\'' {
			wrapped = !wrapped
			continue
		} else if value[i] == ':' {
			if len(valueBuffer) == 0 {
				continue
			} else if valueBuffering && !wrapped {
				values = append(values, string(valueBuffer))
				valueBuffering = false
				valueBuffer = []byte{}
				continue
			}
		}
		valueBuffering = true
		valueBuffer = append(valueBuffer, value[i])
	}
	if len(valueBuffer) > 0 {
		values = append(values, string(valueBuffer))
	}

	return values
}

func FilterCleanup(s *Service, isInitial bool) {
	if len(s.Filter) == 0 {
		return
	}
	var newListOfResources []Resource
	for _, resource := range s.Resources {
		allPredicatesTrue := true
		for _, filter := range s.Filter {
			if filter.isInitial() == isInitial {
				allPredicatesTrue = allPredicatesTrue && filter.Filter(resource)
			}
		}
		if allPredicatesTrue && !ContainsResource(newListOfResources, resource) {
			newListOfResources = append(newListOfResources, resource)
		}
	}
	s.Resources = newListOfResources
}

func ContainsResource(s []Resource, e Resource) bool {
	for _, a := range s {
		if a.InstanceInfo.Id == e.InstanceInfo.Id {
			return true
		}
	}
	return false
}
