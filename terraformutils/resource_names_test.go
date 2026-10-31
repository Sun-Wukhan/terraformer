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

package terraformutils

import (
	"testing"

	"github.com/hashicorp/terraform/terraform"
)

func namedResource(typ, name, id string) Resource {
	return Resource{
		InstanceInfo:  &terraform.InstanceInfo{Type: typ},
		InstanceState: &terraform.InstanceState{ID: id},
		ResourceName:  name,
	}
}

func TestDedupeResourceNames(t *testing.T) {
	imported := map[string][]Resource{
		"virtual_machine_child": {
			namedResource("azurerm_virtual_machine_extension", "tfer--MDE", "/vm1/ext"),
			namedResource("azurerm_virtual_machine_extension", "tfer--MDE", "/vm2/ext"),
			namedResource("azurerm_virtual_machine_extension", "tfer--MDE", "/vm3/ext"),
			namedResource("azurerm_virtual_machine_extension", "tfer--other", "/vm1/other"),
		},
		// Same name in a different type must not be touched.
		"eventhub": {
			namedResource("azurerm_eventhub_consumer_group", "tfer--MDE", "/eh/cg"),
		},
	}

	DedupeResourceNames(imported)

	ext := imported["virtual_machine_child"]
	want := []string{"tfer--MDE", "tfer--MDE-2", "tfer--MDE-3", "tfer--other"}
	for i, w := range want {
		if ext[i].ResourceName != w {
			t.Fatalf("resource %d: got %q, want %q", i, ext[i].ResourceName, w)
		}
	}
	if imported["eventhub"][0].ResourceName != "tfer--MDE" {
		t.Fatalf("different type renamed: %q", imported["eventhub"][0].ResourceName)
	}

	// Idempotent: running again changes nothing.
	DedupeResourceNames(imported)
	for i, w := range want {
		if ext[i].ResourceName != w {
			t.Fatalf("second pass changed resource %d to %q", i, ext[i].ResourceName)
		}
	}
}
