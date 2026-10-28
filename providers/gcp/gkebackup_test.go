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

	"google.golang.org/api/compute/v1"
	gkebackupv1 "google.golang.org/api/gkebackup/v1"
	"google.golang.org/api/option"
)

// TestGkeBackupGenerator feeds canned Backup for GKE responses and
// verifies channels, plans and their IAM trios.
func TestGkeBackupGenerator(t *testing.T) {
	loc := "projects/test-project/locations/us-central1"
	bp := loc + "/backupPlans/bp1"
	rp := loc + "/restorePlans/rp1"
	policy := `{"bindings":[{"role":"roles/gkebackup.viewer","members":["user:a@example.com"]}]}`
	transport := &cannedTransport{routes: []cannedRoute{
		{":getIamPolicy", policy},
		{"/backupChannels", `{"backupChannels":[{"name":"` + loc + `/backupChannels/bc1"}]}`},
		{"/restoreChannels", `{"restoreChannels":[{"name":"` + loc + `/restoreChannels/rc1"}]}`},
		{"/backupPlans", `{"backupPlans":[{"name":"` + bp + `"}]}`},
		{"/restorePlans", `{"restorePlans":[{"name":"` + rp + `"}]}`},
	}}
	ctx := context.Background()
	clientOpts := []option.ClientOption{option.WithHTTPClient(transport.client()), option.WithoutAuthentication()}
	backupService, err := gkebackupv1.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("gkebackup.NewService: %v", err)
	}
	g := &GkeBackupGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project", "region": compute.Region{Name: "us-central1"}})
	if err := g.initResources(ctx, backupService); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_gke_backup_backup_channel":           {loc + "/backupChannels/bc1"},
		"google_gke_backup_backup_plan":              {bp},
		"google_gke_backup_backup_plan_iam_policy":   {bp},
		"google_gke_backup_backup_plan_iam_binding":  {bp},
		"google_gke_backup_backup_plan_iam_member":   {bp},
		"google_gke_backup_restore_channel":          {loc + "/restoreChannels/rc1"},
		"google_gke_backup_restore_plan":             {rp},
		"google_gke_backup_restore_plan_iam_policy":  {rp},
		"google_gke_backup_restore_plan_iam_binding": {rp},
		"google_gke_backup_restore_plan_iam_member":  {rp},
	})
}
