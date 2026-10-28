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
	clouddeployv1 "google.golang.org/api/clouddeploy/v1"
	"google.golang.org/api/compute/v1"
)

var cloudDeployAllowEmptyValues = []string{""}

var cloudDeployAdditionalFields = map[string]interface{}{}

// CloudDeployGenerator enumerates Google Cloud Deploy in the
// configured region: custom target types (with IAM), delivery
// pipelines (with IAM) and their automations, targets (with IAM),
// and deploy policies. Deploy policies carry no provider IAM
// resources, so none are emitted for them.
type CloudDeployGenerator struct {
	GCPService
}

func (g *CloudDeployGenerator) InitResources() error {
	ctx := context.Background()
	deployService, err := clouddeployv1.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, deployService)
}

func (g *CloudDeployGenerator) initResources(ctx context.Context, deployService *clouddeployv1.Service) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + region
	baseAttrs := map[string]string{"project": project, "location": region}

	err := deployService.Projects.Locations.CustomTargetTypes.List(parent).Pages(ctx, func(page *clouddeployv1.ListCustomTargetTypesResponse) error {
		for _, targetType := range page.CustomTargetTypes {
			g.Resources = append(g.Resources, g.newCloudDeployResource(targetType.Name, "google_clouddeploy_custom_target_type", baseAttrs))
			g.loadCloudDeployIam(func() (*clouddeployv1.Policy, error) {
				return deployService.Projects.Locations.CustomTargetTypes.GetIamPolicy(targetType.Name).Context(ctx).Do()
			}, targetType.Name,
				"google_clouddeploy_custom_target_type_iam_policy", "google_clouddeploy_custom_target_type_iam_binding", "google_clouddeploy_custom_target_type_iam_member",
				"custom_target_type_id", baseAttrs)
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = deployService.Projects.Locations.Targets.List(parent).Pages(ctx, func(page *clouddeployv1.ListTargetsResponse) error {
		for _, target := range page.Targets {
			g.Resources = append(g.Resources, g.newCloudDeployResource(target.Name, "google_clouddeploy_target", baseAttrs))
			g.loadCloudDeployIam(func() (*clouddeployv1.Policy, error) {
				return deployService.Projects.Locations.Targets.GetIamPolicy(target.Name).Context(ctx).Do()
			}, target.Name,
				"google_clouddeploy_target_iam_policy", "google_clouddeploy_target_iam_binding", "google_clouddeploy_target_iam_member",
				"target_id", baseAttrs)
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = deployService.Projects.Locations.DeployPolicies.List(parent).Pages(ctx, func(page *clouddeployv1.ListDeployPoliciesResponse) error {
		for _, policy := range page.DeployPolicies {
			g.Resources = append(g.Resources, g.newCloudDeployResource(policy.Name, "google_clouddeploy_deploy_policy", baseAttrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	return deployService.Projects.Locations.DeliveryPipelines.List(parent).Pages(ctx, func(page *clouddeployv1.ListDeliveryPipelinesResponse) error {
		for _, pipeline := range page.DeliveryPipelines {
			g.Resources = append(g.Resources, g.newCloudDeployResource(pipeline.Name, "google_clouddeploy_delivery_pipeline", baseAttrs))
			g.loadCloudDeployIam(func() (*clouddeployv1.Policy, error) {
				return deployService.Projects.Locations.DeliveryPipelines.GetIamPolicy(pipeline.Name).Context(ctx).Do()
			}, pipeline.Name,
				"google_clouddeploy_delivery_pipeline_iam_policy", "google_clouddeploy_delivery_pipeline_iam_binding", "google_clouddeploy_delivery_pipeline_iam_member",
				"delivery_pipeline_id", baseAttrs)
			if err := g.loadAutomations(ctx, deployService, pipeline.Name, baseAttrs); err != nil {
				return err
			}
		}
		return nil
	})
}

func (g *CloudDeployGenerator) newCloudDeployResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
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
		cloudDeployAllowEmptyValues,
		cloudDeployAdditionalFields,
	)
}

func (g *CloudDeployGenerator) loadCloudDeployIam(getPolicy func() (*clouddeployv1.Policy, error), name, policyType, bindingType, memberType, parentKey string, baseAttrs map[string]string) {
	policy, err := getPolicy()
	if err != nil {
		log.Println(err)
		return
	}
	bindings := []iamBinding{}
	for _, binding := range policy.Bindings {
		bindings = append(bindings, iamBinding{Role: binding.Role, Members: binding.Members})
	}
	parentAttrs := map[string]string{parentKey: name}
	for k, v := range baseAttrs {
		parentAttrs[k] = v
	}
	g.Resources = append(g.Resources, expandIamPolicy(
		bindings, name, policyType, bindingType, memberType, parentAttrs, g.newCloudDeployResource)...)
}

func (g *CloudDeployGenerator) loadAutomations(ctx context.Context, deployService *clouddeployv1.Service, pipelineName string, baseAttrs map[string]string) error {
	return deployService.Projects.Locations.DeliveryPipelines.Automations.List(pipelineName).Pages(ctx, func(page *clouddeployv1.ListAutomationsResponse) error {
		for _, automation := range page.Automations {
			g.Resources = append(g.Resources, g.newCloudDeployResource(automation.Name, "google_clouddeploy_automation", baseAttrs))
		}
		return nil
	})
}
