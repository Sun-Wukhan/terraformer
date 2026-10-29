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
	"strconv"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/appconfig"
	"github.com/aws/aws-sdk-go-v2/service/appsync"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
)

// Assembly follow-ups: four types whose cluster workers skipped them only
// because provider-doc fetches were rate-limited (HTTP 429) during the
// cluster runs. Their importer ID formats were verified against the
// provider documentation at assembly time:
//
//   - aws_appsync_function:    api_id-function_id ("xxxxx-yyyyy")
//   - aws_appsync_api_key:     api_id:key_id ("xxxxx:yyyyy")
//   - aws_appconfig_deployment: application_id/environment_id/deployment_number
//   - aws_msk_channel:         channel_arn,cluster_arn (MSK Data Delivery;
//     ListChannels exists in the pinned kafka SDK v1.65.7)

// loadAppSyncFunctionsAndKeys enumerates AppSync functions and API keys for
// every GraphQL API in the region.
func (g *AppSyncGenerator) loadAppSyncFunctionsAndKeys(svc *appsync.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListGraphqlApis(context.TODO(), &appsync.ListGraphqlApisInput{
			NextToken: nextToken,
		})
		if err != nil {
			return err
		}
		for _, api := range out.GraphqlApis {
			apiID := StringValue(api.ApiId)
			if err := g.loadAppSyncFunctions(svc, apiID); err != nil {
				return err
			}
			if err := g.loadAppSyncAPIKeys(svc, apiID); err != nil {
				return err
			}
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

func (g *AppSyncGenerator) loadAppSyncFunctions(svc *appsync.Client, apiID string) error {
	var nextToken *string
	for {
		out, err := svc.ListFunctions(context.TODO(), &appsync.ListFunctionsInput{
			ApiId:     &apiID,
			NextToken: nextToken,
		})
		if err != nil {
			return err
		}
		for _, fn := range out.Functions {
			// Import ID is api_id-function_id.
			id := apiID + "-" + StringValue(fn.FunctionId)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				StringValue(fn.Name),
				"aws_appsync_function",
				"aws",
				appSyncAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

func (g *AppSyncGenerator) loadAppSyncAPIKeys(svc *appsync.Client, apiID string) error {
	var nextToken *string
	for {
		out, err := svc.ListApiKeys(context.TODO(), &appsync.ListApiKeysInput{
			ApiId:     &apiID,
			NextToken: nextToken,
		})
		if err != nil {
			return err
		}
		for _, key := range out.ApiKeys {
			// Import ID is api_id:key_id.
			id := apiID + ":" + StringValue(key.Id)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				StringValue(key.Id),
				"aws_appsync_api_key",
				"aws",
				appSyncAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

// loadAppConfigDeployments enumerates AppConfig deployments for every
// application/environment pair in the region.
func (g *AppConfigGenerator) loadAppConfigDeployments(svc *appconfig.Client) error {
	var appToken *string
	for {
		apps, err := svc.ListApplications(context.TODO(), &appconfig.ListApplicationsInput{
			NextToken: appToken,
		})
		if err != nil {
			return err
		}
		for _, app := range apps.Items {
			appID := StringValue(app.Id)
			var envToken *string
			for {
				envs, err := svc.ListEnvironments(context.TODO(), &appconfig.ListEnvironmentsInput{
					ApplicationId: &appID,
					NextToken:     envToken,
				})
				if err != nil {
					return err
				}
				for _, env := range envs.Items {
					envID := StringValue(env.Id)
					if err := g.loadAppConfigEnvironmentDeployments(svc, appID, envID); err != nil {
						return err
					}
				}
				envToken = envs.NextToken
				if envToken == nil {
					break
				}
			}
		}
		appToken = apps.NextToken
		if appToken == nil {
			break
		}
	}
	return nil
}

func (g *AppConfigGenerator) loadAppConfigEnvironmentDeployments(svc *appconfig.Client, appID, envID string) error {
	var nextToken *string
	for {
		out, err := svc.ListDeployments(context.TODO(), &appconfig.ListDeploymentsInput{
			ApplicationId: &appID,
			EnvironmentId: &envID,
			NextToken:     nextToken,
		})
		if err != nil {
			return err
		}
		for _, dep := range out.Items {
			number := strconv.Itoa(int(dep.DeploymentNumber))
			// Import ID is application_id/environment_id/deployment_number.
			id := appID + "/" + envID + "/" + number
			g.Resources = append(g.Resources, terraformutils.NewResource(
				id,
				number,
				"aws_appconfig_deployment",
				"aws",
				map[string]string{"application_id": appID, "environment_id": envID},
				appConfigAllowEmptyValues,
				map[string]interface{}{}))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

// loadChannels enumerates MSK Data Delivery channels for every provisioned
// cluster. Import ID is channel_arn,cluster_arn.
func (g *MskGenerator) loadChannels(svc *kafka.Client, clusterARNs []string) error {
	for _, clusterARN := range clusterARNs {
		var nextToken *string
		for {
			out, err := svc.ListChannels(context.TODO(), &kafka.ListChannelsInput{
				ClusterArn: &clusterARN,
				NextToken:  nextToken,
			})
			if err != nil {
				return err
			}
			for _, ch := range out.Channels {
				channelARN := StringValue(ch.ChannelArn)
				name := StringValue(ch.ChannelName)
				if name == "" {
					name = arnResourceName(channelARN)
				}
				g.Resources = append(g.Resources, terraformutils.NewResource(
					channelARN+","+clusterARN,
					name,
					"aws_msk_channel",
					"aws",
					map[string]string{"cluster_arn": clusterARN},
					mskAllowEmptyValues,
					map[string]interface{}{}))
			}
			nextToken = out.NextToken
			if nextToken == nil {
				break
			}
		}
	}
	return nil
}
