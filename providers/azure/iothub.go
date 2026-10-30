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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/deviceupdate/armdeviceupdate"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/iotcentral/armiotcentral"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/iothub/armiothub"
)

// IoTHubGenerator covers IoT Hubs and their ARM-listable children.
// Skipped in this family:
//   - azurerm_iothub_dps, azurerm_iothub_dps_certificate and
//     azurerm_iothub_dps_shared_access_policy: no Device Provisioning
//     Service (Microsoft.Devices/provisioningServices) client or List
//     exists in the track-2 armiothub v1.3.0 module.
//   - azurerm_iothub_endpoint_cosmosdb_account, azurerm_iothub_endpoint_eventhub,
//     azurerm_iothub_endpoint_servicebus_queue, azurerm_iothub_endpoint_servicebus_topic,
//     azurerm_iothub_endpoint_storage_container, azurerm_iothub_enrichment,
//     azurerm_iothub_fallback_route and azurerm_iothub_route: embedded
//     arrays in the hub object (properties.routing.endpoints / .enrichments /
//     .routes / .fallbackRoute) with no separate ARM List.
//   - azurerm_iothub_file_upload: an embedded property of the hub object
//     (properties.storageEndpoints.$default.fileUploadNotifications /
//     messaging endpoints) with no separate ARM List.
type IoTHubGenerator struct {
	AzureService
}

func (g *IoTHubGenerator) InitResources() error {
	ctx := context.Background()
	sub := g.getSubscriptionID()
	cred := g.getCredential()
	opts := g.armClientOptions()

	resourceClient, err := armiothub.NewResourceClient(sub, cred, opts)
	if err != nil {
		return err
	}
	var hubs []*armiothub.Description
	if rg := g.getResourceGroup(); rg != "" {
		pager := resourceClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			hubs = append(hubs, page.Value...)
		}
	} else {
		pager := resourceClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			hubs = append(hubs, page.Value...)
		}
	}

	certClient, err := armiothub.NewCertificatesClient(sub, cred, opts)
	if err != nil {
		return err
	}

	for _, hub := range hubs {
		g.AppendSimpleResource(*hub.ID, *hub.Name, "azurerm_iothub")
		parsed, err := ParseAzureResourceID(*hub.ID)
		if err != nil {
			return err
		}
		rg := parsed.ResourceGroup
		hubName := *hub.Name

		certs, err := certClient.ListByIotHub(ctx, rg, hubName, nil)
		if err != nil {
			return err
		}
		for _, cert := range certs.Value {
			g.AppendSimpleResource(*cert.ID, *cert.Name, "azurerm_iothub_certificate")
		}

		cgPager := resourceClient.NewListEventHubConsumerGroupsPager(rg, hubName, "events", nil)
		for cgPager.More() {
			page, err := cgPager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_iothub_consumer_group")
			}
		}

		keyPager := resourceClient.NewListKeysPager(rg, hubName, nil)
		for keyPager.More() {
			page, err := keyPager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				// Shared access policies have no ARM ID of their own; the
				// provider importer ID is <hub ID>/iotHubKeys/<key name>
				// (verified against the provider Import documentation).
				g.AppendSimpleResource(*hub.ID+"/iotHubKeys/"+*item.KeyName, *item.KeyName, "azurerm_iothub_shared_access_policy")
			}
		}
	}
	return nil
}

// IoTDeviceUpdateGenerator covers Device Update for IoT Hub accounts and
// their instances (Microsoft.DeviceUpdate/accounts[/instances]).
type IoTDeviceUpdateGenerator struct {
	AzureService
}

func (g *IoTDeviceUpdateGenerator) InitResources() error {
	ctx := context.Background()
	sub := g.getSubscriptionID()
	cred := g.getCredential()
	opts := g.armClientOptions()

	accountsClient, err := armdeviceupdate.NewAccountsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	var accounts []*armdeviceupdate.Account
	if rg := g.getResourceGroup(); rg != "" {
		pager := accountsClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			accounts = append(accounts, page.Value...)
		}
	} else {
		pager := accountsClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			accounts = append(accounts, page.Value...)
		}
	}

	instancesClient, err := armdeviceupdate.NewInstancesClient(sub, cred, opts)
	if err != nil {
		return err
	}
	for _, account := range accounts {
		g.AppendSimpleResource(*account.ID, *account.Name, "azurerm_iothub_device_update_account")
		parsed, err := ParseAzureResourceID(*account.ID)
		if err != nil {
			return err
		}
		pager := instancesClient.NewListByAccountPager(parsed.ResourceGroup, *account.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_iothub_device_update_instance")
			}
		}
	}
	return nil
}

// IoTCentralGenerator covers IoT Central applications.
// Skipped in this family:
//   - azurerm_iotcentral_application_network_rule_set: an embedded property
//     (properties.networkRuleSets) of the application itself, with no
//     independent ARM object or List.
//   - azurerm_iotcentral_organization: organizations live behind the IoT
//     Central data plane (per-application endpoint), with no ARM List in
//     the track-2 armiotcentral module.
type IoTCentralGenerator struct {
	AzureService
}

func (g *IoTCentralGenerator) InitResources() error {
	ctx := context.Background()
	client, err := armiotcentral.NewAppsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_iotcentral_application")
			}
		}
		return nil
	}
	pager := client.NewListBySubscriptionPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, item := range page.Value {
			g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_iotcentral_application")
		}
	}
	return nil
}
