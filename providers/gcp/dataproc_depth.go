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
	dataprocv1 "google.golang.org/api/dataproc/v1"
)

// This file extends DataprocGenerator (dataproc.go, which enumerates
// clusters) with the rest of the Dataproc surface: autoscaling
// policies and workflow templates with their IAM trios, cluster IAM
// expanded from each cluster's live policy, batches and session
// templates. Provider IDs are the full API resource names.
//
// Deliberately not generated:
//   - google_dataproc_job_iam: jobs are transient execution records
//     (projects.regions.jobs.list returns run history, the same
//     reason google_bigquery_job is not generated in
//     bigquery_depth.go); IAM hung off a finished job is not durable
//     configuration.
//   - google_dataproc_gdc_application_environment,
//     google_dataproc_gdc_service_instance and
//     google_dataproc_gdc_spark_application: these target Dataproc on
//     Google Distributed Cloud (air-gapped) through a separate GDC
//     endpoint, not the public dataproc.googleapis.com API — there
//     is no generated client or public discovery document for them,
//     and their parent GDC service instances are undiscoverable from
//     a public-cloud project scope.
func (g *DataprocGenerator) initDepthResources(ctx context.Context, svc *dataprocv1.Service) {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region)
	regionParent := "projects/" + project + "/regions/" + region.Name
	locationParent := "projects/" + project + "/locations/" + region.Name
	baseAttrs := map[string]string{"project": project, "region": region.Name}

	err := svc.Projects.Regions.AutoscalingPolicies.List(regionParent).Pages(ctx, func(page *dataprocv1.ListAutoscalingPoliciesResponse) error {
		for _, policy := range page.Policies {
			g.Resources = append(g.Resources, g.newDataprocDepthResource(policy.Name, "google_dataproc_autoscaling_policy", baseAttrs))
			g.loadDataprocIam(func() ([]iamBinding, error) {
				p, err := svc.Projects.Regions.AutoscalingPolicies.GetIamPolicy(policy.Name, &dataprocv1.GetIamPolicyRequest{}).Context(ctx).Do()
				if err != nil {
					return nil, err
				}
				return dataprocPolicyBindings(p), nil
			}, policy.Name, "google_dataproc_autoscaling_policy", baseAttrs)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}

	err = svc.Projects.Regions.WorkflowTemplates.List(regionParent).Pages(ctx, func(page *dataprocv1.ListWorkflowTemplatesResponse) error {
		for _, template := range page.Templates {
			g.Resources = append(g.Resources, g.newDataprocDepthResource(template.Name, "google_dataproc_workflow_template", baseAttrs))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}

	err = svc.Projects.Regions.Clusters.List(project, region.Name).Pages(ctx, func(page *dataprocv1.ListClustersResponse) error {
		for _, cluster := range page.Clusters {
			clusterName := regionParent + "/clusters/" + cluster.ClusterName
			g.loadDataprocIam(func() ([]iamBinding, error) {
				p, err := svc.Projects.Regions.Clusters.GetIamPolicy(clusterName, &dataprocv1.GetIamPolicyRequest{}).Context(ctx).Do()
				if err != nil {
					return nil, err
				}
				return dataprocPolicyBindings(p), nil
			}, clusterName, "google_dataproc_cluster", baseAttrs)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}

	err = svc.Projects.Locations.Batches.List(locationParent).Pages(ctx, func(page *dataprocv1.ListBatchesResponse) error {
		for _, batch := range page.Batches {
			g.Resources = append(g.Resources, g.newDataprocDepthResource(batch.Name, "google_dataproc_batch", baseAttrs))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}

	err = svc.Projects.Locations.SessionTemplates.List(locationParent).Pages(ctx, func(page *dataprocv1.ListSessionTemplatesResponse) error {
		for _, template := range page.SessionTemplates {
			g.Resources = append(g.Resources, g.newDataprocDepthResource(template.Name, "google_dataproc_session_template", baseAttrs))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func dataprocPolicyBindings(policy *dataprocv1.Policy) []iamBinding {
	bindings := []iamBinding{}
	for _, binding := range policy.Bindings {
		bindings = append(bindings, iamBinding{Role: binding.Role, Members: binding.Members})
	}
	return bindings
}

func (g *DataprocGenerator) newDataprocDepthResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
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
		dataprocAllowEmptyValues,
		dataprocAdditionalFields,
	)
}

func (g *DataprocGenerator) loadDataprocIam(getBindings func() ([]iamBinding, error), name, resourcePrefix string, baseAttrs map[string]string) {
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
		g.newDataprocDepthResource,
	)...)
}
