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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// ComputeGenerator covers the Microsoft.Compute top-level resources that are
// not part of the pre-existing virtual_machine / scaleset / disk services:
// availability sets, capacity reservations, dedicated hosts, disk access and
// encryption sets, images, proximity placement groups, snapshots, restore
// point collections and the shared image gallery family (galleries, images,
// versions, gallery applications and their versions).
//
// Skipped in this family:
//   - azurerm_virtual_machine_restore_point: the pinned armcompute
//     RestorePointsClient exposes only Get/Create/Delete, no List pager.
type ComputeGenerator struct {
	AzureService
}

func (g *ComputeGenerator) appendSimple(id, name, resourceType string) {
	if id == "" || name == "" {
		return
	}
	g.Resources = append(g.Resources, terraformutils.NewSimpleResource(id, name, resourceType, "azurerm", []string{}))
}

func resourceGroupFromID(id string) string {
	parsed, err := ParseAzureResourceID(id)
	if err != nil {
		return ""
	}
	return parsed.ResourceGroup
}

func (g *ComputeGenerator) InitResources() error {
	ctx := context.Background()
	subscriptionID := g.getSubscriptionID()
	cred := g.getCredential()
	opts := g.armClientOptions()
	rg := g.getResourceGroup()

	// Availability sets
	if client, err := armcompute.NewAvailabilitySetsClient(subscriptionID, cred, opts); err != nil {
		return err
	} else if rg != "" {
		pager := client.NewListPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_availability_set")
			}
		}
	} else {
		pager := client.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_availability_set")
			}
		}
	}

	// Capacity reservation groups + reservations (children)
	crgClient, err := armcompute.NewCapacityReservationGroupsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	crClient, err := armcompute.NewCapacityReservationsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	var crGroups []*armcompute.CapacityReservationGroup
	if rg != "" {
		pager := crgClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			crGroups = append(crGroups, page.Value...)
		}
	} else {
		pager := crgClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			crGroups = append(crGroups, page.Value...)
		}
	}
	for _, group := range crGroups {
		g.appendSimple(*group.ID, *group.Name, "azurerm_capacity_reservation_group")
		groupRG := resourceGroupFromID(*group.ID)
		pager := crClient.NewListByCapacityReservationGroupPager(groupRG, *group.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_capacity_reservation")
			}
		}
	}

	// Dedicated host groups + hosts (children)
	dhgClient, err := armcompute.NewDedicatedHostGroupsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	dhClient, err := armcompute.NewDedicatedHostsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	var dhGroups []*armcompute.DedicatedHostGroup
	if rg != "" {
		pager := dhgClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			dhGroups = append(dhGroups, page.Value...)
		}
	} else {
		pager := dhgClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			dhGroups = append(dhGroups, page.Value...)
		}
	}
	for _, group := range dhGroups {
		g.appendSimple(*group.ID, *group.Name, "azurerm_dedicated_host_group")
		groupRG := resourceGroupFromID(*group.ID)
		pager := dhClient.NewListByHostGroupPager(groupRG, *group.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_dedicated_host")
			}
		}
	}

	// Disk accesses
	if client, err := armcompute.NewDiskAccessesClient(subscriptionID, cred, opts); err != nil {
		return err
	} else if rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_disk_access")
			}
		}
	} else {
		pager := client.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_disk_access")
			}
		}
	}

	// Disk encryption sets
	if client, err := armcompute.NewDiskEncryptionSetsClient(subscriptionID, cred, opts); err != nil {
		return err
	} else if rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_disk_encryption_set")
			}
		}
	} else {
		pager := client.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_disk_encryption_set")
			}
		}
	}

	// Images
	if client, err := armcompute.NewImagesClient(subscriptionID, cred, opts); err != nil {
		return err
	} else if rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_image")
			}
		}
	} else {
		pager := client.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_image")
			}
		}
	}

	// Proximity placement groups
	if client, err := armcompute.NewProximityPlacementGroupsClient(subscriptionID, cred, opts); err != nil {
		return err
	} else if rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_proximity_placement_group")
			}
		}
	} else {
		pager := client.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_proximity_placement_group")
			}
		}
	}

	// Snapshots
	if client, err := armcompute.NewSnapshotsClient(subscriptionID, cred, opts); err != nil {
		return err
	} else if rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_snapshot")
			}
		}
	} else {
		pager := client.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_snapshot")
			}
		}
	}

	// Restore point collections (restore points themselves have no List
	// pager in the pinned SDK, see the type comment above).
	if client, err := armcompute.NewRestorePointCollectionsClient(subscriptionID, cred, opts); err != nil {
		return err
	} else if rg != "" {
		pager := client.NewListPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_virtual_machine_restore_point_collection")
			}
		}
	} else {
		pager := client.NewListAllPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_virtual_machine_restore_point_collection")
			}
		}
	}

	// Shared image galleries + images/versions + gallery applications/versions
	galleryClient, err := armcompute.NewGalleriesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	giClient, err := armcompute.NewGalleryImagesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	givClient, err := armcompute.NewGalleryImageVersionsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	gaClient, err := armcompute.NewGalleryApplicationsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	gavClient, err := armcompute.NewGalleryApplicationVersionsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	var galleries []*armcompute.Gallery
	if rg != "" {
		pager := galleryClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			galleries = append(galleries, page.Value...)
		}
	} else {
		pager := galleryClient.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			galleries = append(galleries, page.Value...)
		}
	}
	for _, gallery := range galleries {
		g.appendSimple(*gallery.ID, *gallery.Name, "azurerm_shared_image_gallery")
		galleryRG := resourceGroupFromID(*gallery.ID)
		// Gallery images + versions
		pager := giClient.NewListByGalleryPager(galleryRG, *gallery.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, image := range page.Value {
				g.appendSimple(*image.ID, *image.Name, "azurerm_shared_image")
				vpager := givClient.NewListByGalleryImagePager(galleryRG, *gallery.Name, *image.Name, nil)
				for vpager.More() {
					vpage, err := vpager.NextPage(ctx)
					if err != nil {
						log.Println(err)
						return err
					}
					for _, version := range vpage.Value {
						g.appendSimple(*version.ID, *version.Name, "azurerm_shared_image_version")
					}
				}
			}
		}
		// Gallery applications + versions
		apager := gaClient.NewListByGalleryPager(galleryRG, *gallery.Name, nil)
		for apager.More() {
			page, err := apager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, app := range page.Value {
				g.appendSimple(*app.ID, *app.Name, "azurerm_gallery_application")
				vpager := gavClient.NewListByGalleryApplicationPager(galleryRG, *gallery.Name, *app.Name, nil)
				for vpager.More() {
					vpage, err := vpager.NextPage(ctx)
					if err != nil {
						log.Println(err)
						return err
					}
					for _, version := range vpage.Value {
						g.appendSimple(*version.ID, *version.Name, "azurerm_gallery_application_version")
					}
				}
			}
		}
	}
	return nil
}
