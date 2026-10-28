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

	"google.golang.org/api/cloudkms/v1"
	"google.golang.org/api/option"
)

// TestKmsDepthGenerator feeds canned Cloud KMS responses through the
// original-estate pass (global key ring and crypto key) and the depth
// pass, verifying the emitted types and IDs: EKM connections, key
// handles, key ring and crypto key IAM policy/binding/member trios,
// import jobs, crypto key versions in both locations, and the
// project/organization singleton configs read over raw REST.
func TestKmsDepthGenerator(t *testing.T) {
	ringG := "projects/test-project/locations/global/keyRings/ring-g"
	keyG := ringG + "/cryptoKeys/key-g"
	ring1 := "projects/test-project/locations/us/keyRings/ring-1"
	key1 := ring1 + "/cryptoKeys/key-1"
	iamPolicy := `{"bindings":[{"role":"roles/cloudkms.cryptoKeyEncrypterDecrypter","members":["user:a@example.com"]}]}`
	transport := &cannedTransport{routes: []cannedRoute{
		{"/locations/us/keyRings/ring-1/cryptoKeys/key-1/cryptoKeyVersions", `{"cryptoKeyVersions":[{"name":"` + key1 + `/cryptoKeyVersions/1"}]}`},
		{"/locations/global/keyRings/ring-g/cryptoKeys/key-g/cryptoKeyVersions", `{"cryptoKeyVersions":[{"name":"` + keyG + `/cryptoKeyVersions/3"}]}`},
		{":getIamPolicy", iamPolicy},
		{"/locations/us/keyRings/ring-1/importJobs", `{"importJobs":[{"name":"` + ring1 + `/importJobs/job-1"}]}`},
		{"/keyRings/ring-1/cryptoKeys", `{"cryptoKeys":[{"name":"` + key1 + `"}]}`},
		{"/keyRings/ring-g/cryptoKeys", `{"cryptoKeys":[{"name":"` + keyG + `"}]}`},
		{"/locations/us/keyRings", `{"keyRings":[{"name":"` + ring1 + `"}]}`},
		{"/locations/global/keyRings", `{"keyRings":[{"name":"` + ringG + `"}]}`},
		{"/locations/us/ekmConnections", `{"ekmConnections":[{"name":"projects/test-project/locations/us/ekmConnections/ekm-1"}]}`},
		{"/locations/us/keyHandles", `{"keyHandles":[{"name":"projects/test-project/locations/us/keyHandles/handle-1"}]}`},
		{"/projects/test-project/autokeyConfig", `{"name":"projects/test-project/autokeyConfig"}`},
		{"/projects/test-project/kajPolicyConfig", `{"name":"projects/test-project/kajPolicyConfig"}`},
		{"/organizations/6789/kajPolicyConfig", `{"name":"organizations/6789/kajPolicyConfig"}`},
		{":getAncestry", `{"ancestor":[{"resourceId":{"type":"organization","id":"6789"}}]}`},
		{"/v1/projects/test-project/locations", `{"locations":[{"name":"projects/test-project/locations/global"},{"name":"projects/test-project/locations/us"}]}`},
	}}

	ctx := context.Background()
	clientOpts := []option.ClientOption{
		option.WithHTTPClient(transport.client()),
		option.WithoutAuthentication(),
	}
	kmsService, err := cloudkms.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("cloudkms.NewService: %v", err)
	}
	crmService := crmTestService(t, transport)

	g := &KmsGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project"})
	g.Resources = g.createKmsRingResources(ctx, kmsService.Projects.Locations.KeyRings.List("projects/test-project/locations/global"), kmsService)
	g.initKmsDepthResources(ctx, kmsService, crmService, transport.client())

	expected := map[string][]string{
		"google_kms_key_ring":                       {"test-project/global/ring-g"},
		"google_kms_crypto_key":                     {keyG},
		"google_kms_ekm_connection":                 {"projects/test-project/locations/us/ekmConnections/ekm-1"},
		"google_kms_ekm_connection_iam_policy":      {"projects/test-project/locations/us/ekmConnections/ekm-1"},
		"google_kms_ekm_connection_iam_binding":     {"projects/test-project/locations/us/ekmConnections/ekm-1"},
		"google_kms_ekm_connection_iam_member":      {"projects/test-project/locations/us/ekmConnections/ekm-1"},
		"google_kms_key_handle":                     {"projects/test-project/locations/us/keyHandles/handle-1"},
		"google_kms_key_ring_iam_policy":            {ringG, ring1},
		"google_kms_key_ring_iam_binding":           {ringG, ring1},
		"google_kms_key_ring_iam_member":            {ringG, ring1},
		"google_kms_key_ring_import_job":            {ring1 + "/importJobs/job-1"},
		"google_kms_crypto_key_iam_policy":          {keyG, key1},
		"google_kms_crypto_key_iam_binding":         {keyG, key1},
		"google_kms_crypto_key_iam_member":          {keyG, key1},
		"google_kms_crypto_key_version":             {keyG + "/cryptoKeyVersions/3", key1 + "/cryptoKeyVersions/1"},
		"google_kms_project_autokey_config":         {"projects/test-project/autokeyConfig"},
		"google_kms_project_kaj_policy_config":      {"projects/test-project/kajPolicyConfig"},
		"google_kms_organization_kaj_policy_config": {"organizations/6789/kajPolicyConfig"},
	}
	assertResourceIDs(t, g.Resources, expected)
}
