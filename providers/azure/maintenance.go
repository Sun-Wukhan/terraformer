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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/maintenance/armmaintenance"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// MaintenanceGenerator covers Microsoft.Maintenance configurations and
// configuration assignments. Assignments are listed subscription-wide
// via ConfigurationAssignmentsWithinSubscriptionClient and split into
// the four provider types by the scope embedded in the assignment ID.
type MaintenanceGenerator struct {
	AzureService
}

func (g *MaintenanceGenerator) appendSimple(id, name, resourceType string) {
	if id == "" || name == "" {
		return
	}
	g.Resources = append(g.Resources, terraformutils.NewSimpleResource(id, name, resourceType, "azurerm", []string{}))
}

// maintenanceAssignmentType maps a configuration assignment ARM ID to
// its provider resource type, or "" when the assignment targets a
// resource class with no dedicated provider type.
func maintenanceAssignmentType(id string) string {
	switch {
	case strings.Contains(id, "/providers/Microsoft.Compute/virtualMachineScaleSets/"):
		return "azurerm_maintenance_assignment_virtual_machine_scale_set"
	case strings.Contains(id, "/providers/Microsoft.Compute/virtualMachines/"):
		return "azurerm_maintenance_assignment_virtual_machine"
	case strings.Contains(id, "/providers/Microsoft.Compute/hostGroups/"):
		return "azurerm_maintenance_assignment_dedicated_host"
	case !strings.Contains(id, "/resourceGroups/") &&
		strings.Contains(id, "/providers/Microsoft.Maintenance/configurationAssignments/"):
		return "azurerm_maintenance_assignment_dynamic_scope"
	default:
		return ""
	}
}

func (g *MaintenanceGenerator) InitResources() error {
	ctx := context.Background()
	subscriptionID := g.getSubscriptionID()
	cred := g.getCredential()
	opts := g.armClientOptions()

	configClient, err := armmaintenance.NewConfigurationsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	pager := configClient.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			if g.getResourceGroup() != "" && resourceGroupFromID(*item.ID) != g.getResourceGroup() {
				continue
			}
			g.appendSimple(*item.ID, *item.Name, "azurerm_maintenance_configuration")
		}
	}

	assignmentClient, err := armmaintenance.NewConfigurationAssignmentsWithinSubscriptionClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	apager := assignmentClient.NewListPager(nil)
	for apager.More() {
		page, err := apager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			if item.ID == nil || item.Name == nil {
				continue
			}
			if g.getResourceGroup() != "" && resourceGroupFromID(*item.ID) != g.getResourceGroup() &&
				maintenanceAssignmentType(*item.ID) != "azurerm_maintenance_assignment_dynamic_scope" {
				continue
			}
			if resourceType := maintenanceAssignmentType(*item.ID); resourceType != "" {
				g.appendSimple(*item.ID, *item.Name, resourceType)
			}
		}
	}
	return nil
}
