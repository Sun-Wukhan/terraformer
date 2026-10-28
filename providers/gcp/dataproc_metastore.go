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
	metastorev1beta "google.golang.org/api/metastore/v1beta"
)

var metastoreAllowEmptyValues = []string{""}

var metastoreAdditionalFields = map[string]interface{}{}

type DataprocMetastoreGenerator struct {
	GCPService
}

// InitResources enumerates Dataproc Metastore resources for the
// configured project and region: services with their IAM trio and
// federations with their IAM trio. Provider IDs are the full API
// resource names, the canonical import format for both (and for
// their IAM resources).
//
// Deliberately not generated:
//   - google_dataproc_metastore_database_iam and
//     google_dataproc_metastore_table_iam: the Metastore API exposes
//     only getIamPolicy/setIamPolicy on databases and tables — there
//     is no list or get for the databases and tables themselves in
//     any API version (they are Hive metastore contents addressed by
//     name), so the parent collections are undiscoverable and their
//     IAM cannot be enumerated.
func (g *DataprocMetastoreGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := metastorev1beta.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, svc)
}

func (g *DataprocMetastoreGenerator) initResources(ctx context.Context, svc *metastorev1beta.APIService) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region)
	locationParent := "projects/" + project + "/locations/" + region.Name
	baseAttrs := map[string]string{"project": project, "location": region.Name}
	err := svc.Projects.Locations.Services.List(locationParent).Pages(ctx, func(page *metastorev1beta.ListServicesResponse) error {
		for _, service := range page.Services {
			g.Resources = append(g.Resources, g.newMetastoreResource(service.Name, "google_dataproc_metastore_service", baseAttrs))
			g.loadMetastoreIam(func() ([]iamBinding, error) {
				policy, err := svc.Projects.Locations.Services.GetIamPolicy(service.Name).Context(ctx).Do()
				if err != nil {
					return nil, err
				}
				return metastorePolicyBindings(policy), nil
			}, service.Name, "google_dataproc_metastore_service", baseAttrs)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = svc.Projects.Locations.Federations.List(locationParent).Pages(ctx, func(page *metastorev1beta.ListFederationsResponse) error {
		for _, federation := range page.Federations {
			g.Resources = append(g.Resources, g.newMetastoreResource(federation.Name, "google_dataproc_metastore_federation", baseAttrs))
			g.loadMetastoreIam(func() ([]iamBinding, error) {
				policy, err := svc.Projects.Locations.Federations.GetIamPolicy(federation.Name).Context(ctx).Do()
				if err != nil {
					return nil, err
				}
				return metastorePolicyBindings(policy), nil
			}, federation.Name, "google_dataproc_metastore_federation", baseAttrs)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	return nil
}

func metastorePolicyBindings(policy *metastorev1beta.Policy) []iamBinding {
	bindings := []iamBinding{}
	for _, binding := range policy.Bindings {
		bindings = append(bindings, iamBinding{Role: binding.Role, Members: binding.Members})
	}
	return bindings
}

func (g *DataprocMetastoreGenerator) newMetastoreResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
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
		metastoreAllowEmptyValues,
		metastoreAdditionalFields,
	)
}

func (g *DataprocMetastoreGenerator) loadMetastoreIam(getBindings func() ([]iamBinding, error), name, resourcePrefix string, baseAttrs map[string]string) {
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
		g.newMetastoreResource,
	)...)
}
