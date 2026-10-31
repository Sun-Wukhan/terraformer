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

package terraformoutput

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

func TestPrintProviderTFDropsDeprecatedProviderVersion(t *testing.T) {
	providerData := map[string]interface{}{
		"provider": map[string]interface{}{
			"aws": map[string]interface{}{
				"region":  "us-east-1",
				"version": "~> 5.0",
			},
		},
	}
	got := string(printProviderTF(providerData))
	if strings.Contains(got, "version") {
		t.Fatalf("provider block must not carry a version constraint:\n%s", got)
	}
	if !strings.Contains(got, `region = "us-east-1"`) {
		t.Fatalf("other attributes must be kept:\n%s", got)
	}
}

// connectionsProvider is a minimal provider whose resource_group connection
// references two attributes (name and location) of the same target.
type connectionsProvider struct {
	terraformutils.Provider
}

// GetName returns the fake provider name.
func (p *connectionsProvider) GetName() string { return "azurerm" }

// InitService is unused by OutputHclFiles.
func (p *connectionsProvider) InitService(string, bool) error { return nil }

// GetProviderData returns an empty provider block.
func (p *connectionsProvider) GetProviderData(...string) map[string]interface{} {
	return map[string]interface{}{"provider": map[string]interface{}{"azurerm": map[string]interface{}{}}}
}

// GetResourceConnections links two attributes of network_security_group to
// resource_group.
func (p *connectionsProvider) GetResourceConnections() map[string]map[string][]string {
	return map[string]map[string][]string{
		"network_security_group": {
			"resource_group": {"resource_group_name", "name", "location", "location"},
		},
	}
}

func TestOutputHclFilesEmitsOutputForEveryConnectionPair(t *testing.T) {
	dir := t.TempDir()
	rg := terraformutils.NewSimpleResource("/rg", "Devtest", "azurerm_resource_group", "azurerm", nil)
	rg.InstanceState.Attributes = map[string]string{"id": "/rg", "name": "Devtest", "location": "eastus"}
	rg.Item = map[string]interface{}{"name": "Devtest", "location": "eastus"}
	resources := []terraformutils.Resource{rg}

	if err := OutputHclFiles(resources, &connectionsProvider{}, dir, "resource_group", false, "hcl", true); err != nil {
		t.Fatalf("OutputHclFiles: %v", err)
	}
	out, err := os.ReadFile(filepath.Join(dir, "outputs.tf"))
	if err != nil {
		t.Fatalf("reading outputs.tf: %v", err)
	}
	for _, want := range []string{
		`output "azurerm_resource_group_tfer--Devtest_id"`,
		`output "azurerm_resource_group_tfer--Devtest_name"`,
		`output "azurerm_resource_group_tfer--Devtest_location"`,
	} {
		if !strings.Contains(string(out), want) {
			t.Errorf("outputs.tf missing %s:\n%s", want, out)
		}
	}
	if got := resources[0].Outputs["azurerm_resource_group_tfer--Devtest_location"]; got == nil || got.Value != "eastus" {
		t.Fatalf("location output not recorded on the resource for the state file: %+v", resources[0].Outputs)
	}
}

func TestPrintProviderTF(t *testing.T) {
	providerData := map[string]interface{}{
		"provider": map[string]interface{}{
			"azurerm": map[string]interface{}{
				"features": map[string]interface{}{},
			},
		},
		"terraform": map[string]interface{}{
			"required_providers": []map[string]interface{}{{
				"azurerm": map[string]interface{}{
					"version": "4.0.0",
					"source":  "hashicorp/azurerm",
				},
			}},
		},
	}
	want := `terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "4.0.0"
    }
  }
}

provider "azurerm" {
  features {}
}
`
	if got := string(printProviderTF(providerData)); got != want {
		t.Errorf("printProviderTF() =\n%s\nwant\n%s", got, want)
	}
}
