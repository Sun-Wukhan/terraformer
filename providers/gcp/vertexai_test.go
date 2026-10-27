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
	"testing"

	"google.golang.org/api/aiplatform/v1"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

const vertexParent = "projects/test-project/locations/us-central1"

func vertexTestTransport() *cannedTransport {
	return &cannedTransport{routes: []cannedRoute{
		{substr: ":getCacheConfig", body: `{"name":"projects/test-project/cacheConfig"}`},
		{substr: ":getIamPolicy", body: `{"bindings":[{"role":"roles/aiplatform.admin","members":["user:jane@example.com"]}]}`},
		{substr: "/datasets", body: `{"datasets":[{"name":"` + vertexParent + `/datasets/123","displayName":"training-data"}]}`},
		{substr: "/endpoints", body: `{"endpoints":[{"name":"` + vertexParent + `/endpoints/456","displayName":"predictor"}]}`},
		{substr: "/indexes", body: `{"indexes":[{"name":"` + vertexParent + `/indexes/789","displayName":"vector-index"}]}`},
		{substr: "/indexEndpoints", body: `{"indexEndpoints":[{"name":"` + vertexParent + `/indexEndpoints/321","displayName":"vector-endpoint","deployedIndexes":[{"id":"deployed_1","index":"` + vertexParent + `/indexes/789","displayName":"deployed"}]}]}`},
		{substr: "/metadataStores", body: `{"metadataStores":[{"name":"` + vertexParent + `/metadataStores/default"}]}`},
		{substr: "/entityTypes/et1/features", body: `{"features":[{"name":"` + vertexParent + `/featurestores/fs1/entityTypes/et1/features/f1"}]}`},
		{substr: "/featurestores/fs1/entityTypes", body: `{"entityTypes":[{"name":"` + vertexParent + `/featurestores/fs1/entityTypes/et1"}]}`},
		{substr: "/featurestores", body: `{"featurestores":[{"name":"` + vertexParent + `/featurestores/fs1"}]}`},
		{substr: "/featureGroups/fg1/features", body: `{"features":[{"name":"` + vertexParent + `/featureGroups/fg1/features/f1"}]}`},
		{substr: "/featureGroups", body: `{"featureGroups":[{"name":"` + vertexParent + `/featureGroups/fg1"}]}`},
		{substr: "/featureOnlineStores/fos1/featureViews", body: `{"featureViews":[{"name":"` + vertexParent + `/featureOnlineStores/fos1/featureViews/fv1"}]}`},
		{substr: "/featureOnlineStores", body: `{"featureOnlineStores":[{"name":"` + vertexParent + `/featureOnlineStores/fos1"}]}`},
		{substr: "/tensorboards/tb1/experiments/exp1/runs", body: `{"tensorboardRuns":[{"name":"` + vertexParent + `/tensorboards/tb1/experiments/exp1/runs/run1","displayName":"run-one"}]}`},
		{substr: "/tensorboards/tb1/experiments", body: `{"tensorboardExperiments":[{"name":"` + vertexParent + `/tensorboards/tb1/experiments/exp1","displayName":"exp-one"}]}`},
		{substr: "/tensorboards", body: `{"tensorboards":[{"name":"` + vertexParent + `/tensorboards/tb1","displayName":"board"}]}`},
		{substr: "/customJobs", body: `{"customJobs":[{"name":"` + vertexParent + `/customJobs/111","displayName":"train"}]}`},
		{substr: "/pipelineJobs", body: `{"pipelineJobs":[{"name":"` + vertexParent + `/pipelineJobs/222","displayName":"pipe"}]}`},
		{substr: "/deploymentResourcePools", body: `{"deploymentResourcePools":[{"name":"` + vertexParent + `/deploymentResourcePools/pool1"}]}`},
		{substr: "/persistentResources", body: `{"persistentResources":[{"name":"` + vertexParent + `/persistentResources/pr1","displayName":"cluster"}]}`},
		{substr: "/ragCorpora/corpus1/ragFiles", body: `{"ragFiles":[{"name":"` + vertexParent + `/ragCorpora/corpus1/ragFiles/file1","displayName":"doc"}]}`},
		{substr: "/ragCorpora", body: `{"ragCorpora":[{"name":"` + vertexParent + `/ragCorpora/corpus1","displayName":"corpus"}]}`},
		{substr: "/reasoningEngines", body: `{"reasoningEngines":[{"name":"` + vertexParent + `/reasoningEngines/999","displayName":"agent"}]}`},
	}}
}

func TestVertexAiGeneratorEnumeration(t *testing.T) {
	ctx := context.Background()
	transport := vertexTestTransport()
	service, err := aiplatform.NewService(ctx,
		option.WithHTTPClient(transport.client()),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	g := &VertexAiGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	if err := g.initResources(ctx, service); err != nil {
		t.Fatalf("initResources: %v", err)
	}

	got := map[string][]string{}
	for _, r := range g.Resources {
		got[r.InstanceInfo.Type] = append(got[r.InstanceInfo.Type], r.InstanceState.ID)
	}

	wantIDs := map[string]string{
		"google_vertex_ai_dataset":              vertexParent + "/datasets/123",
		"google_vertex_ai_endpoint":             vertexParent + "/endpoints/456",
		"google_vertex_ai_index":                vertexParent + "/indexes/789",
		"google_vertex_ai_index_endpoint":       vertexParent + "/indexEndpoints/321",
		"google_vertex_ai_metadata_store":       vertexParent + "/metadataStores/default",
		"google_vertex_ai_featurestore":         vertexParent + "/featurestores/fs1",
		"google_vertex_ai_feature_group":        vertexParent + "/featureGroups/fg1",
		"google_vertex_ai_feature_online_store": vertexParent + "/featureOnlineStores/fos1",
		"google_vertex_ai_tensorboard":          vertexParent + "/tensorboards/tb1",
		"google_vertex_ai_custom_job":           vertexParent + "/customJobs/111",
		"google_vertex_ai_pipeline_job":         vertexParent + "/pipelineJobs/222",
		"google_vertex_ai_rag_corpus":           vertexParent + "/ragCorpora/corpus1",
		"google_vertex_ai_reasoning_engine":     vertexParent + "/reasoningEngines/999",
		"google_vertex_ai_cache_config":         "projects/test-project/cacheConfig",
	}
	for resourceType, wantID := range wantIDs {
		ids := got[resourceType]
		if len(ids) != 1 || ids[0] != wantID {
			t.Errorf("%s: got IDs %v, want [%s]", resourceType, ids, wantID)
		}
	}

	// The deployed index joins the index endpoint name with a singular
	// "deployedIndex" segment, per the provider's ID format.
	deployed := got["google_vertex_ai_index_endpoint_deployed_index"]
	wantDeployed := vertexParent + "/indexEndpoints/321/deployedIndex/deployed_1"
	if len(deployed) != 1 || deployed[0] != wantDeployed {
		t.Errorf("deployed index: got %v, want [%s]", deployed, wantDeployed)
	}

	// Featurestore IAM: policy + one binding + one member from the canned
	// policy (one binding, one member).
	for _, iamType := range []string{
		"google_vertex_ai_featurestore_iam_policy",
		"google_vertex_ai_featurestore_iam_binding",
		"google_vertex_ai_featurestore_iam_member",
		"google_vertex_ai_featurestore_entitytype_iam_policy",
		"google_vertex_ai_feature_group_iam_policy",
		"google_vertex_ai_feature_online_store_iam_policy",
		"google_vertex_ai_feature_online_store_featureview_iam_policy",
	} {
		if len(got[iamType]) != 1 {
			t.Errorf("%s: got %d resources, want 1", iamType, len(got[iamType]))
		}
	}
}
