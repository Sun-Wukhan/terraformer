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
	"testing"

	"github.com/hashicorp/terraform/configs/configschema"
	"github.com/zclconf/go-cty/cty"
)

func TestParseTypeExpr(t *testing.T) {
	mustJSON := func(s string) interface{} {
		var v interface{}
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			t.Fatalf("bad test JSON %q: %v", s, err)
		}
		return v
	}

	testCases := map[string]struct {
		json     string
		expected cty.Type
		wantErr  bool
	}{
		"string":  {`"string"`, cty.String, false},
		"number":  {`"number"`, cty.Number, false},
		"bool":    {`"bool"`, cty.Bool, false},
		"dynamic": {`"dynamic"`, cty.DynamicPseudoType, false},
		"list":    {`["list", "string"]`, cty.List(cty.String), false},
		"set":     {`["set", "number"]`, cty.Set(cty.Number), false},
		"map":     {`["map", "bool"]`, cty.Map(cty.Bool), false},
		"object": {
			`["object", {"name": "string", "count": "number"}]`,
			cty.Object(map[string]cty.Type{"name": cty.String, "count": cty.Number}),
			false,
		},
		"tuple": {
			`["tuple", ["string", "number"]]`,
			cty.Tuple([]cty.Type{cty.String, cty.Number}),
			false,
		},
		"nested list of objects": {
			`["list", ["object", {"id": "string", "tags": ["map", "string"]}]]`,
			cty.List(cty.Object(map[string]cty.Type{
				"id":   cty.String,
				"tags": cty.Map(cty.String),
			})),
			false,
		},
		"nested set of list": {
			`["set", ["list", "bool"]]`,
			cty.Set(cty.List(cty.Bool)),
			false,
		},
		"unknown primitive": {`"float"`, cty.NilType, true},
		"unknown kind":      {`["bag", "string"]`, cty.NilType, true},
		"malformed pair":    {`["list"]`, cty.NilType, true},
		"not a type":        {`42`, cty.NilType, true},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got, err := parseTypeExpr(mustJSON(tc.json))
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got type %s", got.FriendlyName())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !got.Equals(tc.expected) {
				t.Fatalf("got %s, want %s", got.FriendlyName(), tc.expected.FriendlyName())
			}
		})
	}
}

func TestParseProvidersSchemaJSON(t *testing.T) {
	doc := `{
  "format_version": "1.0",
  "provider_schemas": {
    "registry.terraform.io/hashicorp/aws": {
      "provider": {
        "version": 0,
        "block": {
          "attributes": {
            "region": {"type": "string", "description": "The region", "optional": true}
          },
          "block_types": {}
        }
      },
      "resource_schemas": {
        "aws_instance": {
          "version": 1,
          "block": {
            "attributes": {
              "id": {"type": "string", "computed": true},
              "ami": {"type": "string", "optional": true},
              "tags": {"type": ["map", "string"], "optional": true},
              "secret": {"type": "string", "optional": true, "sensitive": true}
            },
            "block_types": {
              "ebs_block_device": {
                "nesting_mode": "set",
                "block": {
                  "attributes": {
                    "device_name": {"type": "string", "required": true},
                    "volume_size": {"type": "number", "optional": true}
                  },
                  "block_types": {}
                }
              }
            }
          }
        }
      },
      "data_source_schemas": {
        "aws_ami": {
          "version": 0,
          "block": {
            "attributes": {
              "id": {"type": "string", "computed": true}
            },
            "block_types": {}
          }
        }
      }
    }
  }
}`
	resp, err := parseProvidersSchemaJSON([]byte(doc), "aws")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Provider block.
	if got := resp.Provider.Block.Attributes["region"]; got == nil || !got.Optional || got.Type != cty.String {
		t.Fatalf("provider region attribute wrong: %+v", got)
	}

	// Resource schema + version.
	rs, ok := resp.ResourceTypes["aws_instance"]
	if !ok {
		t.Fatalf("aws_instance missing from ResourceTypes")
	}
	if rs.Version != 1 {
		t.Fatalf("aws_instance version = %d, want 1", rs.Version)
	}
	attrs := rs.Block.Attributes
	if !attrs["id"].Computed || attrs["id"].Optional || attrs["id"].Required {
		t.Fatalf("id flags wrong: %+v", attrs["id"])
	}
	if !attrs["ami"].Optional || attrs["ami"].Computed {
		t.Fatalf("ami flags wrong: %+v", attrs["ami"])
	}
	if !attrs["secret"].Sensitive {
		t.Fatalf("secret should be sensitive")
	}
	if !attrs["tags"].Type.Equals(cty.Map(cty.String)) {
		t.Fatalf("tags type = %s, want map of string", attrs["tags"].Type.FriendlyName())
	}

	// Nested block type + nesting mode.
	nb, ok := rs.Block.BlockTypes["ebs_block_device"]
	if !ok {
		t.Fatalf("ebs_block_device missing from BlockTypes")
	}
	if nb.Nesting != configschema.NestingSet {
		t.Fatalf("ebs_block_device nesting = %v, want NestingSet", nb.Nesting)
	}
	if !nb.Block.Attributes["device_name"].Required {
		t.Fatalf("device_name should be required")
	}

	// ImpliedType sanity: the whole block must produce a coherent cty type
	// containing the attributes and nested block types.
	implied := rs.Block.ImpliedType()
	if !implied.IsObjectType() {
		t.Fatalf("ImpliedType() is %s, want object", implied.FriendlyName())
	}
	attrTypes := implied.AttributeTypes()
	for _, want := range []string{"id", "ami", "tags", "secret"} {
		if _, ok := attrTypes[want]; !ok {
			t.Fatalf("ImpliedType() missing attribute %q", want)
		}
	}
	if _, ok := attrTypes["ebs_block_device"]; !ok {
		t.Fatalf("ImpliedType() missing nested block type %q", "ebs_block_device")
	}
	if !attrTypes["tags"].Equals(cty.Map(cty.String)) {
		t.Fatalf("ImpliedType() tags = %s, want map of string", attrTypes["tags"].FriendlyName())
	}

	// Data sources.
	if _, ok := resp.DataSources["aws_ami"]; !ok {
		t.Fatalf("aws_ami missing from DataSources")
	}

	// Unknown provider type selects nothing.
	if _, err := parseProvidersSchemaJSON([]byte(doc), "nope"); err == nil {
		t.Fatalf("expected error for unknown provider type")
	}
}
