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

type ChronicleGenerator struct {
	GCPService
}

// InitResources deliberately generates nothing.
//
// Every google_chronicle_* resource in the google provider hangs off a
// provisioned Google SecOps (Chronicle) instance:
// projects/{project}/locations/{location}/instances/{instance_id}/...,
// and the instance is the one parent terraformer cannot discover. The
// Chronicle API exposes no instance enumeration: the instance client of
// cloud.google.com/go/chronicle/apiv1 (and the underlying REST API)
// offers GetInstance only — there is no ListInstances method — and an
// instance ID is an opaque customer UUID assigned when the SecOps
// tenant is onboarded. It cannot be derived from the project, location
// or any other argument terraformer receives, and the google provider
// likewise requires it as user input (there is no
// google_chronicle_instance resource to import it through either).
//
// Until the API grows an instance list (or terraformer grows a way to
// accept a per-service parent argument), the whole family is out of
// reach of enumeration:
//
//	google_chronicle_big_query_export,
//	google_chronicle_case_close_definition,
//	google_chronicle_case_stage_definition,
//	google_chronicle_case_tag_definition,
//	google_chronicle_custom_list, google_chronicle_dashboard_chart,
//	google_chronicle_data_access_label, google_chronicle_data_access_scope,
//	google_chronicle_data_export, google_chronicle_data_table,
//	google_chronicle_data_table_row, google_chronicle_environment,
//	google_chronicle_environment_group, google_chronicle_feed,
//	google_chronicle_findings_refinement,
//	google_chronicle_findings_refinement_deployment,
//	google_chronicle_native_dashboard, google_chronicle_parser,
//	google_chronicle_parser_extension, google_chronicle_reference_list,
//	google_chronicle_retrohunt, google_chronicle_rule,
//	google_chronicle_rule_deployment, google_chronicle_soar_domain,
//	google_chronicle_soar_network, google_chronicle_watchlist.
//
// The generator is registered so the limitation is visible (and logged)
// instead of surfacing as an unsupported-service error.
func (g *ChronicleGenerator) InitResources() error {
	log.Println("chronicle: Google SecOps resources hang off a provisioned instance (projects/{project}/locations/{location}/instances/{instance_id}) and the Chronicle API exposes no instance list (GetInstance only), so no google_chronicle_* resources can be enumerated; skipping")
	return nil
}
