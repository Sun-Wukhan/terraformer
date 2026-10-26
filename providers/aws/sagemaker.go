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
	"errors"
	"strconv"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	sagemakertypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
)

var sagemakerAllowEmptyValues = []string{"tags."}

type SageMakerGenerator struct {
	AWSService
}

// InitResources enumerates the Amazon SageMaker AI control plane: Studio
// domains with their user profiles, spaces and apps, the model training and
// hosting resources (models, endpoints, endpoint configurations, training
// and tuning jobs, pipelines, feature groups, inference components), the
// Studio and Edge Manager registries (images and image versions, app image
// and lifecycle configs, device fleets and devices), hubs and the content
// references copied into them, MLflow apps and tracking servers, Ground
// Truth workforces and workteams, model registry groups and their resource
// policies, and the per-region Service Catalog portfolio status.
func (g *SageMakerGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	client := sagemaker.NewFromConfig(config)

	if err := g.addDomains(client); err != nil {
		return err
	}
	for _, add := range []func(*sagemaker.Client) error{
		g.addAlgorithms,
		g.addAppImageConfigs,
		g.addCodeRepositories,
		g.addDataQualityJobDefinitions,
		g.addDeviceFleets,
		g.addEndpoints,
		g.addEndpointConfigs,
		g.addFeatureGroups,
		g.addFlowDefinitions,
		g.addHubs,
		g.addHumanTaskUis,
		g.addHyperParameterTuningJobs,
		g.addImages,
		g.addInferenceComponents,
		g.addLabelingJobs,
		g.addMlflowApps,
		g.addMlflowTrackingServers,
		g.addModels,
		g.addModelCards,
		g.addModelCardExportJobs,
		g.addModelPackageGroups,
		g.addMonitoringSchedules,
		g.addNotebookInstances,
		g.addNotebookInstanceLifecycleConfigs,
		g.addPipelines,
		g.addProjects,
		g.addStudioLifecycleConfigs,
		g.addTrainingJobs,
		g.addWorkforces,
		g.addWorkteams,
	} {
		if err := add(client); err != nil {
			return err
		}
	}
	return g.addServiceCatalogPortfolioStatus(client, config.Region)
}

// Domains are keyed by domain ID in the provider.
func (g *SageMakerGenerator) addDomains(client *sagemaker.Client) error {
	p := sagemaker.NewListDomainsPaginator(client, &sagemaker.ListDomainsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, domain := range page.Domains {
			domainID := *domain.DomainId
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				domainID,
				*domain.DomainName,
				"aws_sagemaker_domain",
				"aws",
				sagemakerAllowEmptyValues,
			))
			if err := g.addUserProfiles(client, domainID); err != nil {
				return err
			}
			if err := g.addSpaces(client, domainID); err != nil {
				return err
			}
			if err := g.addApps(client, domainID); err != nil {
				return err
			}
		}
	}
	return nil
}

// User profiles are keyed by ARN in the provider; the list API does not
// return ARNs, so each profile is described to obtain it.
func (g *SageMakerGenerator) addUserProfiles(client *sagemaker.Client, domainID string) error {
	p := sagemaker.NewListUserProfilesPaginator(client, &sagemaker.ListUserProfilesInput{
		DomainIdEquals: aws.String(domainID),
	})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, profile := range page.UserProfiles {
			detail, err := client.DescribeUserProfile(context.TODO(), &sagemaker.DescribeUserProfileInput{
				DomainId:        aws.String(domainID),
				UserProfileName: profile.UserProfileName,
			})
			if err != nil {
				return err
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*detail.UserProfileArn,
				*profile.UserProfileName,
				"aws_sagemaker_user_profile",
				"aws",
				sagemakerAllowEmptyValues,
			))
		}
	}
	return nil
}

// Spaces are keyed by ARN in the provider; the list API does not return
// ARNs, so each space is described to obtain it.
func (g *SageMakerGenerator) addSpaces(client *sagemaker.Client, domainID string) error {
	p := sagemaker.NewListSpacesPaginator(client, &sagemaker.ListSpacesInput{
		DomainIdEquals: aws.String(domainID),
	})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, space := range page.Spaces {
			detail, err := client.DescribeSpace(context.TODO(), &sagemaker.DescribeSpaceInput{
				DomainId:  aws.String(domainID),
				SpaceName: space.SpaceName,
			})
			if err != nil {
				return err
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*detail.SpaceArn,
				*space.SpaceName,
				"aws_sagemaker_space",
				"aws",
				sagemakerAllowEmptyValues,
			))
		}
	}
	return nil
}

// Apps are keyed by ARN in the provider; the list API does not return ARNs,
// so each app is described to obtain it.
func (g *SageMakerGenerator) addApps(client *sagemaker.Client, domainID string) error {
	p := sagemaker.NewListAppsPaginator(client, &sagemaker.ListAppsInput{
		DomainIdEquals: aws.String(domainID),
	})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, app := range page.Apps {
			detail, err := client.DescribeApp(context.TODO(), &sagemaker.DescribeAppInput{
				DomainId:        aws.String(domainID),
				AppName:         app.AppName,
				AppType:         app.AppType,
				SpaceName:       app.SpaceName,
				UserProfileName: app.UserProfileName,
			})
			if err != nil {
				return err
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*detail.AppArn,
				*app.AppName,
				"aws_sagemaker_app",
				"aws",
				sagemakerAllowEmptyValues,
			))
		}
	}
	return nil
}

func (g *SageMakerGenerator) addAlgorithms(client *sagemaker.Client) error {
	p := sagemaker.NewListAlgorithmsPaginator(client, &sagemaker.ListAlgorithmsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, algorithm := range page.AlgorithmSummaryList {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*algorithm.AlgorithmName,
				*algorithm.AlgorithmName,
				"aws_sagemaker_algorithm",
				"aws",
				sagemakerAllowEmptyValues,
			))
		}
	}
	return nil
}

func (g *SageMakerGenerator) addAppImageConfigs(client *sagemaker.Client) error {
	p := sagemaker.NewListAppImageConfigsPaginator(client, &sagemaker.ListAppImageConfigsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, config := range page.AppImageConfigs {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*config.AppImageConfigName,
				*config.AppImageConfigName,
				"aws_sagemaker_app_image_config",
				"aws",
				sagemakerAllowEmptyValues,
			))
		}
	}
	return nil
}

func (g *SageMakerGenerator) addCodeRepositories(client *sagemaker.Client) error {
	p := sagemaker.NewListCodeRepositoriesPaginator(client, &sagemaker.ListCodeRepositoriesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, repo := range page.CodeRepositorySummaryList {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*repo.CodeRepositoryName,
				*repo.CodeRepositoryName,
				"aws_sagemaker_code_repository",
				"aws",
				sagemakerAllowEmptyValues,
			))
		}
	}
	return nil
}

func (g *SageMakerGenerator) addDataQualityJobDefinitions(client *sagemaker.Client) error {
	p := sagemaker.NewListDataQualityJobDefinitionsPaginator(client, &sagemaker.ListDataQualityJobDefinitionsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, definition := range page.JobDefinitionSummaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*definition.MonitoringJobDefinitionName,
				*definition.MonitoringJobDefinitionName,
				"aws_sagemaker_data_quality_job_definition",
				"aws",
				sagemakerAllowEmptyValues,
			))
		}
	}
	return nil
}

func (g *SageMakerGenerator) addDeviceFleets(client *sagemaker.Client) error {
	p := sagemaker.NewListDeviceFleetsPaginator(client, &sagemaker.ListDeviceFleetsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, fleet := range page.DeviceFleetSummaries {
			fleetName := *fleet.DeviceFleetName
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				fleetName,
				fleetName,
				"aws_sagemaker_device_fleet",
				"aws",
				sagemakerAllowEmptyValues,
			))
			if err := g.addDevices(client, fleetName); err != nil {
				return err
			}
		}
	}
	return nil
}

// Provider ID format: device fleet name, device name.
func (g *SageMakerGenerator) addDevices(client *sagemaker.Client, fleetName string) error {
	p := sagemaker.NewListDevicesPaginator(client, &sagemaker.ListDevicesInput{
		DeviceFleetName: aws.String(fleetName),
	})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, device := range page.DeviceSummaries {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				fleetName+"/"+*device.DeviceName,
				*device.DeviceName,
				"aws_sagemaker_device",
				"aws",
				map[string]string{"device_fleet_name": fleetName},
				sagemakerAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

func (g *SageMakerGenerator) addEndpoints(client *sagemaker.Client) error {
	p := sagemaker.NewListEndpointsPaginator(client, &sagemaker.ListEndpointsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, endpoint := range page.Endpoints {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*endpoint.EndpointName,
				*endpoint.EndpointName,
				"aws_sagemaker_endpoint",
				"aws",
				sagemakerAllowEmptyValues,
			))
		}
	}
	return nil
}

func (g *SageMakerGenerator) addEndpointConfigs(client *sagemaker.Client) error {
	p := sagemaker.NewListEndpointConfigsPaginator(client, &sagemaker.ListEndpointConfigsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, config := range page.EndpointConfigs {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*config.EndpointConfigName,
				*config.EndpointConfigName,
				"aws_sagemaker_endpoint_configuration",
				"aws",
				sagemakerAllowEmptyValues,
			))
		}
	}
	return nil
}

func (g *SageMakerGenerator) addFeatureGroups(client *sagemaker.Client) error {
	p := sagemaker.NewListFeatureGroupsPaginator(client, &sagemaker.ListFeatureGroupsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, group := range page.FeatureGroupSummaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*group.FeatureGroupName,
				*group.FeatureGroupName,
				"aws_sagemaker_feature_group",
				"aws",
				sagemakerAllowEmptyValues,
			))
		}
	}
	return nil
}

func (g *SageMakerGenerator) addFlowDefinitions(client *sagemaker.Client) error {
	p := sagemaker.NewListFlowDefinitionsPaginator(client, &sagemaker.ListFlowDefinitionsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, definition := range page.FlowDefinitionSummaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*definition.FlowDefinitionName,
				*definition.FlowDefinitionName,
				"aws_sagemaker_flow_definition",
				"aws",
				sagemakerAllowEmptyValues,
			))
		}
	}
	return nil
}

func (g *SageMakerGenerator) addHubs(client *sagemaker.Client) error {
	input := &sagemaker.ListHubsInput{}
	for {
		page, err := client.ListHubs(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, hub := range page.HubSummaries {
			hubName := *hub.HubName
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				hubName,
				hubName,
				"aws_sagemaker_hub",
				"aws",
				sagemakerAllowEmptyValues,
			))
			if err := g.addHubContentReferences(client, hubName); err != nil {
				return err
			}
		}
		if page.NextToken == nil {
			return nil
		}
		input.NextToken = page.NextToken
	}
}

// Provider ID format: hub name, hub content name. Only hub contents of type
// ModelReference are managed by this resource; native hub contents are not.
func (g *SageMakerGenerator) addHubContentReferences(client *sagemaker.Client, hubName string) error {
	input := &sagemaker.ListHubContentsInput{
		HubName:        aws.String(hubName),
		HubContentType: sagemakertypes.HubContentTypeModelReference,
	}
	for {
		page, err := client.ListHubContents(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, content := range page.HubContentSummaries {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				hubName+","+*content.HubContentName,
				*content.HubContentName,
				"aws_sagemaker_hub_content_reference",
				"aws",
				map[string]string{"hub_name": hubName},
				sagemakerAllowEmptyValues,
				map[string]interface{}{},
			))
		}
		if page.NextToken == nil {
			return nil
		}
		input.NextToken = page.NextToken
	}
}

func (g *SageMakerGenerator) addHumanTaskUis(client *sagemaker.Client) error {
	p := sagemaker.NewListHumanTaskUisPaginator(client, &sagemaker.ListHumanTaskUisInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, ui := range page.HumanTaskUiSummaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*ui.HumanTaskUiName,
				*ui.HumanTaskUiName,
				"aws_sagemaker_human_task_ui",
				"aws",
				sagemakerAllowEmptyValues,
			))
		}
	}
	return nil
}

func (g *SageMakerGenerator) addHyperParameterTuningJobs(client *sagemaker.Client) error {
	p := sagemaker.NewListHyperParameterTuningJobsPaginator(client, &sagemaker.ListHyperParameterTuningJobsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, job := range page.HyperParameterTuningJobSummaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*job.HyperParameterTuningJobName,
				*job.HyperParameterTuningJobName,
				"aws_sagemaker_hyper_parameter_tuning_job",
				"aws",
				sagemakerAllowEmptyValues,
			))
		}
	}
	return nil
}

func (g *SageMakerGenerator) addImages(client *sagemaker.Client) error {
	p := sagemaker.NewListImagesPaginator(client, &sagemaker.ListImagesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, image := range page.Images {
			imageName := *image.ImageName
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				imageName,
				imageName,
				"aws_sagemaker_image",
				"aws",
				sagemakerAllowEmptyValues,
			))
			if err := g.addImageVersions(client, imageName); err != nil {
				return err
			}
		}
	}
	return nil
}

// Provider ID format: image name, version.
func (g *SageMakerGenerator) addImageVersions(client *sagemaker.Client, imageName string) error {
	p := sagemaker.NewListImageVersionsPaginator(client, &sagemaker.ListImageVersionsInput{
		ImageName: aws.String(imageName),
	})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, version := range page.ImageVersions {
			versionNumber := strconv.FormatInt(int64(*version.Version), 10)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				imageName+","+versionNumber,
				imageName+"_"+versionNumber,
				"aws_sagemaker_image_version",
				"aws",
				map[string]string{"image_name": imageName},
				sagemakerAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

func (g *SageMakerGenerator) addInferenceComponents(client *sagemaker.Client) error {
	p := sagemaker.NewListInferenceComponentsPaginator(client, &sagemaker.ListInferenceComponentsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, component := range page.InferenceComponents {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*component.InferenceComponentName,
				*component.InferenceComponentName,
				"aws_sagemaker_inference_component",
				"aws",
				sagemakerAllowEmptyValues,
			))
		}
	}
	return nil
}

func (g *SageMakerGenerator) addLabelingJobs(client *sagemaker.Client) error {
	p := sagemaker.NewListLabelingJobsPaginator(client, &sagemaker.ListLabelingJobsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, job := range page.LabelingJobSummaryList {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*job.LabelingJobName,
				*job.LabelingJobName,
				"aws_sagemaker_labeling_job",
				"aws",
				sagemakerAllowEmptyValues,
			))
		}
	}
	return nil
}

// MLflow apps are keyed by ARN in the provider.
func (g *SageMakerGenerator) addMlflowApps(client *sagemaker.Client) error {
	p := sagemaker.NewListMlflowAppsPaginator(client, &sagemaker.ListMlflowAppsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, app := range page.Summaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*app.Arn,
				*app.Name,
				"aws_sagemaker_mlflow_app",
				"aws",
				sagemakerAllowEmptyValues,
			))
		}
	}
	return nil
}

func (g *SageMakerGenerator) addMlflowTrackingServers(client *sagemaker.Client) error {
	p := sagemaker.NewListMlflowTrackingServersPaginator(client, &sagemaker.ListMlflowTrackingServersInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, server := range page.TrackingServerSummaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*server.TrackingServerName,
				*server.TrackingServerName,
				"aws_sagemaker_mlflow_tracking_server",
				"aws",
				sagemakerAllowEmptyValues,
			))
		}
	}
	return nil
}

func (g *SageMakerGenerator) addModels(client *sagemaker.Client) error {
	p := sagemaker.NewListModelsPaginator(client, &sagemaker.ListModelsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, model := range page.Models {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*model.ModelName,
				*model.ModelName,
				"aws_sagemaker_model",
				"aws",
				sagemakerAllowEmptyValues,
			))
		}
	}
	return nil
}

func (g *SageMakerGenerator) addModelCards(client *sagemaker.Client) error {
	p := sagemaker.NewListModelCardsPaginator(client, &sagemaker.ListModelCardsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, card := range page.ModelCardSummaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*card.ModelCardName,
				*card.ModelCardName,
				"aws_sagemaker_model_card",
				"aws",
				sagemakerAllowEmptyValues,
			))
		}
	}
	return nil
}

func (g *SageMakerGenerator) addModelCardExportJobs(client *sagemaker.Client) error {
	p := sagemaker.NewListModelCardExportJobsPaginator(client, &sagemaker.ListModelCardExportJobsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, job := range page.ModelCardExportJobSummaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*job.ModelCardExportJobName,
				*job.ModelCardExportJobName,
				"aws_sagemaker_model_card_export_job",
				"aws",
				sagemakerAllowEmptyValues,
			))
		}
	}
	return nil
}

func (g *SageMakerGenerator) addModelPackageGroups(client *sagemaker.Client) error {
	p := sagemaker.NewListModelPackageGroupsPaginator(client, &sagemaker.ListModelPackageGroupsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, group := range page.ModelPackageGroupSummaryList {
			groupName := *group.ModelPackageGroupName
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				groupName,
				groupName,
				"aws_sagemaker_model_package_group",
				"aws",
				sagemakerAllowEmptyValues,
			))
			if err := g.addModelPackageGroupPolicy(client, groupName); err != nil {
				return err
			}
		}
	}
	return nil
}

// A model package group only has a resource policy if one was explicitly
// attached; the provider keys the policy resource by group name. Groups
// without a policy report ResourceNotFound and are skipped.
func (g *SageMakerGenerator) addModelPackageGroupPolicy(client *sagemaker.Client, groupName string) error {
	out, err := client.GetModelPackageGroupPolicy(context.TODO(), &sagemaker.GetModelPackageGroupPolicyInput{
		ModelPackageGroupName: aws.String(groupName),
	})
	if err != nil {
		var notFound *sagemakertypes.ResourceNotFound
		if errors.As(err, &notFound) {
			return nil
		}
		return err
	}
	if out.ResourcePolicy == nil || *out.ResourcePolicy == "" {
		return nil
	}
	g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
		groupName,
		groupName,
		"aws_sagemaker_model_package_group_policy",
		"aws",
		sagemakerAllowEmptyValues,
	))
	return nil
}

func (g *SageMakerGenerator) addMonitoringSchedules(client *sagemaker.Client) error {
	p := sagemaker.NewListMonitoringSchedulesPaginator(client, &sagemaker.ListMonitoringSchedulesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, schedule := range page.MonitoringScheduleSummaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*schedule.MonitoringScheduleName,
				*schedule.MonitoringScheduleName,
				"aws_sagemaker_monitoring_schedule",
				"aws",
				sagemakerAllowEmptyValues,
			))
		}
	}
	return nil
}

func (g *SageMakerGenerator) addNotebookInstances(client *sagemaker.Client) error {
	p := sagemaker.NewListNotebookInstancesPaginator(client, &sagemaker.ListNotebookInstancesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, instance := range page.NotebookInstances {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*instance.NotebookInstanceName,
				*instance.NotebookInstanceName,
				"aws_sagemaker_notebook_instance",
				"aws",
				sagemakerAllowEmptyValues,
			))
		}
	}
	return nil
}

func (g *SageMakerGenerator) addNotebookInstanceLifecycleConfigs(client *sagemaker.Client) error {
	p := sagemaker.NewListNotebookInstanceLifecycleConfigsPaginator(client, &sagemaker.ListNotebookInstanceLifecycleConfigsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, config := range page.NotebookInstanceLifecycleConfigs {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*config.NotebookInstanceLifecycleConfigName,
				*config.NotebookInstanceLifecycleConfigName,
				"aws_sagemaker_notebook_instance_lifecycle_configuration",
				"aws",
				sagemakerAllowEmptyValues,
			))
		}
	}
	return nil
}

func (g *SageMakerGenerator) addPipelines(client *sagemaker.Client) error {
	p := sagemaker.NewListPipelinesPaginator(client, &sagemaker.ListPipelinesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, pipeline := range page.PipelineSummaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*pipeline.PipelineName,
				*pipeline.PipelineName,
				"aws_sagemaker_pipeline",
				"aws",
				sagemakerAllowEmptyValues,
			))
		}
	}
	return nil
}

func (g *SageMakerGenerator) addProjects(client *sagemaker.Client) error {
	p := sagemaker.NewListProjectsPaginator(client, &sagemaker.ListProjectsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, project := range page.ProjectSummaryList {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*project.ProjectName,
				*project.ProjectName,
				"aws_sagemaker_project",
				"aws",
				sagemakerAllowEmptyValues,
			))
		}
	}
	return nil
}

func (g *SageMakerGenerator) addStudioLifecycleConfigs(client *sagemaker.Client) error {
	p := sagemaker.NewListStudioLifecycleConfigsPaginator(client, &sagemaker.ListStudioLifecycleConfigsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, config := range page.StudioLifecycleConfigs {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*config.StudioLifecycleConfigName,
				*config.StudioLifecycleConfigName,
				"aws_sagemaker_studio_lifecycle_config",
				"aws",
				sagemakerAllowEmptyValues,
			))
		}
	}
	return nil
}

func (g *SageMakerGenerator) addTrainingJobs(client *sagemaker.Client) error {
	p := sagemaker.NewListTrainingJobsPaginator(client, &sagemaker.ListTrainingJobsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, job := range page.TrainingJobSummaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*job.TrainingJobName,
				*job.TrainingJobName,
				"aws_sagemaker_training_job",
				"aws",
				sagemakerAllowEmptyValues,
			))
		}
	}
	return nil
}

func (g *SageMakerGenerator) addWorkforces(client *sagemaker.Client) error {
	p := sagemaker.NewListWorkforcesPaginator(client, &sagemaker.ListWorkforcesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, workforce := range page.Workforces {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*workforce.WorkforceName,
				*workforce.WorkforceName,
				"aws_sagemaker_workforce",
				"aws",
				sagemakerAllowEmptyValues,
			))
		}
	}
	return nil
}

func (g *SageMakerGenerator) addWorkteams(client *sagemaker.Client) error {
	p := sagemaker.NewListWorkteamsPaginator(client, &sagemaker.ListWorkteamsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, workteam := range page.Workteams {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*workteam.WorkteamName,
				*workteam.WorkteamName,
				"aws_sagemaker_workteam",
				"aws",
				sagemakerAllowEmptyValues,
			))
		}
	}
	return nil
}

// The Service Catalog portfolio status is a per-region singleton; the
// provider keys it by region name.
func (g *SageMakerGenerator) addServiceCatalogPortfolioStatus(client *sagemaker.Client, region string) error {
	if _, err := client.GetSagemakerServicecatalogPortfolioStatus(context.TODO(), &sagemaker.GetSagemakerServicecatalogPortfolioStatusInput{}); err != nil {
		return err
	}
	g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
		region,
		"sagemaker_servicecatalog_portfolio_status",
		"aws_sagemaker_servicecatalog_portfolio_status",
		"aws",
		sagemakerAllowEmptyValues,
	))
	return nil
}
