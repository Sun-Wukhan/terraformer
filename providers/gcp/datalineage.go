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

import "log"

type DataLineageGenerator struct {
	GCPService
}

// InitResources deliberately generates nothing.
//
// Deliberately not generated:
//   - google_data_lineage_config: the Data Lineage v1 API surface in
//     the pinned generated client (google.golang.org/api v0.230.0)
//     and its discovery document contains no config resource at all
//     — projects.locations exposes only processes, runs, lineage
//     events and link search; there is no config get/list endpoint,
//     so the provider's per-location config singleton cannot be read
//     or enumerated through any verified endpoint.
//
// The generator is registered so the limitation is visible (and
// logged) instead of surfacing as an unsupported-service error.
func (g *DataLineageGenerator) InitResources() error {
	log.Println("dataLineage: the Data Lineage v1 API has no config resource or endpoint, so google_data_lineage_config cannot be enumerated; skipping")
	return nil
}
