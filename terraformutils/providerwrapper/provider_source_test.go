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

package providerwrapper

import (
	"strings"
	"testing"
)

// everyProviderName lists all terraformer provider names that must resolve to
// a registry source address.
var everyProviderName = []string{
	"alicloud", "auth0", "aws", "azuread", "azuredevops", "azurerm",
	"cloudflare", "commercetools", "datadog", "digitalocean", "fastly",
	"google", "google-beta", "github", "gitlab", "gmailfilter", "grafana",
	"heroku", "honeycombio", "ibm", "ionoscloud", "keycloak", "kubernetes",
	"launchdarkly", "linode", "logzio", "mackerel", "metal", "mikrotik",
	"myrasec", "newrelic", "ns1", "octopusdeploy", "okta", "opal",
	"openstack", "opsgenie", "pagerduty", "panos", "rabbitmq",
	"tencentcloud", "vault", "vultr", "xenorchestra", "yandex",
}

func TestProviderSourceTableComplete(t *testing.T) {
	for _, name := range everyProviderName {
		t.Run(name, func(t *testing.T) {
			source, tfType, err := lookupProviderSource(name)
			if err != nil {
				t.Fatalf("lookup failed: %v", err)
			}
			if source == "" {
				t.Fatalf("empty source address")
			}
			if tfType == "" {
				t.Fatalf("empty provider type")
			}
			parts := strings.Split(source, "/")
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				t.Fatalf("source %q is not namespace/type", source)
			}
			if parts[1] != tfType {
				t.Fatalf("provider type %q != last segment of source %q", tfType, source)
			}
		})
	}
}

func TestProviderSourceTableKnownValues(t *testing.T) {
	cases := map[string]string{
		"aws":          "hashicorp/aws",
		"azurerm":      "hashicorp/azurerm",
		"google":       "hashicorp/google",
		"google-beta":  "hashicorp/google-beta",
		"metal":        "equinix/metal",
		"auth0":        "auth0/auth0",
		"opal":         "opalsecurity/opal",
		"tencentcloud": "tencentcloudstack/tencentcloud",
		"kubernetes":   "hashicorp/kubernetes",
		"vault":        "hashicorp/vault",
	}
	for name, want := range cases {
		if got, _, err := lookupProviderSource(name); err != nil || got != want {
			t.Fatalf("%s: got %q, err %v; want %q", name, got, err, want)
		}
	}
}

func TestLookupProviderSourceUnknown(t *testing.T) {
	if _, _, err := lookupProviderSource("does-not-exist"); err == nil {
		t.Fatalf("expected loud failure for unknown provider name")
	}
}
