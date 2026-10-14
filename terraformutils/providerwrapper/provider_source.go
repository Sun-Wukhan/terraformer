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
	"fmt"
	"strings"
)

// providerSources maps a terraformer provider name (the value returned by
// ProviderGenerator.GetName(), e.g. "aws", "azurerm", "metal") to the provider's
// address on the public Terraform/OpenTofu registries ("namespace/type").
//
// Every entry was verified against https://registry.terraform.io/v1/providers/<ns>/<type>
// (HTTP 200) in October 2026. OpenTofu's registry (registry.opentofu.org)
// mirrors the vast majority of these providers, so the same addresses work
// with `tofu init`.
var providerSources = map[string]string{
	"alicloud":      "aliyun/alicloud",
	"auth0":         "auth0/auth0",
	"aws":           "hashicorp/aws",
	"azuread":       "hashicorp/azuread",
	"azuredevops":   "microsoft/azuredevops",
	"azurerm":       "hashicorp/azurerm",
	"cloudflare":    "cloudflare/cloudflare",
	"commercetools": "labd/commercetools",
	"datadog":       "DataDog/datadog",
	"digitalocean":  "digitalocean/digitalocean",
	"fastly":        "fastly/fastly",
	"google":        "hashicorp/google",
	"google-beta":   "hashicorp/google-beta",
	"github":        "integrations/github",
	"gitlab":        "gitlabhq/gitlab",
	"gmailfilter":   "yamamoto-febc/gmailfilter",
	"grafana":       "grafana/grafana",
	"heroku":        "heroku/heroku",
	"honeycombio":   "honeycombio/honeycombio",
	"ibm":           "IBM-Cloud/ibm",
	"ionoscloud":    "ionos-cloud/ionoscloud",
	"keycloak":      "mrparkers/keycloak",
	"kubernetes":    "hashicorp/kubernetes",
	"launchdarkly":  "launchdarkly/launchdarkly",
	"linode":        "linode/linode",
	"logzio":        "logzio/logzio",
	"mackerel":      "mackerelio-labs/mackerel",
	"metal":         "equinix/metal",
	"mikrotik":      "ddelnano/mikrotik",
	"myrasec":       "Myra-Security-GmbH/myrasec",
	"newrelic":      "newrelic/newrelic",
	"ns1":           "ns1-terraform/ns1",
	"octopusdeploy": "OctopusDeployLabs/octopusdeploy",
	"okta":          "okta/okta",
	"opal":          "opalsecurity/opal",
	"openstack":     "terraform-provider-openstack/openstack",
	"opsgenie":      "opsgenie/opsgenie",
	"pagerduty":     "PagerDuty/pagerduty",
	"panos":         "PaloAltoNetworks/panos",
	"rabbitmq":      "cyrilgdn/rabbitmq",
	"tencentcloud":  "tencentcloudstack/tencentcloud",
	"vault":         "hashicorp/vault",
	"vultr":         "vultr/vultr",
	"xenorchestra":  "terra-farm/xenorchestra",
	"yandex":        "yandex-cloud/yandex",
}

// lookupProviderSource returns the registry source address ("namespace/type")
// and the terraform provider type (the last segment of the source) for a
// terraformer provider name. It fails loudly for unknown names: silently
// guessing a registry address would install the wrong provider binary.
func lookupProviderSource(name string) (source, tfType string, err error) {
	source, ok := providerSources[name]
	if !ok || source == "" {
		return "", "", fmt.Errorf("unknown terraformer provider %q: no registry source mapping in providerSources (terraformutils/providerwrapper/provider_source.go)", name)
	}
	parts := strings.Split(source, "/")
	return source, parts[len(parts)-1], nil
}
