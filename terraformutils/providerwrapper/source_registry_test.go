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
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

// TestOpenTofuRegistryCoverage verifies every provider source address in the
// table resolves on registry.opentofu.org, so `tofu init` can install it.
// This is a live-network test: it only runs with
// TERRAFORMER_INTEGRATION_TEST=1.
func TestOpenTofuRegistryCoverage(t *testing.T) {
	if os.Getenv("TERRAFORMER_INTEGRATION_TEST") != "1" {
		t.Skip("live registry test; set TERRAFORMER_INTEGRATION_TEST=1 to run")
	}
	client := &http.Client{Timeout: 30 * time.Second}
	var missing []string
	for name, src := range providerSources {
		// The /versions endpoint is what `tofu init` actually resolves.
		url := fmt.Sprintf("https://registry.opentofu.org/v1/providers/%s/versions", src)
		resp, err := client.Get(url)
		if err != nil {
			t.Errorf("provider %q (%s): request failed: %v", name, src, err)
			continue
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			missing = append(missing, fmt.Sprintf("%s (%s): HTTP %d", name, src, resp.StatusCode))
			continue
		}
		t.Logf("provider %q: %s OK", name, src)
	}
	if len(missing) > 0 {
		t.Errorf("providers missing from the OpenTofu registry:\n  %s", joinLines(missing))
	}
}

func joinLines(lines []string) string {
	out := ""
	for i, l := range lines {
		if i > 0 {
			out += "\n  "
		}
		out += l
	}
	return out
}
