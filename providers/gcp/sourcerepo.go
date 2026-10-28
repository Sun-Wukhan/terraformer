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

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	sourcerepo "google.golang.org/api/sourcerepo/v1"
)

// SourceRepoGenerator enumerates Cloud Source Repositories via the
// generated sourcerepo v1 client, with the combined
// google_sourcerepo_repository_iam resource when a repository
// carries IAM bindings.
type SourceRepoGenerator struct {
	GCPService
}

func (g *SourceRepoGenerator) InitResources() error {
	ctx := context.Background()
	service, err := sourcerepo.NewService(ctx)
	if err != nil {
		return err
	}
	g.initResources(ctx, service)
	return nil
}

func (g *SourceRepoGenerator) initResources(ctx context.Context, service *sourcerepo.Service) {
	project := g.GetArgs()["project"].(string)
	if err := service.Projects.Repos.List("projects/"+project).Pages(ctx, func(page *sourcerepo.ListReposResponse) error {
		for _, repo := range page.Repos {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				repo.Name, opsLastSegment(repo.Name), "google_sourcerepo_repository", g.ProviderName,
				map[string]string{"project": project},
				[]string{}, map[string]interface{}{},
			))
			policy, err := service.Projects.Repos.GetIamPolicy(repo.Name).Context(ctx).Do()
			if err != nil {
				log.Println(err)
				continue
			}
			if len(policy.Bindings) > 0 {
				g.Resources = append(g.Resources, terraformutils.NewResource(
					repo.Name, opsLastSegment(repo.Name), "google_sourcerepo_repository_iam", g.ProviderName,
					map[string]string{"project": project},
					[]string{}, map[string]interface{}{},
				))
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
}
