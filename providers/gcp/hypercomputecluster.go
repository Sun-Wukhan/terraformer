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

type HypercomputeClusterGenerator struct {
	GCPService
}

// InitResources deliberately generates nothing.
//
// Deliberately not generated:
//   - google_hypercomputecluster_cluster: the Hypercompute Cluster
//     (Cluster Director) API is a pre-GA surface with no generated
//     client in the pinned google.golang.org/api module (v0.230.0)
//     and no published discovery document exposing a stable v1 list
//     endpoint that could be verified for raw REST enumeration, so
//     neither the collection endpoint nor the provider's expected
//     ID form can be confirmed. Emitting guessed IDs is worse than
//     skipping; revisit once the API publishes a discovery document
//     or a generated client.
//
// The generator is registered so the limitation is visible (and
// logged) instead of surfacing as an unsupported-service error.
func (g *HypercomputeClusterGenerator) InitResources() error {
	log.Println("hypercomputecluster: the Cluster Director API has no generated client in the pinned google.golang.org/api module and no verifiable published list endpoint, so google_hypercomputecluster_cluster cannot be enumerated; skipping")
	return nil
}
