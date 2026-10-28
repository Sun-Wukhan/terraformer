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
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"google.golang.org/api/compute/v1"
	datacatalogv1 "google.golang.org/api/datacatalog/v1"
)

var dataCatalogAllowEmptyValues = []string{""}

var dataCatalogAdditionalFields = map[string]interface{}{}

type DataCatalogGenerator struct {
	GCPService
}

// InitResources enumerates Data Catalog resources for the configured
// project and region: entry groups with their IAM trio, entries per
// group, tags on groups and on entries, and taxonomies with their
// IAM trio and policy tags with theirs. Provider IDs are the full
// API resource names, the canonical import format for all of these.
//
// Deliberately not generated:
//   - google_data_catalog_tag_template and
//     google_data_catalog_tag_template_iam: the Data Catalog v1 API
//     has no tag template list method (create/get/patch/delete and
//     IAM methods on a known name only, in both the pinned generated
//     client and its discovery document), so tag templates — and
//     therefore their IAM — are undiscoverable from a location.
func (g *DataCatalogGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := datacatalogv1.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, svc)
}

func (g *DataCatalogGenerator) initResources(ctx context.Context, svc *datacatalogv1.Service) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region)
	locationParent := "projects/" + project + "/locations/" + region.Name
	baseAttrs := map[string]string{"project": project, "location": region.Name}
	g.loadEntryGroups(ctx, svc, locationParent, baseAttrs)
	g.loadTaxonomies(ctx, svc, locationParent, baseAttrs)
	return nil
}

func dataCatalogPolicyBindings(policy *datacatalogv1.Policy) []iamBinding {
	bindings := []iamBinding{}
	for _, binding := range policy.Bindings {
		bindings = append(bindings, iamBinding{Role: binding.Role, Members: binding.Members})
	}
	return bindings
}

func (g *DataCatalogGenerator) newDataCatalogResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
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
		dataCatalogAllowEmptyValues,
		dataCatalogAdditionalFields,
	)
}

func (g *DataCatalogGenerator) loadDataCatalogIam(getBindings func() ([]iamBinding, error), name, resourcePrefix string, baseAttrs map[string]string) {
	bindings, err := getBindings()
	if err != nil {
		log.Println(err)
		return
	}
	g.Resources = append(g.Resources, expandIamPolicy(
		bindings,
		name,
		resourcePrefix+"_iam_policy",
		resourcePrefix+"_iam_binding",
		resourcePrefix+"_iam_member",
		baseAttrs,
		g.newDataCatalogResource,
	)...)
}

func (g *DataCatalogGenerator) loadEntryGroups(ctx context.Context, svc *datacatalogv1.Service, locationParent string, baseAttrs map[string]string) {
	err := svc.Projects.Locations.EntryGroups.List(locationParent).Pages(ctx, func(page *datacatalogv1.GoogleCloudDatacatalogV1ListEntryGroupsResponse) error {
		for _, group := range page.EntryGroups {
			g.Resources = append(g.Resources, g.newDataCatalogResource(group.Name, "google_data_catalog_entry_group", baseAttrs))
			g.loadDataCatalogIam(func() ([]iamBinding, error) {
				policy, err := svc.Projects.Locations.EntryGroups.GetIamPolicy(group.Name, &datacatalogv1.GetIamPolicyRequest{}).Context(ctx).Do()
				if err != nil {
					return nil, err
				}
				return dataCatalogPolicyBindings(policy), nil
			}, group.Name, "google_data_catalog_entry_group", baseAttrs)
			g.loadTags(func() error {
				return svc.Projects.Locations.EntryGroups.Tags.List(group.Name).Pages(ctx, func(page *datacatalogv1.GoogleCloudDatacatalogV1ListTagsResponse) error {
					for _, tag := range page.Tags {
						g.Resources = append(g.Resources, g.newDataCatalogResource(tag.Name, "google_data_catalog_tag", baseAttrs))
					}
					return nil
				})
			})
			g.loadEntries(ctx, svc, group.Name, baseAttrs)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (g *DataCatalogGenerator) loadTags(listFn func() error) {
	if err := listFn(); err != nil {
		log.Println(err)
	}
}

func (g *DataCatalogGenerator) loadEntries(ctx context.Context, svc *datacatalogv1.Service, groupName string, baseAttrs map[string]string) {
	err := svc.Projects.Locations.EntryGroups.Entries.List(groupName).Pages(ctx, func(page *datacatalogv1.GoogleCloudDatacatalogV1ListEntriesResponse) error {
		for _, entry := range page.Entries {
			g.Resources = append(g.Resources, g.newDataCatalogResource(entry.Name, "google_data_catalog_entry", baseAttrs))
			g.loadTags(func() error {
				return svc.Projects.Locations.EntryGroups.Entries.Tags.List(entry.Name).Pages(ctx, func(page *datacatalogv1.GoogleCloudDatacatalogV1ListTagsResponse) error {
					for _, tag := range page.Tags {
						g.Resources = append(g.Resources, g.newDataCatalogResource(tag.Name, "google_data_catalog_tag", baseAttrs))
					}
					return nil
				})
			})
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (g *DataCatalogGenerator) loadTaxonomies(ctx context.Context, svc *datacatalogv1.Service, locationParent string, baseAttrs map[string]string) {
	err := svc.Projects.Locations.Taxonomies.List(locationParent).Pages(ctx, func(page *datacatalogv1.GoogleCloudDatacatalogV1ListTaxonomiesResponse) error {
		for _, taxonomy := range page.Taxonomies {
			g.Resources = append(g.Resources, g.newDataCatalogResource(taxonomy.Name, "google_data_catalog_taxonomy", baseAttrs))
			g.loadDataCatalogIam(func() ([]iamBinding, error) {
				policy, err := svc.Projects.Locations.Taxonomies.GetIamPolicy(taxonomy.Name, &datacatalogv1.GetIamPolicyRequest{}).Context(ctx).Do()
				if err != nil {
					return nil, err
				}
				return dataCatalogPolicyBindings(policy), nil
			}, taxonomy.Name, "google_data_catalog_taxonomy", baseAttrs)
			tagErr := svc.Projects.Locations.Taxonomies.PolicyTags.List(taxonomy.Name).Pages(ctx, func(page *datacatalogv1.GoogleCloudDatacatalogV1ListPolicyTagsResponse) error {
				for _, policyTag := range page.PolicyTags {
					g.Resources = append(g.Resources, g.newDataCatalogResource(policyTag.Name, "google_data_catalog_policy_tag", baseAttrs))
					g.loadDataCatalogIam(func() ([]iamBinding, error) {
						policy, err := svc.Projects.Locations.Taxonomies.PolicyTags.GetIamPolicy(policyTag.Name, &datacatalogv1.GetIamPolicyRequest{}).Context(ctx).Do()
						if err != nil {
							return nil, err
						}
						return dataCatalogPolicyBindings(policy), nil
					}, policyTag.Name, "google_data_catalog_policy_tag", baseAttrs)
				}
				return nil
			})
			if tagErr != nil {
				log.Println(tagErr)
			}
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}
