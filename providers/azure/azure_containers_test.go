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
	"strings"
	"testing"

	armappcontainers "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers/v4"
	armappplatform "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appplatform/armappplatform/v2"
	armcontainerservice "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/hdinsight/armhdinsight"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/streamanalytics/armstreamanalytics"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/webpubsub/armwebpubsub"
)

func strPtr(s string) *string { return &s }

func TestBotChannelTerraformType(t *testing.T) {
	cases := map[string]string{
		"AlexaChannel":            "azurerm_bot_channel_alexa",
		"DirectLineSpeechChannel": "azurerm_bot_channel_direct_line_speech",
		"DirectLineChannel":       "azurerm_bot_channel_directline",
		"EmailChannel":            "azurerm_bot_channel_email",
		"FacebookChannel":         "azurerm_bot_channel_facebook",
		"LineChannel":             "azurerm_bot_channel_line",
		"MsTeamsChannel":          "azurerm_bot_channel_ms_teams",
		"SlackChannel":            "azurerm_bot_channel_slack",
		"SmsChannel":              "azurerm_bot_channel_sms",
		"WebChatChannel":          "azurerm_bot_channel_web_chat",
		"TelegramChannel":         "",
	}
	for in, want := range cases {
		if got := botChannelTerraformType(in); got != want {
			t.Errorf("botChannelTerraformType(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestHDInsightTerraformType(t *testing.T) {
	mk := func(kind string) *armhdinsight.Cluster {
		return &armhdinsight.Cluster{Properties: &armhdinsight.ClusterGetProperties{
			ClusterDefinition: &armhdinsight.ClusterDefinition{Kind: strPtr(kind)},
		}}
	}
	if got := hdinsightTerraformType(mk("kafka")); got != "azurerm_hdinsight_kafka_cluster" {
		t.Errorf("kafka = %q", got)
	}
	if got := hdinsightTerraformType(mk("InteractiveQuery")); got != "azurerm_hdinsight_interactive_query_cluster" {
		t.Errorf("interactivequery = %q", got)
	}
	if got := hdinsightTerraformType(mk("trino")); got != "" {
		t.Errorf("trino = %q, want empty", got)
	}
}

func TestKubernetesClusterTerraformType(t *testing.T) {
	automatic := armcontainerservice.ManagedClusterSKUName("Automatic")
	cluster := &armcontainerservice.ManagedCluster{SKU: &armcontainerservice.ManagedClusterSKU{Name: &automatic}}
	if got := kubernetesClusterTerraformType(cluster); got != "azurerm_kubernetes_automatic_cluster" {
		t.Errorf("automatic = %q", got)
	}
	if got := kubernetesClusterTerraformType(&armcontainerservice.ManagedCluster{}); got != "azurerm_kubernetes_cluster" {
		t.Errorf("plain = %q", got)
	}
}

func TestWebPubSubTerraformType(t *testing.T) {
	socketIO := armwebpubsub.ServiceKindSocketIO
	if got := webPubSubTerraformType(&armwebpubsub.ResourceInfo{Kind: &socketIO}); got != "azurerm_web_pubsub_socketio" {
		t.Errorf("socketio = %q", got)
	}
	if got := webPubSubTerraformType(&armwebpubsub.ResourceInfo{}); got != "azurerm_web_pubsub" {
		t.Errorf("plain = %q", got)
	}
}

func TestSpringDiscriminators(t *testing.T) {
	apm := &armappplatform.ApmResource{Properties: &armappplatform.ApmProperties{Type: strPtr("NewRelic")}}
	if got := springAPMTerraformType(apm); got != "azurerm_spring_cloud_new_relic_application_performance_monitoring" {
		t.Errorf("apm = %q", got)
	}
	binding := &armappplatform.BindingResource{Properties: &armappplatform.BindingResourceProperties{ResourceType: strPtr("Microsoft.DocumentDB/databaseAccounts")}}
	if got := springBindingTerraformType(binding); got != "azurerm_spring_cloud_app_cosmosdb_association" {
		t.Errorf("binding = %q", got)
	}
	deployment := &armappplatform.DeploymentResource{Properties: &armappplatform.DeploymentResourceProperties{
		Source: &armappplatform.BuildResultUserSourceInfo{Type: strPtr("BuildResult")},
	}}
	if got := springDeploymentTerraformType(deployment); got != "azurerm_spring_cloud_build_deployment" {
		t.Errorf("deployment = %q", got)
	}
}

func TestStreamAnalyticsDiscriminators(t *testing.T) {
	input := &armstreamanalytics.Input{Properties: &armstreamanalytics.StreamInputProperties{
		Datasource: &armstreamanalytics.EventHubV2StreamInputDataSource{},
	}}
	if got := streamInputTerraformType(input); got != "azurerm_stream_analytics_stream_input_eventhub_v2" {
		t.Errorf("stream input = %q", got)
	}
	ref := &armstreamanalytics.Input{Properties: &armstreamanalytics.ReferenceInputProperties{
		Datasource: &armstreamanalytics.BlobReferenceInputDataSource{},
	}}
	if got := referenceInputTerraformType(ref); got != "azurerm_stream_analytics_reference_input_blob" {
		t.Errorf("reference input = %q", got)
	}
	output := &armstreamanalytics.Output{Properties: &armstreamanalytics.OutputProperties{
		Datasource: &armstreamanalytics.AzureSynapseOutputDataSource{},
	}}
	if got := outputTerraformType(output); got != "azurerm_stream_analytics_output_synapse" {
		t.Errorf("output = %q", got)
	}
}

func TestContainerAppCustomDomainResources(t *testing.T) {
	app := &armappcontainers.ContainerApp{
		ID:   strPtr("/subscriptions/s/resourceGroups/rg/providers/Microsoft.App/containerApps/app1"),
		Name: strPtr("app1"),
		Properties: &armappcontainers.ContainerAppProperties{
			Configuration: &armappcontainers.Configuration{
				Ingress: &armappcontainers.Ingress{CustomDomains: []*armappcontainers.CustomDomain{
					{Name: strPtr("example.com")},
				}},
			},
		},
	}
	resources := containerAppCustomDomainResources(app, "azurerm")
	if len(resources) != 1 {
		t.Fatalf("got %d resources, want 1", len(resources))
	}
	if resources[0].InstanceInfo.Type != "azurerm_container_app_custom_domain" {
		t.Errorf("type = %q", resources[0].InstanceInfo.Type)
	}
	if !strings.HasSuffix(resources[0].InstanceState.ID, "/customDomainName/example.com") {
		t.Errorf("id = %q", resources[0].InstanceState.ID)
	}
}

const mockHDInsightJSON = `{"value":[
	{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.HDInsight/clusters/hdi-mock-1","name":"hdi-mock-1","location":"eastus","type":"Microsoft.HDInsight/clusters","properties":{"clusterDefinition":{"kind":"spark"}}}
]}`

func TestMockTrack2HDInsightClusters(t *testing.T) {
	tr := &cannedTransport{body: mockHDInsightJSON}
	client, err := armhdinsight.NewClustersClient(mockSubscriptionID, stubCredential{}, mockClientOptions(tr))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	pager := client.NewListPager(nil)
	ctx := context.Background()
	g := &HDInsightGenerator{}
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			t.Fatalf("next page: %v", err)
		}
		for _, cluster := range page.Value {
			if resourceType := hdinsightTerraformType(cluster); resourceType != "" {
				g.AppendSimpleResource(*cluster.ID, *cluster.Name, resourceType)
			}
		}
	}
	if len(g.Resources) != 1 {
		t.Fatalf("got %d resources, want 1", len(g.Resources))
	}
	if g.Resources[0].InstanceInfo.Type != "azurerm_hdinsight_spark_cluster" {
		t.Errorf("type = %q, want azurerm_hdinsight_spark_cluster", g.Resources[0].InstanceInfo.Type)
	}
	if len(tr.urls) == 0 || !strings.Contains(tr.urls[0], "Microsoft.HDInsight") {
		t.Errorf("expected a Microsoft.HDInsight call, got %v", tr.urls)
	}
}
