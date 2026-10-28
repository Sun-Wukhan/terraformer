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

package gcp

import "testing"

// TestOsLoginGeneratorEmitsNothing pins the documented skip:
// google_os_login_ssh_public_key is a per-user collection with no
// project-level enumeration path.
func TestOsLoginGeneratorEmitsNothing(t *testing.T) {
	g := &OsLoginGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project"})
	if err := g.InitResources(); err != nil {
		t.Fatalf("InitResources: %v", err)
	}
	if len(g.Resources) != 0 {
		t.Fatalf("expected no resources, got %v", g.Resources)
	}
}
