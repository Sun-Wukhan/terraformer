// Copyright 2026 The Terraformer Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"strings"
	"testing"
)

func TestBackendValidate(t *testing.T) {
	cases := []struct {
		name    string
		backend backendConfig
		wantErr string
	}{
		{"local default", backendConfig{}, ""},
		{"local explicit", backendConfig{Kind: "local"}, ""},
		{"hcp needs org", backendConfig{Kind: "hcp"}, "--backend-org"},
		{"hcp ok", backendConfig{Kind: "hcp", Organization: "my-org"}, ""},
		{"gcs needs bucket", backendConfig{Kind: "gcs"}, "--backend-bucket"},
		{"gcs ok", backendConfig{Kind: "gcs", GCSBucket: "b"}, ""},
		{"s3 needs region", backendConfig{Kind: "s3", S3Bucket: "b"}, "--backend-region"},
		{"s3 ok", backendConfig{Kind: "s3", S3Bucket: "b", S3Region: "us-east-1"}, ""},
		{"http needs address", backendConfig{Kind: "http"}, "--backend-address"},
		{"http ok", backendConfig{Kind: "http", HTTPAddress: "https://state.example.com"}, ""},
		{"unknown", backendConfig{Kind: "consul"}, "unknown backend"},
		{"shell expansion rejected", backendConfig{Kind: "gcs", GCSBucket: "b$(touch /tmp/pwned)"}, "must not contain"},
		{"backtick rejected", backendConfig{Kind: "hcp", Organization: "org`id`"}, "must not contain"},
		{"newline rejected", backendConfig{Kind: "s3", S3Bucket: "b", S3Region: "us-east-1\nrm -rf ~"}, "must not contain"},
	}
	for _, tc := range cases {
		err := tc.backend.Validate()
		if tc.wantErr == "" && err != nil {
			t.Errorf("%s: unexpected error: %v", tc.name, err)
		}
		if tc.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tc.wantErr)) {
			t.Errorf("%s: error = %v, want substring %q", tc.name, err, tc.wantErr)
		}
	}
}

func TestBackendTFPerDirectory(t *testing.T) {
	dir := "google/my-project/gcs/global"

	hcp := backendConfig{Kind: "hcp", Organization: "my-org"}.TF(dir)
	for _, want := range []string{"cloud {", `organization = "my-org"`, `name = "terraformer-google-my-project-gcs-global"`} {
		if !strings.Contains(hcp, want) {
			t.Errorf("hcp TF missing %q:\n%s", want, hcp)
		}
	}

	gcs := backendConfig{Kind: "gcs", GCSBucket: "state-bucket"}.TF(dir)
	for _, want := range []string{`backend "gcs"`, `bucket = "state-bucket"`, `prefix = "terraformer/google/my-project/gcs/global"`} {
		if !strings.Contains(gcs, want) {
			t.Errorf("gcs TF missing %q:\n%s", want, gcs)
		}
	}

	s3 := backendConfig{Kind: "s3", S3Bucket: "state-bucket", S3Region: "us-east-1", S3DynamoDBTable: "locks"}.TF(dir)
	for _, want := range []string{`backend "s3"`, `key     = "terraformer/google/my-project/gcs/global/terraform.tfstate"`, "encrypt = true", `dynamodb_table = "locks"`} {
		if !strings.Contains(s3, want) {
			t.Errorf("s3 TF missing %q:\n%s", want, s3)
		}
	}

	http := backendConfig{Kind: "http", HTTPAddress: "https://state.example.com/tf", HTTPLockAddress: "https://state.example.com/tf/lock"}.TF(dir)
	for _, want := range []string{`backend "http"`, `address = "https://state.example.com/tf/google/my-project/gcs/global"`, `lock_address   = "https://state.example.com/tf/lock/google/my-project/gcs/global"`} {
		if !strings.Contains(http, want) {
			t.Errorf("http TF missing %q:\n%s", want, http)
		}
	}

	if got := (backendConfig{Kind: "local"}).TF(dir); got != "" {
		t.Errorf("local backend must render no TF, got %q", got)
	}
}

func TestBackendScriptLines(t *testing.T) {
	cfg := wizardConfig{
		Provider:   "google",
		PathOutput: "generated",
		Backend:    backendConfig{Kind: "gcs", GCSBucket: "state-bucket"},
	}
	script := strings.Join(cfg.BackendScriptLines(), "\n")
	for _, want := range []string{"OUTROOT=generated", "terraform.tfstate", `prefix = "terraformer/$key"`, "terraform init -migrate-state"} {
		if !strings.Contains(script, want) {
			t.Errorf("script missing %q:\n%s", want, script)
		}
	}

	hcpCfg := wizardConfig{
		Provider:   "google",
		PathOutput: "generated",
		Backend:    backendConfig{Kind: "hcp", Organization: "my-org"},
	}
	hcpScript := strings.Join(hcpCfg.BackendScriptLines(), "\n")
	if !strings.Contains(hcpScript, "wkey=") || !strings.Contains(hcpScript, `name = "terraformer-$wkey"`) {
		t.Errorf("hcp script must derive a workspace-safe key:\n%s", hcpScript)
	}

	local := wizardConfig{Provider: "google", Backend: backendConfig{Kind: "local"}}
	if lines := local.BackendScriptLines(); lines != nil {
		t.Errorf("local backend must add no script lines, got %v", lines)
	}
}

func TestWizardValidateChecksBackend(t *testing.T) {
	cfg := wizardConfig{
		Provider:  "google",
		Resources: []string{"networks"},
		Projects:  []string{"my-project"},
		Backend:   backendConfig{Kind: "hcp"}, // missing organization
	}
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "--backend-org") {
		t.Errorf("Validate must surface the backend error, got %v", err)
	}
	cfg.Backend.Organization = "my-org"
	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate with complete backend: %v", err)
	}
}
