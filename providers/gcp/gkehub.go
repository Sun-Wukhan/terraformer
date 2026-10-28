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
	"net/http"
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"google.golang.org/api/compute/v1"
	gkehubv1 "google.golang.org/api/gkehub/v1"
)

var gkeHubAllowEmptyValues = []string{""}

var gkeHubAdditionalFields = map[string]interface{}{}

const (
	gkeHubBaseURL      = "https://gkehub.googleapis.com/v1/"
	gkeHubAlphaBaseURL = "https://gkehub.googleapis.com/v1alpha/"
)

// GkeHubGenerator enumerates GKE Hub (fleet) resources: fleets,
// features (with IAM and per-membership feature state), memberships
// (with IAM, bindings and RBAC role bindings), scopes (with IAM,
// namespaces and RBAC role bindings) and rollout sequences.
//
// Fleet resources live in the "global" location; memberships also
// exist in regional locations, so memberships are listed in both the
// global location and the configured region. Features, fleets,
// scopes and rollout sequences are global-only. Feature memberships
// and rollout sequences have no service in the pinned
// google.golang.org/api gkehub/v1 client (feature memberships are not
// a generated service at all, and rollout sequences exist only in
// the v1alpha API), so both run over raw REST.
type GkeHubGenerator struct {
	GCPService
}

func (g *GkeHubGenerator) InitResources() error {
	ctx := context.Background()
	hubService, err := gkehubv1.NewService(ctx)
	if err != nil {
		return err
	}
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, hubService, restClient)
}

func (g *GkeHubGenerator) initResources(ctx context.Context, hubService *gkehubv1.Service, restClient *http.Client) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	globalParent := "projects/" + project + "/locations/global"
	baseAttrs := map[string]string{"project": project}

	if err := g.loadFleets(ctx, hubService, globalParent, baseAttrs); err != nil {
		return err
	}
	if err := g.loadFeatures(ctx, hubService, globalParent, baseAttrs); err != nil {
		return err
	}
	g.loadFeatureMemberships(ctx, restClient)
	if err := g.loadScopes(ctx, hubService, globalParent, baseAttrs); err != nil {
		return err
	}
	for _, parent := range []string{globalParent, "projects/" + project + "/locations/" + region} {
		if err := g.loadMemberships(ctx, hubService, parent, baseAttrs); err != nil {
			return err
		}
	}
	g.loadRolloutSequences(ctx, restClient, globalParent, baseAttrs)
	return nil
}

func (g *GkeHubGenerator) newGkeHubResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
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
		gkeHubAllowEmptyValues,
		gkeHubAdditionalFields,
	)
}

func (g *GkeHubGenerator) loadGkeHubIam(getPolicy func() (*gkehubv1.Policy, error), name, policyType, bindingType, memberType, parentKey string, baseAttrs map[string]string) {
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
		bindings, name, policyType, bindingType, memberType, parentAttrs, g.newGkeHubResource)...)
}

func (g *GkeHubGenerator) loadFleets(ctx context.Context, hubService *gkehubv1.Service, parent string, baseAttrs map[string]string) error {
	return hubService.Projects.Locations.Fleets.List(parent).Pages(ctx, func(page *gkehubv1.ListFleetsResponse) error {
		for _, fleet := range page.Fleets {
			g.Resources = append(g.Resources, g.newGkeHubResource(fleet.Name, "google_gke_hub_fleet", baseAttrs))
		}
		return nil
	})
}

func (g *GkeHubGenerator) loadFeatures(ctx context.Context, hubService *gkehubv1.Service, parent string, baseAttrs map[string]string) error {
	return hubService.Projects.Locations.Features.List(parent).Pages(ctx, func(page *gkehubv1.ListFeaturesResponse) error {
		for _, feature := range page.Resources {
			g.Resources = append(g.Resources, g.newGkeHubResource(feature.Name, "google_gke_hub_feature", baseAttrs))
			g.loadGkeHubIam(func() (*gkehubv1.Policy, error) {
				return hubService.Projects.Locations.Features.GetIamPolicy(feature.Name).Context(ctx).Do()
			}, feature.Name,
				"google_gke_hub_feature_iam_policy", "google_gke_hub_feature_iam_binding", "google_gke_hub_feature_iam_member",
				"location", baseAttrs)
		}
		return nil
	})
}

// loadFeatureMemberships lists per-membership feature state over raw
// REST for every feature resource emitted so far: the pinned client
// has no feature memberships service.
func (g *GkeHubGenerator) loadFeatureMemberships(ctx context.Context, restClient *http.Client) {
	features := []string{}
	for _, resource := range g.Resources {
		if resource.InstanceInfo.Type == "google_gke_hub_feature" {
			features = append(features, resource.InstanceState.ID)
		}
	}
	for _, feature := range features {
		items, err := restListResources(ctx, restClient, gkeHubBaseURL+feature+"/memberships")
		if err != nil {
			log.Println(err)
			continue
		}
		for _, item := range items {
			g.Resources = append(g.Resources, g.newGkeHubResource(item.Name, "google_gke_hub_feature_membership", map[string]string{"project": g.GetArgs()["project"].(string)}))
		}
	}
}

func (g *GkeHubGenerator) loadMemberships(ctx context.Context, hubService *gkehubv1.Service, parent string, baseAttrs map[string]string) error {
	return hubService.Projects.Locations.Memberships.List(parent).Pages(ctx, func(page *gkehubv1.ListMembershipsResponse) error {
		for _, membership := range page.Resources {
			g.Resources = append(g.Resources, g.newGkeHubResource(membership.Name, "google_gke_hub_membership", baseAttrs))
			g.loadGkeHubIam(func() (*gkehubv1.Policy, error) {
				return hubService.Projects.Locations.Memberships.GetIamPolicy(membership.Name).Context(ctx).Do()
			}, membership.Name,
				"google_gke_hub_membership_iam_policy", "google_gke_hub_membership_iam_binding", "google_gke_hub_membership_iam_member",
				"membership_id", baseAttrs)
			if err := g.loadMembershipChildren(ctx, hubService, membership.Name, baseAttrs); err != nil {
				return err
			}
		}
		return nil
	})
}

func (g *GkeHubGenerator) loadMembershipChildren(ctx context.Context, hubService *gkehubv1.Service, membershipName string, baseAttrs map[string]string) error {
	err := hubService.Projects.Locations.Memberships.Bindings.List(membershipName).Pages(ctx, func(page *gkehubv1.ListMembershipBindingsResponse) error {
		for _, binding := range page.MembershipBindings {
			g.Resources = append(g.Resources, g.newGkeHubResource(binding.Name, "google_gke_hub_membership_binding", baseAttrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	return hubService.Projects.Locations.Memberships.Rbacrolebindings.List(membershipName).Pages(ctx, func(page *gkehubv1.ListMembershipRBACRoleBindingsResponse) error {
		for _, binding := range page.Rbacrolebindings {
			g.Resources = append(g.Resources, g.newGkeHubResource(binding.Name, "google_gke_hub_membership_rbac_role_binding", baseAttrs))
		}
		return nil
	})
}

func (g *GkeHubGenerator) loadScopes(ctx context.Context, hubService *gkehubv1.Service, parent string, baseAttrs map[string]string) error {
	return hubService.Projects.Locations.Scopes.List(parent).Pages(ctx, func(page *gkehubv1.ListScopesResponse) error {
		for _, scope := range page.Scopes {
			g.Resources = append(g.Resources, g.newGkeHubResource(scope.Name, "google_gke_hub_scope", baseAttrs))
			g.loadGkeHubIam(func() (*gkehubv1.Policy, error) {
				return hubService.Projects.Locations.Scopes.GetIamPolicy(scope.Name).Context(ctx).Do()
			}, scope.Name,
				"google_gke_hub_scope_iam_policy", "google_gke_hub_scope_iam_binding", "google_gke_hub_scope_iam_member",
				"scope_id", baseAttrs)
			if err := g.loadScopeChildren(ctx, hubService, scope.Name, baseAttrs); err != nil {
				return err
			}
		}
		return nil
	})
}

func (g *GkeHubGenerator) loadScopeChildren(ctx context.Context, hubService *gkehubv1.Service, scopeName string, baseAttrs map[string]string) error {
	err := hubService.Projects.Locations.Scopes.Namespaces.List(scopeName).Pages(ctx, func(page *gkehubv1.ListScopeNamespacesResponse) error {
		for _, namespace := range page.ScopeNamespaces {
			g.Resources = append(g.Resources, g.newGkeHubResource(namespace.Name, "google_gke_hub_namespace", baseAttrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	return hubService.Projects.Locations.Scopes.Rbacrolebindings.List(scopeName).Pages(ctx, func(page *gkehubv1.ListScopeRBACRoleBindingsResponse) error {
		for _, binding := range page.Rbacrolebindings {
			g.Resources = append(g.Resources, g.newGkeHubResource(binding.Name, "google_gke_hub_scope_rbac_role_binding", baseAttrs))
		}
		return nil
	})
}

// loadRolloutSequences lists fleet rollout sequences over raw REST:
// they exist only in the v1alpha API, which the pinned module does
// not generate a client for.
func (g *GkeHubGenerator) loadRolloutSequences(ctx context.Context, restClient *http.Client, parent string, baseAttrs map[string]string) {
	items, err := restListResources(ctx, restClient, gkeHubAlphaBaseURL+parent+"/rolloutSequences")
	if err != nil {
		log.Println(err)
		return
	}
	for _, item := range items {
		g.Resources = append(g.Resources, g.newGkeHubResource(item.Name, "google_gke_hub_rollout_sequence", baseAttrs))
	}
}
