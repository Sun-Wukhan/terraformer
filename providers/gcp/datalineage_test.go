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

// TestDataLineageGeneratorIsNoOp pins the documented skip: the Data
// Lineage API exposes no config resource, so nothing is emitted.
func TestDataLineageGeneratorIsNoOp(t *testing.T) {
	g := &DataLineageGenerator{}
	if err := g.InitResources(); err != nil {
		t.Fatalf("InitResources: %v", err)
	}
	if len(g.Resources) != 0 {
		t.Errorf("resources = %d, want 0 (documented skip)", len(g.Resources))
	}
}
