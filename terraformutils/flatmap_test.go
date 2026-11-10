package terraformutils

import (
	"regexp"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestZeroValueHandling(t *testing.T) {
	ty := cty.Object(map[string]cty.Type{
		"plain_zero":    cty.Number, // Optional only: a reported 0 is data
		"oc_zero":       cty.Number, // Optional+Computed: dropped, recorded
		"required_zero": cty.Number, // Required: kept even at 0
		"oc_false":      cty.Bool,   // Optional+Computed false: dropped, recorded
		"plain_five":    cty.Number,
	})
	attributes := map[string]string{
		"plain_zero":    "0",
		"oc_zero":       "0",
		"required_zero": "0",
		"oc_false":      "false",
		"plain_five":    "5",
	}
	parser := NewFlatmapParser(attributes, nil,
		[]*regexp.Regexp{regexp.MustCompile(`^required_zero$`)}).
		WithDropZeroKeys([]*regexp.Regexp{regexp.MustCompile(`^oc_`)})
	result, err := parser.Parse(ty)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := result["plain_zero"]; !ok {
		t.Errorf("plain numeric 0 must be kept, got %v", result)
	}
	if _, ok := result["required_zero"]; !ok {
		t.Errorf("required numeric 0 must be kept, got %v", result)
	}
	if _, ok := result["oc_zero"]; ok {
		t.Errorf("Optional+Computed numeric 0 must be dropped, got %v", result)
	}
	if _, ok := result["oc_false"]; ok {
		t.Errorf("Optional+Computed false must be dropped, got %v", result)
	}
	if result["plain_five"] != "5" {
		t.Errorf("non-zero values must pass through, got %v", result)
	}
	dropped := parser.DroppedZeros()
	if len(dropped) != 2 {
		t.Fatalf("DroppedZeros = %v, want oc_zero and oc_false", dropped)
	}
	seen := map[string]bool{}
	for _, d := range dropped {
		seen[d] = true
	}
	if !seen["oc_zero"] || !seen["oc_false"] {
		t.Fatalf("DroppedZeros = %v, want oc_zero and oc_false", dropped)
	}
}

func TestNestedAttributeFiltering(t *testing.T) {
	attributes := map[string]string{
		"attribute":        "value1",
		"nested.attribute": "value2",
	}

	ignoreKeys := []*regexp.Regexp{
		regexp.MustCompile(`^attribute$`),
	}
	parser := NewFlatmapParser(attributes, ignoreKeys, []*regexp.Regexp{})

	attributesType := cty.Object(map[string]cty.Type{
		"attribute": cty.String,
		"nested": cty.Object(map[string]cty.Type{
			"attribute": cty.String,
		}),
	})

	result, _ := parser.Parse(attributesType)

	if _, ok := result["attribute"]; ok {
		t.Errorf("failed to resolve %v", result)
	}
	if val, ok := result["nested"].(map[string]interface{})["attribute"]; !ok && val != "value2" {
		t.Errorf("failed to resolve %v", result)
	}
}
