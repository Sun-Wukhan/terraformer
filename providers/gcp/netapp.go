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

package gcp

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"google.golang.org/api/compute/v1"
	netappv1 "google.golang.org/api/netapp/v1"
)

var netappAllowEmptyValues = []string{""}

var netappAdditionalFields = map[string]interface{}{}

const netappBaseURL = "https://netapp.googleapis.com/v1/"

// NetappGenerator enumerates Google Cloud NetApp Volumes in the
// configured region: active directories, KMS configs, storage
// pools, backup policies, backup vaults with their backups, volumes
// with their quota rules, replications and snapshots — plus host
// groups over raw REST (the pinned netapp/v1 client has no host
// groups service; the provider ID is the full resource name,
// projects/{project}/locations/{location}/hostGroups/{name}).
type NetappGenerator struct {
	GCPService
}

func (g *NetappGenerator) InitResources() error {
	ctx := context.Background()
	netappService, err := netappv1.NewService(ctx)
	if err != nil {
		return err
	}
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, netappService, restClient)
}

func (g *NetappGenerator) initResources(ctx context.Context, netappService *netappv1.Service, restClient *http.Client) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + region
	baseAttrs := map[string]string{"project": project, "location": region}

	err := netappService.Projects.Locations.ActiveDirectories.List(parent).Pages(ctx, func(page *netappv1.ListActiveDirectoriesResponse) error {
		for _, item := range page.ActiveDirectories {
			g.Resources = append(g.Resources, g.newNetappResource(item.Name, "google_netapp_active_directory", baseAttrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = netappService.Projects.Locations.KmsConfigs.List(parent).Pages(ctx, func(page *netappv1.ListKmsConfigsResponse) error {
		for _, item := range page.KmsConfigs {
			g.Resources = append(g.Resources, g.newNetappResource(item.Name, "google_netapp_kmsconfig", baseAttrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = netappService.Projects.Locations.StoragePools.List(parent).Pages(ctx, func(page *netappv1.ListStoragePoolsResponse) error {
		for _, item := range page.StoragePools {
			g.Resources = append(g.Resources, g.newNetappResource(item.Name, "google_netapp_storage_pool", baseAttrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = netappService.Projects.Locations.BackupPolicies.List(parent).Pages(ctx, func(page *netappv1.ListBackupPoliciesResponse) error {
		for _, item := range page.BackupPolicies {
			g.Resources = append(g.Resources, g.newNetappResource(item.Name, "google_netapp_backup_policy", baseAttrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = netappService.Projects.Locations.BackupVaults.List(parent).Pages(ctx, func(page *netappv1.ListBackupVaultsResponse) error {
		for _, vault := range page.BackupVaults {
			g.Resources = append(g.Resources, g.newNetappResource(vault.Name, "google_netapp_backup_vault", baseAttrs))
			if err := g.loadNetappBackups(ctx, netappService, vault.Name, baseAttrs); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = netappService.Projects.Locations.Volumes.List(parent).Pages(ctx, func(page *netappv1.ListVolumesResponse) error {
		for _, volume := range page.Volumes {
			g.Resources = append(g.Resources, g.newNetappResource(volume.Name, "google_netapp_volume", baseAttrs))
			if err := g.loadNetappVolumeChildren(ctx, netappService, volume.Name, baseAttrs); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	g.loadHostGroups(ctx, restClient, parent, baseAttrs)
	return nil
}

func (g *NetappGenerator) newNetappResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
	if attrs == nil {
		attrs = map[string]string{}
	}
	parts := strings.Split(id, "/")
	return terraformutils.NewResource(
		id,
		parts[len(parts)-1],
		resourceType,
		g.ProviderName,
		attrs,
		netappAllowEmptyValues,
		netappAdditionalFields,
	)
}

func (g *NetappGenerator) loadNetappBackups(ctx context.Context, netappService *netappv1.Service, vaultName string, baseAttrs map[string]string) error {
	return netappService.Projects.Locations.BackupVaults.Backups.List(vaultName).Pages(ctx, func(page *netappv1.ListBackupsResponse) error {
		for _, item := range page.Backups {
			g.Resources = append(g.Resources, g.newNetappResource(item.Name, "google_netapp_backup", baseAttrs))
		}
		return nil
	})
}

func (g *NetappGenerator) loadNetappVolumeChildren(ctx context.Context, netappService *netappv1.Service, volumeName string, baseAttrs map[string]string) error {
	err := netappService.Projects.Locations.Volumes.QuotaRules.List(volumeName).Pages(ctx, func(page *netappv1.ListQuotaRulesResponse) error {
		for _, item := range page.QuotaRules {
			g.Resources = append(g.Resources, g.newNetappResource(item.Name, "google_netapp_volume_quota_rule", baseAttrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = netappService.Projects.Locations.Volumes.Replications.List(volumeName).Pages(ctx, func(page *netappv1.ListReplicationsResponse) error {
		for _, item := range page.Replications {
			g.Resources = append(g.Resources, g.newNetappResource(item.Name, "google_netapp_volume_replication", baseAttrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	return netappService.Projects.Locations.Volumes.Snapshots.List(volumeName).Pages(ctx, func(page *netappv1.ListSnapshotsResponse) error {
		for _, item := range page.Snapshots {
			g.Resources = append(g.Resources, g.newNetappResource(item.Name, "google_netapp_volume_snapshot", baseAttrs))
		}
		return nil
	})
}

func (g *NetappGenerator) loadHostGroups(ctx context.Context, restClient *http.Client, parent string, baseAttrs map[string]string) {
	items, err := restListResources(ctx, restClient, netappBaseURL+parent+"/hostGroups")
	if err != nil {
		log.Println(err)
		return
	}
	for _, item := range items {
		g.Resources = append(g.Resources, g.newNetappResource(item.Name, "google_netapp_host_group", baseAttrs))
	}
}
