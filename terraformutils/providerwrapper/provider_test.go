package providerwrapper //nolint

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/configs/configschema"
	"github.com/hashicorp/terraform/providers"
	"github.com/zclconf/go-cty/cty"
)

func TestIgnoredAttributes(t *testing.T) {
	attributes := map[string]*configschema.Attribute{
		"computed_attribute": {
			Type:     cty.Number,
			Computed: true,
		},
		"required_attribute": {
			Type:     cty.String,
			Required: true,
		},
	}

	testCases := map[string]struct {
		block                map[string]*configschema.NestedBlock
		ignoredAttributes    []string
		notIgnoredAttributes []string
	}{
		"nesting_set": {map[string]*configschema.NestedBlock{
			"attribute_one": {
				Block: configschema.Block{
					Attributes: attributes,
				},
				Nesting: configschema.NestingSet,
			},
		}, []string{"nesting_set.attribute_one.computed_attribute"},
			[]string{"nesting_set.attribute_one.required_attribute"}},
		"nesting_list": {map[string]*configschema.NestedBlock{
			"attribute_one": {
				Block: configschema.Block{
					Attributes: map[string]*configschema.Attribute{},
					BlockTypes: map[string]*configschema.NestedBlock{
						"attribute_two_nested": {
							Nesting: configschema.NestingList,
							Block: configschema.Block{
								Attributes: attributes,
							},
						},
					},
				},
				Nesting: configschema.NestingList,
			},
		}, []string{"nesting_list.0.attribute_one.0.attribute_two_nested.computed_attribute"},
			[]string{"nesting_list.0.attribute_one.0.attribute_two_nested.required_attribute"}},
	}

	for key, tc := range testCases {
		t.Run(key, func(t *testing.T) {
			provider := ProviderWrapper{}
			readOnlyAttributes := provider.readObjBlocks(tc.block, []string{}, key)
			for _, attr := range tc.ignoredAttributes {
				if ignored := isAttributeIgnored(attr, readOnlyAttributes); !ignored {
					t.Errorf("attribute \"%s\" was not ignored. Pattern list: %s", attr, readOnlyAttributes)
				}
			}

			for _, attr := range tc.notIgnoredAttributes {
				if ignored := isAttributeIgnored(attr, readOnlyAttributes); ignored {
					t.Errorf("attribute \"%s\" was ignored. Pattern list: %s", attr, readOnlyAttributes)
				}
			}
		})
	}
}

func TestRequiredAttributePatterns(t *testing.T) {
	block := &configschema.Block{
		Attributes: map[string]*configschema.Attribute{
			"name":     {Type: cty.String, Required: true},
			"location": {Type: cty.String, Optional: true},
			"os_disk":  {Type: cty.String, Optional: true, Computed: true},
		},
		BlockTypes: map[string]*configschema.NestedBlock{
			"subnet": {
				Nesting: configschema.NestingSet,
				Block: configschema.Block{
					Attributes: map[string]*configschema.Attribute{
						"address_prefix": {Type: cty.String, Required: true},
						"security_group": {Type: cty.String, Optional: true},
						"custom_runtime": {Type: cty.Bool, Optional: true, Computed: true},
					},
				},
			},
		},
	}
	patterns := attributePatterns(block, "^", func(a *configschema.Attribute) bool { return a.Required })

	for _, want := range []string{"name", "subnet.0.address_prefix", "subnet.12.address_prefix"} {
		if !isAttributeIgnored(want, patterns) {
			t.Errorf("%q should match a required pattern: %v", want, patterns)
		}
	}
	for _, notWant := range []string{"location", "subnet.0.security_group", "other.name", "subnet.name"} {
		if isAttributeIgnored(notWant, patterns) {
			t.Errorf("%q must not match a required pattern: %v", notWant, patterns)
		}
	}

	p := &ProviderWrapper{schema: &providers.GetSchemaResponse{
		ResourceTypes: map[string]providers.Schema{"x_vm": {Block: block}},
	}}
	oc := p.GetOptionalComputedAttributes("x_vm")
	for _, want := range []string{"os_disk", "subnet.3.custom_runtime"} {
		if !isAttributeIgnored(want, oc) {
			t.Errorf("%q should match an optional+computed pattern: %v", want, oc)
		}
	}
	for _, notWant := range []string{"name", "location", "subnet.0.address_prefix"} {
		if isAttributeIgnored(notWant, oc) {
			t.Errorf("%q must not match an optional+computed pattern: %v", notWant, oc)
		}
	}
	if p.GetOptionalComputedAttributes("missing") != nil {
		t.Errorf("unknown type must yield nil")
	}
}

func TestParseDeprecatedAttributes(t *testing.T) {
	doc := `{"format_version":"1.0","provider_schemas":{"registry.terraform.io/hashicorp/azurerm":{"provider":{"block":{}},
	  "resource_schemas":{"azurerm_storage_account":{"version":0,"block":{"attributes":{
	    "public_network_access_enabled":{"type":"bool","optional":true,"computed":true,"deprecated":true},
	    "public_network_access":{"type":"string","optional":true,"computed":true}},
	    "block_types":{"network_rules":{"nesting_mode":"list","block":{"attributes":{
	      "old_flag":{"type":"bool","optional":true,"deprecated":true}}}}}}}}}}}`
	got, err := parseDeprecatedAttributes([]byte(doc), "azurerm")
	if err != nil {
		t.Fatal(err)
	}
	patterns := got["azurerm_storage_account"]
	for _, want := range []string{"public_network_access_enabled", "network_rules.0.old_flag"} {
		if !isAttributeIgnored(want, patterns) {
			t.Errorf("%q should be deprecated: %v", want, patterns)
		}
	}
	if isAttributeIgnored("public_network_access", patterns) {
		t.Errorf("public_network_access must not be deprecated: %v", patterns)
	}
	if (&ProviderWrapper{deprecatedAttrs: got}).GetDeprecatedAttributes("other") != nil {
		t.Errorf("unknown type must yield nil")
	}
}

func TestProviderSourceAndDefaults(t *testing.T) {
	if got := GetProviderSource("azurerm"); got != "hashicorp/azurerm" {
		t.Fatalf("GetProviderSource(azurerm) = %q", got)
	}
	if got := GetProviderSource("no-such-provider"); got != "" {
		t.Fatalf("unknown provider must yield empty source, got %q", got)
	}
	if GetCLIVersion() == "" || GetRegistryHost() == "" {
		t.Fatalf("CLI version/registry host must always have a value")
	}
}

func isAttributeIgnored(name string, patterns []string) bool {
	ignored := false
	for _, pattern := range patterns {
		if match, _ := regexp.MatchString(pattern, name); match {
			ignored = true
			break
		}
	}
	return ignored
}
