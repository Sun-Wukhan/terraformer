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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/iothub/armiothub"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"
)

// SecurityCenterGenerator covers the subscription-level Microsoft.Security
// resources beyond contacts and subscription pricing.
// Skipped in this family:
//   - azurerm_security_center_server_vulnerability_assessment_virtual_machine:
//     armsecurity exposes it only per extended resource (per virtual
//     machine, ServerVulnerabilityAssessmentClient.ListByExtendedResource /
//     Get), with no subscription- or resource-group-level List.
//   - azurerm_security_center_storage_defender: armsecurity exposes
//     DefenderForStorage settings only scoped to a single resource ID
//     (per storage account), with no subscription- or resource-group-level
//     List.
type SecurityCenterGenerator struct {
	AzureService
}

func (g *SecurityCenterGenerator) InitResources() error {
	ctx := context.Background()
	sub := g.getSubscriptionID()
	cred := g.getCredential()
	opts := g.armClientOptions()

	assessmentsClient, err := armsecurity.NewAssessmentsClient(cred, opts)
	if err != nil {
		return err
	}
	scope := "subscriptions/" + sub
	pager := assessmentsClient.NewListPager(scope, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, item := range page.Value {
			g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_security_center_assessment")
		}
	}

	metadataClient, err := armsecurity.NewAssessmentsMetadataClient(sub, cred, opts)
	if err != nil {
		return err
	}
	metaPager := metadataClient.NewListPager(nil)
	for metaPager.More() {
		page, err := metaPager.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, item := range page.Value {
			g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_security_center_assessment_policy")
		}
	}

	automationsClient, err := armsecurity.NewAutomationsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	if rg := g.getResourceGroup(); rg != "" {
		autoPager := automationsClient.NewListByResourceGroupPager(rg, nil)
		for autoPager.More() {
			page, err := autoPager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_security_center_automation")
			}
		}
	} else {
		autoPager := automationsClient.NewListPager(nil)
		for autoPager.More() {
			page, err := autoPager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_security_center_automation")
			}
		}
	}

	settingsClient, err := armsecurity.NewSettingsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	settingsPager := settingsClient.NewListPager(nil)
	for settingsPager.More() {
		page, err := settingsPager.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, item := range page.Value {
			setting := item.GetSetting()
			if setting == nil || setting.ID == nil || setting.Name == nil {
				continue
			}
			g.AppendSimpleResource(*setting.ID, *setting.Name, "azurerm_security_center_setting")
		}
	}

	workspaceClient, err := armsecurity.NewWorkspaceSettingsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	wsPager := workspaceClient.NewListPager(nil)
	for wsPager.More() {
		page, err := wsPager.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, item := range page.Value {
			g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_security_center_workspace")
		}
	}

	svaClient, err := armsecurity.NewServerVulnerabilityAssessmentsSettingsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	svaPager := svaClient.NewListBySubscriptionPager(nil)
	for svaPager.More() {
		page, err := svaPager.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, item := range page.Value {
			setting := item.GetServerVulnerabilityAssessmentsSetting()
			if setting == nil || setting.ID == nil || setting.Name == nil {
				continue
			}
			g.AppendSimpleResource(*setting.ID, *setting.Name, "azurerm_security_center_server_vulnerability_assessments_setting")
		}
	}
	return nil
}

// IoTSecurityGenerator covers Defender for IoT solutions and the device
// security groups attached to IoT Hubs.
type IoTSecurityGenerator struct {
	AzureService
}

func (g *IoTSecurityGenerator) InitResources() error {
	ctx := context.Background()
	sub := g.getSubscriptionID()
	cred := g.getCredential()
	opts := g.armClientOptions()

	solutionsClient, err := armsecurity.NewIotSecuritySolutionClient(sub, cred, opts)
	if err != nil {
		return err
	}
	if rg := g.getResourceGroup(); rg != "" {
		pager := solutionsClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_iot_security_solution")
			}
		}
	} else {
		pager := solutionsClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_iot_security_solution")
			}
		}
	}

	hubClient, err := armiothub.NewResourceClient(sub, cred, opts)
	if err != nil {
		return err
	}
	var hubs []*armiothub.Description
	if rg := g.getResourceGroup(); rg != "" {
		pager := hubClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			hubs = append(hubs, page.Value...)
		}
	} else {
		pager := hubClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			hubs = append(hubs, page.Value...)
		}
	}

	dsgClient, err := armsecurity.NewDeviceSecurityGroupsClient(cred, opts)
	if err != nil {
		return err
	}
	for _, hub := range hubs {
		pager := dsgClient.NewListPager(*hub.ID, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_iot_security_device_group")
			}
		}
	}
	return nil
}
