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
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"google.golang.org/api/compute/v1"
	transcoder "google.golang.org/api/transcoder/v1"
)

type TranscoderGenerator struct {
	GCPService
}

// InitResources enumerates Transcoder API jobs and job templates in
// the configured region. Provider IDs are the full resource names.
func (g *TranscoderGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := transcoder.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, svc)
}

func (g *TranscoderGenerator) initResources(ctx context.Context, svc *transcoder.Service) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + region
	if err := svc.Projects.Locations.Jobs.List(parent).Pages(ctx, func(page *transcoder.ListJobsResponse) error {
		for _, job := range page.Jobs {
			g.appendResource(job.Name, "google_transcoder_job")
		}
		return nil
	}); err != nil {
		return err
	}
	return svc.Projects.Locations.JobTemplates.List(parent).Pages(ctx, func(page *transcoder.ListJobTemplatesResponse) error {
		for _, tpl := range page.JobTemplates {
			g.appendResource(tpl.Name, "google_transcoder_job_template")
		}
		return nil
	})
}

func (g *TranscoderGenerator) appendResource(name, resourceType string) {
	parts := strings.Split(name, "/")
	g.Resources = append(g.Resources, terraformutils.NewResource(
		name, parts[len(parts)-1], resourceType, g.ProviderName,
		map[string]string{}, []string{}, map[string]interface{}{}))
}
