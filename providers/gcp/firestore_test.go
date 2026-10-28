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

	firestorev1 "google.golang.org/api/firestore/v1"
	"google.golang.org/api/option"
)

// TestFirestoreGenerator feeds canned Firestore Admin responses and
// verifies databases, backup schedules, user creds, fields and
// indexes are emitted with full resource names as IDs.
func TestFirestoreGenerator(t *testing.T) {
	db := "projects/test-project/databases/(default)"
	group := db + "/collectionGroups/users"
	transport := &cannedTransport{routes: []cannedRoute{
		{"/backupSchedules", `{"backupSchedules":[{"name":"` + db + `/backupSchedules/sched1"}]}`},
		{"/userCreds", `{"userCreds":[{"name":"` + db + `/userCreds/cred1"}]}`},
		{"/fields", `{"fields":[{"name":"` + group + `/fields/email"}]}`},
		{"/indexes", `{"indexes":[{"name":"` + group + `/indexes/idx1"}]}`},
		{"/v1/projects/test-project/databases", `{"databases":[{"name":"` + db + `"}]}`},
	}}
	ctx := context.Background()
	svc, err := firestorev1.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("firestore.NewService: %v", err)
	}
	g := &FirestoreGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project"})
	if err := g.initResources(ctx, svc); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_firestore_database":        {db},
		"google_firestore_backup_schedule": {db + "/backupSchedules/sched1"},
		"google_firestore_user_creds":      {db + "/userCreds/cred1"},
		"google_firestore_field":           {group + "/fields/email"},
		"google_firestore_index":           {group + "/indexes/idx1"},
	})
}
