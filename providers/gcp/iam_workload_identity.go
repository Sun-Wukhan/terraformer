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
	"google.golang.org/api/iam/v1"
)

// initWorkloadIdentityResources enumerates Workload Identity Federation
// configuration for the project with the iam v1 REST client: workload
// identity pools (with their IAM policy, bindings and members), pool
// namespaces, pool managed identities, pool providers, and the global
// OAuth clients with their credentials. Pools and OAuth clients live in
// the "global" location.
//
// Note: the iam/v2 deny policies, the iam/v3beta policy bindings and
// principal access boundary policies, and the workforce pool family
// are enumerated by the IamOrgGenerator (iam_org.go) and by
// initWorkforceResources (iam_workforce.go) through organization
// ancestry — the pinned iam/v2 client does expose ListPolicies and
// workforce pools are listable under the organization, so the older
// project-scope-only skip note for those families no longer applies.
func (g *IamGenerator) initWorkloadIdentityResources(ctx context.Context, service *iam.Service) error {
	project := g.GetArgs()["project"].(string)
	parent := "projects/" + project + "/locations/global"
	if err := g.loadWorkloadIdentityPools(ctx, service, parent); err != nil {
		return err
	}
	return g.loadOauthClients(ctx, service, parent)
}

func workloadIdentityLastSegment(name string) string {
	parts := strings.Split(name, "/")
	return parts[len(parts)-1]
}

func (g *IamGenerator) newWorkloadIdentityResource(id, name, resourceType string, attrs map[string]string) terraformutils.Resource {
	if attrs == nil {
		attrs = map[string]string{}
	}
	if _, ok := attrs["project"]; !ok {
		attrs["project"] = g.GetArgs()["project"].(string)
	}
	return terraformutils.NewResource(
		id,
		name,
		resourceType,
		g.ProviderName,
		attrs,
		IamAllowEmptyValues,
		IamAdditionalFields,
	)
}

func (g *IamGenerator) loadWorkloadIdentityPools(ctx context.Context, service *iam.Service, parent string) error {
	return service.Projects.Locations.WorkloadIdentityPools.List(parent).Pages(ctx, func(page *iam.ListWorkloadIdentityPoolsResponse) error {
		for _, pool := range page.WorkloadIdentityPools {
			poolID := workloadIdentityLastSegment(pool.Name)
			g.Resources = append(g.Resources, g.newWorkloadIdentityResource(
				pool.Name,
				poolID,
				"google_iam_workload_identity_pool",
				map[string]string{"workload_identity_pool_id": poolID},
			))
			g.Resources = append(g.Resources, g.loadWorkloadIdentityPoolIam(ctx, service, pool.Name, poolID)...)
			if err := g.loadWorkloadIdentityPoolNamespaces(ctx, service, pool.Name, poolID); err != nil {
				return err
			}
			if err := g.loadWorkloadIdentityPoolProviders(ctx, service, pool.Name, poolID); err != nil {
				return err
			}
		}
		return nil
	})
}

// loadWorkloadIdentityPoolIam appends the
// google_iam_workload_identity_pool_iam_policy/_binding/_member resources
// for one pool; the provider ID for all three is the pool's full name.
func (g *IamGenerator) loadWorkloadIdentityPoolIam(ctx context.Context, service *iam.Service, poolName, poolID string) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	policy, err := service.Projects.Locations.WorkloadIdentityPools.GetIamPolicy(poolName, &iam.GetIamPolicyRequest{}).Context(ctx).Do()
	if err != nil {
		log.Println(err)
		return resources
	}
	poolAttrs := map[string]string{"workload_identity_pool_id": poolID}
	resources = append(resources, g.newWorkloadIdentityResource(
		poolName,
		poolID,
		"google_iam_workload_identity_pool_iam_policy",
		poolAttrs,
	))
	for _, binding := range policy.Bindings {
		bindingAttrs := map[string]string{"role": binding.Role}
		for k, v := range poolAttrs {
			bindingAttrs[k] = v
		}
		resources = append(resources, g.newWorkloadIdentityResource(
			poolName,
			poolID,
			"google_iam_workload_identity_pool_iam_binding",
			bindingAttrs,
		))
		for _, member := range binding.Members {
			memberAttrs := map[string]string{"role": binding.Role, "member": member}
			for k, v := range poolAttrs {
				memberAttrs[k] = v
			}
			resources = append(resources, g.newWorkloadIdentityResource(
				poolName,
				poolID,
				"google_iam_workload_identity_pool_iam_member",
				memberAttrs,
			))
		}
	}
	return resources
}

func (g *IamGenerator) loadWorkloadIdentityPoolNamespaces(ctx context.Context, service *iam.Service, poolName, poolID string) error {
	return service.Projects.Locations.WorkloadIdentityPools.Namespaces.List(poolName).Pages(ctx, func(page *iam.ListWorkloadIdentityPoolNamespacesResponse) error {
		for _, namespace := range page.WorkloadIdentityPoolNamespaces {
			namespaceID := workloadIdentityLastSegment(namespace.Name)
			g.Resources = append(g.Resources, g.newWorkloadIdentityResource(
				namespace.Name,
				namespaceID,
				"google_iam_workload_identity_pool_namespace",
				map[string]string{
					"workload_identity_pool_id": poolID,
					"name":                      namespaceID,
				},
			))
			if err := g.loadWorkloadIdentityPoolManagedIdentities(ctx, service, namespace.Name, poolID, namespaceID); err != nil {
				return err
			}
		}
		return nil
	})
}

func (g *IamGenerator) loadWorkloadIdentityPoolManagedIdentities(ctx context.Context, service *iam.Service, namespaceName, poolID, namespaceID string) error {
	return service.Projects.Locations.WorkloadIdentityPools.Namespaces.ManagedIdentities.List(namespaceName).Pages(ctx, func(page *iam.ListWorkloadIdentityPoolManagedIdentitiesResponse) error {
		for _, identity := range page.WorkloadIdentityPoolManagedIdentities {
			g.Resources = append(g.Resources, g.newWorkloadIdentityResource(
				identity.Name,
				workloadIdentityLastSegment(identity.Name),
				"google_iam_workload_identity_pool_managed_identity",
				map[string]string{
					"workload_identity_pool_id":           poolID,
					"workload_identity_pool_namespace_id": namespaceID,
					"name":                                workloadIdentityLastSegment(identity.Name),
				},
			))
		}
		return nil
	})
}

func (g *IamGenerator) loadWorkloadIdentityPoolProviders(ctx context.Context, service *iam.Service, poolName, poolID string) error {
	return service.Projects.Locations.WorkloadIdentityPools.Providers.List(poolName).Pages(ctx, func(page *iam.ListWorkloadIdentityPoolProvidersResponse) error {
		for _, provider := range page.WorkloadIdentityPoolProviders {
			g.Resources = append(g.Resources, g.newWorkloadIdentityResource(
				provider.Name,
				workloadIdentityLastSegment(provider.Name),
				"google_iam_workload_identity_pool_provider",
				map[string]string{
					"workload_identity_pool_id":          poolID,
					"workload_identity_pool_provider_id": workloadIdentityLastSegment(provider.Name),
				},
			))
		}
		return nil
	})
}

func (g *IamGenerator) loadOauthClients(ctx context.Context, service *iam.Service, parent string) error {
	return service.Projects.Locations.OauthClients.List(parent).Pages(ctx, func(page *iam.ListOauthClientsResponse) error {
		for _, client := range page.OauthClients {
			clientID := workloadIdentityLastSegment(client.Name)
			g.Resources = append(g.Resources, g.newWorkloadIdentityResource(
				client.Name,
				clientID,
				"google_iam_oauth_client",
				map[string]string{
					"location":        "global",
					"oauth_client_id": clientID,
				},
			))
			if err := g.loadOauthClientCredentials(ctx, service, client.Name, clientID); err != nil {
				return err
			}
		}
		return nil
	})
}

func (g *IamGenerator) loadOauthClientCredentials(ctx context.Context, service *iam.Service, clientName, clientID string) error {
	// The credentials list call has no page token in the discovery
	// document, so it is a single request rather than a Pages loop.
	page, err := service.Projects.Locations.OauthClients.Credentials.List(clientName).Context(ctx).Do()
	if err != nil {
		return err
	}
	for _, credential := range page.OauthClientCredentials {
		g.Resources = append(g.Resources, g.newWorkloadIdentityResource(
			credential.Name,
			workloadIdentityLastSegment(credential.Name),
			"google_iam_oauth_client_credential",
			map[string]string{
				"location":        "global",
				"oauth_client_id": clientID,
				"name":            workloadIdentityLastSegment(credential.Name),
			},
		))
	}
	return nil
}
