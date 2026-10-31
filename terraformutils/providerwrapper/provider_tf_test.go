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

package providerwrapper //nolint

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestWriteProviderTFRendersObjectsAsBlocks(t *testing.T) {
	p := &ProviderWrapper{workDir: t.TempDir(), tfType: "azurerm"}
	config := cty.ObjectVal(map[string]cty.Value{
		"features":        cty.EmptyObjectVal,
		"subscription_id": cty.StringVal("sub-1"),
		"skipped":         cty.NullVal(cty.String),
	})

	if err := p.writeProviderTF(config); err != nil {
		t.Fatalf("writeProviderTF: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(p.workDir, "provider.tf"))
	if err != nil {
		t.Fatalf("reading provider.tf: %v", err)
	}
	want := "provider \"azurerm\" {\n  features {\n  }\n  subscription_id = \"sub-1\"\n}\n"
	if string(got) != want {
		t.Fatalf("provider.tf mismatch\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestWriteProviderTFSkipsAllNullConfig(t *testing.T) {
	p := &ProviderWrapper{workDir: t.TempDir(), tfType: "aws"}
	if err := p.writeProviderTF(cty.ObjectVal(map[string]cty.Value{"region": cty.NullVal(cty.String)})); err != nil {
		t.Fatalf("writeProviderTF: %v", err)
	}
	if _, err := os.Stat(filepath.Join(p.workDir, "provider.tf")); !os.IsNotExist(err) {
		t.Fatalf("provider.tf written for an all-null config (stat err: %v)", err)
	}
}
