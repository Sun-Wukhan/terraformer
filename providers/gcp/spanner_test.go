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
	"context"
	"testing"

	"google.golang.org/api/option"
	spannerv1 "google.golang.org/api/spanner/v1"
)

// TestSpannerGenerator feeds canned Spanner responses and verifies
// instance configs, instances and databases with their IAM trios,
// backup schedules and instance partitions.
func TestSpannerGenerator(t *testing.T) {
	proj := "projects/test-project"
	inst := proj + "/instances/inst1"
	db := inst + "/databases/db1"
	transport := &cannedTransport{routes: []cannedRoute{
		{":getIamPolicy", `{"bindings":[{"role":"roles/spanner.databaseUser","members":["user:a@example.com"]}]}`},
		{"/instances/inst1/databases/db1/backupSchedules", `{"backupSchedules":[{"name":"` + db + `/backupSchedules/sched1"}]}`},
		{"/instances/inst1/databases", `{"databases":[{"name":"` + db + `"}]}`},
		{"/instances/inst1/instancePartitions", `{"instancePartitions":[{"name":"` + inst + `/instancePartitions/part1"}]}`},
		{"/v1/projects/test-project/instances", `{"instances":[{"name":"` + inst + `"}]}`},
		{"/instanceConfigs", `{"instanceConfigs":[{"name":"` + proj + `/instanceConfigs/nam3"}]}`},
	}}
	ctx := context.Background()
	svc, err := spannerv1.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("spanner.NewService: %v", err)
	}
	g := &SpannerGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project"})
	if err := g.initResources(ctx, svc); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_spanner_instance_config":      {proj + "/instanceConfigs/nam3"},
		"google_spanner_instance":             {inst},
		"google_spanner_instance_iam_policy":  {inst},
		"google_spanner_instance_iam_binding": {inst},
		"google_spanner_instance_iam_member":  {inst},
		"google_spanner_database":             {db},
		"google_spanner_database_iam_policy":  {db},
		"google_spanner_database_iam_binding": {db},
		"google_spanner_database_iam_member":  {db},
		"google_spanner_backup_schedule":      {db + "/backupSchedules/sched1"},
		"google_spanner_instance_partition":   {inst + "/instancePartitions/part1"},
	})
}
