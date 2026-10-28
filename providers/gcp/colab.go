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
	"google.golang.org/api/aiplatform/v1"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

// ColabGenerator enumerates the Colab Enterprise surface of one
// project/location. Colab Enterprise is served by the Vertex AI
// (aiplatform v1) API: notebook runtime templates (with their IAM
// trio, expanded with literal type names), notebook runtimes,
// notebook execution jobs, and schedules — restricted to schedules
// whose payload is a notebook execution job request, which is what
// google_colab_schedule models (pipeline-job schedules belong to
// Vertex AI Pipelines and are not Colab resources). Every provider
// ID is the resource's full API name.
type ColabGenerator struct {
	GCPService
}

func (g *ColabGenerator) InitResources() error {
	ctx := context.Background()
	region := g.GetArgs()["region"].(compute.Region).Name
	options := []option.ClientOption{}
	if region != "" {
		options = append(options, option.WithEndpoint("https://"+region+"-aiplatform.googleapis.com/v1/"))
	}
	service, err := aiplatform.NewService(ctx, options...)
	if err != nil {
		return err
	}
	g.initResources(ctx, service)
	return nil
}

func (g *ColabGenerator) initResources(ctx context.Context, service *aiplatform.Service) {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + region

	templates, err := cxListNames(func(pageToken string) ([]string, string, error) {
		resp, err := service.Projects.Locations.NotebookRuntimeTemplates.List(parent).PageToken(pageToken).Context(ctx).Do()
		if err != nil {
			return nil, "", err
		}
		names := []string{}
		for _, template := range resp.NotebookRuntimeTemplates {
			names = append(names, template.Name)
		}
		return names, resp.NextPageToken, nil
	})
	if err != nil {
		log.Println(err)
	} else {
		for _, template := range templates {
			g.Resources = append(g.Resources, g.newColabResource(template, "google_colab_runtime_template"))
			policy, err := service.Projects.Locations.NotebookRuntimeTemplates.GetIamPolicy(template).Context(ctx).Do()
			if err != nil {
				log.Println(err)
				continue
			}
			bindings := []iamBinding{}
			for _, binding := range policy.Bindings {
				bindings = append(bindings, iamBinding{Role: binding.Role, Members: binding.Members})
			}
			g.Resources = append(g.Resources, expandIamPolicy(bindings, template,
				"google_colab_runtime_template_iam_policy",
				"google_colab_runtime_template_iam_binding",
				"google_colab_runtime_template_iam_member",
				map[string]string{"runtime_template": template},
				func(id, resourceType string, attrs map[string]string) terraformutils.Resource {
					merged := map[string]string{}
					for k, v := range attrs {
						merged[k] = v
					}
					parts := strings.Split(id, "/")
					return terraformutils.NewResource(id, parts[len(parts)-1], resourceType, g.ProviderName, merged, []string{}, map[string]interface{}{})
				})...)
		}
	}

	runtimes, err := cxListNames(func(pageToken string) ([]string, string, error) {
		resp, err := service.Projects.Locations.NotebookRuntimes.List(parent).PageToken(pageToken).Context(ctx).Do()
		if err != nil {
			return nil, "", err
		}
		names := []string{}
		for _, runtime := range resp.NotebookRuntimes {
			names = append(names, runtime.Name)
		}
		return names, resp.NextPageToken, nil
	})
	if err != nil {
		log.Println(err)
	} else {
		for _, runtime := range runtimes {
			g.Resources = append(g.Resources, g.newColabResource(runtime, "google_colab_runtime"))
		}
	}

	executions, err := cxListNames(func(pageToken string) ([]string, string, error) {
		resp, err := service.Projects.Locations.NotebookExecutionJobs.List(parent).PageToken(pageToken).Context(ctx).Do()
		if err != nil {
			return nil, "", err
		}
		names := []string{}
		for _, job := range resp.NotebookExecutionJobs {
			names = append(names, job.Name)
		}
		return names, resp.NextPageToken, nil
	})
	if err != nil {
		log.Println(err)
	} else {
		for _, execution := range executions {
			g.Resources = append(g.Resources, g.newColabResource(execution, "google_colab_notebook_execution"))
		}
	}

	err = service.Projects.Locations.Schedules.List(parent).Pages(ctx, func(page *aiplatform.GoogleCloudAiplatformV1ListSchedulesResponse) error {
		for _, schedule := range page.Schedules {
			if schedule.CreateNotebookExecutionJobRequest != nil {
				g.Resources = append(g.Resources, g.newColabResource(schedule.Name, "google_colab_schedule"))
			}
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (g *ColabGenerator) newColabResource(id, resourceType string) terraformutils.Resource {
	parts := strings.Split(id, "/")
	return terraformutils.NewResource(
		id,
		parts[len(parts)-1],
		resourceType,
		g.ProviderName,
		map[string]string{
			"project": g.GetArgs()["project"].(string),
			"region":  g.GetArgs()["region"].(compute.Region).Name,
		},
		[]string{},
		map[string]interface{}{},
	)
}
