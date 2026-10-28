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
)

var memorystoreAllowEmptyValues = []string{""}

var memorystoreAdditionalFields = map[string]interface{}{}

const memorystoreBaseURL = "https://memorystore.googleapis.com/v1/"

type MemorystoreGenerator struct {
	GCPService
}

// InitResources enumerates Memorystore (the unified memorystore
// .googleapis.com API, distinct from the legacy redis.googleapis.com
// API behind MemoryStoreGenerator) resources for the configured
// project and region: instances and ACL policies. There is no
// generated client for memorystore.googleapis.com in the pinned
// google.golang.org/api module, so enumeration is raw REST through
// the shared helper; both collections are list endpoints in the
// Memorystore discovery document
// (projects.locations.instances.list / projects.locations
// .aclPolicies.list). Provider IDs are the full resource names.
//
// Deliberately not generated:
//   - google_memorystore_instance_desired_user_created_endpoints:
//     desired user-created endpoints are a field of the instance
//     object itself (its desired endpoint configuration), not a
//     separate API collection — the provider resource is an overlay
//     on google_memorystore_instance with the instance's own ID, so
//     emitting it would double-manage the instance's endpoint spec.
func (g *MemorystoreGenerator) InitResources() error {
	ctx := context.Background()
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, restClient)
}

func (g *MemorystoreGenerator) initResources(ctx context.Context, restClient *http.Client) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region)
	locationParent := "projects/" + project + "/locations/" + region.Name
	baseAttrs := map[string]string{"project": project, "location": region.Name}
	for _, collection := range []struct {
		suffix       string
		resourceType string
	}{
		{"/instances", "google_memorystore_instance"},
		{"/aclPolicies", "google_memorystore_acl_policy"},
	} {
		items, err := restListResources(ctx, restClient, memorystoreBaseURL+locationParent+collection.suffix)
		if err != nil {
			log.Println(err)
			continue
		}
		for _, item := range items {
			g.Resources = append(g.Resources, g.newMemorystoreResource(item.Name, collection.resourceType, baseAttrs))
		}
	}
	return nil
}

func (g *MemorystoreGenerator) newMemorystoreResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
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
		memorystoreAllowEmptyValues,
		memorystoreAdditionalFields,
	)
}
