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

	cloudresourcemanagerv1 "google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/compute/v1"
	storagev1 "google.golang.org/api/storage/v1"
	storagebatchoperationsv1 "google.golang.org/api/storagebatchoperations/v1"
)

// Storage depth enumeration (Batch 4), extending the hand-written
// GCS generator the way bigquery_depth.go extends BigQuery: per
// bucket it adds Anywhere caches, bucket and default-object ACL
// entries, folders, managed folders (with their combined IAM
// resource), objects with their ACL entries and whole-object ACLs;
// per project it adds HMAC keys and Storage Batch Operations jobs.
//
// Storage Control singletons and collections (intelligence configs
// for the project, its ancestor folders and its organization, and
// project feature configs) and the Storage FTP service (SFTP
// servers and their users) have no generated client surface in the
// pinned module and run over raw REST through gcp_rest.go, with
// endpoints and IDs from the provider's Magic Modules definitions
// (storagecontrol and storageftp products).
//
// Skipped, with reason:
//   - google_storage_bucket_iam: the legacy combined IAM resource
//     covers the same bucket IAM policy object, with the same ID
//     (the bucket name), as the google_storage_bucket_iam_policy /
//     _binding / _member resources the GCS generator already emits —
//     a duplicate over the same API object and ID.

const (
	storageControlBaseURL = "https://storage.googleapis.com/storage/v1/"
	storageFTPBaseURL     = "https://ftp.googleapis.com/v1/"
)

func (g *GcsGenerator) newStorageDepthResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
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
		GcsAllowEmptyValues,
		GcsAdditionalFields,
	)
}

// initStorageDepthResources runs the depth enumeration. Any client
// may be nil (construction failure in InitResources), disabling only
// its own section.
func (g *GcsGenerator) initStorageDepthResources(ctx context.Context, storageService *storagev1.Service, batchService *storagebatchoperationsv1.Service, crmService *cloudresourcemanagerv1.Service, restClient *http.Client) {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	if storageService != nil {
		bucketList := storageService.Buckets.List(project)
		if err := bucketList.Pages(ctx, func(page *storagev1.Buckets) error {
			for _, bucket := range page.Items {
				g.storageBucketDepth(ctx, storageService, bucket.Name)
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
		g.storageHmacKeys(ctx, storageService, project)
	}
	if batchService != nil {
		call := batchService.Projects.Locations.Jobs.List("projects/" + project + "/locations/global")
		if err := call.Pages(ctx, func(page *storagebatchoperationsv1.ListJobsResponse) error {
			for _, job := range page.Jobs {
				g.Resources = append(g.Resources, g.newStorageDepthResource(
					job.Name,
					"google_storage_batch_operations_job",
					nil,
				))
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
	}
	if restClient != nil {
		g.storageControlConfigs(ctx, restClient, crmService, project)
		g.storageFTPServers(ctx, restClient, project, region)
	}
}

// storageBucketDepth enumerates the depth collections of one bucket.
func (g *GcsGenerator) storageBucketDepth(ctx context.Context, storageService *storagev1.Service, bucket string) {
	anywhereCall := storageService.AnywhereCaches.List(bucket)
	if err := anywhereCall.Pages(ctx, func(page *storagev1.AnywhereCaches) error {
		for _, cache := range page.Items {
			g.Resources = append(g.Resources, g.newStorageDepthResource(
				bucket+"/"+cache.AnywhereCacheId,
				"google_storage_anywhere_cache",
				map[string]string{"bucket": bucket},
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if acls, err := storageService.BucketAccessControls.List(bucket).Context(ctx).Do(); err == nil {
		for _, acl := range acls.Items {
			g.Resources = append(g.Resources, g.newStorageDepthResource(
				bucket+"/"+acl.Entity,
				"google_storage_bucket_access_control",
				map[string]string{"bucket": bucket, "entity": acl.Entity},
			))
		}
	} else {
		log.Println(err)
	}
	if acls, err := storageService.DefaultObjectAccessControls.List(bucket).Context(ctx).Do(); err == nil {
		for _, acl := range acls.Items {
			g.Resources = append(g.Resources, g.newStorageDepthResource(
				bucket+"/"+acl.Entity,
				"google_storage_default_object_access_control",
				map[string]string{"bucket": bucket, "entity": acl.Entity},
			))
		}
	} else {
		log.Println(err)
	}
	foldersCall := storageService.Folders.List(bucket)
	if err := foldersCall.Pages(ctx, func(page *storagev1.Folders) error {
		for _, folder := range page.Items {
			g.Resources = append(g.Resources, g.newStorageDepthResource(
				bucket+"/"+folder.Name,
				"google_storage_folder",
				map[string]string{"bucket": bucket},
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	managedCall := storageService.ManagedFolders.List(bucket)
	if err := managedCall.Pages(ctx, func(page *storagev1.ManagedFolders) error {
		for _, folder := range page.Items {
			id := bucket + "/" + folder.Name
			g.Resources = append(g.Resources, g.newStorageDepthResource(
				id,
				"google_storage_managed_folder",
				map[string]string{"bucket": bucket},
			))
			// The combined managed-folder IAM resource shares the
			// folder's ID; emit it when the folder has a policy.
			if _, err := storageService.ManagedFolders.GetIamPolicy(bucket, folder.Name).Context(ctx).Do(); err == nil {
				g.Resources = append(g.Resources, g.newStorageDepthResource(
					id,
					"google_storage_managed_folder_iam",
					map[string]string{"bucket": bucket, "managed_folder": folder.Name},
				))
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	g.storageBucketObjects(ctx, storageService, bucket)
}

// storageBucketObjects enumerates a bucket's objects, each object's
// ACL entries, and the whole-object ACL resource per object that has
// ACL entries.
func (g *GcsGenerator) storageBucketObjects(ctx context.Context, storageService *storagev1.Service, bucket string) {
	objectsCall := storageService.Objects.List(bucket)
	if err := objectsCall.Pages(ctx, func(page *storagev1.Objects) error {
		for _, object := range page.Items {
			objectID := bucket + "/" + object.Name
			g.Resources = append(g.Resources, g.newStorageDepthResource(
				objectID,
				"google_storage_bucket_object",
				map[string]string{"bucket": bucket, "name": object.Name},
			))
			acls, err := storageService.ObjectAccessControls.List(bucket, object.Name).Context(ctx).Do()
			if err != nil {
				log.Println(err)
				continue
			}
			if len(acls.Items) > 0 {
				g.Resources = append(g.Resources, g.newStorageDepthResource(
					objectID,
					"google_storage_object_acl",
					map[string]string{"bucket": bucket, "object": object.Name},
				))
			}
			for _, acl := range acls.Items {
				g.Resources = append(g.Resources, g.newStorageDepthResource(
					objectID+"/"+acl.Entity,
					"google_storage_object_access_control",
					map[string]string{"bucket": bucket, "object": object.Name, "entity": acl.Entity},
				))
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
}

// storageHmacKeys enumerates the project's HMAC keys. The provider
// ID is projects/{project}/hmacKeys/{accessId}.
func (g *GcsGenerator) storageHmacKeys(ctx context.Context, storageService *storagev1.Service, project string) {
	call := storageService.Projects.HmacKeys.List(project)
	if err := call.Pages(ctx, func(page *storagev1.HmacKeysMetadata) error {
		for _, key := range page.Items {
			g.Resources = append(g.Resources, g.newStorageDepthResource(
				"projects/"+project+"/hmacKeys/"+key.AccessId,
				"google_storage_hmac_key",
				map[string]string{"project": project},
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
}

// storageControlConfigs probes the Storage Control intelligence
// config singletons for the project, each ancestor folder and the
// organization (ancestry via Cloud Resource Manager), and lists the
// project's feature configs.
func (g *GcsGenerator) storageControlConfigs(ctx context.Context, restClient *http.Client, crmService *cloudresourcemanagerv1.Service, project string) {
	projectConfig := "projects/" + project + "/locations/global/intelligenceConfig"
	if _, err := restDo(ctx, restClient, http.MethodGet, storageControlBaseURL+projectConfig, nil); err == nil {
		g.Resources = append(g.Resources, g.newStorageDepthResource(
			projectConfig,
			"google_storage_control_project_intelligence_config",
			nil,
		))
	}
	featureConfigs, err := restListResources(ctx, restClient, storageControlBaseURL+"projects/"+project+"/locations/global/featureConfigs")
	if err != nil {
		log.Println(err)
	}
	for _, config := range featureConfigs {
		g.Resources = append(g.Resources, g.newStorageDepthResource(
			config.Name,
			"google_storage_control_project_feature_config",
			nil,
		))
	}
	if crmService == nil {
		return
	}
	ancestry, err := crmService.Projects.GetAncestry(project, &cloudresourcemanagerv1.GetAncestryRequest{}).Context(ctx).Do()
	if err != nil {
		log.Println(err)
		return
	}
	for _, ancestor := range ancestry.Ancestor {
		if ancestor.ResourceId == nil {
			continue
		}
		var configName, resourceType string
		switch ancestor.ResourceId.Type {
		case "folder":
			configName = "folders/" + ancestor.ResourceId.Id + "/locations/global/intelligenceConfig"
			resourceType = "google_storage_control_folder_intelligence_config"
		case "organization":
			configName = "organizations/" + ancestor.ResourceId.Id + "/locations/global/intelligenceConfig"
			resourceType = "google_storage_control_organization_intelligence_config"
		default:
			continue
		}
		if _, err := restDo(ctx, restClient, http.MethodGet, storageControlBaseURL+configName, nil); err == nil {
			g.Resources = append(g.Resources, g.newStorageDepthResource(
				configName,
				resourceType,
				nil,
			))
		}
	}
}

// storageFTPServers lists Storage FTP (SFTP) servers in the
// configured region and each server's users, over raw REST
// (ftp.googleapis.com/v1 — no generated client).
func (g *GcsGenerator) storageFTPServers(ctx context.Context, restClient *http.Client, project, region string) {
	servers, err := restListResources(ctx, restClient, storageFTPBaseURL+"projects/"+project+"/locations/"+region+"/servers")
	if err != nil {
		log.Println(err)
		return
	}
	for _, server := range servers {
		g.Resources = append(g.Resources, g.newStorageDepthResource(
			server.Name,
			"google_storage_ftp_server",
			nil,
		))
		users, err := restListResources(ctx, restClient, storageFTPBaseURL+server.Name+"/users")
		if err != nil {
			log.Println(err)
			continue
		}
		for _, user := range users {
			g.Resources = append(g.Resources, g.newStorageDepthResource(
				user.Name,
				"google_storage_ftp_user",
				nil,
			))
		}
	}
}
