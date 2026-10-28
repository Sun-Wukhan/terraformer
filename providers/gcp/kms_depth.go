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
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"google.golang.org/api/cloudkms/v1"
	"google.golang.org/api/cloudresourcemanager/v1"
)

const kmsBaseURL = "https://cloudkms.googleapis.com/v1/"

// initKmsDepthResources is the KMS depth pass (the kms_depth.go sibling
// of kms.go, following the storage_depth.go pattern): for every KMS
// location of the project it enumerates EKM connections and key handles,
// and for every key ring the ring's IAM trio, its import jobs, and for
// every crypto key the key's IAM trio and its crypto key versions. It
// also reads the project-scoped singleton configs (project Autokey
// config and project KAJ policy config) and — when the project's
// organization resolves through Cloud Resource Manager ancestry — the
// organization KAJ policy config. Those singletons are read over raw
// REST: the pinned cloudkms client surfaces folders.getAutokeyConfig
// only, while projects.getAutokeyConfig and the getKajPolicyConfig
// methods exist in the live v1 API (verified against the Cloud KMS
// REST reference).
//
// Deliberately not generated:
//   - google_kms_autokey_config: the folder-scoped Autokey singleton
//     (folders/{folder}/autokeyConfig). A project's folders are not
//     discoverable from the project scope available to this import, so
//     the collection cannot be enumerated.
//   - google_kms_folder_kaj_policy_config: the folder-scoped KAJ
//     policy singleton (folders/{folder}/kajPolicyConfig); same
//     undiscoverable-folder reason.
//   - google_kms_secret_ciphertext: a local-only helper resource that
//     encrypts a plaintext at apply time. It has no API object, no ID
//     in Cloud KMS and no import support; nothing exists server-side
//     to enumerate.
func (g *KmsGenerator) initKmsDepthResources(ctx context.Context, kmsService *cloudkms.Service, crmService *cloudresourcemanager.Service, restClient *http.Client) {
	project := g.GetArgs()["project"].(string)
	projectParent := "projects/" + project

	err := kmsService.Projects.Locations.List(projectParent).Pages(ctx, func(page *cloudkms.ListLocationsResponse) error {
		for _, location := range page.Locations {
			g.loadKmsLocation(ctx, kmsService, location.Name, project)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}

	if restClient != nil {
		if name, err := restGetSingletonName(ctx, restClient, kmsBaseURL+projectParent+"/autokeyConfig"); err != nil {
			log.Println(err)
		} else if name != "" {
			g.Resources = append(g.Resources, g.newKmsDepthResource(name, "google_kms_project_autokey_config", map[string]string{"project": project}))
		}
		if name, err := restGetSingletonName(ctx, restClient, kmsBaseURL+projectParent+"/kajPolicyConfig"); err != nil {
			log.Println(err)
		} else if name != "" {
			g.Resources = append(g.Resources, g.newKmsDepthResource(name, "google_kms_project_kaj_policy_config", map[string]string{"project": project}))
		}
		if orgID := orgIDFromAncestry(ctx, crmService, project); orgID != "" {
			orgName := "organizations/" + orgID
			if name, err := restGetSingletonName(ctx, restClient, kmsBaseURL+orgName+"/kajPolicyConfig"); err != nil {
				log.Println(err)
			} else if name != "" {
				g.Resources = append(g.Resources, g.newKmsDepthResource(name, "google_kms_organization_kaj_policy_config", map[string]string{"organization": orgID}))
			}
		}
	}
}

// orgIDFromAncestry resolves the organization ID of a project through
// Cloud Resource Manager ancestry, following networksecurity.go.
func orgIDFromAncestry(ctx context.Context, crmService *cloudresourcemanager.Service, project string) string {
	if crmService == nil {
		return ""
	}
	ancestry, err := crmService.Projects.GetAncestry(project, &cloudresourcemanager.GetAncestryRequest{}).Context(ctx).Do()
	if err != nil {
		log.Println(err)
		return ""
	}
	orgID := ""
	for _, ancestor := range ancestry.Ancestor {
		if ancestor.ResourceId != nil && ancestor.ResourceId.Type == "organization" {
			orgID = ancestor.ResourceId.Id
		}
	}
	return orgID
}

// restGetSingletonName GETs a singleton resource over raw REST and
// returns its name (empty when the singleton is not present).
func restGetSingletonName(ctx context.Context, client *http.Client, url string) (string, error) {
	body, err := restDo(ctx, client, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	var singleton struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(body, &singleton); err != nil {
		return "", err
	}
	return singleton.Name, nil
}

func (g *KmsGenerator) newKmsDepthResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
	parts := strings.Split(id, "/")
	return terraformutils.NewResource(
		id,
		parts[len(parts)-1],
		resourceType,
		g.ProviderName,
		attrs,
		kmsAllowEmptyValues,
		kmsAdditionalFields,
	)
}

func (g *KmsGenerator) loadKmsLocation(ctx context.Context, kmsService *cloudkms.Service, locationName, project string) {
	parts := strings.Split(locationName, "/")
	location := parts[len(parts)-1]
	attrs := map[string]string{"project": project, "location": location}

	err := kmsService.Projects.Locations.EkmConnections.List(locationName).Pages(ctx, func(page *cloudkms.ListEkmConnectionsResponse) error {
		for _, connection := range page.EkmConnections {
			g.Resources = append(g.Resources, g.newKmsDepthResource(connection.Name, "google_kms_ekm_connection", attrs))
			g.loadKmsEkmConnectionIam(ctx, kmsService, connection.Name)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = kmsService.Projects.Locations.KeyHandles.List(locationName).Pages(ctx, func(page *cloudkms.ListKeyHandlesResponse) error {
		for _, handle := range page.KeyHandles {
			g.Resources = append(g.Resources, g.newKmsDepthResource(handle.Name, "google_kms_key_handle", attrs))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = kmsService.Projects.Locations.KeyRings.List(locationName).Pages(ctx, func(page *cloudkms.ListKeyRingsResponse) error {
		for _, ring := range page.KeyRings {
			g.loadKmsKeyRingIam(ctx, kmsService, ring.Name)
			g.loadKmsImportJobs(ctx, kmsService, ring.Name, attrs)
			g.loadKmsCryptoKeyDepth(ctx, kmsService, ring.Name, attrs)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (g *KmsGenerator) loadKmsImportJobs(ctx context.Context, kmsService *cloudkms.Service, ringName string, attrs map[string]string) {
	err := kmsService.Projects.Locations.KeyRings.ImportJobs.List(ringName).Pages(ctx, func(page *cloudkms.ListImportJobsResponse) error {
		for _, job := range page.ImportJobs {
			g.Resources = append(g.Resources, g.newKmsDepthResource(job.Name, "google_kms_key_ring_import_job", attrs))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (g *KmsGenerator) loadKmsCryptoKeyDepth(ctx context.Context, kmsService *cloudkms.Service, ringName string, attrs map[string]string) {
	err := kmsService.Projects.Locations.KeyRings.CryptoKeys.List(ringName).Pages(ctx, func(page *cloudkms.ListCryptoKeysResponse) error {
		for _, key := range page.CryptoKeys {
			g.loadKmsCryptoKeyIam(ctx, kmsService, key.Name)
			versionErr := kmsService.Projects.Locations.KeyRings.CryptoKeys.CryptoKeyVersions.List(key.Name).Pages(ctx, func(page *cloudkms.ListCryptoKeyVersionsResponse) error {
				for _, version := range page.CryptoKeyVersions {
					g.Resources = append(g.Resources, g.newKmsDepthResource(version.Name, "google_kms_crypto_key_version", attrs))
				}
				return nil
			})
			if versionErr != nil {
				log.Println(versionErr)
			}
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (g *KmsGenerator) loadKmsKeyRingIam(ctx context.Context, kmsService *cloudkms.Service, ringName string) {
	policy, err := kmsService.Projects.Locations.KeyRings.GetIamPolicy(ringName).Context(ctx).Do()
	if err != nil {
		log.Println(err)
		return
	}
	g.Resources = append(g.Resources, expandIamPolicy(kmsBindings(policy), ringName,
		"google_kms_key_ring_iam_policy", "google_kms_key_ring_iam_binding", "google_kms_key_ring_iam_member",
		map[string]string{"key_ring_id": ringName},
		func(id, resourceType string, attrs map[string]string) terraformutils.Resource {
			return g.newKmsDepthResource(id, resourceType, attrs)
		})...)
}

func (g *KmsGenerator) loadKmsCryptoKeyIam(ctx context.Context, kmsService *cloudkms.Service, keyName string) {
	policy, err := kmsService.Projects.Locations.KeyRings.CryptoKeys.GetIamPolicy(keyName).Context(ctx).Do()
	if err != nil {
		log.Println(err)
		return
	}
	g.Resources = append(g.Resources, expandIamPolicy(kmsBindings(policy), keyName,
		"google_kms_crypto_key_iam_policy", "google_kms_crypto_key_iam_binding", "google_kms_crypto_key_iam_member",
		map[string]string{"crypto_key_id": keyName},
		func(id, resourceType string, attrs map[string]string) terraformutils.Resource {
			return g.newKmsDepthResource(id, resourceType, attrs)
		})...)
}

func (g *KmsGenerator) loadKmsEkmConnectionIam(ctx context.Context, kmsService *cloudkms.Service, connectionName string) {
	policy, err := kmsService.Projects.Locations.EkmConnections.GetIamPolicy(connectionName).Context(ctx).Do()
	if err != nil {
		log.Println(err)
		return
	}
	g.Resources = append(g.Resources, expandIamPolicy(kmsBindings(policy), connectionName,
		"google_kms_ekm_connection_iam_policy", "google_kms_ekm_connection_iam_binding", "google_kms_ekm_connection_iam_member",
		map[string]string{"ekm_connection_id": connectionName},
		func(id, resourceType string, attrs map[string]string) terraformutils.Resource {
			return g.newKmsDepthResource(id, resourceType, attrs)
		})...)
}

func kmsBindings(policy *cloudkms.Policy) []iamBinding {
	bindings := []iamBinding{}
	if policy == nil {
		return bindings
	}
	for _, binding := range policy.Bindings {
		bindings = append(bindings, iamBinding{Role: binding.Role, Members: binding.Members})
	}
	return bindings
}
