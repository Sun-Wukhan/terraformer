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

import (
	"log"
)

// AgenticApplicationsGenerator is a deliberate no-op, following the
// ChronicleGenerator precedent.
//
// Deliberately not generated:
//   - google_agentic_applications_analyst_agent_persona: the
//     resource exists only in the google-beta provider and its API
//     (agenticapplications.googleapis.com) has no publicly
//     documented discovery document or list endpoint as of this
//     sweep — no enumeration path could be verified against any
//     published API surface, so personas cannot be listed from the
//     scopes available to terraformer. The generator is registered
//     so the service key exists and logs the limitation instead of
//     failing.
type AgenticApplicationsGenerator struct {
	GCPService
}

func (g *AgenticApplicationsGenerator) InitResources() error {
	log.Println("agenticApplications: google_agentic_applications_analyst_agent_persona is not generated: agenticapplications.googleapis.com exposes no publicly documented list endpoint or discovery document to enumerate personas")
	return nil
}
