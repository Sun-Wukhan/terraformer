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

package azure

import (
	"context"
	"log"

	armappcontainers "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers/v4"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// ContainerAppGenerator covers Microsoft.App: container apps, jobs and
// managed environments (plus their child resources).
type ContainerAppGenerator struct {
	AzureService
}

func (g *ContainerAppGenerator) listEnvironments() ([]*armappcontainers.ManagedEnvironment, error) {
	client, err := armappcontainers.NewManagedEnvironmentsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	var out []*armappcontainers.ManagedEnvironment
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return out, err
			}
			out = append(out, page.Value...)
		}
	} else {
		pager := client.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return out, err
			}
			out = append(out, page.Value...)
		}
	}
	return out, nil
}

func (g *ContainerAppGenerator) listApps() ([]*armappcontainers.ContainerApp, error) {
	client, err := armappcontainers.NewContainerAppsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	var out []*armappcontainers.ContainerApp
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return out, err
			}
			out = append(out, page.Value...)
		}
	} else {
		pager := client.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return out, err
			}
			out = append(out, page.Value...)
		}
	}
	return out, nil
}

func (g *ContainerAppGenerator) listJobs() ([]*armappcontainers.Job, error) {
	client, err := armappcontainers.NewJobsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	var out []*armappcontainers.Job
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return out, err
			}
			out = append(out, page.Value...)
		}
	} else {
		pager := client.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return out, err
			}
			out = append(out, page.Value...)
		}
	}
	return out, nil
}

func (g *ContainerAppGenerator) appendEnvironmentChildren(env *armappcontainers.ManagedEnvironment) error {
	ctx := context.Background()
	parsed, err := ParseAzureResourceID(*env.ID)
	if err != nil {
		return err
	}
	rg, envName := parsed.ResourceGroup, *env.Name

	// The environment custom domain suffix is a singleton configuration on
	// the environment itself: its Terraform ID is the environment ID and it
	// only exists when a DNS suffix has been configured.
	if env.Properties != nil && env.Properties.CustomDomainConfiguration != nil &&
		env.Properties.CustomDomainConfiguration.DNSSuffix != nil && *env.Properties.CustomDomainConfiguration.DNSSuffix != "" {
		g.AppendSimpleResource(*env.ID, *env.Name, "azurerm_container_app_environment_custom_domain")
	}

	certClient, err := armappcontainers.NewCertificatesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	pager := certClient.NewListPager(rg, envName, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_container_app_environment_certificate")
		}
	}

	managedCertClient, err := armappcontainers.NewManagedCertificatesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	mcPager := managedCertClient.NewListPager(rg, envName, nil)
	for mcPager.More() {
		page, err := mcPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_container_app_environment_managed_certificate")
		}
	}

	daprClient, err := armappcontainers.NewDaprComponentsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	daprPager := daprClient.NewListPager(rg, envName, nil)
	for daprPager.More() {
		page, err := daprPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_container_app_environment_dapr_component")
		}
	}

	storageClient, err := armappcontainers.NewManagedEnvironmentsStoragesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	storages, err := storageClient.List(ctx, rg, envName, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	for _, item := range storages.Value {
		g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_container_app_environment_storage")
	}
	return nil
}

// containerAppCustomDomainResources expands the custom domains embedded in
// a container app's ingress configuration. Their Terraform ID is the app ID
// plus /customDomainName/<domain> (see the provider's import docs).
func containerAppCustomDomainResources(app *armappcontainers.ContainerApp, providerName string) []terraformutils.Resource {
	var resources []terraformutils.Resource
	if app == nil || app.ID == nil || app.Properties == nil || app.Properties.Configuration == nil ||
		app.Properties.Configuration.Ingress == nil {
		return resources
	}
	for _, domain := range app.Properties.Configuration.Ingress.CustomDomains {
		if domain == nil || domain.Name == nil {
			continue
		}
		resources = append(resources, terraformutils.NewSimpleResource(
			*app.ID+"/customDomainName/"+*domain.Name,
			*domain.Name,
			"azurerm_container_app_custom_domain",
			providerName,
			[]string{}))
	}
	return resources
}

func (g *ContainerAppGenerator) InitResources() error {
	envs, err := g.listEnvironments()
	if err != nil {
		return err
	}
	for _, env := range envs {
		g.AppendSimpleResource(*env.ID, *env.Name, "azurerm_container_app_environment")
		if err := g.appendEnvironmentChildren(env); err != nil {
			return err
		}
	}

	apps, err := g.listApps()
	if err != nil {
		return err
	}
	for _, app := range apps {
		g.AppendSimpleResource(*app.ID, *app.Name, "azurerm_container_app")
		g.Resources = append(g.Resources, containerAppCustomDomainResources(app, g.ProviderName)...)
	}

	jobs, err := g.listJobs()
	if err != nil {
		return err
	}
	for _, job := range jobs {
		g.AppendSimpleResource(*job.ID, *job.Name, "azurerm_container_app_job")
	}
	return nil
}
