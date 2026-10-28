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
)

var filestoreAllowEmptyValues = []string{""}

var filestoreAdditionalFields = map[string]interface{}{}

const filestoreBaseURL = "https://file.googleapis.com/v1/"

// FilestoreGenerator enumerates Filestore in the configured region
// over raw REST — the Cloud Filestore API (file.googleapis.com)
// has no generated client in the pinned google.golang.org/api
// module: instances with their snapshots, and backups.
type FilestoreGenerator struct {
	GCPService
}

func (g *FilestoreGenerator) InitResources() error {
	ctx := context.Background()
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		return err
	}
	g.initResources(ctx, restClient)
	return nil
}

func (g *FilestoreGenerator) initResources(ctx context.Context, restClient *http.Client) {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + region
	baseAttrs := map[string]string{"project": project, "location": region}

	instances, err := restListResources(ctx, restClient, filestoreBaseURL+parent+"/instances")
	if err != nil {
		log.Println(err)
		return
	}
	for _, instance := range instances {
		g.Resources = append(g.Resources, g.newFilestoreResource(instance.Name, "google_filestore_instance", baseAttrs))
		snapshots, err := restListResources(ctx, restClient, filestoreBaseURL+instance.Name+"/snapshots")
		if err != nil {
			log.Println(err)
			continue
		}
		for _, snapshot := range snapshots {
			g.Resources = append(g.Resources, g.newFilestoreResource(snapshot.Name, "google_filestore_snapshot", baseAttrs))
		}
	}
	backups, err := restListResources(ctx, restClient, filestoreBaseURL+parent+"/backups")
	if err != nil {
		log.Println(err)
		return
	}
	for _, backup := range backups {
		g.Resources = append(g.Resources, g.newFilestoreResource(backup.Name, "google_filestore_backup", baseAttrs))
	}
}

func (g *FilestoreGenerator) newFilestoreResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
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
		filestoreAllowEmptyValues,
		filestoreAdditionalFields,
	)
}
