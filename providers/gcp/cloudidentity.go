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
	cloudidentity "google.golang.org/api/cloudidentity/v1"
)

// CloudIdentityGenerator enumerates Cloud Identity via the generated
// cloudidentity v1 client. Groups and policies are customer-scoped,
// not project-scoped: enumeration uses the API's own-organization
// alias parent "customers/my_customer" (documented on the generated
// client), which resolves to the customer of the calling identity.
// For every group, its memberships are listed as well.
type CloudIdentityGenerator struct {
	GCPService
}

const cloudIdentityCustomerParent = "customers/my_customer"

func (g *CloudIdentityGenerator) InitResources() error {
	ctx := context.Background()
	service, err := cloudidentity.NewService(ctx)
	if err != nil {
		return err
	}
	g.initResources(ctx, service)
	return nil
}

func (g *CloudIdentityGenerator) newResource(id, resourceType string) terraformutils.Resource {
	return terraformutils.NewResource(
		id, opsLastSegment(id), resourceType, g.ProviderName,
		map[string]string{}, []string{}, map[string]interface{}{},
	)
}

func (g *CloudIdentityGenerator) initResources(ctx context.Context, service *cloudidentity.Service) {
	if err := service.Groups.List().Parent(cloudIdentityCustomerParent).Pages(ctx, func(page *cloudidentity.ListGroupsResponse) error {
		for _, group := range page.Groups {
			g.Resources = append(g.Resources, g.newResource(group.Name, "google_cloud_identity_group"))
			if err := service.Groups.Memberships.List(group.Name).Pages(ctx, func(mPage *cloudidentity.ListMembershipsResponse) error {
				for _, membership := range mPage.Memberships {
					g.Resources = append(g.Resources, g.newResource(membership.Name, "google_cloud_identity_group_membership"))
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
	// Policies have no parent parameter; they are filtered by
	// customer with a CEL filter clause.
	if err := service.Policies.List().Filter("customer == \""+cloudIdentityCustomerParent+"\"").Pages(ctx, func(page *cloudidentity.ListPoliciesResponse) error {
		for _, policy := range page.Policies {
			g.Resources = append(g.Resources, g.newResource(policy.Name, "google_cloud_identity_policy"))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
}
