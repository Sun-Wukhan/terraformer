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
	runv1 "google.golang.org/api/run/v1"
	runv2 "google.golang.org/api/run/v2"
)

var cloudRunAllowEmptyValues = []string{"tags."}

var cloudRunAdditionalFields = map[string]interface{}{}

type CloudRunGenerator struct {
	GCPService
}

// InitResources enumerates Cloud Run in the configured region with both
// generations of the API: v1 services (with IAM) and domain mappings, and
// v2 services (with IAM), jobs (with IAM) and worker pools.
//
// Deliberately not generated:
//   - google_cloud_run_v2_instance and google_cloud_run_v2_instance_iam:
//     the GA run/v2 client has no instances collection.
//   - google_cloud_run_v2_worker_pool_iam: worker pools expose no
//     getIamPolicy in the GA run/v2 client.
func (g *CloudRunGenerator) InitResources() error {
	ctx := context.Background()
	v1Service, err := runv1.NewService(ctx)
	if err != nil {
		return err
	}
	v2Service, err := runv2.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, v1Service, v2Service)
}

func (g *CloudRunGenerator) initResources(ctx context.Context, v1Service *runv1.APIService, v2Service *runv2.Service) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + region
	if err := g.loadV1Services(ctx, v1Service, parent); err != nil {
		return err
	}
	if err := g.loadDomainMappings(ctx, v1Service, parent); err != nil {
		return err
	}
	if err := g.loadV2Services(ctx, v2Service, parent); err != nil {
		return err
	}
	if err := g.loadV2Jobs(ctx, v2Service, parent); err != nil {
		return err
	}
	return g.loadV2WorkerPools(ctx, v2Service, parent)
}

func cloudRunLastSegment(name string) string {
	parts := strings.Split(name, "/")
	return parts[len(parts)-1]
}

func (g *CloudRunGenerator) newCloudRunResource(id, name, resourceType string, attrs map[string]string) terraformutils.Resource {
	if attrs == nil {
		attrs = map[string]string{}
	}
	if _, ok := attrs["project"]; !ok {
		attrs["project"] = g.GetArgs()["project"].(string)
	}
	if _, ok := attrs["location"]; !ok {
		attrs["location"] = g.GetArgs()["region"].(compute.Region).Name
	}
	return terraformutils.NewResource(
		id,
		name,
		resourceType,
		g.ProviderName,
		attrs,
		cloudRunAllowEmptyValues,
		cloudRunAdditionalFields,
	)
}

// appendCloudRunIam is the Cloud Run analogue of the GCS bucket IAM
// pattern: the policy resource ID is the parent resource's full name, and
// bindings and members are expanded from the parent's live IAM policy.
// Callers convert their client generation's policy bindings into
// cloudRunBinding values.
func appendCloudRunIam(g *CloudRunGenerator, resources []terraformutils.Resource, parentName, resourcePrefix string, bindings []cloudRunBinding, parentAttrs map[string]string) []terraformutils.Resource {
	resources = append(resources, g.newCloudRunResource(
		parentName,
		cloudRunLastSegment(parentName),
		resourcePrefix+"_iam_policy",
		parentAttrs,
	))
	for _, binding := range bindings {
		bindingAttrs := map[string]string{"role": binding.Role}
		for k, v := range parentAttrs {
			bindingAttrs[k] = v
		}
		resources = append(resources, g.newCloudRunResource(
			parentName,
			cloudRunLastSegment(parentName),
			resourcePrefix+"_iam_binding",
			bindingAttrs,
		))
		for _, member := range binding.Members {
			memberAttrs := map[string]string{"role": binding.Role, "member": member}
			for k, v := range parentAttrs {
				memberAttrs[k] = v
			}
			resources = append(resources, g.newCloudRunResource(
				parentName,
				cloudRunLastSegment(parentName),
				resourcePrefix+"_iam_member",
				memberAttrs,
			))
		}
	}
	return resources
}

type cloudRunBinding struct {
	Role    string
	Members []string
}

func (g *CloudRunGenerator) loadV1Services(ctx context.Context, service *runv1.APIService, parent string) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	// The generated v1 list calls have no Pages helper (Knative-style
	// continue tokens instead of page tokens), so this is a single call.
	page, err := service.Projects.Locations.Services.List(parent).Context(ctx).Do()
	if err != nil {
		return err
	}
	for _, svc := range page.Items {
		// The v1 API is Knative-shaped: the service identity lives
		// in metadata, and the provider ID is
		// locations/{location}/namespaces/{project}/services/{name}.
		name := svc.Metadata.Name
		namespace := svc.Metadata.Namespace
		if namespace == "" {
			namespace = project
		}
		id := "locations/" + region + "/namespaces/" + namespace + "/services/" + name
		g.Resources = append(g.Resources, g.newCloudRunResource(
			id,
			name,
			"google_cloud_run_service",
			map[string]string{"name": name},
		))
		iamResource := parent + "/services/" + name
		policy, err := service.Projects.Locations.Services.GetIamPolicy(iamResource).Context(ctx).Do()
		if err != nil {
			log.Println(err)
			continue
		}
		bindings := []cloudRunBinding{}
		for _, b := range policy.Bindings {
			bindings = append(bindings, cloudRunBinding{Role: b.Role, Members: b.Members})
		}
		g.Resources = appendCloudRunIam(g, g.Resources, id,
			"google_cloud_run_service", bindings, map[string]string{"service": name, "location": region})
	}
	return nil
}

func (g *CloudRunGenerator) loadDomainMappings(ctx context.Context, service *runv1.APIService, parent string) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	// Single call, like loadV1Services: no Pages helper on v1 list calls.
	page, err := service.Projects.Locations.Domainmappings.List(parent).Context(ctx).Do()
	if err != nil {
		return err
	}
	for _, mapping := range page.Items {
		namespace := mapping.Metadata.Namespace
		if namespace == "" {
			namespace = project
		}
		// Provider ID:
		// locations/{location}/namespaces/{project}/domainmappings/{name}.
		id := "locations/" + region + "/namespaces/" + namespace + "/domainmappings/" + mapping.Metadata.Name
		g.Resources = append(g.Resources, g.newCloudRunResource(
			id,
			mapping.Metadata.Name,
			"google_cloud_run_domain_mapping",
			map[string]string{"name": mapping.Metadata.Name},
		))
	}
	return nil
}

func (g *CloudRunGenerator) loadV2Services(ctx context.Context, service *runv2.Service, parent string) error {
	return service.Projects.Locations.Services.List(parent).Pages(ctx, func(page *runv2.GoogleCloudRunV2ListServicesResponse) error {
		for _, svc := range page.Services {
			name := cloudRunLastSegment(svc.Name)
			g.Resources = append(g.Resources, g.newCloudRunResource(
				svc.Name,
				name,
				"google_cloud_run_v2_service",
				map[string]string{"name": name},
			))
			policy, err := service.Projects.Locations.Services.GetIamPolicy(svc.Name).Context(ctx).Do()
			if err != nil {
				log.Println(err)
				continue
			}
			bindings := []cloudRunBinding{}
			for _, b := range policy.Bindings {
				bindings = append(bindings, cloudRunBinding{Role: b.Role, Members: b.Members})
			}
			g.Resources = appendCloudRunIam(g, g.Resources, svc.Name,
				"google_cloud_run_v2_service", bindings, map[string]string{"service": name})
		}
		return nil
	})
}

func (g *CloudRunGenerator) loadV2Jobs(ctx context.Context, service *runv2.Service, parent string) error {
	return service.Projects.Locations.Jobs.List(parent).Pages(ctx, func(page *runv2.GoogleCloudRunV2ListJobsResponse) error {
		for _, job := range page.Jobs {
			name := cloudRunLastSegment(job.Name)
			g.Resources = append(g.Resources, g.newCloudRunResource(
				job.Name,
				name,
				"google_cloud_run_v2_job",
				map[string]string{"name": name},
			))
			policy, err := service.Projects.Locations.Jobs.GetIamPolicy(job.Name).Context(ctx).Do()
			if err != nil {
				log.Println(err)
				continue
			}
			bindings := []cloudRunBinding{}
			for _, b := range policy.Bindings {
				bindings = append(bindings, cloudRunBinding{Role: b.Role, Members: b.Members})
			}
			g.Resources = appendCloudRunIam(g, g.Resources, job.Name,
				"google_cloud_run_v2_job", bindings, map[string]string{"job": name})
		}
		return nil
	})
}

func (g *CloudRunGenerator) loadV2WorkerPools(ctx context.Context, service *runv2.Service, parent string) error {
	return service.Projects.Locations.WorkerPools.List(parent).Pages(ctx, func(page *runv2.GoogleCloudRunV2ListWorkerPoolsResponse) error {
		for _, pool := range page.WorkerPools {
			g.Resources = append(g.Resources, g.newCloudRunResource(
				pool.Name,
				cloudRunLastSegment(pool.Name),
				"google_cloud_run_v2_worker_pool",
				map[string]string{"name": cloudRunLastSegment(pool.Name)},
			))
		}
		return nil
	})
}
