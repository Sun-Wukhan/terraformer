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
	composer "google.golang.org/api/composer/v1"
	"google.golang.org/api/compute/v1"
)

// ComposerGenerator enumerates Cloud Composer environments in the
// configured region via the generated composer v1 client, plus each
// environment's user workloads ConfigMaps and Secrets.
type ComposerGenerator struct {
	GCPService
}

func (g *ComposerGenerator) InitResources() error {
	ctx := context.Background()
	service, err := composer.NewService(ctx)
	if err != nil {
		return err
	}
	g.initResources(ctx, service)
	return nil
}

func (g *ComposerGenerator) newResource(id, resourceType string) terraformutils.Resource {
	return terraformutils.NewResource(
		id, opsLastSegment(id), resourceType, g.ProviderName,
		map[string]string{"project": g.GetArgs()["project"].(string)},
		[]string{}, map[string]interface{}{},
	)
}

func (g *ComposerGenerator) initResources(ctx context.Context, service *composer.Service) {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + region
	if err := service.Projects.Locations.Environments.List(parent).Pages(ctx, func(page *composer.ListEnvironmentsResponse) error {
		for _, env := range page.Environments {
			g.Resources = append(g.Resources, g.newResource(env.Name, "google_composer_environment"))
			if err := service.Projects.Locations.Environments.UserWorkloadsConfigMaps.List(env.Name).Pages(ctx, func(cmPage *composer.ListUserWorkloadsConfigMapsResponse) error {
				for _, cm := range cmPage.UserWorkloadsConfigMaps {
					g.Resources = append(g.Resources, g.newResource(cm.Name, "google_composer_user_workloads_config_map"))
				}
				return nil
			}); err != nil {
				log.Println(err)
			}
			if err := service.Projects.Locations.Environments.UserWorkloadsSecrets.List(env.Name).Pages(ctx, func(sPage *composer.ListUserWorkloadsSecretsResponse) error {
				for _, secret := range sPage.UserWorkloadsSecrets {
					g.Resources = append(g.Resources, g.newResource(secret.Name, "google_composer_user_workloads_secret"))
				}
				return nil
			}); err != nil {
				log.Println(err)
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
}
