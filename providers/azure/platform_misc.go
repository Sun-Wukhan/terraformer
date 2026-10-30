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
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/advisor/armadvisor"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/blueprint/armblueprint"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/customproviders/armcustomproviders"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/databoxedge/armdataboxedge"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/extendedlocation/armextendedlocation"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/fabric/armfabric"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/guestconfiguration/armguestconfiguration"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/healthbot/armhealthbot"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/loadtestservice/armloadtestservice"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/marketplaceordering/armmarketplaceordering"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/msi/armmsi"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/playwrighttesting/armplaywrighttesting"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/portal/armportal"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/powerbidedicated/armpowerbidedicated"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armdeploymentscripts"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/trustedsigning/armtrustedsigning"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/workloads/armworkloads"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// PlatformMiscGenerator covers the platform singleton and misc
// subscription-scope resource types that don't belong to another family.
//
// Skipped in this area:
//   - azurerm_subscription: singleton — it is the target subscription
//     terraformer itself operates in, and its importer ID is a
//     Microsoft.Subscription alias ID, not the subscription's ARM ID.
//   - azurerm_resource_provider_registration: subscription provider
//     registration is a side-effect-only state toggle (register/unregister),
//     not an enumerable resource object.
//   - azurerm_resource_provider_feature_registration: same state-toggle
//     shape, and armresources has no Features client to list them with.
//   - azurerm_resource_management_private_link and
//     azurerm_resource_management_private_link_association: no
//     ResourceManagementPrivateLinks or PrivateLinkAssociations client
//     exists in the track-2 armauthorization (v1.0.0) or armresources
//     (v1.2.0) modules.
//   - azurerm_advanced_threat_protection: armsecurity's
//     AdvancedThreatProtectionClient exposes only Create/Get against a
//     caller-supplied target resource ID — there is no list call, and the
//     settings live under arbitrary individual resources.
//   - azurerm_aadb2c_directory: no track-2 SDK module exists for
//     Microsoft.AzureActiveDirectory (the directory is a tenant-level
//     object managed through a preview API surface with no arm* module).
type PlatformMiscGenerator struct {
	AzureService
}

func (g *PlatformMiscGenerator) add(id *string, name *string, resourceType string) {
	if id == nil || name == nil {
		return
	}
	g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
		*id,
		*name,
		resourceType,
		g.ProviderName,
		[]string{}))
}

func (g *PlatformMiscGenerator) InitResources() error {
	ctx := context.Background()
	subscriptionID := g.getSubscriptionID()
	cred := g.getCredential()
	opts := g.armClientOptions()

	identitiesClient, err := armmsi.NewUserAssignedIdentitiesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	identPager := identitiesClient.NewListBySubscriptionPager(nil)
	for identPager.More() {
		page, err := identPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			break
		}
		for _, item := range page.Value {
			g.add(item.ID, item.Name, "azurerm_user_assigned_identity")
		}
	}

	dashboardsClient, err := armportal.NewDashboardsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	dashPager := dashboardsClient.NewListBySubscriptionPager(nil)
	for dashPager.More() {
		page, err := dashPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			break
		}
		for _, item := range page.Value {
			g.add(item.ID, item.Name, "azurerm_portal_dashboard")
		}
	}

	tenantConfigClient, err := armportal.NewTenantConfigurationsClient(cred, opts)
	if err != nil {
		return err
	}
	tenantPager := tenantConfigClient.NewListPager(nil)
	for tenantPager.More() {
		page, err := tenantPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			break
		}
		for _, item := range page.Value {
			g.add(item.ID, item.Name, "azurerm_portal_tenant_configuration")
		}
	}

	agreementsClient, err := armmarketplaceordering.NewMarketplaceAgreementsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	agreements, err := agreementsClient.List(ctx, nil)
	if err != nil {
		log.Println(err)
	} else {
		for _, item := range agreements.AgreementTermsArray {
			g.add(item.ID, item.Name, "azurerm_marketplace_agreement")
		}
	}

	blueprintClient, err := armblueprint.NewAssignmentsClient(cred, opts)
	if err != nil {
		return err
	}
	blueprintPager := blueprintClient.NewListPager("/subscriptions/"+subscriptionID, nil)
	for blueprintPager.More() {
		page, err := blueprintPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			break
		}
		for _, item := range page.Value {
			g.add(item.ID, item.Name, "azurerm_blueprint_assignment")
		}
	}

	suppressionsClient, err := armadvisor.NewSuppressionsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	suppressionPager := suppressionsClient.NewListPager(nil)
	for suppressionPager.More() {
		page, err := suppressionPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			break
		}
		for _, item := range page.Value {
			g.add(item.ID, item.Name, "azurerm_advisor_suppression")
		}
	}

	customProvidersClient, err := armcustomproviders.NewCustomResourceProviderClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	cpPager := customProvidersClient.NewListBySubscriptionPager(nil)
	for cpPager.More() {
		page, err := cpPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			break
		}
		for _, item := range page.Value {
			g.add(item.ID, item.Name, "azurerm_custom_provider")
		}
	}

	customLocationsClient, err := armextendedlocation.NewCustomLocationsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	clPager := customLocationsClient.NewListBySubscriptionPager(nil)
	for clPager.More() {
		page, err := clPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			break
		}
		for _, item := range page.Value {
			g.add(item.ID, item.Name, "azurerm_extended_location_custom_location")
		}
	}

	fabricClient, err := armfabric.NewCapacitiesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	fabricPager := fabricClient.NewListBySubscriptionPager(nil)
	for fabricPager.More() {
		page, err := fabricPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			break
		}
		for _, item := range page.Value {
			g.add(item.ID, item.Name, "azurerm_fabric_capacity")
		}
	}

	healthBotClient, err := armhealthbot.NewBotsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	botPager := healthBotClient.NewListPager(nil)
	for botPager.More() {
		page, err := botPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			break
		}
		for _, item := range page.Value {
			g.add(item.ID, item.Name, "azurerm_healthbot")
		}
	}

	loadTestClient, err := armloadtestservice.NewLoadTestsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	ltPager := loadTestClient.NewListBySubscriptionPager(nil)
	for ltPager.More() {
		page, err := ltPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			break
		}
		for _, item := range page.Value {
			g.add(item.ID, item.Name, "azurerm_load_test")
		}
	}

	playwrightClient, err := armplaywrighttesting.NewAccountsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	pwPager := playwrightClient.NewListBySubscriptionPager(nil)
	for pwPager.More() {
		page, err := pwPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			break
		}
		for _, item := range page.Value {
			g.add(item.ID, item.Name, "azurerm_playwright_workspace")
		}
	}

	powerBIClient, err := armpowerbidedicated.NewCapacitiesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	pbiPager := powerBIClient.NewListPager(nil)
	for pbiPager.More() {
		page, err := pbiPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			break
		}
		for _, item := range page.Value {
			g.add(item.ID, item.Name, "azurerm_powerbi_embedded")
		}
	}

	signingClient, err := armtrustedsigning.NewCodeSigningAccountsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	signingPager := signingClient.NewListBySubscriptionPager(nil)
	for signingPager.More() {
		page, err := signingPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			break
		}
		for _, item := range page.Value {
			g.add(item.ID, item.Name, "azurerm_trusted_signing_account")
		}
	}

	databoxClient, err := armdataboxedge.NewDevicesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	databoxPager := databoxClient.NewListBySubscriptionPager(nil)
	for databoxPager.More() {
		page, err := databoxPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			break
		}
		for _, item := range page.Value {
			g.add(item.ID, item.Name, "azurerm_databox_edge_device")
		}
	}

	scriptsClient, err := armdeploymentscripts.NewClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	scriptPager := scriptsClient.NewListBySubscriptionPager(nil)
	for scriptPager.More() {
		page, err := scriptPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			break
		}
		for _, item := range page.Value {
			switch script := item.(type) {
			case *armdeploymentscripts.AzureCliScript:
				g.add(script.ID, script.Name, "azurerm_resource_deployment_script_azure_cli")
			case *armdeploymentscripts.AzurePowerShellScript:
				g.add(script.ID, script.Name, "azurerm_resource_deployment_script_azure_power_shell")
			}
		}
	}

	guestConfigClient, err := armguestconfiguration.NewAssignmentsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	guestPager := guestConfigClient.NewSubscriptionListPager(nil)
	for guestPager.More() {
		page, err := guestPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			break
		}
		for _, item := range page.Value {
			// The subscription list also returns VMSS and Arc machine
			// assignments; only VM assignments map to this azurerm type.
			if item.ID == nil || !strings.Contains(*item.ID, "/providers/Microsoft.Compute/virtualMachines/") {
				continue
			}
			g.add(item.ID, item.Name, "azurerm_policy_virtual_machine_configuration_assignment")
		}
	}

	sapClient, err := armworkloads.NewSAPVirtualInstancesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	sapPager := sapClient.NewListBySubscriptionPager(nil)
	for sapPager.More() {
		page, err := sapPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			break
		}
		for _, item := range page.Value {
			if resourceType := sapVirtualInstanceType(item.Properties); resourceType != "" {
				g.add(item.ID, item.Name, resourceType)
			}
		}
	}
	return nil
}

// sapVirtualInstanceType maps an enumerated SAP virtual instance to the
// azurerm type matching how it was created: the three azurerm types share
// one ARM resource and differ only in the configuration payload.
func sapVirtualInstanceType(properties *armworkloads.SAPVirtualInstanceProperties) string {
	if properties == nil {
		return ""
	}
	switch config := properties.Configuration.(type) {
	case *armworkloads.DiscoveryConfiguration:
		return "azurerm_workloads_sap_discovery_virtual_instance"
	case *armworkloads.DeploymentConfiguration:
		return sapInfrastructureType(config.InfrastructureConfiguration)
	case *armworkloads.DeploymentWithOSConfiguration:
		return sapInfrastructureType(config.InfrastructureConfiguration)
	}
	return ""
}

func sapInfrastructureType(infrastructure armworkloads.InfrastructureConfigurationClassification) string {
	switch infrastructure.(type) {
	case *armworkloads.SingleServerConfiguration:
		return "azurerm_workloads_sap_single_node_virtual_instance"
	case *armworkloads.ThreeTierConfiguration:
		return "azurerm_workloads_sap_three_tier_virtual_instance"
	}
	return ""
}
