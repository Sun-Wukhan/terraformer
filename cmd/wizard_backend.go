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
	"fmt"
	"strconv"
	"strings"
)

// backendConfig is where the imported state should live afterwards.
// Terraformer always imports into local state files first (one per
// generated service directory); a non-local backend here makes the
// wizard emit the extra step that writes a backend.tf into every
// generated directory and migrates its state with
// `terraform init -migrate-state`. The wizard never handles backend
// credentials: each backend reads its own from the environment.
type backendConfig struct {
	Kind string // local (default), hcp, gcs, s3, http

	// hcp (HCP Terraform / Terraform Cloud)
	Organization    string
	WorkspacePrefix string

	// gcs
	GCSBucket string
	GCSPrefix string

	// s3
	S3Bucket        string
	S3Region        string
	S3KeyPrefix     string
	S3DynamoDBTable string

	// http
	HTTPAddress      string
	HTTPLockAddress  string
	HTTPUnlockAddres string
	HTTPUsername     string
}

func backendKinds() []string {
	return []string{"local", "hcp", "gcs", "s3", "http"}
}

func (b backendConfig) IsLocal() bool {
	return b.Kind == "" || b.Kind == "local"
}

// backendFields lists every user-supplied string that ends up inside
// the generated migration script (in an unquoted heredoc, where the
// per-directory $key must expand at runtime). A value containing $, `
// or a newline could break out of its HCL string and execute when the
// user runs the script, so reject those characters up front.
func (b backendConfig) backendFields() []string {
	return []string{b.Organization, b.WorkspacePrefix, b.GCSBucket, b.GCSPrefix,
		b.S3Bucket, b.S3Region, b.S3KeyPrefix, b.S3DynamoDBTable,
		b.HTTPAddress, b.HTTPLockAddress, b.HTTPUnlockAddres, b.HTTPUsername}
}

func (b backendConfig) Validate() error {
	for _, f := range b.backendFields() {
		if strings.ContainsAny(f, "$`\n\r") {
			return fmt.Errorf("backend values must not contain $, backticks or newlines (got %q)", f)
		}
	}
	switch b.Kind {
	case "", "local":
		return nil
	case "hcp":
		if b.Organization == "" {
			return fmt.Errorf("backend hcp needs --backend-org (your HCP Terraform organization)")
		}
	case "gcs":
		if b.GCSBucket == "" {
			return fmt.Errorf("backend gcs needs --backend-bucket (the GCS bucket that will hold state)")
		}
	case "s3":
		if b.S3Bucket == "" {
			return fmt.Errorf("backend s3 needs --backend-bucket (the S3 bucket that will hold state)")
		}
		if b.S3Region == "" {
			return fmt.Errorf("backend s3 needs --backend-region (the bucket's AWS region)")
		}
	case "http":
		if b.HTTPAddress == "" {
			return fmt.Errorf("backend http needs --backend-address (the state endpoint URL)")
		}
	default:
		return fmt.Errorf("unknown backend %q (valid: %s)", b.Kind, strings.Join(backendKinds(), ", "))
	}
	return nil
}

// sanitizeKey turns a generated directory path into a suffix that is
// safe inside a workspace name, object key or URL path segment.
// A leading "$" marks a shell variable placeholder rendered for the
// wizard's script output; those pass through untouched.
func sanitizeKey(dirKey string, sep string) string {
	if strings.HasPrefix(dirKey, "$") {
		return dirKey
	}
	key := strings.Trim(strings.ReplaceAll(dirKey, "\\", "/"), "/")
	if sep == "-" {
		var sb strings.Builder
		for _, r := range strings.ToLower(key) {
			if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' {
				sb.WriteRune(r)
			} else {
				sb.WriteRune('-')
			}
		}
		key = strings.Trim(sb.String(), "-")
	}
	return key
}

// TF renders the backend.tf content for one generated service
// directory. dirKey identifies the directory (its path under the
// output root, e.g. "google/my-project/gcs/global") so every directory
// gets its own workspace / prefix / key and states never collide.
func (b backendConfig) TF(dirKey string) string {
	switch b.Kind {
	case "hcp":
		prefix := b.WorkspacePrefix
		if prefix == "" {
			prefix = "terraformer"
		}
		name := prefix + "-" + sanitizeKey(dirKey, "-")
		return fmt.Sprintf(`terraform {
  cloud {
    organization = %s

    workspaces {
      name = %s
    }
  }
}
`, strconv.Quote(b.Organization), strconv.Quote(name))
	case "gcs":
		prefix := b.GCSPrefix
		if prefix == "" {
			prefix = "terraformer"
		}
		return fmt.Sprintf(`terraform {
  backend "gcs" {
    bucket = %s
    prefix = %s
  }
}
`, strconv.Quote(b.GCSBucket), strconv.Quote(prefix+"/"+sanitizeKey(dirKey, "/")))
	case "s3":
		prefix := b.S3KeyPrefix
		if prefix == "" {
			prefix = "terraformer"
		}
		var sb strings.Builder
		sb.WriteString("terraform {\n  backend \"s3\" {\n")
		fmt.Fprintf(&sb, "    bucket  = %s\n", strconv.Quote(b.S3Bucket))
		fmt.Fprintf(&sb, "    key     = %s\n", strconv.Quote(prefix+"/"+sanitizeKey(dirKey, "/")+"/terraform.tfstate"))
		fmt.Fprintf(&sb, "    region  = %s\n", strconv.Quote(b.S3Region))
		sb.WriteString("    encrypt = true\n")
		if b.S3DynamoDBTable != "" {
			fmt.Fprintf(&sb, "    dynamodb_table = %s\n", strconv.Quote(b.S3DynamoDBTable))
		}
		sb.WriteString("  }\n}\n")
		return sb.String()
	case "http":
		addr := strings.TrimSuffix(b.HTTPAddress, "/") + "/" + sanitizeKey(dirKey, "/")
		var sb strings.Builder
		sb.WriteString("terraform {\n  backend \"http\" {\n")
		fmt.Fprintf(&sb, "    address = %s\n", strconv.Quote(addr))
		if b.HTTPLockAddress != "" {
			fmt.Fprintf(&sb, "    lock_address   = %s\n", strconv.Quote(strings.TrimSuffix(b.HTTPLockAddress, "/")+"/"+sanitizeKey(dirKey, "/")))
		}
		if b.HTTPUnlockAddres != "" {
			fmt.Fprintf(&sb, "    unlock_address = %s\n", strconv.Quote(strings.TrimSuffix(b.HTTPUnlockAddres, "/")+"/"+sanitizeKey(dirKey, "/")))
		}
		if b.HTTPUsername != "" {
			fmt.Fprintf(&sb, "    username = %s\n", strconv.Quote(b.HTTPUsername))
		}
		sb.WriteString("  }\n}\n")
		return sb.String()
	}
	return ""
}

// ScriptLines returns the shell lines that configure the chosen
// backend in every generated service directory and migrate each local
// state. They run AFTER the import command. Local backend: no lines.
func (c wizardConfig) BackendScriptLines() []string {
	b := c.Backend
	if b.IsLocal() {
		return nil
	}
	out := c.PathOutput
	if out == "" {
		out = DefaultPathOutput
	}
	lines := []string{
		"",
		fmt.Sprintf("# Move state to the %s backend (runs after the import):", b.Kind),
		fmt.Sprintf("OUTROOT=%s", shellQuote(out)),
		`find "$OUTROOT/` + c.Provider + `" -name terraform.tfstate | while read -r state; do`,
		`  d=$(dirname "$state")`,
		`  key=$(printf '%s' "$d" | sed "s#^$OUTROOT/##")`,
	}
	placeholder := "$key"
	if b.Kind == "hcp" {
		// Workspace names cannot contain slashes; derive a safe form.
		lines = append(lines, `  wkey=$(printf '%s' "$key" | sed 's#[^a-zA-Z0-9-]#-#g')`)
		placeholder = "$wkey"
	}
	lines = append(lines, `  cat > "$d/backend.tf" <<EOF`)
	// The heredoc body must use the key variables at script runtime,
	// so render the backend block with a placeholder key.
	body := b.TF(placeholder)
	lines = append(lines, strings.Split(strings.TrimSuffix(body, "\n"), "\n")...)
	lines = append(lines,
		`EOF`,
		`  (cd "$d" && terraform init -migrate-state -input=false)`,
		`done`,
	)
	return lines
}

// BackendNotes are human-facing cautions for the chosen backend,
// printed with the wizard summary. No secret values, ever.
func (c wizardConfig) BackendNotes() []string {
	switch c.Backend.Kind {
	case "hcp":
		return []string{"State will live in HCP Terraform: one workspace per generated service directory.",
			"Log in first with: terraform login (the wizard never asks for your token).",
			"You can browse every workspace and its state at https://app.terraform.io"}
	case "gcs":
		return []string{"State will live in the GCS bucket you named, one prefix per service directory.",
			"Your GOOGLE_APPLICATION_CREDENTIALS identity needs write access to that bucket."}
	case "s3":
		return []string{"State will live in the S3 bucket you named, one key per service directory, encrypted at rest.",
			"Without a DynamoDB table there is no state locking — set --backend-dynamodb-table for teams."}
	case "http":
		return []string{"State will be served by your HTTP backend, one address per service directory.",
			"Set TF_HTTP_PASSWORD in your environment if the endpoint needs a password — never in the config.",
			"Without lock/unlock addresses the backend cannot lock state."}
	}
	return nil
}
