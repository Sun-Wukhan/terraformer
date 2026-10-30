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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/healthcareapis/armhealthcareapis"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// HealthcareGenerator covers Azure Health Data Services via
// armhealthcareapis: legacy services, workspaces, and per-workspace DICOM,
// FHIR and MedTech (IoT connector) services plus MedTech FHIR destinations.
type HealthcareGenerator struct {
	AzureService
}

func (g *HealthcareGenerator) InitResources() error {
	ctx := context.Background()
	sub := g.getSubscriptionID()
	cred := g.getCredential()
	opts := g.armClientOptions()
	rgFilter := g.getResourceGroup()
	var resources []terraformutils.Resource
	add := func(id, name, typ string) {
		if id != "" && name != "" {
			resources = append(resources, terraformutils.NewSimpleResource(id, name, typ, "azurerm", []string{}))
		}
	}

	servicesClient, err := armhealthcareapis.NewServicesClient(sub, cred, opts)
	if err != nil {
		return err
	}
	if rgFilter != "" {
		pager := servicesClient.NewListByResourceGroupPager(rgFilter, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_healthcare_service")
				}
			}
		}
	} else {
		pager := servicesClient.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_healthcare_service")
				}
			}
		}
	}

	workspacesClient, err := armhealthcareapis.NewWorkspacesClient(sub, cred, opts)
	if err != nil {
		return err
	}
	type wsRef struct{ rg, name string }
	var workspaces []wsRef
	if rgFilter != "" {
		pager := workspacesClient.NewListByResourceGroupPager(rgFilter, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_healthcare_workspace")
					workspaces = append(workspaces, wsRef{rgFilter, *item.Name})
				}
			}
		}
	} else {
		pager := workspacesClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				if item.ID == nil || item.Name == nil {
					continue
				}
				add(*item.ID, *item.Name, "azurerm_healthcare_workspace")
				if parsed, err := ParseAzureResourceID(*item.ID); err == nil {
					workspaces = append(workspaces, wsRef{parsed.ResourceGroup, *item.Name})
				}
			}
		}
	}

	dicomClient, err := armhealthcareapis.NewDicomServicesClient(sub, cred, opts)
	if err != nil {
		return err
	}
	fhirClient, err := armhealthcareapis.NewFhirServicesClient(sub, cred, opts)
	if err != nil {
		return err
	}
	iotClient, err := armhealthcareapis.NewIotConnectorsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	destinationsClient, err := armhealthcareapis.NewFhirDestinationsClient(sub, cred, opts)
	if err != nil {
		return err
	}

	for _, ws := range workspaces {
		dPager := dicomClient.NewListByWorkspacePager(ws.rg, ws.name, nil)
		for dPager.More() {
			page, err := dPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_healthcare_dicom_service")
				}
			}
		}
		fPager := fhirClient.NewListByWorkspacePager(ws.rg, ws.name, nil)
		for fPager.More() {
			page, err := fPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_healthcare_fhir_service")
				}
			}
		}
		iPager := iotClient.NewListByWorkspacePager(ws.rg, ws.name, nil)
		for iPager.More() {
			page, err := iPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID == nil || item.Name == nil {
					continue
				}
				add(*item.ID, *item.Name, "azurerm_healthcare_medtech_service")
				fdPager := destinationsClient.NewListByIotConnectorPager(ws.rg, ws.name, *item.Name, nil)
				for fdPager.More() {
					fdPage, err := fdPager.NextPage(ctx)
					if err != nil {
						log.Println(err)
						break
					}
					for _, dest := range fdPage.Value {
						if dest.ID != nil && dest.Name != nil {
							add(*dest.ID, *dest.Name, "azurerm_healthcare_medtech_service_fhir_destination")
						}
					}
				}
			}
		}
	}
	g.Resources = resources
	return nil
}
