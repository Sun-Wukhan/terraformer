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

package providerwrapper

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/configs/configschema"
	"github.com/hashicorp/terraform/providers"
	"github.com/zclconf/go-cty/cty"
)

// This file converts the output of `terraform providers schema -json`
// (format_version "1.0", also produced by `tofu providers schema -json`)
// into the providers.GetSchemaResponse type from the Terraform 0.12 module,
// which the rest of terraformer consumes as its schema model. The 0.12 module
// is therefore kept as a types-only dependency; no 0.12 runtime behavior is
// used.

// --- JSON document shapes (format_version "1.0") ---

type schemaJSONDoc struct {
	FormatVersion   string                        `json:"format_version"`
	ProviderSchemas map[string]providerSchemaJSON `json:"provider_schemas"`
}

type providerSchemaJSON struct {
	Provider          versionedBlockJSON            `json:"provider"`
	ResourceSchemas   map[string]versionedBlockJSON `json:"resource_schemas"`
	DataSourceSchemas map[string]versionedBlockJSON `json:"data_source_schemas"`
}

type versionedBlockJSON struct {
	Version int       `json:"version"`
	Block   blockJSON `json:"block"`
}

type blockJSON struct {
	Attributes  map[string]attributeJSON `json:"attributes"`
	BlockTypes  map[string]blockTypeJSON `json:"block_types"`
	Description string                   `json:"description"`
}

type attributeJSON struct {
	Type        interface{}     `json:"type"`
	NestedType  *nestedTypeJSON `json:"nested_type"`
	Description string          `json:"description"`
	Required    bool            `json:"required"`
	Optional    bool            `json:"optional"`
	Computed    bool            `json:"computed"`
	Sensitive   bool            `json:"sensitive"`
}

// nestedTypeJSON is the schema shape for protocol-6 nested attributes
// (terraform-plugin-framework SingleNestedAttribute etc.): the attribute
// carries nested_type instead of a type expression.
type nestedTypeJSON struct {
	Attributes  map[string]attributeJSON `json:"attributes"`
	NestingMode string                   `json:"nesting_mode"`
}

type blockTypeJSON struct {
	NestingMode string    `json:"nesting_mode"`
	Block       blockJSON `json:"block"`
	MinItems    int       `json:"min_items"`
	MaxItems    int       `json:"max_items"`
}

// parseProvidersSchemaJSON converts raw `providers schema -json` output into a
// providers.GetSchemaResponse, selecting the provider whose registry address
// ends with "/<tfType>" (e.g. registry.terraform.io/hashicorp/aws for "aws").
func parseProvidersSchemaJSON(data []byte, tfType string) (*providers.GetSchemaResponse, error) {
	var doc schemaJSONDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parsing providers schema JSON: %w", err)
	}

	var match *providerSchemaJSON
	for key, ps := range doc.ProviderSchemas {
		parts := strings.Split(key, "/")
		if parts[len(parts)-1] == tfType {
			c := ps
			match = &c
			break
		}
	}
	if match == nil {
		return nil, fmt.Errorf("no provider schema found for provider type %q", tfType)
	}

	resp := &providers.GetSchemaResponse{
		ResourceTypes: make(map[string]providers.Schema, len(match.ResourceSchemas)),
		DataSources:   make(map[string]providers.Schema, len(match.DataSourceSchemas)),
	}

	providerBlock, err := convertBlockJSON(match.Provider.Block)
	if err != nil {
		return nil, fmt.Errorf("provider schema block: %w", err)
	}
	resp.Provider = providers.Schema{
		Version: int64(match.Provider.Version),
		Block:   providerBlock,
	}

	for name, rs := range match.ResourceSchemas {
		block, err := convertBlockJSON(rs.Block)
		if err != nil {
			return nil, fmt.Errorf("resource schema %q: %w", name, err)
		}
		resp.ResourceTypes[name] = providers.Schema{
			Version: int64(rs.Version),
			Block:   block,
		}
	}
	for name, ds := range match.DataSourceSchemas {
		block, err := convertBlockJSON(ds.Block)
		if err != nil {
			return nil, fmt.Errorf("data source schema %q: %w", name, err)
		}
		resp.DataSources[name] = providers.Schema{
			Version: int64(ds.Version),
			Block:   block,
		}
	}
	return resp, nil
}

// convertBlockJSON converts a schema JSON block into a configschema.Block.
func convertBlockJSON(b blockJSON) (*configschema.Block, error) {
	out := &configschema.Block{
		Attributes: make(map[string]*configschema.Attribute, len(b.Attributes)),
		BlockTypes: make(map[string]*configschema.NestedBlock, len(b.BlockTypes)),
	}
	for name, a := range b.Attributes {
		ty, err := convertAttributeType(name, a)
		if err != nil {
			return nil, err
		}
		out.Attributes[name] = &configschema.Attribute{
			Type:        ty,
			Description: a.Description,
			Required:    a.Required,
			Optional:    a.Optional,
			Computed:    a.Computed,
			Sensitive:   a.Sensitive,
		}
	}
	for name, bt := range b.BlockTypes {
		inner, err := convertBlockJSON(bt.Block)
		if err != nil {
			return nil, fmt.Errorf("block type %q: %w", name, err)
		}
		var nesting configschema.NestingMode
		switch bt.NestingMode {
		case "single":
			nesting = configschema.NestingSingle
		case "group":
			nesting = configschema.NestingGroup
		case "list":
			nesting = configschema.NestingList
		case "set":
			nesting = configschema.NestingSet
		case "map":
			nesting = configschema.NestingMap
		default:
			return nil, fmt.Errorf("block type %q: unknown nesting mode %q", name, bt.NestingMode)
		}
		out.BlockTypes[name] = &configschema.NestedBlock{
			Block:    *inner,
			Nesting:  nesting,
			MinItems: bt.MinItems,
			MaxItems: bt.MaxItems,
		}
	}
	return out, nil
}

// convertAttributeType resolves an attribute's cty.Type from either a type
// expression or a protocol-6 nested_type object.
func convertAttributeType(name string, a attributeJSON) (cty.Type, error) {
	if a.NestedType != nil {
		ty, err := convertNestedTypeJSON(a.NestedType)
		if err != nil {
			return cty.NilType, fmt.Errorf("attribute %q: %w", name, err)
		}
		return ty, nil
	}
	ty, err := parseTypeExpr(a.Type)
	if err != nil {
		return cty.NilType, fmt.Errorf("attribute %q: %w", name, err)
	}
	return ty, nil
}

// convertNestedTypeJSON synthesizes the cty.Type for a nested_type attribute:
// a single object, or a list/set/map of objects for the other nesting modes.
func convertNestedTypeJSON(nt *nestedTypeJSON) (cty.Type, error) {
	attrTypes := make(map[string]cty.Type, len(nt.Attributes))
	for name, a := range nt.Attributes {
		ty, err := convertAttributeType(name, a)
		if err != nil {
			return cty.NilType, err
		}
		attrTypes[name] = ty
	}
	obj := cty.Object(attrTypes)
	switch nt.NestingMode {
	case "single":
		return obj, nil
	case "list":
		return cty.List(obj), nil
	case "set":
		return cty.Set(obj), nil
	case "map":
		return cty.Map(obj), nil
	default:
		return cty.NilType, fmt.Errorf("unknown nested_type nesting mode %q", nt.NestingMode)
	}
}

// parseTypeExpr parses a Terraform type expression as it appears in
// `providers schema -json` output:
//
//	"string" | "number" | "bool" | "dynamic"
//	["list", et] | ["set", et] | ["map", et]
//	["object", {"attr": <typeexpr>, ...}]
//	["tuple", [<typeexpr>, ...]]
func parseTypeExpr(v interface{}) (cty.Type, error) {
	switch tv := v.(type) {
	case string:
		switch tv {
		case "string":
			return cty.String, nil
		case "number":
			return cty.Number, nil
		case "bool":
			return cty.Bool, nil
		case "dynamic":
			return cty.DynamicPseudoType, nil
		default:
			return cty.NilType, fmt.Errorf("unknown primitive type %q", tv)
		}
	case []interface{}:
		if len(tv) != 2 {
			return cty.NilType, fmt.Errorf("malformed type expression (want [kind, element]): %v", v)
		}
		kind, ok := tv[0].(string)
		if !ok {
			return cty.NilType, fmt.Errorf("malformed type expression kind: %v", tv[0])
		}
		switch kind {
		case "list", "set", "map":
			elem, err := parseTypeExpr(tv[1])
			if err != nil {
				return cty.NilType, err
			}
			switch kind {
			case "list":
				return cty.List(elem), nil
			case "set":
				return cty.Set(elem), nil
			default:
				return cty.Map(elem), nil
			}
		case "object":
			attrs, ok := tv[1].(map[string]interface{})
			if !ok {
				return cty.NilType, fmt.Errorf("malformed object type expression: %v", tv[1])
			}
			attrTypes := make(map[string]cty.Type, len(attrs))
			for name, at := range attrs {
				aty, err := parseTypeExpr(at)
				if err != nil {
					return cty.NilType, fmt.Errorf("object attribute %q: %w", name, err)
				}
				attrTypes[name] = aty
			}
			return cty.Object(attrTypes), nil
		case "tuple":
			elems, ok := tv[1].([]interface{})
			if !ok {
				return cty.NilType, fmt.Errorf("malformed tuple type expression: %v", tv[1])
			}
			elemTypes := make([]cty.Type, len(elems))
			for i, e := range elems {
				ety, err := parseTypeExpr(e)
				if err != nil {
					return cty.NilType, fmt.Errorf("tuple element %d: %w", i, err)
				}
				elemTypes[i] = ety
			}
			return cty.Tuple(elemTypes), nil
		default:
			return cty.NilType, fmt.Errorf("unknown collection type kind %q", kind)
		}
	default:
		return cty.NilType, fmt.Errorf("unsupported type expression: %v", v)
	}
}
