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

package azure

import (
	"context"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/streamanalytics/armstreamanalytics"
)

// StreamAnalyticsGenerator covers Microsoft.StreamAnalytics clusters and
// streaming jobs. Job inputs, outputs and functions are single ARM resource
// kinds whose Terraform type is decided by the data-source / function kind
// reported by the API, mapped by the helpers below.
//
// Skipped (see results-containers.md):
//   - azurerm_stream_analytics_job_schedule: side-effect-only — creating it
//     starts the job via the streamingJobs start action with a synthetic
//     .../schedule/default ID; armstreamanalytics has no Schedule client
//     with a List/Get surface.
type StreamAnalyticsGenerator struct {
	AzureService
}

// streamInputTerraformType maps a stream input's data source to its
// Terraform type ("" when there is no matching type in this sweep).
func streamInputTerraformType(input *armstreamanalytics.Input) string {
	if input == nil || input.Properties == nil {
		return ""
	}
	streamProps, ok := input.Properties.(*armstreamanalytics.StreamInputProperties)
	if !ok || streamProps.Datasource == nil {
		return ""
	}
	switch streamProps.Datasource.(type) {
	case *armstreamanalytics.BlobStreamInputDataSource:
		return "azurerm_stream_analytics_stream_input_blob"
	case *armstreamanalytics.EventHubStreamInputDataSource:
		return "azurerm_stream_analytics_stream_input_eventhub"
	case *armstreamanalytics.EventHubV2StreamInputDataSource:
		return "azurerm_stream_analytics_stream_input_eventhub_v2"
	case *armstreamanalytics.IoTHubStreamInputDataSource:
		return "azurerm_stream_analytics_stream_input_iothub"
	}
	return ""
}

// referenceInputTerraformType maps a reference input's data source to its
// Terraform type ("" when there is no matching type in this sweep).
func referenceInputTerraformType(input *armstreamanalytics.Input) string {
	if input == nil || input.Properties == nil {
		return ""
	}
	refProps, ok := input.Properties.(*armstreamanalytics.ReferenceInputProperties)
	if !ok || refProps.Datasource == nil {
		return ""
	}
	switch refProps.Datasource.(type) {
	case *armstreamanalytics.BlobReferenceInputDataSource:
		return "azurerm_stream_analytics_reference_input_blob"
	case *armstreamanalytics.AzureSQLReferenceInputDataSource:
		return "azurerm_stream_analytics_reference_input_mssql"
	}
	return ""
}

// outputTerraformType maps an output's data source to its Terraform type
// ("" when there is no matching type in this sweep, e.g. Data Lake Store).
func outputTerraformType(output *armstreamanalytics.Output) string {
	if output == nil || output.Properties == nil || output.Properties.Datasource == nil {
		return ""
	}
	switch output.Properties.Datasource.(type) {
	case *armstreamanalytics.BlobOutputDataSource:
		return "azurerm_stream_analytics_output_blob"
	case *armstreamanalytics.DocumentDbOutputDataSource:
		return "azurerm_stream_analytics_output_cosmosdb"
	case *armstreamanalytics.EventHubOutputDataSource, *armstreamanalytics.EventHubV2OutputDataSource:
		return "azurerm_stream_analytics_output_eventhub"
	case *armstreamanalytics.AzureFunctionOutputDataSource:
		return "azurerm_stream_analytics_output_function"
	case *armstreamanalytics.AzureSQLDatabaseOutputDataSource:
		return "azurerm_stream_analytics_output_mssql"
	case *armstreamanalytics.PowerBIOutputDataSource:
		return "azurerm_stream_analytics_output_powerbi"
	case *armstreamanalytics.ServiceBusQueueOutputDataSource:
		return "azurerm_stream_analytics_output_servicebus_queue"
	case *armstreamanalytics.ServiceBusTopicOutputDataSource:
		return "azurerm_stream_analytics_output_servicebus_topic"
	case *armstreamanalytics.AzureSynapseOutputDataSource:
		return "azurerm_stream_analytics_output_synapse"
	case *armstreamanalytics.AzureTableOutputDataSource:
		return "azurerm_stream_analytics_output_table"
	}
	return ""
}

// functionTerraformType maps a function to its Terraform type: scalar
// JavaScript functions are UDFs, aggregate JavaScript functions are UDAs.
// Non-JavaScript (e.g. Azure ML) functions have no type in this sweep.
func functionTerraformType(fn *armstreamanalytics.Function) string {
	if fn == nil || fn.Properties == nil {
		return ""
	}
	props := fn.Properties.GetFunctionProperties()
	if props == nil || props.Properties == nil {
		return ""
	}
	if _, isJS := props.Properties.Binding.(*armstreamanalytics.JavaScriptFunctionBinding); !isJS {
		return ""
	}
	switch fn.Properties.(type) {
	case *armstreamanalytics.ScalarFunctionProperties:
		return "azurerm_stream_analytics_function_javascript_udf"
	case *armstreamanalytics.AggregateFunctionProperties:
		return "azurerm_stream_analytics_function_javascript_uda"
	}
	return ""
}

func (g *StreamAnalyticsGenerator) appendJobChildren(ctx context.Context, rg string, job *armstreamanalytics.StreamingJob) error {
	jobName := *job.Name

	// The job storage account is a configuration block on the job itself;
	// its Terraform ID is the job ID and it only exists when configured.
	if job.Properties != nil && job.Properties.JobStorageAccount != nil {
		g.AppendSimpleResource(*job.ID, jobName, "azurerm_stream_analytics_job_storage_account")
	}

	inputsClient, err := armstreamanalytics.NewInputsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	inputPager := inputsClient.NewListByStreamingJobPager(rg, jobName, nil)
	for inputPager.More() {
		page, err := inputPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, input := range page.Value {
			if resourceType := streamInputTerraformType(input); resourceType != "" {
				g.AppendSimpleResource(*input.ID, *input.Name, resourceType)
			} else if resourceType := referenceInputTerraformType(input); resourceType != "" {
				g.AppendSimpleResource(*input.ID, *input.Name, resourceType)
			}
		}
	}

	outputsClient, err := armstreamanalytics.NewOutputsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	outputPager := outputsClient.NewListByStreamingJobPager(rg, jobName, nil)
	for outputPager.More() {
		page, err := outputPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, output := range page.Value {
			if resourceType := outputTerraformType(output); resourceType != "" {
				g.AppendSimpleResource(*output.ID, *output.Name, resourceType)
			}
		}
	}

	functionsClient, err := armstreamanalytics.NewFunctionsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	functionPager := functionsClient.NewListByStreamingJobPager(rg, jobName, nil)
	for functionPager.More() {
		page, err := functionPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, fn := range page.Value {
			if resourceType := functionTerraformType(fn); resourceType != "" {
				g.AppendSimpleResource(*fn.ID, *fn.Name, resourceType)
			}
		}
	}
	return nil
}

func (g *StreamAnalyticsGenerator) InitResources() error {
	ctx := context.Background()

	clustersClient, err := armstreamanalytics.NewClustersClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	var clusters []*armstreamanalytics.Cluster
	if rg := g.getResourceGroup(); rg != "" {
		pager := clustersClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			clusters = append(clusters, page.Value...)
		}
	} else {
		pager := clustersClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			clusters = append(clusters, page.Value...)
		}
	}

	privateEndpointsClient, err := armstreamanalytics.NewPrivateEndpointsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	for _, cluster := range clusters {
		g.AppendSimpleResource(*cluster.ID, *cluster.Name, "azurerm_stream_analytics_cluster")
		parsed, err := ParseAzureResourceID(*cluster.ID)
		if err != nil {
			return err
		}
		pePager := privateEndpointsClient.NewListByClusterPager(parsed.ResourceGroup, *cluster.Name, nil)
		for pePager.More() {
			page, err := pePager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_stream_analytics_managed_private_endpoint")
			}
		}
	}

	jobsClient, err := armstreamanalytics.NewStreamingJobsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	var jobs []*armstreamanalytics.StreamingJob
	if rg := g.getResourceGroup(); rg != "" {
		pager := jobsClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			jobs = append(jobs, page.Value...)
		}
	} else {
		pager := jobsClient.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			jobs = append(jobs, page.Value...)
		}
	}
	for _, job := range jobs {
		g.AppendSimpleResource(*job.ID, *job.Name, "azurerm_stream_analytics_job")
		parsed, err := ParseAzureResourceID(*job.ID)
		if err != nil {
			return err
		}
		if err := g.appendJobChildren(ctx, parsed.ResourceGroup, job); err != nil {
			return err
		}
	}
	return nil
}
