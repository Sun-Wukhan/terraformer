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
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/logic/armlogic"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// LogicAppGenerator imports Logic Apps (Consumption) workflows, their
// definition-embedded actions and triggers, and Integration Accounts with
// their child resources, via the track-2 armlogic SDK.
//
// Actions and triggers are not independent ARM child resources: they live
// in the workflow definition. They are enumerated by reading each
// workflow's definition (WorkflowsClient.Get) and mapping each named
// action/trigger to the azurerm type matching its definition type, which
// reproduces the provider importer IDs
// (<workflow id>/actions/<name> and <workflow id>/triggers/<name>).
type LogicAppGenerator struct {
	AzureService
}

func (g *LogicAppGenerator) simple(id, name, resourceType string) (terraformutils.Resource, bool) {
	if id == "" || name == "" {
		return terraformutils.Resource{}, false
	}
	return terraformutils.NewSimpleResource(id, name, resourceType, g.ProviderName, []string{}), true
}

func (g *LogicAppGenerator) InitResources() error {
	ctx := context.Background()
	if err := g.listWorkflows(ctx); err != nil {
		return err
	}
	return g.listIntegrationAccounts(ctx)
}

// logicDefinitionResources maps the actions and triggers embedded in a
// workflow definition to azurerm resources under the workflow ID.
func logicDefinitionResources(providerName, workflowID string, definition interface{}) []terraformutils.Resource {
	var resources []terraformutils.Resource
	def, ok := definition.(map[string]interface{})
	if !ok {
		return resources
	}
	if actions, ok := def["actions"].(map[string]interface{}); ok {
		for name, raw := range actions {
			actionType := ""
			if action, ok := raw.(map[string]interface{}); ok {
				actionType, _ = action["type"].(string)
			}
			resourceType := "azurerm_logic_app_action_custom"
			if strings.EqualFold(actionType, "Http") {
				resourceType = "azurerm_logic_app_action_http"
			}
			id := workflowID + "/actions/" + name
			resources = append(resources, terraformutils.NewSimpleResource(id, name, resourceType, providerName, []string{}))
		}
	}
	if triggers, ok := def["triggers"].(map[string]interface{}); ok {
		for name, raw := range triggers {
			triggerType := ""
			if trigger, ok := raw.(map[string]interface{}); ok {
				triggerType, _ = trigger["type"].(string)
			}
			resourceType := "azurerm_logic_app_trigger_custom"
			switch {
			case strings.EqualFold(triggerType, "Recurrence"):
				resourceType = "azurerm_logic_app_trigger_recurrence"
			case strings.EqualFold(triggerType, "Request"):
				resourceType = "azurerm_logic_app_trigger_http_request"
			}
			id := workflowID + "/triggers/" + name
			resources = append(resources, terraformutils.NewSimpleResource(id, name, resourceType, providerName, []string{}))
		}
	}
	return resources
}

func (g *LogicAppGenerator) listWorkflows(ctx context.Context) error {
	client, err := armlogic.NewWorkflowsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		return collectLogicWorkflows(ctx, g, client, pager, func(p armlogic.WorkflowsClientListByResourceGroupResponse) []*armlogic.Workflow { return p.Value })
	}
	pager := client.NewListBySubscriptionPager(nil)
	return collectLogicWorkflows(ctx, g, client, pager, func(p armlogic.WorkflowsClientListBySubscriptionResponse) []*armlogic.Workflow { return p.Value })
}

func collectLogicWorkflows[R any](ctx context.Context, g *LogicAppGenerator, client *armlogic.WorkflowsClient, pager *runtime.Pager[R], values func(R) []*armlogic.Workflow) error {
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, wf := range values(page) {
			if wf.ID == nil || wf.Name == nil {
				continue
			}
			if res, ok := g.simple(*wf.ID, *wf.Name, "azurerm_logic_app_workflow"); ok {
				g.Resources = append(g.Resources, res)
			}
			parsed, err := ParseAzureResourceID(*wf.ID)
			if err != nil {
				continue
			}
			full, err := client.Get(ctx, parsed.ResourceGroup, *wf.Name, nil)
			if err != nil {
				return err
			}
			if full.Properties != nil {
				g.Resources = append(g.Resources, logicDefinitionResources(g.ProviderName, *wf.ID, full.Properties.Definition)...)
			}
		}
	}
	return nil
}

func (g *LogicAppGenerator) listIntegrationAccounts(ctx context.Context) error {
	client, err := armlogic.NewIntegrationAccountsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	var accounts []*armlogic.IntegrationAccount
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			accounts = append(accounts, page.Value...)
		}
	} else {
		pager := client.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			accounts = append(accounts, page.Value...)
		}
	}
	for _, account := range accounts {
		if account.ID == nil || account.Name == nil {
			continue
		}
		if res, ok := g.simple(*account.ID, *account.Name, "azurerm_logic_app_integration_account"); ok {
			g.Resources = append(g.Resources, res)
		}
		parsed, err := ParseAzureResourceID(*account.ID)
		if err != nil {
			continue
		}
		if err := g.listIntegrationAccountChildren(ctx, parsed.ResourceGroup, *account.Name); err != nil {
			return err
		}
	}
	return nil
}

func (g *LogicAppGenerator) listIntegrationAccountChildren(ctx context.Context, resourceGroup, accountName string) error {
	agreementsClient, err := armlogic.NewIntegrationAccountAgreementsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	resources, err := drainPager(ctx, agreementsClient.NewListPager(resourceGroup, accountName, nil),
		func(p armlogic.IntegrationAccountAgreementsClientListResponse) []*armlogic.IntegrationAccountAgreement {
			return p.Value
		},
		func(item *armlogic.IntegrationAccountAgreement) (terraformutils.Resource, bool) {
			if item.ID == nil || item.Name == nil {
				return terraformutils.Resource{}, false
			}
			return g.simple(*item.ID, *item.Name, "azurerm_logic_app_integration_account_agreement")
		})
	if err != nil {
		return err
	}
	g.Resources = append(g.Resources, resources...)

	assembliesClient, err := armlogic.NewIntegrationAccountAssembliesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	resources, err = drainPager(ctx, assembliesClient.NewListPager(resourceGroup, accountName, nil),
		func(p armlogic.IntegrationAccountAssembliesClientListResponse) []*armlogic.AssemblyDefinition {
			return p.Value
		},
		func(item *armlogic.AssemblyDefinition) (terraformutils.Resource, bool) {
			if item.ID == nil || item.Name == nil {
				return terraformutils.Resource{}, false
			}
			return g.simple(*item.ID, *item.Name, "azurerm_logic_app_integration_account_assembly")
		})
	if err != nil {
		return err
	}
	g.Resources = append(g.Resources, resources...)

	batchClient, err := armlogic.NewIntegrationAccountBatchConfigurationsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	resources, err = drainPager(ctx, batchClient.NewListPager(resourceGroup, accountName, nil),
		func(p armlogic.IntegrationAccountBatchConfigurationsClientListResponse) []*armlogic.BatchConfiguration {
			return p.Value
		},
		func(item *armlogic.BatchConfiguration) (terraformutils.Resource, bool) {
			if item.ID == nil || item.Name == nil {
				return terraformutils.Resource{}, false
			}
			return g.simple(*item.ID, *item.Name, "azurerm_logic_app_integration_account_batch_configuration")
		})
	if err != nil {
		return err
	}
	g.Resources = append(g.Resources, resources...)

	certificatesClient, err := armlogic.NewIntegrationAccountCertificatesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	resources, err = drainPager(ctx, certificatesClient.NewListPager(resourceGroup, accountName, nil),
		func(p armlogic.IntegrationAccountCertificatesClientListResponse) []*armlogic.IntegrationAccountCertificate {
			return p.Value
		},
		func(item *armlogic.IntegrationAccountCertificate) (terraformutils.Resource, bool) {
			if item.ID == nil || item.Name == nil {
				return terraformutils.Resource{}, false
			}
			return g.simple(*item.ID, *item.Name, "azurerm_logic_app_integration_account_certificate")
		})
	if err != nil {
		return err
	}
	g.Resources = append(g.Resources, resources...)

	mapsClient, err := armlogic.NewIntegrationAccountMapsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	resources, err = drainPager(ctx, mapsClient.NewListPager(resourceGroup, accountName, nil),
		func(p armlogic.IntegrationAccountMapsClientListResponse) []*armlogic.IntegrationAccountMap {
			return p.Value
		},
		func(item *armlogic.IntegrationAccountMap) (terraformutils.Resource, bool) {
			if item.ID == nil || item.Name == nil {
				return terraformutils.Resource{}, false
			}
			return g.simple(*item.ID, *item.Name, "azurerm_logic_app_integration_account_map")
		})
	if err != nil {
		return err
	}
	g.Resources = append(g.Resources, resources...)

	partnersClient, err := armlogic.NewIntegrationAccountPartnersClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	resources, err = drainPager(ctx, partnersClient.NewListPager(resourceGroup, accountName, nil),
		func(p armlogic.IntegrationAccountPartnersClientListResponse) []*armlogic.IntegrationAccountPartner {
			return p.Value
		},
		func(item *armlogic.IntegrationAccountPartner) (terraformutils.Resource, bool) {
			if item.ID == nil || item.Name == nil {
				return terraformutils.Resource{}, false
			}
			return g.simple(*item.ID, *item.Name, "azurerm_logic_app_integration_account_partner")
		})
	if err != nil {
		return err
	}
	g.Resources = append(g.Resources, resources...)

	schemasClient, err := armlogic.NewIntegrationAccountSchemasClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	resources, err = drainPager(ctx, schemasClient.NewListPager(resourceGroup, accountName, nil),
		func(p armlogic.IntegrationAccountSchemasClientListResponse) []*armlogic.IntegrationAccountSchema {
			return p.Value
		},
		func(item *armlogic.IntegrationAccountSchema) (terraformutils.Resource, bool) {
			if item.ID == nil || item.Name == nil {
				return terraformutils.Resource{}, false
			}
			return g.simple(*item.ID, *item.Name, "azurerm_logic_app_integration_account_schema")
		})
	if err != nil {
		return err
	}
	g.Resources = append(g.Resources, resources...)

	sessionsClient, err := armlogic.NewIntegrationAccountSessionsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	resources, err = drainPager(ctx, sessionsClient.NewListPager(resourceGroup, accountName, nil),
		func(p armlogic.IntegrationAccountSessionsClientListResponse) []*armlogic.IntegrationAccountSession {
			return p.Value
		},
		func(item *armlogic.IntegrationAccountSession) (terraformutils.Resource, bool) {
			if item.ID == nil || item.Name == nil {
				return terraformutils.Resource{}, false
			}
			return g.simple(*item.ID, *item.Name, "azurerm_logic_app_integration_account_session")
		})
	if err != nil {
		return err
	}
	g.Resources = append(g.Resources, resources...)
	return nil
}
