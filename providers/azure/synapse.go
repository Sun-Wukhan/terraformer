// Copyright 2021 The Terraformer Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package azure

import (
	"context"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/synapse/armsynapse"
	"github.com/Azure/azure-sdk-for-go/services/preview/synapse/2019-06-01-preview/managedvirtualnetwork"
	// "github.com/Azure/azure-sdk-for-go/services/preview/synapse/2020-08-01-preview/accesscontrol"
)

type SynapseGenerator struct {
	AzureService
}

func (az *SynapseGenerator) listWorkspaces() ([]armsynapse.Workspace, error) {
	client, err := armsynapse.NewWorkspacesClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	var resources []armsynapse.Workspace
	if rg := az.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return resources, err
			}
			for _, workspace := range page.Value {
				resources = append(resources, *workspace)
			}
		}
	} else {
		pager := client.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return resources, err
			}
			for _, workspace := range page.Value {
				resources = append(resources, *workspace)
			}
		}
	}
	return resources, nil
}

func (az *SynapseGenerator) appendWorkspace(workspace *armsynapse.Workspace) {
	az.AppendSimpleResource(*workspace.ID, *workspace.Name, "azurerm_synapse_workspace")
}

func (az *SynapseGenerator) appendSQLPools(workspace *armsynapse.Workspace, workspaceRg *ResourceID) error {
	client, err := armsynapse.NewSQLPoolsClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
	if err != nil {
		return err
	}
	ctx := context.Background()
	pager := client.NewListByWorkspacePager(workspaceRg.ResourceGroup, *workspace.Name, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			az.AppendSimpleResource(*item.ID, *item.Name, "azurerm_synapse_sql_pool")
			if err := az.appendSQLPoolChildren(workspace, workspaceRg, *item.Name); err != nil {
				return err
			}
		}
	}
	return nil
}

// Sweep additions (data cluster). Skipped in this family:
//   - azurerm_synapse_workspace_aad_admin and
//     azurerm_synapse_workspace_sql_aad_admin: singleton settings;
//     armsynapse WorkspaceAADAdminsClient / WorkspaceSQLAADAdminsClient are
//     Get-only (fixed name "activeDirectory"), no List pager.
//   - azurerm_synapse_sql_pool_vulnerability_assessment_baseline: armsynapse
//     SQLPoolVulnerabilityAssessmentRuleBaselinesClient is Get-only per
//     rule/baseline name, no List pager.
//   - azurerm_synapse_linked_service and azurerm_synapse_role_assignment:
//     data-plane only (managed through the workspace's Synapse artifacts /
//     RBAC endpoints at {workspace}.dev.azuresynapse.net); armsynapse has
//     no ARM client for either.

// synapseIntegrationRuntimeType maps the ARM integration runtime type to
// the provider resource type.
func synapseIntegrationRuntimeType(properties armsynapse.IntegrationRuntimeClassification) string {
	if properties == nil {
		return ""
	}
	base := properties.GetIntegrationRuntime()
	if base == nil || base.Type == nil {
		return ""
	}
	switch *base.Type {
	case armsynapse.IntegrationRuntimeTypeManaged:
		return "azurerm_synapse_integration_runtime_azure"
	case armsynapse.IntegrationRuntimeTypeSelfHosted:
		return "azurerm_synapse_integration_runtime_self_hosted"
	}
	return ""
}

func (az *SynapseGenerator) appendIntegrationRuntimes(workspace *armsynapse.Workspace, workspaceRg *ResourceID) error {
	client, err := armsynapse.NewIntegrationRuntimesClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
	if err != nil {
		return err
	}
	ctx := context.Background()
	pager := client.NewListByWorkspacePager(workspaceRg.ResourceGroup, *workspace.Name, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			resourceType := synapseIntegrationRuntimeType(item.Properties)
			if resourceType == "" {
				continue
			}
			az.AppendSimpleResource(*item.ID, *item.Name, resourceType)
		}
	}
	return nil
}

func (az *SynapseGenerator) appendWorkspaceKeys(workspace *armsynapse.Workspace, workspaceRg *ResourceID) error {
	client, err := armsynapse.NewKeysClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
	if err != nil {
		return err
	}
	ctx := context.Background()
	pager := client.NewListByWorkspacePager(workspaceRg.ResourceGroup, *workspace.Name, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			az.AppendSimpleResource(*item.ID, *item.Name, "azurerm_synapse_workspace_key")
		}
	}
	return nil
}

// appendWorkspacePolicies enumerates the workspace-level (managed SQL
// server) singleton policies; each List endpoint returns the single
// "default"/"Default" policy whose ID is the provider importer ID.
func (az *SynapseGenerator) appendWorkspacePolicies(workspace *armsynapse.Workspace, workspaceRg *ResourceID) error {
	ctx := context.Background()

	extendedAuditingClient, err := armsynapse.NewWorkspaceManagedSQLServerExtendedBlobAuditingPoliciesClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
	if err != nil {
		return err
	}
	extendedAuditingPager := extendedAuditingClient.NewListByWorkspacePager(workspaceRg.ResourceGroup, *workspace.Name, nil)
	for extendedAuditingPager.More() {
		page, err := extendedAuditingPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			az.AppendSimpleResource(*item.ID, *item.Name+"-"+*workspace.Name, "azurerm_synapse_workspace_extended_auditing_policy")
		}
	}

	securityAlertClient, err := armsynapse.NewWorkspaceManagedSQLServerSecurityAlertPolicyClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
	if err != nil {
		return err
	}
	securityAlertPager := securityAlertClient.NewListPager(workspaceRg.ResourceGroup, *workspace.Name, nil)
	for securityAlertPager.More() {
		page, err := securityAlertPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			az.AppendSimpleResource(*item.ID, *item.Name+"-"+*workspace.Name, "azurerm_synapse_workspace_security_alert_policy")
		}
	}

	vulnerabilityClient, err := armsynapse.NewWorkspaceManagedSQLServerVulnerabilityAssessmentsClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
	if err != nil {
		return err
	}
	vulnerabilityPager := vulnerabilityClient.NewListPager(workspaceRg.ResourceGroup, *workspace.Name, nil)
	for vulnerabilityPager.More() {
		page, err := vulnerabilityPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			az.AppendSimpleResource(*item.ID, *item.Name+"-"+*workspace.Name, "azurerm_synapse_workspace_vulnerability_assessment")
		}
	}

	return nil
}

// appendSQLPoolChildren enumerates the per-pool policies and workload
// groups/classifiers; the policy List endpoints return the single
// "default" policy whose ID is the provider importer ID.
func (az *SynapseGenerator) appendSQLPoolChildren(workspace *armsynapse.Workspace, workspaceRg *ResourceID, sqlPoolName string) error {
	ctx := context.Background()

	extendedAuditingClient, err := armsynapse.NewExtendedSQLPoolBlobAuditingPoliciesClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
	if err != nil {
		return err
	}
	extendedAuditingPager := extendedAuditingClient.NewListBySQLPoolPager(workspaceRg.ResourceGroup, *workspace.Name, sqlPoolName, nil)
	for extendedAuditingPager.More() {
		page, err := extendedAuditingPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			az.AppendSimpleResource(*item.ID, *item.Name+"-"+sqlPoolName, "azurerm_synapse_sql_pool_extended_auditing_policy")
		}
	}

	securityAlertClient, err := armsynapse.NewSQLPoolSecurityAlertPoliciesClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
	if err != nil {
		return err
	}
	securityAlertPager := securityAlertClient.NewListPager(workspaceRg.ResourceGroup, *workspace.Name, sqlPoolName, nil)
	for securityAlertPager.More() {
		page, err := securityAlertPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			az.AppendSimpleResource(*item.ID, *item.Name+"-"+sqlPoolName, "azurerm_synapse_sql_pool_security_alert_policy")
		}
	}

	vulnerabilityClient, err := armsynapse.NewSQLPoolVulnerabilityAssessmentsClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
	if err != nil {
		return err
	}
	vulnerabilityPager := vulnerabilityClient.NewListPager(workspaceRg.ResourceGroup, *workspace.Name, sqlPoolName, nil)
	for vulnerabilityPager.More() {
		page, err := vulnerabilityPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			az.AppendSimpleResource(*item.ID, *item.Name+"-"+sqlPoolName, "azurerm_synapse_sql_pool_vulnerability_assessment")
		}
	}

	workloadGroupsClient, err := armsynapse.NewSQLPoolWorkloadGroupClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
	if err != nil {
		return err
	}
	workloadClassifiersClient, err := armsynapse.NewSQLPoolWorkloadClassifierClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
	if err != nil {
		return err
	}
	workloadGroupsPager := workloadGroupsClient.NewListPager(workspaceRg.ResourceGroup, *workspace.Name, sqlPoolName, nil)
	for workloadGroupsPager.More() {
		page, err := workloadGroupsPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, group := range page.Value {
			az.AppendSimpleResource(*group.ID, *group.Name, "azurerm_synapse_sql_pool_workload_group")
			classifiersPager := workloadClassifiersClient.NewListPager(workspaceRg.ResourceGroup, *workspace.Name, sqlPoolName, *group.Name, nil)
			for classifiersPager.More() {
				classifiersPage, err := classifiersPager.NextPage(ctx)
				if err != nil {
					log.Println(err)
					return err
				}
				for _, classifier := range classifiersPage.Value {
					az.AppendSimpleResource(*classifier.ID, *classifier.Name, "azurerm_synapse_sql_pool_workload_classifier")
				}
			}
		}
	}

	return nil
}

func (az *SynapseGenerator) appendSparkPools(workspace *armsynapse.Workspace, workspaceRg *ResourceID) error {
	client, err := armsynapse.NewBigDataPoolsClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
	if err != nil {
		return err
	}
	ctx := context.Background()
	pager := client.NewListByWorkspacePager(workspaceRg.ResourceGroup, *workspace.Name, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			az.AppendSimpleResource(*item.ID, *item.Name, "azurerm_synapse_spark_pool")
		}
	}
	return nil
}

func (az *SynapseGenerator) appendFirewallRule(workspace *armsynapse.Workspace, workspaceRg *ResourceID) error {
	client, err := armsynapse.NewIPFirewallRulesClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
	if err != nil {
		return err
	}
	ctx := context.Background()
	pager := client.NewListByWorkspacePager(workspaceRg.ResourceGroup, *workspace.Name, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			az.AppendSimpleResource(*item.ID, *item.Name, "azurerm_synapse_firewall_rule")
		}
	}
	return nil
}

func (az *SynapseGenerator) appendManagedPrivateEndpoint(workspace *armsynapse.Workspace) error {
	// armsynapse@v0.8.0 has no ManagedPrivateEndpointsClient; keep this one
	// listing on track-1 (the workspace check uses the track-2 Properties).
	if workspace.Properties == nil || workspace.Properties.ManagedVirtualNetwork == nil {
		return nil
	}
	virtualNetworkName := *workspace.Properties.ManagedVirtualNetwork
	if virtualNetworkName == "" || virtualNetworkName == "default" {
		return nil
	}
	subscriptionID, _, authorizer, _ := az.getClientArgs()
	// ManagedPrivateEndpointsClient does not have a ...WithBaseURI function, why is this different?
	client := managedvirtualnetwork.NewManagedPrivateEndpointsClient(subscriptionID)
	client.Authorizer = authorizer
	ctx := context.Background()
	iterator, err := client.ListComplete(ctx, virtualNetworkName)
	if err != nil {
		return err
	}
	for iterator.NotDone() {
		item := iterator.Value()
		az.AppendSimpleResource(*item.ID, *item.Name, "azurerm_synapse_managed_private_endpoint")
		if err := iterator.NextWithContext(ctx); err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}

func (az *SynapseGenerator) listPrivateLinkHubs() ([]armsynapse.PrivateLinkHub, error) {
	client, err := armsynapse.NewPrivateLinkHubsClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	var resources []armsynapse.PrivateLinkHub
	if rg := az.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return resources, err
			}
			for _, hub := range page.Value {
				resources = append(resources, *hub)
			}
		}
	} else {
		pager := client.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return resources, err
			}
			for _, hub := range page.Value {
				resources = append(resources, *hub)
			}
		}
	}
	return resources, nil
}

func (az *SynapseGenerator) appendtPrivateLinkHubs(workspace *armsynapse.PrivateLinkHub) {
	az.AppendSimpleResource(*workspace.ID, *workspace.Name, "azurerm_synapse_private_link_hub")
}

func (az *SynapseGenerator) InitResources() error {

	workspaces, err := az.listWorkspaces()
	if err != nil {
		return err
	}
	for _, workspace := range workspaces {
		az.appendWorkspace(&workspace)
		workspaceRg, err := ParseAzureResourceID(*workspace.ID)
		if err != nil {
			return err
		}
		err = az.appendSQLPools(&workspace, workspaceRg)
		if err != nil {
			return err
		}
		err = az.appendSparkPools(&workspace, workspaceRg)
		if err != nil {
			return err
		}
		err = az.appendFirewallRule(&workspace, workspaceRg)
		if err != nil {
			return err
		}
		err = az.appendManagedPrivateEndpoint(&workspace)
		if err != nil {
			return err
		}
		err = az.appendIntegrationRuntimes(&workspace, workspaceRg)
		if err != nil {
			return err
		}
		err = az.appendWorkspaceKeys(&workspace, workspaceRg)
		if err != nil {
			return err
		}
		err = az.appendWorkspacePolicies(&workspace, workspaceRg)
		if err != nil {
			return err
		}
	}

	hubs, err := az.listPrivateLinkHubs()
	if err == nil {
		for _, hub := range hubs {
			az.appendtPrivateLinkHubs(&hub)
		}
	} else {
		az.RecordFailure("azurerm_synapse_private_link_hub", err)
	}
	return nil
}
