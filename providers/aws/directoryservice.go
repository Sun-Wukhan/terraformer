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

package aws

import (
	"context"
	"log"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/directoryservice"
	dstypes "github.com/aws/aws-sdk-go-v2/service/directoryservice/types"
	"github.com/aws/aws-sdk-go-v2/service/directoryservicedata"
)

var directoryServiceAllowEmptyValues = []string{"tags."}

type DirectoryServiceGenerator struct {
	AWSService
}

// InitResources enumerates Directory Service directories in the
// configured region (DescribeDirectories) and, per directory:
// conditional forwarders, directory settings, IP routes, log
// subscriptions, RADIUS settings, additional regions, shared
// directories and trusts, plus Directory Service Data users and
// groups.
//
// Provider importer ID formats (terraform-provider-aws docs):
//   - aws_directory_service_directory: the directory ID.
//   - aws_directory_service_conditional_forwarder:
//     directory_id:remote_domain_name.
//   - aws_directory_service_directory_setting: the directory ID (the
//     settings collection of one directory), emitted when the
//     directory has settings entries.
//   - aws_directory_service_ip_route: directory_id:cidr_ip.
//   - aws_directory_service_log_subscription: the directory ID.
//   - aws_directory_service_radius_settings: the directory ID,
//     emitted when RADIUS properties are configured.
//   - aws_directory_service_region: directory_id:region_name.
//   - aws_directory_service_shared_directory: the shared directory ID.
//   - aws_directory_service_trust: the trust ID.
//   - aws_directoryservicedata_user / aws_directoryservicedata_group:
//     directory_id,sam_account_name (verified for user; group uses
//     the same documented identity format).
//
// Intentionally not generated:
//   - aws_directory_service_ip_routes_exclusive conflicts with the
//     individual aws_directory_service_ip_route resources generated
//     here, so emitting both would double-manage the routes.
//   - aws_directory_service_shared_directory_accepter is the
//     accepter side of a directory share in the receiving account.
func (g *DirectoryServiceGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := directoryservice.NewFromConfig(config)
	dataSvc := directoryservicedata.NewFromConfig(config)

	p := directoryservice.NewDescribeDirectoriesPaginator(svc, &directoryservice.DescribeDirectoriesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, directory := range page.DirectoryDescriptions {
			directoryID := StringValue(directory.DirectoryId)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				directoryID,
				StringValue(directory.Name),
				"aws_directory_service_directory",
				"aws",
				directoryServiceAllowEmptyValues))
			g.addConditionalForwarders(svc, directoryID)
			g.addDirectorySettings(svc, directoryID)
			g.addIPRoutes(svc, directoryID)
			g.addLogSubscriptions(svc, directoryID)
			g.addRadiusSettings(directory, directoryID)
			g.addRegions(svc, directoryID)
			g.addSharedDirectories(svc, directoryID)
			g.addTrusts(svc, directoryID)
			g.addDataUsersAndGroups(dataSvc, directoryID)
		}
	}
	return nil
}

func (g *DirectoryServiceGenerator) addConditionalForwarders(svc *directoryservice.Client, directoryID string) {
	out, err := svc.DescribeConditionalForwarders(context.TODO(), &directoryservice.DescribeConditionalForwardersInput{
		DirectoryId: &directoryID,
	})
	if err != nil {
		log.Println(err)
		return
	}
	for _, forwarder := range out.ConditionalForwarders {
		domain := StringValue(forwarder.RemoteDomainName)
		g.Resources = append(g.Resources, terraformutils.NewResource(
			directoryID+":"+domain,
			domain,
			"aws_directory_service_conditional_forwarder",
			"aws",
			map[string]string{"directory_id": directoryID},
			directoryServiceAllowEmptyValues,
			map[string]interface{}{}))
	}
}

func (g *DirectoryServiceGenerator) addDirectorySettings(svc *directoryservice.Client, directoryID string) {
	input := &directoryservice.DescribeSettingsInput{
		DirectoryId: &directoryID,
	}
	hasSettings := false
	for {
		page, err := svc.DescribeSettings(context.TODO(), input)
		if err != nil {
			log.Println(err)
			return
		}
		if len(page.SettingEntries) > 0 {
			hasSettings = true
		}
		if page.NextToken == nil {
			break
		}
		input.NextToken = page.NextToken
	}
	if !hasSettings {
		return
	}
	g.Resources = append(g.Resources, terraformutils.NewResource(
		directoryID,
		directoryID,
		"aws_directory_service_directory_setting",
		"aws",
		map[string]string{"directory_id": directoryID},
		directoryServiceAllowEmptyValues,
		map[string]interface{}{}))
}

func (g *DirectoryServiceGenerator) addIPRoutes(svc *directoryservice.Client, directoryID string) {
	out, err := svc.ListIpRoutes(context.TODO(), &directoryservice.ListIpRoutesInput{
		DirectoryId: &directoryID,
	})
	if err != nil {
		log.Println(err)
		return
	}
	for _, route := range out.IpRoutesInfo {
		cidr := StringValue(route.CidrIp)
		g.Resources = append(g.Resources, terraformutils.NewResource(
			directoryID+":"+cidr,
			cidr,
			"aws_directory_service_ip_route",
			"aws",
			map[string]string{"directory_id": directoryID},
			directoryServiceAllowEmptyValues,
			map[string]interface{}{}))
	}
}

func (g *DirectoryServiceGenerator) addLogSubscriptions(svc *directoryservice.Client, directoryID string) {
	out, err := svc.ListLogSubscriptions(context.TODO(), &directoryservice.ListLogSubscriptionsInput{
		DirectoryId: &directoryID,
	})
	if err != nil {
		log.Println(err)
		return
	}
	if len(out.LogSubscriptions) == 0 {
		return
	}
	g.Resources = append(g.Resources, terraformutils.NewResource(
		directoryID,
		directoryID,
		"aws_directory_service_log_subscription",
		"aws",
		map[string]string{"directory_id": directoryID},
		directoryServiceAllowEmptyValues,
		map[string]interface{}{}))
}

func (g *DirectoryServiceGenerator) addRadiusSettings(directory dstypes.DirectoryDescription, directoryID string) {
	if directory.RadiusSettings == nil || len(directory.RadiusSettings.RadiusServers) == 0 {
		return
	}
	g.Resources = append(g.Resources, terraformutils.NewResource(
		directoryID,
		directoryID,
		"aws_directory_service_radius_settings",
		"aws",
		map[string]string{"directory_id": directoryID},
		directoryServiceAllowEmptyValues,
		map[string]interface{}{}))
}

func (g *DirectoryServiceGenerator) addRegions(svc *directoryservice.Client, directoryID string) {
	out, err := svc.DescribeRegions(context.TODO(), &directoryservice.DescribeRegionsInput{
		DirectoryId: &directoryID,
	})
	if err != nil {
		log.Println(err)
		return
	}
	for _, region := range out.RegionsDescription {
		regionName := StringValue(region.RegionName)
		g.Resources = append(g.Resources, terraformutils.NewResource(
			directoryID+":"+regionName,
			regionName,
			"aws_directory_service_region",
			"aws",
			map[string]string{"directory_id": directoryID},
			directoryServiceAllowEmptyValues,
			map[string]interface{}{}))
	}
}

func (g *DirectoryServiceGenerator) addSharedDirectories(svc *directoryservice.Client, directoryID string) {
	out, err := svc.DescribeSharedDirectories(context.TODO(), &directoryservice.DescribeSharedDirectoriesInput{
		OwnerDirectoryId: &directoryID,
	})
	if err != nil {
		log.Println(err)
		return
	}
	for _, shared := range out.SharedDirectories {
		sharedID := StringValue(shared.SharedDirectoryId)
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			sharedID,
			sharedID,
			"aws_directory_service_shared_directory",
			"aws",
			directoryServiceAllowEmptyValues))
	}
}

func (g *DirectoryServiceGenerator) addTrusts(svc *directoryservice.Client, directoryID string) {
	out, err := svc.DescribeTrusts(context.TODO(), &directoryservice.DescribeTrustsInput{
		DirectoryId: &directoryID,
	})
	if err != nil {
		log.Println(err)
		return
	}
	for _, trust := range out.Trusts {
		trustID := StringValue(trust.TrustId)
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			trustID,
			StringValue(trust.RemoteDomainName),
			"aws_directory_service_trust",
			"aws",
			directoryServiceAllowEmptyValues))
	}
}

func (g *DirectoryServiceGenerator) addDataUsersAndGroups(svc *directoryservicedata.Client, directoryID string) {
	users, err := svc.ListUsers(context.TODO(), &directoryservicedata.ListUsersInput{
		DirectoryId: &directoryID,
	})
	if err != nil {
		log.Println(err)
	} else {
		for _, user := range users.Users {
			sam := StringValue(user.SAMAccountName)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				directoryID+","+sam,
				sam,
				"aws_directoryservicedata_user",
				"aws",
				map[string]string{"directory_id": directoryID},
				directoryServiceAllowEmptyValues,
				map[string]interface{}{}))
		}
	}
	groups, err := svc.ListGroups(context.TODO(), &directoryservicedata.ListGroupsInput{
		DirectoryId: &directoryID,
	})
	if err != nil {
		log.Println(err)
		return
	}
	for _, group := range groups.Groups {
		sam := StringValue(group.SAMAccountName)
		g.Resources = append(g.Resources, terraformutils.NewResource(
			directoryID+","+sam,
			sam,
			"aws_directoryservicedata_group",
			"aws",
			map[string]string{"directory_id": directoryID},
			directoryServiceAllowEmptyValues,
			map[string]interface{}{}))
	}
}
