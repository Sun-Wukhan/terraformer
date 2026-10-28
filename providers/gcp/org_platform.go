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
	apikeysv2 "google.golang.org/api/apikeys/v2"
	assuredworkloadsv1 "google.golang.org/api/assuredworkloads/v1"
	crmv1 "google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/compute/v1"
	websecurityscannerv1 "google.golang.org/api/websecurityscanner/v1"
)

var orgPlatformAllowEmptyValues = []string{""}

var orgPlatformAdditionalFields = map[string]interface{}{}

// OrgPlatformGenerator enumerates the long-tail platform services of
// the iam-org cluster:
//
//   - API keys (apikeys/v2): projects/{p}/locations/global/keys,
//     provider ID is the key resource name.
//   - Public CA external account keys (raw REST; the pinned publicca
//     client has only a Create call for external account keys, but
//     the v1 discovery document has the list endpoint):
//     projects/{p}/locations/global/externalAccountKeys.
//   - Web Security Scanner scan configs (websecurityscanner/v1):
//     projects/{p}/scanConfigs.
//   - Assured Workloads (assuredworkloads/v1): workloads of the
//     project's organization in the configured region,
//     organizations/{org}/locations/{region}/workloads.
//   - Cloud Quotas (raw REST; no generated cloudquotas client exists
//     in google.golang.org/api): quota preferences and the quota
//     adjuster settings singleton under
//     projects/{p}/locations/global.
//   - License Manager (raw REST; no generated licensemanager client
//     exists): projects/{p}/locations/global/configurations.
//   - Secure Source Manager (raw REST; no generated client exists):
//     locations are listed through the standard locations endpoint,
//     then instances per location, repositories per instance with
//     their branch rules and hooks, and the instance/repository IAM
//     trios via :getIamPolicy.
//
// Deliberately not generated:
//   - google_service_usage_consumer_quota_override: overrides exist
//     only under a fully qualified
//     projects/{p}/services/{service}/consumerQuotaMetrics/{metric}/limits/{limit}
//     parent; the serviceusage/v1 client has no project-level (or
//     service-level) override list call, so the metric/limit parents
//     cannot be enumerated.
//   - google_service_usage_v2_consumer_policy: a consumer policy is a
//     per-service get-only object in the v2 API; there is no list
//     endpoint in the v2 discovery document and no v2 client in
//     google.golang.org/api.
//   - google_workload_identity_service_agent: the workload identity
//     service agent is provisioned through a generateServiceIdentity
//     call for a workload identity pool; there is no listable or
//     gettable object and reading it would mutate state (same
//     reasoning as the service identities in iam_org.go).
type OrgPlatformGenerator struct {
	GCPService
}

func (g *OrgPlatformGenerator) InitResources() error {
	ctx := context.Background()
	akSvc, err := apikeysv2.NewService(ctx)
	if err != nil {
		return err
	}
	scanSvc, err := websecurityscannerv1.NewService(ctx)
	if err != nil {
		log.Println(err)
		scanSvc = nil
	}
	awSvc, err := assuredworkloadsv1.NewService(ctx)
	if err != nil {
		log.Println(err)
		awSvc = nil
	}
	crmSvc, err := crmv1.NewService(ctx)
	if err != nil {
		log.Println(err)
		crmSvc = nil
	}
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, akSvc, scanSvc, awSvc, crmSvc, restClient)
}

func (g *OrgPlatformGenerator) newOrgPlatformResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
	if attrs == nil {
		attrs = map[string]string{}
	}
	parts := strings.Split(id, "/")
	return terraformutils.NewResource(id, parts[len(parts)-1], resourceType, g.ProviderName, attrs, orgPlatformAllowEmptyValues, orgPlatformAdditionalFields)
}

func (g *OrgPlatformGenerator) initResources(ctx context.Context, akSvc *apikeysv2.Service, scanSvc *websecurityscannerv1.Service, awSvc *assuredworkloadsv1.Service, crmSvc *crmv1.Service, restClient *http.Client) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region)
	regionName := region.Name

	if akSvc != nil {
		err := akSvc.Projects.Locations.Keys.List("projects/"+project+"/locations/global").Pages(ctx, func(page *apikeysv2.V2ListKeysResponse) error {
			for _, key := range page.Keys {
				g.Resources = append(g.Resources, g.newOrgPlatformResource(key.Name, "google_apikeys_key", map[string]string{"project": project}))
			}
			return nil
		})
		if err != nil {
			log.Println(err)
		}
	}
	if scanSvc != nil {
		err := scanSvc.Projects.ScanConfigs.List("projects/"+project).Pages(ctx, func(page *websecurityscannerv1.ListScanConfigsResponse) error {
			for _, cfg := range page.ScanConfigs {
				g.Resources = append(g.Resources, g.newOrgPlatformResource(cfg.Name, "google_security_scanner_scan_config", map[string]string{"project": project}))
			}
			return nil
		})
		if err != nil {
			log.Println(err)
		}
	}
	g.loadAssuredWorkloads(ctx, awSvc, crmSvc, project, regionName)
	g.loadRawREST(ctx, restClient, project)
	return nil
}

func (g *OrgPlatformGenerator) loadAssuredWorkloads(ctx context.Context, awSvc *assuredworkloadsv1.Service, crmSvc *crmv1.Service, project, regionName string) {
	if awSvc == nil || crmSvc == nil || regionName == "" {
		return
	}
	ancestry, err := crmSvc.Projects.GetAncestry(project, &crmv1.GetAncestryRequest{}).Context(ctx).Do()
	if err != nil {
		log.Println(err)
		return
	}
	orgID := ""
	for _, ancestor := range ancestry.Ancestor {
		if ancestor.ResourceId != nil && ancestor.ResourceId.Type == "organization" {
			orgID = ancestor.ResourceId.Id
		}
	}
	if orgID == "" {
		return
	}
	parent := "organizations/" + orgID + "/locations/" + regionName
	err = awSvc.Organizations.Locations.Workloads.List(parent).Pages(ctx, func(page *assuredworkloadsv1.GoogleCloudAssuredworkloadsV1ListWorkloadsResponse) error {
		for _, workload := range page.Workloads {
			g.Resources = append(g.Resources, g.newOrgPlatformResource(workload.Name, "google_assured_workloads_workload", map[string]string{"organization": orgID, "location": regionName}))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (g *OrgPlatformGenerator) loadRawREST(ctx context.Context, restClient *http.Client, project string) {
	if restClient == nil {
		return
	}
	projectParent := "projects/" + project
	// Public CA external account keys.
	if items, err := restListResources(ctx, restClient, "https://publicca.googleapis.com/v1/"+projectParent+"/locations/global/externalAccountKeys"); err == nil {
		for _, item := range items {
			g.Resources = append(g.Resources, g.newOrgPlatformResource(item.Name, "google_public_ca_external_account_key", map[string]string{"project": project}))
		}
	} else {
		log.Println(err)
	}
	// Cloud Quotas preferences + adjuster settings singleton.
	if items, err := restListResources(ctx, restClient, "https://cloudquotas.googleapis.com/v1/"+projectParent+"/locations/global/quotaPreferences"); err == nil {
		for _, item := range items {
			g.Resources = append(g.Resources, g.newOrgPlatformResource(item.Name, "google_cloud_quotas_quota_preference", map[string]string{"parent": projectParent}))
		}
	} else {
		log.Println(err)
	}
	if body, err := restDo(ctx, restClient, http.MethodGet, "https://cloudquotas.googleapis.com/v1/"+projectParent+"/locations/global/quotaAdjusterSettings", nil); err == nil && len(body) > 2 {
		g.Resources = append(g.Resources, g.newOrgPlatformResource(projectParent+"/locations/global/quotaAdjusterSettings", "google_cloud_quotas_quota_adjuster_settings", map[string]string{"parent": projectParent}))
	} else if err != nil {
		log.Println(err)
	}
	// License Manager configurations.
	if items, err := restListResources(ctx, restClient, "https://licensemanager.googleapis.com/v1/"+projectParent+"/locations/global/configurations"); err == nil {
		for _, item := range items {
			g.Resources = append(g.Resources, g.newOrgPlatformResource(item.Name, "google_license_manager_configuration", map[string]string{"project": project, "location": "global"}))
		}
	} else {
		log.Println(err)
	}
	g.loadSecureSourceManager(ctx, restClient, project)
}

func (g *OrgPlatformGenerator) loadSecureSourceManager(ctx context.Context, restClient *http.Client, project string) {
	base := "https://securesourcemanager.googleapis.com/v1/projects/" + project
	locations, err := restListResources(ctx, restClient, base+"/locations")
	if err != nil {
		log.Println(err)
		return
	}
	for _, loc := range locations {
		locID := loc.Name[strings.LastIndex(loc.Name, "/")+1:]
		instances, err := restListResources(ctx, restClient, "https://securesourcemanager.googleapis.com/v1/"+loc.Name+"/instances")
		if err != nil {
			log.Println(err)
			continue
		}
		for _, instance := range instances {
			g.Resources = append(g.Resources, g.newOrgPlatformResource(instance.Name, "google_secure_source_manager_instance", map[string]string{"project": project, "location": locID}))
			g.loadSecureSourceManagerIam(ctx, restClient, instance.Name,
				"google_secure_source_manager_instance_iam_policy", "google_secure_source_manager_instance_iam_binding", "google_secure_source_manager_instance_iam_member",
				map[string]string{"project": project, "location": locID})
			g.loadSecureSourceManagerRepositories(ctx, restClient, instance.Name, project, locID)
		}
	}
}

func (g *OrgPlatformGenerator) loadSecureSourceManagerRepositories(ctx context.Context, restClient *http.Client, instanceName, project, locID string) {
	repos, err := restListResources(ctx, restClient, "https://securesourcemanager.googleapis.com/v1/"+instanceName+"/repositories")
	if err != nil {
		log.Println(err)
		return
	}
	for _, repo := range repos {
		g.Resources = append(g.Resources, g.newOrgPlatformResource(repo.Name, "google_secure_source_manager_repository", map[string]string{"project": project, "location": locID}))
		g.loadSecureSourceManagerIam(ctx, restClient, repo.Name,
			"google_secure_source_manager_repository_iam_policy", "google_secure_source_manager_repository_iam_binding", "google_secure_source_manager_repository_iam_member",
			map[string]string{"project": project, "location": locID})
		if rules, err := restListResources(ctx, restClient, "https://securesourcemanager.googleapis.com/v1/"+repo.Name+"/branchRules"); err == nil {
			for _, rule := range rules {
				g.Resources = append(g.Resources, g.newOrgPlatformResource(rule.Name, "google_secure_source_manager_branch_rule", map[string]string{"project": project, "location": locID}))
			}
		} else {
			log.Println(err)
		}
		if hooks, err := restListResources(ctx, restClient, "https://securesourcemanager.googleapis.com/v1/"+repo.Name+"/hooks"); err == nil {
			for _, hook := range hooks {
				g.Resources = append(g.Resources, g.newOrgPlatformResource(hook.Name, "google_secure_source_manager_hook", map[string]string{"project": project, "location": locID}))
			}
		} else {
			log.Println(err)
		}
	}
}

func (g *OrgPlatformGenerator) loadSecureSourceManagerIam(ctx context.Context, restClient *http.Client, parentName, policyType, bindingType, memberType string, attrs map[string]string) {
	bindings, err := restGetIamPolicy(ctx, restClient, "https://securesourcemanager.googleapis.com/v1/"+parentName+":getIamPolicy")
	if err != nil {
		log.Println(err)
		return
	}
	g.Resources = append(g.Resources, expandIamPolicy(bindings, parentName, policyType, bindingType, memberType, attrs, g.newOrgPlatformResource)...)
}
