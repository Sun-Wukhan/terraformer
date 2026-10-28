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
	binauthzv1 "google.golang.org/api/binaryauthorization/v1"
)

var binaryAuthorizationAllowEmptyValues = []string{"tags."}

var binaryAuthorizationAdditionalFields = map[string]interface{}{}

type BinaryAuthorizationGenerator struct {
	GCPService
}

// InitResources enumerates Binary Authorization resources: the
// project's attestors (with the IAM policy, bindings and members
// expanded from each attestor's live IAM policy) and the project's
// singleton Binary Authorization policy, whose provider ID is the
// project itself.
func (g *BinaryAuthorizationGenerator) InitResources() error {
	ctx := context.Background()
	service, err := binauthzv1.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, service)
}

func (g *BinaryAuthorizationGenerator) initResources(ctx context.Context, service *binauthzv1.Service) error {
	project := g.GetArgs()["project"].(string)
	projectParent := "projects/" + project
	attrs := map[string]string{"project": project}

	if _, err := service.Projects.GetPolicy(projectParent).Context(ctx).Do(); err != nil {
		log.Println(err)
	} else {
		g.Resources = append(g.Resources, terraformutils.NewResource(
			project,
			project,
			"google_binary_authorization_policy",
			g.ProviderName,
			attrs,
			binaryAuthorizationAllowEmptyValues,
			binaryAuthorizationAdditionalFields,
		))
	}

	err := service.Projects.Attestors.List(projectParent).Pages(ctx, func(page *binauthzv1.ListAttestorsResponse) error {
		for _, attestor := range page.Attestors {
			g.Resources = append(g.Resources, g.newBinAuthzResource(attestor.Name, "google_binary_authorization_attestor", attrs))
			g.loadAttestorIam(ctx, service, attestor.Name)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (g *BinaryAuthorizationGenerator) newBinAuthzResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
	parts := strings.Split(id, "/")
	return terraformutils.NewResource(
		id,
		parts[len(parts)-1],
		resourceType,
		g.ProviderName,
		attrs,
		binaryAuthorizationAllowEmptyValues,
		binaryAuthorizationAdditionalFields,
	)
}

func (g *BinaryAuthorizationGenerator) loadAttestorIam(ctx context.Context, service *binauthzv1.Service, attestorName string) {
	policy, err := service.Projects.Attestors.GetIamPolicy(attestorName).Context(ctx).Do()
	if err != nil {
		log.Println(err)
		return
	}
	bindings := []iamBinding{}
	for _, binding := range policy.Bindings {
		bindings = append(bindings, iamBinding{Role: binding.Role, Members: binding.Members})
	}
	g.Resources = append(g.Resources, expandIamPolicy(bindings, attestorName,
		"google_binary_authorization_attestor_iam_policy", "google_binary_authorization_attestor_iam_binding", "google_binary_authorization_attestor_iam_member",
		map[string]string{"attestor": attestorName},
		func(id, resourceType string, attrs map[string]string) terraformutils.Resource {
			return g.newBinAuthzResource(id, resourceType, attrs)
		})...)
}
