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

	cloudresourcemanagerv1 "google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
	storagev1 "google.golang.org/api/storage/v1"
	storagebatchoperationsv1 "google.golang.org/api/storagebatchoperations/v1"
)

// TestStorageDepthGenerator feeds canned Cloud Storage, Batch
// Operations, Resource Manager, Storage Control and Storage FTP
// responses through the GCS depth enumeration and verifies the
// emitted types and composite IDs: anywhere cache, bucket and
// default-object ACL entries, folder, managed folder + its combined
// IAM resource, bucket object + object ACL + ACL entry, HMAC key,
// batch job, the three intelligence config scopes, the project
// feature config, and the FTP server + user.
func TestStorageDepthGenerator(t *testing.T) {
	proj := "projects/test-project"
	transport := &cannedTransport{routes: []cannedRoute{
		// Object ACLs before the objects list (its path is a prefix).
		{"/b/bkt1/o/obj1/acl", `{"items":[{"entity":"user-a@example.com","role":"READER"}]}`},
		{"/b/bkt1/o", `{"items":[{"name":"obj1","bucket":"bkt1"}]}`},
		{"/b/bkt1/anywhereCaches", `{"items":[{"anywhereCacheId":"cache1","bucket":"bkt1"}]}`},
		{"/b/bkt1/acl", `{"items":[{"entity":"user-a@example.com","role":"READER"}]}`},
		{"/b/bkt1/defaultObjectAcl", `{"items":[{"entity":"user-a@example.com","role":"READER"}]}`},
		{"/b/bkt1/folders", `{"items":[{"name":"fld1/","bucket":"bkt1"}]}`},
		{"/b/bkt1/managedFolders", `{"items":[{"name":"mf1/","bucket":"bkt1"}]}`},
		// Bucket list last among the /storage/v1/b routes: its path
		// is a prefix of every bucket-scoped path above.
		{"/storage/v1/b", `{"items":[{"name":"bkt1"}]}`},
		{"/projects/test-project/hmacKeys", `{"items":[{"accessId":"GOOG123"}]}`},
		{"/locations/global/jobs", `{"jobs":[{"name":"` + proj + `/locations/global/jobs/job1"}]}`},
		{":getAncestry", `{"ancestor":[{"resourceId":{"type":"folder","id":"12345"}},{"resourceId":{"type":"organization","id":"6789"}}]}`},
		{"/locations/global/featureConfigs", `{"featureConfigs":[{"name":"` + proj + `/locations/global/featureConfigs/fc1"}]}`},
		// FTP users before the servers list (path prefix again).
		{"/servers/srv1/users", `{"users":[{"name":"` + proj + `/locations/us-central1/servers/srv1/users/u1"}]}`},
		{"/locations/us-central1/servers", `{"servers":[{"name":"` + proj + `/locations/us-central1/servers/srv1"}]}`},
	}}

	ctx := context.Background()
	clientOpts := []option.ClientOption{
		option.WithHTTPClient(transport.client()),
		option.WithoutAuthentication(),
	}
	storageService, err := storagev1.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("storage.NewService: %v", err)
	}
	batchService, err := storagebatchoperationsv1.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("storagebatchoperations.NewService: %v", err)
	}
	crmService, err := cloudresourcemanagerv1.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("cloudresourcemanager.NewService: %v", err)
	}

	g := &GcsGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	g.initStorageDepthResources(ctx, storageService, batchService, crmService, transport.client())

	expected := map[string][]string{
		"google_storage_anywhere_cache":                           {"bkt1/cache1"},
		"google_storage_bucket_access_control":                    {"bkt1/user-a@example.com"},
		"google_storage_default_object_access_control":            {"bkt1/user-a@example.com"},
		"google_storage_folder":                                   {"bkt1/fld1/"},
		"google_storage_managed_folder":                           {"bkt1/mf1/"},
		"google_storage_managed_folder_iam":                       {"bkt1/mf1/"},
		"google_storage_bucket_object":                            {"bkt1/obj1"},
		"google_storage_object_acl":                               {"bkt1/obj1"},
		"google_storage_object_access_control":                    {"bkt1/obj1/user-a@example.com"},
		"google_storage_hmac_key":                                 {proj + "/hmacKeys/GOOG123"},
		"google_storage_batch_operations_job":                     {proj + "/locations/global/jobs/job1"},
		"google_storage_control_project_intelligence_config":      {proj + "/locations/global/intelligenceConfig"},
		"google_storage_control_project_feature_config":           {proj + "/locations/global/featureConfigs/fc1"},
		"google_storage_control_folder_intelligence_config":       {"folders/12345/locations/global/intelligenceConfig"},
		"google_storage_control_organization_intelligence_config": {"organizations/6789/locations/global/intelligenceConfig"},
		"google_storage_ftp_server":                               {proj + "/locations/us-central1/servers/srv1"},
		"google_storage_ftp_user":                                 {proj + "/locations/us-central1/servers/srv1/users/u1"},
	}
	assertResourceIDs(t, g.Resources, expected)
}
