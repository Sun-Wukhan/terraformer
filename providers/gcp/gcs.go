// Copyright 2018 The Terraformer Authors.
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
	"fmt"
	"log"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"

	cloudresourcemanagerv1 "google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/storage/v1"
	storagebatchoperationsv1 "google.golang.org/api/storagebatchoperations/v1"
)

var GcsAllowEmptyValues = []string{"labels.", "created_before"}

var GcsAdditionalFields = map[string]interface{}{}

type GcsGenerator struct {
	GCPService
}

func (g *GcsGenerator) createBucketsResources(ctx context.Context, gcsService *storage.Service) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	bucketList := gcsService.Buckets.List(g.GetArgs()["project"].(string))
	if err := bucketList.Pages(ctx, func(page *storage.Buckets) error {
		for _, bucket := range page.Items {
			resources = append(resources, terraformutils.NewResource(
				bucket.Name,
				bucket.Name,
				"google_storage_bucket",
				g.ProviderName,
				map[string]string{
					"name":          bucket.Name,
					"force_destroy": "false",
				},
				GcsAllowEmptyValues,
				GcsAdditionalFields,
			))
			// The whole-bucket ACL resources (google_storage_bucket_acl
			// and google_storage_default_object_acl) are deliberately not
			// emitted: the google provider has no importer for them
			// (verified against provider v8.11), so every bucket produced
			// two guaranteed import failures. The same ACL data is still
			// imported per entity via the
			// google_storage_bucket_access_control and
			// google_storage_default_object_access_control resources in
			// the storage depth pass, which do support import.
			resources = append(resources, terraformutils.NewResource(
				bucket.Name,
				bucket.Name,
				"google_storage_bucket_iam_policy",
				g.ProviderName,
				map[string]string{
					"bucket": bucket.Name,
				},
				GcsAllowEmptyValues,
				GcsAdditionalFields,
			))

			if iam, err := gcsService.Buckets.GetIamPolicy(bucket.Name).Do(); err == nil {
				for _, binding := range iam.Bindings {
					resources = append(resources, terraformutils.NewResource(
						bucket.Name,
						bucket.Name,
						"google_storage_bucket_iam_binding",
						g.ProviderName,
						map[string]string{
							"bucket": bucket.Name,
							"role":   binding.Role,
						},
						GcsAllowEmptyValues,
						GcsAdditionalFields,
					))

					for _, member := range binding.Members {
						resources = append(resources, terraformutils.NewResource(
							bucket.Name,
							bucket.Name,
							"google_storage_bucket_iam_member",
							g.ProviderName,
							map[string]string{
								"bucket": bucket.Name,
								"role":   binding.Role,
								"member": member,
							},
							GcsAllowEmptyValues,
							GcsAdditionalFields,
						))
					}
				}
			}

			resources = append(resources, g.createNotificationResources(gcsService, bucket)...)
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return resources
}

func (g *GcsGenerator) createNotificationResources(gcsService *storage.Service, bucket *storage.Bucket) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	notificationList, err := gcsService.Notifications.List(bucket.Name).Do()
	if err != nil {
		log.Println(err)
		return resources
	}
	for _, notification := range notificationList.Items {
		resources = append(resources, terraformutils.NewResource(
			bucket.Name+"/notificationConfigs/"+notification.Id,
			bucket.Name+"/"+notification.Id,
			"google_storage_notification",
			g.ProviderName,
			map[string]string{},
			GcsAllowEmptyValues,
			GcsAdditionalFields,
		))
	}
	return resources
}

/*
func (g *GcsGenerator) createTransferJobsResources(ctx context.Context, storageTransferService *storagetransfer.Service) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	transferJobsList := storageTransferService.TransferJobs.List()
	err := transferJobsList.Pages(ctx, func(page *storagetransfer.ListTransferJobsResponse) error {
		log.Println(page.TransferJobs)
		for _, transferJob := range page.TransferJobs {
			resources = append(resources, terraformutils.NewResource(
				transferJob.Name,
				transferJob.Name,
				"google_storage_transfer_job",
				g.ProviderName,
				map[string]string{
					"name": transferJob.Name,
				},
				GcsAllowEmptyValues,
				GcsAdditionalFields,
			))
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	return resources
}
*/

// Generate TerraformResources from GCP API,
// from each bucket  create 1 TerraformResource
// Need bucket name as ID for terraform resource
func (g *GcsGenerator) InitResources() error {
	ctx := context.Background()
	gcsService, err := storage.NewService(ctx)
	if err != nil {
		log.Print(err)
		return err
	}
	g.Resources = g.createBucketsResources(ctx, gcsService)

	// Depth collections (storage_depth.go): per-bucket ACLs, folders,
	// managed folders, objects and anywhere caches, HMAC keys, batch
	// operations jobs, Storage Control configs and FTP servers. A
	// satellite client that fails to construct only disables its own
	// section.
	batchService, err := storagebatchoperationsv1.NewService(ctx)
	if err != nil {
		log.Println(err)
		batchService = nil
	}
	crmService, err := cloudresourcemanagerv1.NewService(ctx)
	if err != nil {
		log.Println(err)
		crmService = nil
	}
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		log.Println(err)
		restClient = nil
	}
	g.initStorageDepthResources(ctx, gcsService, batchService, crmService, restClient)

	// TODO find bug with storageTransferService.TransferJobs.List().Pages
	// storageTransferService, err := storagetransfer.NewService(ctx)
	// if err != nil {
	// 	log.Print(err)
	// 		return err
	// 	}
	// g.Resources = append(g.Resources, g.createTransferJobsResources(ctx, storageTransferService)...)
	return nil
}

// PostGenerateHook for add bucket policy json as heredoc
// support only bucket with policy
func (g *GcsGenerator) PostConvertHook() error {
	for i, resource := range g.Resources {
		if resource.InstanceInfo.Type != "google_storage_bucket_iam_policy" {
			continue
		}
		if _, exist := resource.Item["policy_data"]; exist {
			policy := resource.Item["policy_data"].(string)
			g.Resources[i].Item["policy_data"] = fmt.Sprintf(`<<POLICY
%s
POLICY`, policy)
		}
	}
	return nil
}
