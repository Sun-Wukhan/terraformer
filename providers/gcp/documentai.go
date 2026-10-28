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
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"google.golang.org/api/compute/v1"
	documentaiv1 "google.golang.org/api/documentai/v1"
	documentaiv1beta3 "google.golang.org/api/documentai/v1beta3"
)

// DocumentAiGenerator enumerates the Document AI surface of one
// project: processors in the configured region and in the "us" and
// "eu" multi-regions, each processor's default-version assignment
// (emitted when the processor reports a default version) and each
// processor's dataset schema singleton (read through the v1beta3
// client, the only version exposing the dataset schema).
//
// Deliberately not generated:
//   - google_document_ai_schema_version: the dataset schema is a
//     single versioned document under a processor's dataset; neither
//     the v1 nor the v1beta3 Document AI discovery document exposes
//     an addressable schema-version collection (no list or get), so
//     individual schema versions cannot be enumerated.
type DocumentAiGenerator struct {
	GCPService
}

func (g *DocumentAiGenerator) InitResources() error {
	ctx := context.Background()
	service, err := documentaiv1.NewService(ctx)
	if err != nil {
		return err
	}
	betaService, err := documentaiv1beta3.NewService(ctx)
	if err != nil {
		return err
	}
	g.initResources(ctx, service, betaService)
	return nil
}

func (g *DocumentAiGenerator) initResources(ctx context.Context, service *documentaiv1.Service, betaService *documentaiv1beta3.Service) {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	locations := []string{}
	for _, location := range []string{region, "us", "eu"} {
		if location == "" {
			continue
		}
		seen := false
		for _, existing := range locations {
			if existing == location {
				seen = true
			}
		}
		if !seen {
			locations = append(locations, location)
		}
	}
	for _, location := range locations {
		parent := "projects/" + project + "/locations/" + location
		processors, err := cxListNames(func(pageToken string) ([]string, string, error) {
			resp, err := service.Projects.Locations.Processors.List(parent).PageToken(pageToken).Context(ctx).Do()
			if err != nil {
				return nil, "", err
			}
			names := []string{}
			for _, processor := range resp.Processors {
				names = append(names, processor.Name)
				if processor.DefaultProcessorVersion != "" {
					// The default-version resource is keyed by the
					// processor's own name in the provider.
					g.Resources = append(g.Resources, g.newDocumentAiResource(processor.Name, "google_document_ai_processor_default_version"))
				}
			}
			return names, resp.NextPageToken, nil
		})
		if err != nil {
			log.Println(err)
			continue
		}
		for _, processor := range processors {
			g.Resources = append(g.Resources, g.newDocumentAiResource(processor, "google_document_ai_processor"))
			schemaName := processor + "/dataset/datasetSchema"
			if _, err := betaService.Projects.Locations.Processors.Dataset.GetDatasetSchema(schemaName).Context(ctx).Do(); err != nil {
				log.Println(err)
			} else {
				g.Resources = append(g.Resources, g.newDocumentAiResource(schemaName, "google_document_ai_schema"))
			}
		}
	}
}

func (g *DocumentAiGenerator) newDocumentAiResource(id, resourceType string) terraformutils.Resource {
	parts := strings.Split(id, "/")
	return terraformutils.NewResource(
		id,
		parts[len(parts)-1],
		resourceType,
		g.ProviderName,
		map[string]string{"project": g.GetArgs()["project"].(string)},
		[]string{},
		map[string]interface{}{},
	)
}
