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

type DataflowGenerator struct {
	GCPService
}

// InitResources deliberately generates nothing.
//
// Deliberately not generated:
//   - google_dataflow_job and google_dataflow_flex_template_job:
//     both provider resources launch a Dataflow job; a job is a
//     transient execution record, not durable configuration — the
//     Dataflow API's jobs.list returns run history (including jobs
//     launched outside Terraform, by other pipelines and by Google
//     services), a completed job cannot be re-imported into the
//     launching resource's lifecycle (the provider resources manage
//     the act of launching and draining/cancelling, and neither
//     documents an import format for an existing job), and the job's
//     template/pipeline definition is not recoverable from the job
//     record. This matches the treatment of google_bigquery_job in
//     bigquery_depth.go.
//
// The generator is registered so the limitation is visible (and
// logged) instead of surfacing as an unsupported-service error.
func (g *DataflowGenerator) InitResources() error {
	log.Println("dataflow: google_dataflow_job and google_dataflow_flex_template_job launch transient jobs (jobs.list is run history) and neither resource documents import of an existing job, so no Dataflow resources can be enumerated; skipping")
	return nil
}
