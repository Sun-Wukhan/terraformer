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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/automation/armautomation"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// AutomationGenerator covers Microsoft.Automation automationAccounts and
// their child resources.
//
// Skipped in this service (see ledger):
//   - azurerm_automation_variable_bool / _datetime / _int / _object /
//     _string: all five share the single ARM variables/{name} resource;
//     armautomation VariableProperties exposes only Value *string with no
//     type discriminator, so the five Terraform types cannot be
//     distinguished from the ListByAutomationAccount response.
//   - azurerm_automation_powershell72_module: the provider ID uses the
//     powerShell72Modules segment, but armautomation v1.0.0 has no
//     PowerShell72Modules client / List operation.
type AutomationGenerator struct {
	AzureService
}

// classifyAutomationConnection maps an Automation Connection's connection
// type name (ConnectionProperties.ConnectionType.Name) to the Terraform
// type the azurerm provider uses for it. Connections whose type is not one
// of the specialised types are the generic azurerm_automation_connection.
func classifyAutomationConnection(connectionTypeName string) string {
	switch connectionTypeName {
	case "Azure":
		return "azurerm_automation_connection_certificate"
	case "AzureClassicCertificate":
		return "azurerm_automation_connection_classic_certificate"
	case "AzureServicePrincipal":
		return "azurerm_automation_connection_service_principal"
	default:
		return "azurerm_automation_connection"
	}
}

func (g *AutomationGenerator) listAccounts() ([]*armautomation.Account, error) {
	client, err := armautomation.NewAccountClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	var out []*armautomation.Account
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
		pager := client.NewListPager(nil)
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

func appendAutomationResource(resources []terraformutils.Resource, id, name *string, tfType, provider string) []terraformutils.Resource {
	if id == nil || name == nil {
		return resources
	}
	return append(resources, terraformutils.NewSimpleResource(*id, *name, tfType, provider, []string{}))
}

// listAccountChildren enumerates every child collection of one automation
// account. Errors are logged and collected into the returned resources
// best-effort: a failure in one collection does not drop the others.
func (g *AutomationGenerator) listAccountChildren(rg, account string) []terraformutils.Resource {
	var resources []terraformutils.Resource
	ctx := context.Background()
	sub, cred, opts := g.getSubscriptionID(), g.getCredential(), g.armClientOptions()

	if client, err := armautomation.NewCertificateClient(sub, cred, opts); err == nil {
		pager := client.NewListByAutomationAccountPager(rg, account, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				resources = appendAutomationResource(resources, item.ID, item.Name, "azurerm_automation_certificate", g.ProviderName)
			}
		}
	}
	if client, err := armautomation.NewConnectionClient(sub, cred, opts); err == nil {
		pager := client.NewListByAutomationAccountPager(rg, account, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				typeName := ""
				if item.Properties != nil && item.Properties.ConnectionType != nil && item.Properties.ConnectionType.Name != nil {
					typeName = *item.Properties.ConnectionType.Name
				}
				resources = appendAutomationResource(resources, item.ID, item.Name, classifyAutomationConnection(typeName), g.ProviderName)
			}
		}
	}
	if client, err := armautomation.NewConnectionTypeClient(sub, cred, opts); err == nil {
		pager := client.NewListByAutomationAccountPager(rg, account, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				resources = appendAutomationResource(resources, item.ID, item.Name, "azurerm_automation_connection_type", g.ProviderName)
			}
		}
	}
	if client, err := armautomation.NewCredentialClient(sub, cred, opts); err == nil {
		pager := client.NewListByAutomationAccountPager(rg, account, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				resources = appendAutomationResource(resources, item.ID, item.Name, "azurerm_automation_credential", g.ProviderName)
			}
		}
	}
	if client, err := armautomation.NewDscConfigurationClient(sub, cred, opts); err == nil {
		pager := client.NewListByAutomationAccountPager(rg, account, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				resources = appendAutomationResource(resources, item.ID, item.Name, "azurerm_automation_dsc_configuration", g.ProviderName)
			}
		}
	}
	if client, err := armautomation.NewDscNodeConfigurationClient(sub, cred, opts); err == nil {
		pager := client.NewListByAutomationAccountPager(rg, account, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				resources = appendAutomationResource(resources, item.ID, item.Name, "azurerm_automation_dsc_nodeconfiguration", g.ProviderName)
			}
		}
	}
	if client, err := armautomation.NewJobScheduleClient(sub, cred, opts); err == nil {
		pager := client.NewListByAutomationAccountPager(rg, account, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				resources = appendAutomationResource(resources, item.ID, item.Name, "azurerm_automation_job_schedule", g.ProviderName)
			}
		}
	}
	if client, err := armautomation.NewModuleClient(sub, cred, opts); err == nil {
		pager := client.NewListByAutomationAccountPager(rg, account, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				resources = appendAutomationResource(resources, item.ID, item.Name, "azurerm_automation_module", g.ProviderName)
			}
		}
	}
	if client, err := armautomation.NewPython3PackageClient(sub, cred, opts); err == nil {
		pager := client.NewListByAutomationAccountPager(rg, account, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				resources = appendAutomationResource(resources, item.ID, item.Name, "azurerm_automation_python3_package", g.ProviderName)
			}
		}
	}
	if client, err := armautomation.NewRunbookClient(sub, cred, opts); err == nil {
		pager := client.NewListByAutomationAccountPager(rg, account, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				resources = appendAutomationResource(resources, item.ID, item.Name, "azurerm_automation_runbook", g.ProviderName)
			}
		}
	}
	if client, err := armautomation.NewScheduleClient(sub, cred, opts); err == nil {
		pager := client.NewListByAutomationAccountPager(rg, account, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				resources = appendAutomationResource(resources, item.ID, item.Name, "azurerm_automation_schedule", g.ProviderName)
			}
		}
	}
	if client, err := armautomation.NewSourceControlClient(sub, cred, opts); err == nil {
		pager := client.NewListByAutomationAccountPager(rg, account, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				resources = appendAutomationResource(resources, item.ID, item.Name, "azurerm_automation_source_control", g.ProviderName)
			}
		}
	}
	if client, err := armautomation.NewWatcherClient(sub, cred, opts); err == nil {
		pager := client.NewListByAutomationAccountPager(rg, account, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				resources = appendAutomationResource(resources, item.ID, item.Name, "azurerm_automation_watcher", g.ProviderName)
			}
		}
	}
	if client, err := armautomation.NewWebhookClient(sub, cred, opts); err == nil {
		pager := client.NewListByAutomationAccountPager(rg, account, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				resources = appendAutomationResource(resources, item.ID, item.Name, "azurerm_automation_webhook", g.ProviderName)
			}
		}
	}
	// Hybrid runbook worker groups, then workers within each group.
	if client, err := armautomation.NewHybridRunbookWorkerGroupClient(sub, cred, opts); err == nil {
		workersClient, werr := armautomation.NewHybridRunbookWorkersClient(sub, cred, opts)
		pager := client.NewListByAutomationAccountPager(rg, account, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				resources = appendAutomationResource(resources, item.ID, item.Name, "azurerm_automation_hybrid_runbook_worker_group", g.ProviderName)
				if werr != nil || item.Name == nil {
					continue
				}
				wPager := workersClient.NewListByHybridRunbookWorkerGroupPager(rg, account, *item.Name, nil)
				for wPager.More() {
					wPage, err := wPager.NextPage(ctx)
					if err != nil {
						log.Println(err)
						break
					}
					for _, w := range wPage.Value {
						resources = appendAutomationResource(resources, w.ID, w.Name, "azurerm_automation_hybrid_runbook_worker", g.ProviderName)
					}
				}
			}
		}
	}
	// Runtime environments, then packages within each environment.
	if client, err := armautomation.NewRuntimeEnvironmentsClient(sub, cred, opts); err == nil {
		pkgClient, perr := armautomation.NewPackageClient(sub, cred, opts)
		pager := client.NewListByAutomationAccountPager(rg, account, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				resources = appendAutomationResource(resources, item.ID, item.Name, "azurerm_automation_runtime_environment", g.ProviderName)
				if perr != nil || item.Name == nil {
					continue
				}
				pPager := pkgClient.NewListByRuntimeEnvironmentPager(rg, account, *item.Name, nil)
				for pPager.More() {
					pPage, err := pPager.NextPage(ctx)
					if err != nil {
						log.Println(err)
						break
					}
					for _, p := range pPage.Value {
						resources = appendAutomationResource(resources, p.ID, p.Name, "azurerm_automation_runtime_environment_package", g.ProviderName)
					}
				}
			}
		}
	}
	return resources
}

func (g *AutomationGenerator) InitResources() error {
	accounts, err := g.listAccounts()
	if err != nil {
		return err
	}
	for _, account := range accounts {
		if account.ID == nil || account.Name == nil {
			continue
		}
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(*account.ID, *account.Name, "azurerm_automation_account", g.ProviderName, []string{}))
		parsed, err := ParseAzureResourceID(*account.ID)
		if err != nil {
			return err
		}
		g.Resources = append(g.Resources, g.listAccountChildren(parsed.ResourceGroup, *account.Name)...)
	}
	return nil
}
