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

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"google.golang.org/api/compute/v1"
	siteverification "google.golang.org/api/siteverification/v1"
)

type SiteVerificationGenerator struct {
	GCPService
}

// InitResources enumerates Google Site Verification web resources
// visible to the caller's credentials (the API is user-scoped, not
// project-scoped), together with one owner resource per verified
// owner of each web resource. Provider IDs follow the provider's
// import formats: webResource/{id} for web resources and
// webResource/{id}/{email} for owners.
func (g *SiteVerificationGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := siteverification.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, svc)
}

func (g *SiteVerificationGenerator) initResources(ctx context.Context, svc *siteverification.Service) error {
	_ = g.GetArgs()["project"].(string)
	_ = g.GetArgs()["region"].(compute.Region)
	list, err := svc.WebResource.List().Context(ctx).Do()
	if err != nil {
		return err
	}
	for _, item := range list.Items {
		id := "webResource/" + item.Id
		name := item.Id
		if item.Site != nil && item.Site.Identifier != "" {
			name = item.Site.Identifier
		}
		g.Resources = append(g.Resources, terraformutils.NewResource(
			id, name, "google_site_verification_web_resource",
			g.ProviderName, map[string]string{}, []string{}, map[string]interface{}{}))
		for _, owner := range item.Owners {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				id+"/"+owner, owner, "google_site_verification_owner",
				g.ProviderName, map[string]string{}, []string{}, map[string]interface{}{}))
		}
	}
	return nil
}
