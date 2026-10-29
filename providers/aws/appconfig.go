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
)

var appConfigAllowEmptyValues = []string{"tags."}

// AppConfigGenerator covers AWS AppConfig. Import IDs follow the provider's
// documented formats: resource IDs for applications, deployment strategies,
// extensions and extension associations; application/profile and
// application/environment composites; app/profile/version for hosted
// configuration versions (verified against the provider docs).
//
// aws_appconfig_deployment is generated via loadAppConfigDeployments
// (assembly_followup.go); its importer ID format
// (application_id/environment_id/deployment_number) was verified against
// the provider docs at assembly time after the cluster run's doc fetches
// were rate-limited.
type AppConfigGenerator struct {
	AWSService
}

func (g *AppConfigGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := appconfig.NewFromConfig(config)

	if err := g.loadAppConfigApplications(svc); err != nil {
		return err
	}
	if err := g.loadAppConfigDeploymentStrategies(svc); err != nil {
		return err
	}
	if err := g.loadAppConfigExtensions(svc); err != nil {
		return err
	}
	return g.loadAppConfigDeployments(svc)
}

func (g *AppConfigGenerator) loadAppConfigApplications(svc *appconfig.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListApplications(context.TODO(), &appconfig.ListApplicationsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, app := range out.Items {
			appID := StringValue(app.Id)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				appID,
				StringValue(app.Name),
				"aws_appconfig_application",
				"aws",
				appConfigAllowEmptyValues))
			if err := g.loadAppConfigProfiles(svc, appID); err != nil {
				return err
			}
			if err := g.loadAppConfigEnvironments(svc, appID); err != nil {
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

func (g *AppConfigGenerator) loadAppConfigProfiles(svc *appconfig.Client, appID string) error {
	var nextToken *string
	for {
		out, err := svc.ListConfigurationProfiles(context.TODO(), &appconfig.ListConfigurationProfilesInput{
			ApplicationId: &appID,
			NextToken:     nextToken,
		})
		if err != nil {
			return err
		}
		for _, profile := range out.Items {
			profileID := StringValue(profile.Id)
			id := appID + "/" + profileID
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				StringValue(profile.Name),
				"aws_appconfig_configuration_profile",
				"aws",
				appConfigAllowEmptyValues))
			if err := g.loadHostedConfigurationVersions(svc, appID, profileID); err != nil {
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

func (g *AppConfigGenerator) loadHostedConfigurationVersions(svc *appconfig.Client, appID, profileID string) error {
	var nextToken *string
	for {
		out, err := svc.ListHostedConfigurationVersions(context.TODO(), &appconfig.ListHostedConfigurationVersionsInput{
			ApplicationId:          &appID,
			ConfigurationProfileId: &profileID,
			NextToken:              nextToken,
		})
		if err != nil {
			return err
		}
		for _, version := range out.Items {
			// Import ID is application_id/configuration_profile_id/version_number.
			id := appID + "/" + profileID + "/" + strconv.Itoa(int(version.VersionNumber))
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				id,
				"aws_appconfig_hosted_configuration_version",
				"aws",
				appConfigAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

func (g *AppConfigGenerator) loadAppConfigEnvironments(svc *appconfig.Client, appID string) error {
	var nextToken *string
	for {
		out, err := svc.ListEnvironments(context.TODO(), &appconfig.ListEnvironmentsInput{
			ApplicationId: &appID,
			NextToken:     nextToken,
		})
		if err != nil {
			return err
		}
		for _, env := range out.Items {
			id := appID + "/" + StringValue(env.Id)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				StringValue(env.Name),
				"aws_appconfig_environment",
				"aws",
				appConfigAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

func (g *AppConfigGenerator) loadAppConfigDeploymentStrategies(svc *appconfig.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListDeploymentStrategies(context.TODO(), &appconfig.ListDeploymentStrategiesInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, strategy := range out.Items {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(strategy.Id),
				StringValue(strategy.Name),
				"aws_appconfig_deployment_strategy",
				"aws",
				appConfigAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

func (g *AppConfigGenerator) loadAppConfigExtensions(svc *appconfig.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListExtensions(context.TODO(), &appconfig.ListExtensionsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, extension := range out.Items {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(extension.Id),
				StringValue(extension.Name),
				"aws_appconfig_extension",
				"aws",
				appConfigAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}

	nextToken = nil
	for {
		out, err := svc.ListExtensionAssociations(context.TODO(), &appconfig.ListExtensionAssociationsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, assoc := range out.Items {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(assoc.Id),
				StringValue(assoc.Id),
				"aws_appconfig_extension_association",
				"aws",
				appConfigAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
