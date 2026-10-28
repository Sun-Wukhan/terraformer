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
	"google.golang.org/api/compute/v1"
	containeranalysisv1 "google.golang.org/api/containeranalysis/v1"
)

// This file is the depth extension of the existing GKE generator
// (the bigquery_depth.go / storage_depth.go pattern): the container
// family's remainders beyond clusters and node pools.
//
// Deliberately not generated:
//   - google_container_registry: there is no Container Registry
//     API resource to enumerate. The provider resource is a
//     project-level singleton whose ID is the project itself and
//     whose backing object is the registry's Cloud Storage bucket
//     (already emitted by the gcs service); it has no get/list
//     surface of its own, and emitting it unconditionally for every
//     project would claim a registry that may not exist.

// initContainerDepth enumerates Container Analysis notes (with
// IAM) and occurrences for the project, and GKE attached clusters
// over raw REST from the GKE multicloud API (which has no generated
// client in the pinned module). Failures are logged rather than
// returned: the depth collections must not lose the clusters the
// main generator already found.
func (g *GkeGenerator) initContainerDepth(ctx context.Context) {
	caService, err := containeranalysisv1.NewService(ctx)
	if err != nil {
		log.Println(err)
		return
	}
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		log.Println(err)
		return
	}
	g.initContainerDepthWith(ctx, caService, restClient)
}

func (g *GkeGenerator) initContainerDepthWith(ctx context.Context, caService *containeranalysisv1.Service, restClient *http.Client) {
	project := g.GetArgs()["project"].(string)
	parent := "projects/" + project
	baseAttrs := map[string]string{"project": project}

	err := caService.Projects.Notes.List(parent).Pages(ctx, func(page *containeranalysisv1.ListNotesResponse) error {
		for _, note := range page.Notes {
			g.Resources = append(g.Resources, g.newContainerDepthResource(note.Name, "google_container_analysis_note", baseAttrs))
			policy, err := caService.Projects.Notes.GetIamPolicy(note.Name, &containeranalysisv1.GetIamPolicyRequest{}).Context(ctx).Do()
			if err != nil {
				log.Println(err)
				continue
			}
			bindings := []iamBinding{}
			for _, binding := range policy.Bindings {
				bindings = append(bindings, iamBinding{Role: binding.Role, Members: binding.Members})
			}
			parentAttrs := map[string]string{"note": note.Name, "project": project}
			g.Resources = append(g.Resources, expandIamPolicy(
				bindings, note.Name,
				"google_container_analysis_note_iam_policy", "google_container_analysis_note_iam_binding", "google_container_analysis_note_iam_member",
				parentAttrs, g.newContainerDepthResource)...)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = caService.Projects.Occurrences.List(parent).Pages(ctx, func(page *containeranalysisv1.ListOccurrencesResponse) error {
		for _, occurrence := range page.Occurrences {
			g.Resources = append(g.Resources, g.newContainerDepthResource(occurrence.Name, "google_container_analysis_occurrence", baseAttrs))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}

	// Attached clusters are regional on the multicloud API; the GKE
	// generator's region argument is a compute.Region value.
	regionName := ""
	if region, ok := g.GetArgs()["region"].(compute.Region); ok {
		regionName = region.Name
	}
	if regionName == "" {
		return
	}
	locationParent := "projects/" + project + "/locations/" + regionName
	attached, err := restListResources(ctx, restClient, multicloudBaseURL+locationParent+"/attachedClusters")
	if err != nil {
		log.Println(err)
		return
	}
	attachedAttrs := map[string]string{"project": project, "location": regionName}
	for _, cluster := range attached {
		g.Resources = append(g.Resources, g.newContainerDepthResource(cluster.Name, "google_container_attached_cluster", attachedAttrs))
	}
}

func (g *GkeGenerator) newContainerDepthResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
	if attrs == nil {
		attrs = map[string]string{}
	}
	parts := strings.Split(id, "/")
	return terraformutils.NewResource(
		id,
		parts[len(parts)-1],
		resourceType,
		g.ProviderName,
		attrs,
		GkeAllowEmptyValues,
		GkeAdditionalFields,
	)
}
