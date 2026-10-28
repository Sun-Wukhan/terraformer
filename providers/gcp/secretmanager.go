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
	secretmanager "google.golang.org/api/secretmanager/v1"
)

var secretManagerAllowEmptyValues = []string{""}

var secretManagerAdditionalFields = map[string]interface{}{}

type SecretManagerGenerator struct {
	GCPService
}

// InitResources enumerates Secret Manager in the configured project:
// global secrets and regional secrets (in the configured region), each
// with its versions and with the IAM policy, bindings and members
// expanded from the secret's live IAM policy. Provider IDs are the
// full resource names (projects/{p}/secrets/{s},
// projects/{p}/secrets/{s}/versions/{v},
// projects/{p}/locations/{l}/secrets/{s} and
// projects/{p}/locations/{l}/secrets/{s}/versions/{v}).
func (g *SecretManagerGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := secretmanager.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, svc)
}

func (g *SecretManagerGenerator) initResources(ctx context.Context, svc *secretmanager.Service) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	if err := svc.Projects.Secrets.List("projects/"+project).Pages(ctx, func(page *secretmanager.ListSecretsResponse) error {
		for _, secret := range page.Secrets {
			g.appendSecretIam(ctx, svc, secret.Name, "google_secret_manager_secret")
			if err := svc.Projects.Secrets.Versions.List(secret.Name).Pages(ctx, func(vpage *secretmanager.ListSecretVersionsResponse) error {
				for _, version := range vpage.Versions {
					g.Resources = append(g.Resources, g.newSecretManagerResource(version.Name, "google_secret_manager_secret_version"))
				}
				return nil
			}); err != nil {
				log.Println(err)
			}
		}
		return nil
	}); err != nil {
		return err
	}
	locationParent := "projects/" + project + "/locations/" + region
	return svc.Projects.Locations.Secrets.List(locationParent).Pages(ctx, func(page *secretmanager.ListSecretsResponse) error {
		for _, secret := range page.Secrets {
			g.appendRegionalSecretIam(ctx, svc, secret.Name)
			if err := svc.Projects.Locations.Secrets.Versions.List(secret.Name).Pages(ctx, func(vpage *secretmanager.ListSecretVersionsResponse) error {
				for _, version := range vpage.Versions {
					g.Resources = append(g.Resources, g.newSecretManagerResource(version.Name, "google_secret_manager_regional_secret_version"))
				}
				return nil
			}); err != nil {
				log.Println(err)
			}
		}
		return nil
	})
}

func (g *SecretManagerGenerator) newSecretManagerResource(id, resourceType string) terraformutils.Resource {
	return g.newSecretManagerResourceWithAttrs(id, resourceType, map[string]string{})
}

func (g *SecretManagerGenerator) newSecretManagerResourceWithAttrs(id, resourceType string, attrs map[string]string) terraformutils.Resource {
	parts := strings.Split(id, "/")
	return terraformutils.NewResource(
		id,
		parts[len(parts)-1],
		resourceType,
		g.ProviderName,
		attrs,
		secretManagerAllowEmptyValues,
		secretManagerAdditionalFields,
	)
}

func (g *SecretManagerGenerator) appendIam(id, resourcePrefix string, bindings []*secretmanager.Binding) {
	resources := []terraformutils.Resource{g.newSecretManagerResource(id, resourcePrefix)}
	converted := []iamBinding{}
	for _, b := range bindings {
		converted = append(converted, iamBinding{Role: b.Role, Members: b.Members})
	}
	resources = append(resources, expandIamPolicy(converted, id,
		resourcePrefix+"_iam_policy", resourcePrefix+"_iam_binding", resourcePrefix+"_iam_member",
		map[string]string{}, func(id, resourceType string, attrs map[string]string) terraformutils.Resource {
			return g.newSecretManagerResourceWithAttrs(id, resourceType, attrs)
		})...)
	g.Resources = append(g.Resources, resources...)
}

func (g *SecretManagerGenerator) appendSecretIam(ctx context.Context, svc *secretmanager.Service, name, resourcePrefix string) {
	policy, err := svc.Projects.Secrets.GetIamPolicy(name).Context(ctx).Do()
	if err != nil {
		log.Println(err)
		g.Resources = append(g.Resources, g.newSecretManagerResource(name, resourcePrefix))
		return
	}
	g.appendIam(name, resourcePrefix, policy.Bindings)
}

func (g *SecretManagerGenerator) appendRegionalSecretIam(ctx context.Context, svc *secretmanager.Service, name string) {
	policy, err := svc.Projects.Locations.Secrets.GetIamPolicy(name).Context(ctx).Do()
	if err != nil {
		log.Println(err)
		g.Resources = append(g.Resources, g.newSecretManagerResource(name, "google_secret_manager_regional_secret"))
		return
	}
	g.appendIam(name, "google_secret_manager_regional_secret", policy.Bindings)
}
