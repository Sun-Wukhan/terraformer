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
	bigtableadminv2 "google.golang.org/api/bigtableadmin/v2"
)

var bigtableAllowEmptyValues = []string{""}

var bigtableAdditionalFields = map[string]interface{}{}

type BigtableGenerator struct {
	GCPService
}

// InitResources enumerates Cloud Bigtable resources for the configured
// project: instances with their IAM trio, app profiles, tables with
// their IAM trio, authorized views per table, and logical and
// materialized views per instance. Provider IDs are the full API
// resource names (projects/{project}/instances/{instance}/...), which
// is the canonical import format for every resource emitted here.
//
// Deliberately not generated:
//   - google_bigtable_gc_policy: a GC policy is not a separate API
//     object; it is the gcRule embedded in a column family of the
//     parent table (projects.instances.tables has no gcPolicy
//     collection, list or get), so it is a sub-resource embedded in
//     the emitted google_bigtable_table's spec.
//   - google_bigtable_memory_layer: the Bigtable Admin v2 API has no
//     memory layer collection — neither the pinned generated client
//     (google.golang.org/api v0.230.0) nor its discovery document
//     contains a memoryLayers list/get endpoint.
//   - google_bigtable_schema_bundle: schema bundles likewise have no
//     surface in the pinned generated client or its discovery
//     document (no schemaBundles collection), so there is no verified
//     list endpoint to enumerate them through.
func (g *BigtableGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := bigtableadminv2.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, svc)
}

func (g *BigtableGenerator) initResources(ctx context.Context, svc *bigtableadminv2.Service) error {
	project := g.GetArgs()["project"].(string)
	baseAttrs := map[string]string{"project": project}
	err := svc.Projects.Instances.List("projects/"+project).Pages(ctx, func(page *bigtableadminv2.ListInstancesResponse) error {
		for _, instance := range page.Instances {
			g.Resources = append(g.Resources, g.newBigtableResource(instance.Name, "google_bigtable_instance", baseAttrs))
			g.loadBigtableIam(func() ([]iamBinding, error) {
				policy, err := svc.Projects.Instances.GetIamPolicy(instance.Name, &bigtableadminv2.GetIamPolicyRequest{}).Context(ctx).Do()
				if err != nil {
					return nil, err
				}
				return bigtablePolicyBindings(policy), nil
			}, instance.Name, "google_bigtable_instance", "instance", baseAttrs)
			g.loadAppProfiles(ctx, svc, instance.Name, baseAttrs)
			g.loadLogicalViews(ctx, svc, instance.Name, baseAttrs)
			g.loadMaterializedViews(ctx, svc, instance.Name, baseAttrs)
			g.loadBigtableTables(ctx, svc, instance.Name, baseAttrs)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	return nil
}

func bigtablePolicyBindings(policy *bigtableadminv2.Policy) []iamBinding {
	bindings := []iamBinding{}
	for _, binding := range policy.Bindings {
		bindings = append(bindings, iamBinding{Role: binding.Role, Members: binding.Members})
	}
	return bindings
}

func (g *BigtableGenerator) newBigtableResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
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
		bigtableAllowEmptyValues,
		bigtableAdditionalFields,
	)
}

func (g *BigtableGenerator) loadBigtableIam(getBindings func() ([]iamBinding, error), name, resourcePrefix, parentKey string, baseAttrs map[string]string) {
	bindings, err := getBindings()
	if err != nil {
		log.Println(err)
		return
	}
	parentAttrs := map[string]string{parentKey: name}
	for k, v := range baseAttrs {
		parentAttrs[k] = v
	}
	g.Resources = append(g.Resources, expandIamPolicy(
		bindings,
		name,
		resourcePrefix+"_iam_policy",
		resourcePrefix+"_iam_binding",
		resourcePrefix+"_iam_member",
		parentAttrs,
		g.newBigtableResource,
	)...)
}

func (g *BigtableGenerator) loadAppProfiles(ctx context.Context, svc *bigtableadminv2.Service, instanceName string, baseAttrs map[string]string) {
	err := svc.Projects.Instances.AppProfiles.List(instanceName).Pages(ctx, func(page *bigtableadminv2.ListAppProfilesResponse) error {
		for _, profile := range page.AppProfiles {
			g.Resources = append(g.Resources, g.newBigtableResource(profile.Name, "google_bigtable_app_profile", baseAttrs))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (g *BigtableGenerator) loadLogicalViews(ctx context.Context, svc *bigtableadminv2.Service, instanceName string, baseAttrs map[string]string) {
	err := svc.Projects.Instances.LogicalViews.List(instanceName).Pages(ctx, func(page *bigtableadminv2.ListLogicalViewsResponse) error {
		for _, view := range page.LogicalViews {
			g.Resources = append(g.Resources, g.newBigtableResource(view.Name, "google_bigtable_logical_view", baseAttrs))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (g *BigtableGenerator) loadMaterializedViews(ctx context.Context, svc *bigtableadminv2.Service, instanceName string, baseAttrs map[string]string) {
	err := svc.Projects.Instances.MaterializedViews.List(instanceName).Pages(ctx, func(page *bigtableadminv2.ListMaterializedViewsResponse) error {
		for _, view := range page.MaterializedViews {
			g.Resources = append(g.Resources, g.newBigtableResource(view.Name, "google_bigtable_materialized_view", baseAttrs))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (g *BigtableGenerator) loadBigtableTables(ctx context.Context, svc *bigtableadminv2.Service, instanceName string, baseAttrs map[string]string) {
	err := svc.Projects.Instances.Tables.List(instanceName).Pages(ctx, func(page *bigtableadminv2.ListTablesResponse) error {
		for _, table := range page.Tables {
			g.Resources = append(g.Resources, g.newBigtableResource(table.Name, "google_bigtable_table", baseAttrs))
			g.loadBigtableIam(func() ([]iamBinding, error) {
				policy, err := svc.Projects.Instances.Tables.GetIamPolicy(table.Name, &bigtableadminv2.GetIamPolicyRequest{}).Context(ctx).Do()
				if err != nil {
					return nil, err
				}
				return bigtablePolicyBindings(policy), nil
			}, table.Name, "google_bigtable_table", "table", baseAttrs)
			viewErr := svc.Projects.Instances.Tables.AuthorizedViews.List(table.Name).Pages(ctx, func(page *bigtableadminv2.ListAuthorizedViewsResponse) error {
				for _, view := range page.AuthorizedViews {
					g.Resources = append(g.Resources, g.newBigtableResource(view.Name, "google_bigtable_authorized_view", baseAttrs))
				}
				return nil
			})
			if viewErr != nil {
				log.Println(viewErr)
			}
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}
