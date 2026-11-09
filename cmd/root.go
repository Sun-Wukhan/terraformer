// Copyright 2018 The Terraformer Authors.
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

package cmd

import (
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/spf13/cobra"
)

func NewCmdRoot() *cobra.Command {
	cmd := &cobra.Command{
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version,
	}
	cmd.AddCommand(newImportCmd())
	cmd.AddCommand(newWizardCmd())
	cmd.AddCommand(newPlanCmd())
	cmd.AddCommand(versionCmd)
	return cmd
}

func Execute() error {
	cmd := NewCmdRoot()
	return cmd.Execute()
}

func providerImporterSubcommands() []func(options ImportOptions) *cobra.Command {
	return []func(options ImportOptions) *cobra.Command{
		// Major Cloud
		newCmdGoogleImporter,
		newCmdAwsImporter,
		newCmdAzureImporter,
		newCmdAliCloudImporter,
		newCmdIbmImporter,
		// Cloud
		newCmdDigitalOceanImporter,
		newCmdEquinixMetalImporter,
		newCmdHerokuImporter,
		newCmdLaunchDarklyImporter,
		newCmdLinodeImporter,
		newCmdOpenStackImporter,
		newCmdTencentCloudImporter,
		newCmdVultrImporter,
		newCmdYandexImporter,
		newCmdIonosCloudImporter,
		// Infrastructure Software
		newCmdKubernetesImporter,
		newCmdOctopusDeployImporter,
		newCmdRabbitMQImporter,
		// Network
		newCmdMyrasecImporter,
		newCmdCloudflareImporter,
		newCmdFastlyImporter,
		newCmdNs1Importer,
		newCmdPanosImporter,
		// VCS
		newCmdAzureDevOpsImporter,
		newCmdAzureADImporter,
		newCmdGithubImporter,
		newCmdGitLabImporter,
		// Monitoring & System Management
		newCmdDatadogImporter,
		newCmdNewRelicImporter,
		newCmdMackerelImporter,
		newCmdGrafanaImporter,
		newCmdPagerDutyImporter,
		newCmdOpsgenieImporter,
		newCmdHoneycombioImporter,
		newCmdOpalImporter,
		// Community
		newCmdKeycloakImporter,
		newCmdLogzioImporter,
		newCmdCommercetoolsImporter,
		newCmdMikrotikImporter,
		newCmdXenorchestraImporter,
		newCmdGmailfilterImporter,
		newCmdVaultImporter,
		newCmdOktaImporter,
		newCmdAuth0Importer,
	}
}

func providerGenerators() map[string]func() terraformutils.ProviderGenerator {
	list := make(map[string]func() terraformutils.ProviderGenerator)
	for _, providerGen := range []func() terraformutils.ProviderGenerator{
		// Major Cloud
		newGoogleProvider,
		newAWSProvider,
		newAzureProvider,
		newAliCloudProvider,
		newIbmProvider,
		// Cloud
		newDigitalOceanProvider,
		newEquinixMetalProvider,
		newFastlyProvider,
		newHerokuProvider,
		newLaunchDarklyProvider,
		newLinodeProvider,
		newNs1Provider,
		newOpenStackProvider,
		newTencentCloudProvider,
		newVultrProvider,
		// Infrastructure Software
		newKubernetesProvider,
		newOctopusDeployProvider,
		newRabbitMQProvider,
		// Network
		newMyrasecProvider,
		newCloudflareProvider,
		// VCS
		newAzureDevOpsProvider,
		newAzureADProvider,
		newGitHubProvider,
		newGitLabProvider,
		// Monitoring & System Management
		newDataDogProvider,
		newNewRelicProvider,
		newPagerDutyProvider,
		newHoneycombioProvider,
		newOpalProvider,
		// Community
		newKeycloakProvider,
		newLogzioProvider,
		newCommercetoolsProvider,
		newMikrotikProvider,
		newXenorchestraProvider,
		newGmailfilterProvider,
		newVaultProvider,
		newOktaProvider,
		newAuth0Provider,
	} {
		list[providerGen().GetName()] = providerGen
	}
	return list
}

// wizardProviders maps an `import` subcommand name to its generator, for
// the import wizard in wizard.go. Keys are subcommand names (note: the
// equinix metal subcommand is "metal"), not Provider.GetName() values,
// because generated commands must match the CLI exactly.
//
// This map deliberately lives in root.go, one entry per line, like
// providerGenerators above: build/multi-build builds each per-provider
// binary by moving the other provider_cmd_*.go files away and commenting
// out the matching new<Provider>Provider lines in this file, so the
// wizard's registry shrinks to the providers actually compiled into a
// per-provider binary instead of breaking its build.
func wizardProviders() map[string]func() terraformutils.ProviderGenerator {
	return map[string]func() terraformutils.ProviderGenerator{
		"alicloud":      newAliCloudProvider,
		"auth0":         newAuth0Provider,
		"aws":           newAWSProvider,
		"azure":         newAzureProvider,
		"azuread":       newAzureADProvider,
		"azuredevops":   newAzureDevOpsProvider,
		"cloudflare":    newCloudflareProvider,
		"commercetools": newCommercetoolsProvider,
		"datadog":       newDataDogProvider,
		"digitalocean":  newDigitalOceanProvider,
		"fastly":        newFastlyProvider,
		"github":        newGitHubProvider,
		"gitlab":        newGitLabProvider,
		"gmailfilter":   newGmailfilterProvider,
		"google":        newGoogleProvider,
		"grafana":       newGrafanaProvider,
		"heroku":        newHerokuProvider,
		"honeycombio":   newHoneycombioProvider,
		"ibm":           newIbmProvider,
		"ionoscloud":    newIonosCloudProvider,
		"keycloak":      newKeycloakProvider,
		"kubernetes":    newKubernetesProvider,
		"launchdarkly":  newLaunchDarklyProvider,
		"linode":        newLinodeProvider,
		"logzio":        newLogzioProvider,
		"mackerel":      newMackerelProvider,
		"metal":         newEquinixMetalProvider,
		"mikrotik":      newMikrotikProvider,
		"myrasec":       newMyrasecProvider,
		"newrelic":      newNewRelicProvider,
		"ns1":           newNs1Provider,
		"octopusdeploy": newOctopusDeployProvider,
		"okta":          newOktaProvider,
		"opal":          newOpalProvider,
		"openstack":     newOpenStackProvider,
		"opsgenie":      newOpsgenieProvider,
		"pagerduty":     newPagerDutyProvider,
		"panos":         newPanosProvider,
		"rabbitmq":      newRabbitMQProvider,
		"tencentcloud":  newTencentCloudProvider,
		"vault":         newVaultProvider,
		"vultr":         newVultrProvider,
		"xenorchestra":  newXenorchestraProvider,
		"yandex":        newYandexProvider,
	}
}
