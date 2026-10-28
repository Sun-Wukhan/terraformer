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

	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/logging/v2"
)

// Logging depth enumeration (ops sweep), extending the project-scoped
// LoggingGenerator the way storage_depth.go extends GCS: the folder and
// organization variants of sinks, exclusions, log bucket configs and
// settings, discovered through the project's Cloud Resource Manager
// ancestry (the same path SccGenerator uses for its folder/org scopes).
//
// The billing-account variants (google_logging_billing_account_sink,
// google_logging_billing_account_exclusion and
// google_logging_billing_account_bucket_config) remain deliberately
// not generated, as documented in logging.go: a project's ancestry
// contains folders and its organization, never its billing account, so
// there is no discovery path to a billing-account parent from a
// project-scoped import.

// initLoggingDepth enumerates the folder and organization scopes of
// the configured project. Failures in a non-project scope are logged
// rather than aborting the import.
func (g *LoggingGenerator) initLoggingDepth(ctx context.Context, service *logging.Service, crmService *cloudresourcemanager.Service) {
	project := g.GetArgs()["project"].(string)
	for _, scope := range opsScopes(ctx, crmService, project) {
		if scope.kind == "project" {
			continue
		}
		if err := g.loadLoggingScope(ctx, service, scope); err != nil {
			log.Println(err)
		}
	}
}

func (g *LoggingGenerator) loadLoggingScope(ctx context.Context, service *logging.Service, scope sccScope) error {
	switch scope.kind {
	case "folder":
		return g.loadLoggingFolder(ctx, service, scope)
	case "organization":
		return g.loadLoggingOrganization(ctx, service, scope)
	}
	return nil
}

func (g *LoggingGenerator) loadLoggingFolder(ctx context.Context, service *logging.Service, scope sccScope) error {
	parent := scope.parent
	if err := service.Folders.Sinks.List(parent).Pages(ctx, func(page *logging.ListSinksResponse) error {
		for _, sink := range page.Sinks {
			g.Resources = append(g.Resources, g.newLoggingResource(
				parent+"/sinks/"+sink.Name, sink.Name, "google_logging_folder_sink",
				map[string]string{"name": sink.Name, "folder": scope.id},
			))
		}
		return nil
	}); err != nil {
		return err
	}
	if err := service.Folders.Exclusions.List(parent).Pages(ctx, func(page *logging.ListExclusionsResponse) error {
		for _, exclusion := range page.Exclusions {
			g.Resources = append(g.Resources, g.newLoggingResource(
				parent+"/exclusions/"+exclusion.Name, exclusion.Name, "google_logging_folder_exclusion",
				map[string]string{"name": exclusion.Name, "folder": scope.id},
			))
		}
		return nil
	}); err != nil {
		return err
	}
	if err := service.Folders.Locations.Buckets.List(parent+"/locations/-").Pages(ctx, func(page *logging.ListBucketsResponse) error {
		for _, bucket := range page.Buckets {
			location, bucketID := loggingParentAttrs(bucket.Name)
			g.Resources = append(g.Resources, g.newLoggingResource(
				bucket.Name, bucketID, "google_logging_folder_bucket_config",
				map[string]string{"location": location, "bucket_id": bucketID, "folder": scope.id},
			))
		}
		return nil
	}); err != nil {
		return err
	}
	// Folder settings are a singleton: a successful get means the
	// settings resource exists (it always does for an active folder).
	if _, err := service.Folders.GetSettings(parent).Context(ctx).Do(); err == nil {
		g.Resources = append(g.Resources, g.newLoggingResource(
			parent, scope.id, "google_logging_folder_settings",
			map[string]string{"folder": scope.id},
		))
	} else {
		log.Println(err)
	}
	return nil
}

func (g *LoggingGenerator) loadLoggingOrganization(ctx context.Context, service *logging.Service, scope sccScope) error {
	parent := scope.parent
	if err := service.Organizations.Sinks.List(parent).Pages(ctx, func(page *logging.ListSinksResponse) error {
		for _, sink := range page.Sinks {
			g.Resources = append(g.Resources, g.newLoggingResource(
				parent+"/sinks/"+sink.Name, sink.Name, "google_logging_organization_sink",
				map[string]string{"name": sink.Name, "organization": scope.id},
			))
		}
		return nil
	}); err != nil {
		return err
	}
	if err := service.Organizations.Exclusions.List(parent).Pages(ctx, func(page *logging.ListExclusionsResponse) error {
		for _, exclusion := range page.Exclusions {
			g.Resources = append(g.Resources, g.newLoggingResource(
				parent+"/exclusions/"+exclusion.Name, exclusion.Name, "google_logging_organization_exclusion",
				map[string]string{"name": exclusion.Name, "organization": scope.id},
			))
		}
		return nil
	}); err != nil {
		return err
	}
	if err := service.Organizations.Locations.Buckets.List(parent+"/locations/-").Pages(ctx, func(page *logging.ListBucketsResponse) error {
		for _, bucket := range page.Buckets {
			location, bucketID := loggingParentAttrs(bucket.Name)
			g.Resources = append(g.Resources, g.newLoggingResource(
				bucket.Name, bucketID, "google_logging_organization_bucket_config",
				map[string]string{"location": location, "bucket_id": bucketID, "organization": scope.id},
			))
		}
		return nil
	}); err != nil {
		return err
	}
	if _, err := service.Organizations.GetSettings(parent).Context(ctx).Do(); err == nil {
		g.Resources = append(g.Resources, g.newLoggingResource(
			parent, scope.id, "google_logging_organization_settings",
			map[string]string{"organization": scope.id},
		))
	} else {
		log.Println(err)
	}
	return nil
}
