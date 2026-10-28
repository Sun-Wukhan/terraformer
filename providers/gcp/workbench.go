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
	notebooksv2 "google.golang.org/api/notebooks/v2"
	"google.golang.org/api/option"
)

// WorkbenchGenerator enumerates the Vertex AI Workbench surface of
// one project/location through the Notebooks v2 client: instances
// and each instance's IAM policy, expanded into the provider's
// _iam_policy/_iam_binding/_iam_member trio with literal type names.
// Every provider ID is the instance's full API name.
type WorkbenchGenerator struct {
	GCPService
}

func (g *WorkbenchGenerator) InitResources() error {
	ctx := context.Background()
	region := g.GetArgs()["region"].(compute.Region).Name
	options := []option.ClientOption{}
	if region != "" {
		options = append(options, option.WithEndpoint("https://"+region+"-notebooks.googleapis.com/v2/"))
	}
	service, err := notebooksv2.NewService(ctx, options...)
	if err != nil {
		return err
	}
	g.initResources(ctx, service)
	return nil
}

func (g *WorkbenchGenerator) initResources(ctx context.Context, service *notebooksv2.Service) {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + region
	instances, err := cxListNames(func(pageToken string) ([]string, string, error) {
		resp, err := service.Projects.Locations.Instances.List(parent).PageToken(pageToken).Context(ctx).Do()
		if err != nil {
			return nil, "", err
		}
		names := []string{}
		for _, instance := range resp.Instances {
			names = append(names, instance.Name)
		}
		return names, resp.NextPageToken, nil
	})
	if err != nil {
		log.Println(err)
		return
	}
	for _, instance := range instances {
		g.Resources = append(g.Resources, g.newWorkbenchResource(instance, "google_workbench_instance", map[string]string{}))
		policy, err := service.Projects.Locations.Instances.GetIamPolicy(instance).Context(ctx).Do()
		if err != nil {
			log.Println(err)
			continue
		}
		bindings := []iamBinding{}
		for _, binding := range policy.Bindings {
			bindings = append(bindings, iamBinding{Role: binding.Role, Members: binding.Members})
		}
		g.Resources = append(g.Resources, expandIamPolicy(bindings, instance,
			"google_workbench_instance_iam_policy",
			"google_workbench_instance_iam_binding",
			"google_workbench_instance_iam_member",
			map[string]string{"instance_name": instance},
			func(id, resourceType string, attrs map[string]string) terraformutils.Resource {
				return g.newWorkbenchResource(id, resourceType, attrs)
			})...)
	}
}

func (g *WorkbenchGenerator) newWorkbenchResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
	parts := strings.Split(id, "/")
	merged := map[string]string{
		"project":  g.GetArgs()["project"].(string),
		"location": g.GetArgs()["region"].(compute.Region).Name,
	}
	for k, v := range attrs {
		merged[k] = v
	}
	return terraformutils.NewResource(
		id,
		parts[len(parts)-1],
		resourceType,
		g.ProviderName,
		merged,
		[]string{},
		map[string]interface{}{},
	)
}
