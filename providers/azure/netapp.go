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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/netapp/armnetapp"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// NetAppGenerator covers Azure NetApp Files via armnetapp: accounts, capacity
// pools, volumes, snapshots, snapshot policies, backup policies, backup
// vaults and application volume groups (split by ApplicationType).
//
// Skipped in this family (see docs/azure.md "Known limits"):
//   - azurerm_netapp_account_encryption imports by the NetApp account ID
//     itself; it is a singleton property of the account object.
//   - azurerm_netapp_volume_bucket and
//     azurerm_netapp_volume_bucket_with_server: armnetapp v1.0.0 has no
//     Buckets client/List.
//   - azurerm_netapp_volume_quota_rule: armnetapp v1.0.0 has no
//     VolumeQuotaRules client/List.
type NetAppGenerator struct {
	AzureService
}

// netappVolumeGroupType maps an armnetapp ApplicationType to the azurerm
// volume-group resource type, or "" when it is neither SAP HANA nor Oracle.
func netappVolumeGroupType(applicationType string) string {
	switch {
	case strings.EqualFold(applicationType, "SAP-HANA"):
		return "azurerm_netapp_volume_group_sap_hana"
	case strings.EqualFold(applicationType, "Oracle"):
		return "azurerm_netapp_volume_group_oracle"
	default:
		return ""
	}
}

func (g *NetAppGenerator) InitResources() error {
	ctx := context.Background()
	sub := g.getSubscriptionID()
	opts := g.armClientOptions()
	var resources []terraformutils.Resource
	add := func(id, name, typ string) {
		if id != "" && name != "" {
			resources = append(resources, terraformutils.NewSimpleResource(id, name, typ, "azurerm", []string{}))
		}
	}

	accountsClient, err := armnetapp.NewAccountsClient(sub, g.getCredential(), opts)
	if err != nil {
		return err
	}
	poolsClient, err := armnetapp.NewPoolsClient(sub, g.getCredential(), opts)
	if err != nil {
		return err
	}
	volumesClient, err := armnetapp.NewVolumesClient(sub, g.getCredential(), opts)
	if err != nil {
		return err
	}
	snapshotsClient, err := armnetapp.NewSnapshotsClient(sub, g.getCredential(), opts)
	if err != nil {
		return err
	}
	snapPoliciesClient, err := armnetapp.NewSnapshotPoliciesClient(sub, g.getCredential(), opts)
	if err != nil {
		return err
	}
	backupPoliciesClient, err := armnetapp.NewBackupPoliciesClient(sub, g.getCredential(), opts)
	if err != nil {
		return err
	}
	vaultsClient, err := armnetapp.NewVaultsClient(sub, g.getCredential(), opts)
	if err != nil {
		return err
	}
	volumeGroupsClient, err := armnetapp.NewVolumeGroupsClient(sub, g.getCredential(), opts)
	if err != nil {
		return err
	}

	type accountRef struct{ rg, name string }
	var accounts []accountRef
	if rg := g.getResourceGroup(); rg != "" {
		pager := accountsClient.NewListPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, a := range page.Value {
				if a.ID != nil && a.Name != nil {
					add(*a.ID, *a.Name, "azurerm_netapp_account")
					accounts = append(accounts, accountRef{rg, *a.Name})
				}
			}
		}
	} else {
		pager := accountsClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, a := range page.Value {
				if a.ID == nil || a.Name == nil {
					continue
				}
				add(*a.ID, *a.Name, "azurerm_netapp_account")
				if parsed, err := ParseAzureResourceID(*a.ID); err == nil {
					accounts = append(accounts, accountRef{parsed.ResourceGroup, *a.Name})
				}
			}
		}
	}

	for _, acct := range accounts {
		bpPager := backupPoliciesClient.NewListPager(acct.rg, acct.name, nil)
		for bpPager.More() {
			page, err := bpPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_netapp_backup_policy")
				}
			}
		}
		vPager := vaultsClient.NewListPager(acct.rg, acct.name, nil)
		for vPager.More() {
			page, err := vPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_netapp_backup_vault")
				}
			}
		}
		spPager := snapPoliciesClient.NewListPager(acct.rg, acct.name, nil)
		for spPager.More() {
			page, err := spPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_netapp_snapshot_policy")
				}
			}
		}
		vgPager := volumeGroupsClient.NewListByNetAppAccountPager(acct.rg, acct.name, nil)
		for vgPager.More() {
			page, err := vgPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID == nil || item.Name == nil || item.Properties == nil || item.Properties.GroupMetaData == nil || item.Properties.GroupMetaData.ApplicationType == nil {
					continue
				}
				if typ := netappVolumeGroupType(string(*item.Properties.GroupMetaData.ApplicationType)); typ != "" {
					add(*item.ID, *item.Name, typ)
				}
			}
		}
		pPager := poolsClient.NewListPager(acct.rg, acct.name, nil)
		for pPager.More() {
			page, err := pPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, pool := range page.Value {
				if pool.ID == nil || pool.Name == nil {
					continue
				}
				add(*pool.ID, *pool.Name, "azurerm_netapp_pool")
				volPager := volumesClient.NewListPager(acct.rg, acct.name, *pool.Name, nil)
				for volPager.More() {
					vPage, err := volPager.NextPage(ctx)
					if err != nil {
						log.Println(err)
						break
					}
					for _, vol := range vPage.Value {
						if vol.ID == nil || vol.Name == nil {
							continue
						}
						add(*vol.ID, *vol.Name, "azurerm_netapp_volume")
						snPager := snapshotsClient.NewListPager(acct.rg, acct.name, *pool.Name, *vol.Name, nil)
						for snPager.More() {
							sPage, err := snPager.NextPage(ctx)
							if err != nil {
								log.Println(err)
								break
							}
							for _, snap := range sPage.Value {
								if snap.ID != nil && snap.Name != nil {
									add(*snap.ID, *snap.Name, "azurerm_netapp_snapshot")
								}
							}
						}
					}
				}
			}
		}
	}
	g.Resources = resources
	return nil
}
