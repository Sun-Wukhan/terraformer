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

import "testing"

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
