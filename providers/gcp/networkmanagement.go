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
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/compute/v1"
	netmgmtv1 "google.golang.org/api/networkmanagement/v1"
)

const networkManagementBaseURL = "https://networkmanagement.googleapis.com/v1/"
const networkManagementAlphaBaseURL = "https://networkmanagement.googleapis.com/v1alpha1/"

var networkManagementAllowEmptyValues = []string{"tags."}

var networkManagementAdditionalFields = map[string]interface{}{}

type NetworkManagementGenerator struct {
	GCPService
}

// InitResources enumerates Network Management (Network Intelligence
// Center) resources: connectivity tests and project VPC flow logs
// configs through the v1 client; organization VPC flow logs configs
// over raw REST (the pinned v1/v1beta1 clients expose no organization
// vpcFlowLogsConfigs surface, but the v1 REST method
// organizations.locations.vpcFlowLogsConfigs.list exists — verified
// against the Network Management REST reference — with the
// organization resolved through Cloud Resource Manager ancestry); and
// network monitoring providers over raw REST v1alpha1 (the only API
// version that serves projects.locations.networkMonitoringProviders,
// likewise verified against the REST reference).
func (g *NetworkManagementGenerator) InitResources() error {
	ctx := context.Background()
	service, err := netmgmtv1.NewService(ctx)
	if err != nil {
		return err
	}
	crmService, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		log.Println(err)
		crmService = nil
	}
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		log.Println(err)
		restClient = nil
	}
	return g.initResources(ctx, service, crmService, restClient)
}

func (g *NetworkManagementGenerator) initResources(ctx context.Context, service *netmgmtv1.Service, crmService *cloudresourcemanager.Service, restClient *http.Client) error {
	project := g.GetArgs()["project"].(string)
	globalParent := "projects/" + project + "/locations/global"
	attrs := map[string]string{"project": project}

	err := service.Projects.Locations.Global.ConnectivityTests.List(globalParent).Pages(ctx, func(page *netmgmtv1.ListConnectivityTestsResponse) error {
		for _, test := range page.Resources {
			g.Resources = append(g.Resources, g.newNetworkManagementResource(test.Name, "google_network_management_connectivity_test", attrs))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = service.Projects.Locations.VpcFlowLogsConfigs.List(globalParent).Pages(ctx, func(page *netmgmtv1.ListVpcFlowLogsConfigsResponse) error {
		for _, config := range page.VpcFlowLogsConfigs {
			g.Resources = append(g.Resources, g.newNetworkManagementResource(config.Name, "google_network_management_vpc_flow_logs_config", attrs))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}

	if restClient != nil {
		for _, location := range []string{"global", g.regionName()} {
			if location == "" {
				continue
			}
			providers, err := restListResources(ctx, restClient, networkManagementAlphaBaseURL+"projects/"+project+"/locations/"+location+"/networkMonitoringProviders")
			if err != nil {
				log.Println(err)
				continue
			}
			for _, provider := range providers {
				g.Resources = append(g.Resources, g.newNetworkManagementResource(provider.Name, "google_network_management_network_monitoring_provider", attrs))
			}
		}
		if orgID := orgIDFromAncestry(ctx, crmService, project); orgID != "" {
			configs, err := restListResources(ctx, restClient, networkManagementBaseURL+"organizations/"+orgID+"/locations/global/vpcFlowLogsConfigs")
			if err != nil {
				log.Println(err)
			}
			for _, config := range configs {
				g.Resources = append(g.Resources, g.newNetworkManagementResource(config.Name, "google_network_management_organization_vpc_flow_logs_config", map[string]string{"organization": orgID}))
			}
		}
	}
	return nil
}

// regionName returns the configured region's name (empty when the
// import runs without a region).
func (g *NetworkManagementGenerator) regionName() string {
	if arg, ok := g.GetArgs()["region"]; ok {
		if region, ok := arg.(compute.Region); ok {
			return region.Name
		}
	}
	return ""
}

func (g *NetworkManagementGenerator) newNetworkManagementResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
	parts := strings.Split(id, "/")
	return terraformutils.NewResource(
		id,
		parts[len(parts)-1],
		resourceType,
		g.ProviderName,
		attrs,
		networkManagementAllowEmptyValues,
		networkManagementAdditionalFields,
	)
}
