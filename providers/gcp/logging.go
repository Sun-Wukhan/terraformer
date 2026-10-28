// Copyright 2018 The Terraformer Authors.
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
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/iterator"
	"google.golang.org/api/logging/v2"

	"cloud.google.com/go/logging/logadmin"
)

var loggingAllowEmptyValues = []string{}

var loggingAdditionalFields = map[string]interface{}{}

type LoggingGenerator struct {
	GCPService
}

func (g *LoggingGenerator) loadLoggingMetrics(ctx context.Context, client *logadmin.Client) error {
	metricIterator := client.Metrics(ctx)

	for {
		metric, err := metricIterator.Next()

		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		g.Resources = append(g.Resources, terraformutils.NewResource(
			metric.ID,
			metric.ID,
			"google_logging_metric",
			g.ProviderName,
			map[string]string{
				"name":    metric.ID,
				"project": g.GetArgs()["project"].(string),
			},
			loggingAllowEmptyValues,
			loggingAdditionalFields,
		))
	}
	return nil
}

// Generate TerraformResources from GCP API
func (g *LoggingGenerator) InitResources() error {
	project := g.GetArgs()["project"].(string)
	ctx := context.Background()
	client, err := logadmin.NewClient(ctx, project)
	if err != nil {
		return err
	}

	if err := g.loadLoggingMetrics(ctx, client); err != nil {
		return err
	}

	service, err := logging.NewService(ctx)
	if err != nil {
		return err
	}
	if err := g.initResources(ctx, service); err != nil {
		return err
	}
	// Folder/organization depth (logging_depth.go); a CRM failure only
	// disables the depth pass, never the project enumeration.
	crmService, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		log.Println(err)
		return nil
	}
	g.initLoggingDepth(ctx, service, crmService)
	return nil
}

// initResources enumerates the project-scoped Cloud Logging configuration
// with the logging v2 REST client: sinks, exclusions, log buckets with
// their linked datasets and views (including view IAM), log scopes and
// saved queries.
//
// Deliberately not generated: the billing-account variants
// (google_logging_billing_account_sink/_exclusion/_bucket_config).
// A project's Cloud Resource Manager ancestry contains its folders and
// organization (enumerated by the depth pass in logging_depth.go) but
// never its billing account, so there is no discovery path to
// billing-account parents from a project import.
func (g *LoggingGenerator) initResources(ctx context.Context, service *logging.Service) error {
	project := g.GetArgs()["project"].(string)
	projectParent := "projects/" + project
	if err := g.loadProjectSinks(ctx, service, projectParent); err != nil {
		return err
	}
	if err := g.loadProjectExclusions(ctx, service, projectParent); err != nil {
		return err
	}
	if err := g.loadProjectBuckets(ctx, service, projectParent); err != nil {
		return err
	}
	if err := g.loadLogScopes(ctx, service, projectParent); err != nil {
		return err
	}
	return g.loadSavedQueries(ctx, service, projectParent)
}

func (g *LoggingGenerator) newLoggingResource(id, name, resourceType string, attrs map[string]string) terraformutils.Resource {
	if attrs == nil {
		attrs = map[string]string{}
	}
	if _, ok := attrs["project"]; !ok {
		attrs["project"] = g.GetArgs()["project"].(string)
	}
	return terraformutils.NewResource(
		id,
		name,
		resourceType,
		g.ProviderName,
		attrs,
		loggingAllowEmptyValues,
		loggingAdditionalFields,
	)
}

// loggingParentAttrs derives the location and bucket ID from a full
// resource name of the form projects/{p}/locations/{l}/buckets/{b}[/...].
func loggingParentAttrs(name string) (location, bucket string) {
	parts := strings.Split(name, "/")
	for i, part := range parts {
		if part == "locations" && i+1 < len(parts) {
			location = parts[i+1]
		}
		if part == "buckets" && i+1 < len(parts) {
			bucket = parts[i+1]
		}
	}
	return location, bucket
}

func loggingLastSegment(name string) string {
	parts := strings.Split(name, "/")
	return parts[len(parts)-1]
}

func (g *LoggingGenerator) loadProjectSinks(ctx context.Context, service *logging.Service, projectParent string) error {
	return service.Projects.Sinks.List(projectParent).Pages(ctx, func(page *logging.ListSinksResponse) error {
		for _, sink := range page.Sinks {
			// Sink names are short IDs within the project; the provider ID
			// is projects/{project}/sinks/{sink}.
			g.Resources = append(g.Resources, g.newLoggingResource(
				projectParent+"/sinks/"+sink.Name,
				sink.Name,
				"google_logging_project_sink",
				map[string]string{"name": sink.Name},
			))
		}
		return nil
	})
}

func (g *LoggingGenerator) loadProjectExclusions(ctx context.Context, service *logging.Service, projectParent string) error {
	return service.Projects.Exclusions.List(projectParent).Pages(ctx, func(page *logging.ListExclusionsResponse) error {
		for _, exclusion := range page.Exclusions {
			g.Resources = append(g.Resources, g.newLoggingResource(
				projectParent+"/exclusions/"+exclusion.Name,
				exclusion.Name,
				"google_logging_project_exclusion",
				map[string]string{"name": exclusion.Name},
			))
		}
		return nil
	})
}

func (g *LoggingGenerator) loadProjectBuckets(ctx context.Context, service *logging.Service, projectParent string) error {
	// "-" is the API's location wildcard: buckets in every location.
	return service.Projects.Locations.Buckets.List(projectParent+"/locations/-").Pages(ctx, func(page *logging.ListBucketsResponse) error {
		for _, bucket := range page.Buckets {
			location, bucketID := loggingParentAttrs(bucket.Name)
			g.Resources = append(g.Resources, g.newLoggingResource(
				bucket.Name,
				bucketID,
				"google_logging_project_bucket_config",
				map[string]string{
					"location":  location,
					"bucket_id": bucketID,
					"project":   g.GetArgs()["project"].(string),
				},
			))
			if err := g.loadBucketLinks(ctx, service, bucket.Name); err != nil {
				return err
			}
			if err := g.loadBucketViews(ctx, service, bucket.Name); err != nil {
				return err
			}
		}
		return nil
	})
}

func (g *LoggingGenerator) loadBucketLinks(ctx context.Context, service *logging.Service, bucketName string) error {
	return service.Projects.Locations.Buckets.Links.List(bucketName).Pages(ctx, func(page *logging.ListLinksResponse) error {
		for _, link := range page.Links {
			// Only links backed by a BigQuery dataset are
			// google_logging_linked_dataset resources.
			if link.BigqueryDataset == nil {
				continue
			}
			location, bucketID := loggingParentAttrs(link.Name)
			g.Resources = append(g.Resources, g.newLoggingResource(
				link.Name,
				loggingLastSegment(link.Name),
				"google_logging_linked_dataset",
				map[string]string{
					"name":           loggingLastSegment(link.Name),
					"location":       location,
					"bucket":         bucketID,
					"linked_dataset": link.BigqueryDataset.DatasetId,
				},
			))
		}
		return nil
	})
}

func (g *LoggingGenerator) loadBucketViews(ctx context.Context, service *logging.Service, bucketName string) error {
	project := g.GetArgs()["project"].(string)
	return service.Projects.Locations.Buckets.Views.List(bucketName).Pages(ctx, func(page *logging.ListViewsResponse) error {
		for _, view := range page.Views {
			location, bucketID := loggingParentAttrs(view.Name)
			viewAttrs := map[string]string{
				"parent":   "projects/" + project,
				"location": location,
				"bucket":   bucketID,
				"name":     loggingLastSegment(view.Name),
			}
			g.Resources = append(g.Resources, g.newLoggingResource(
				view.Name,
				loggingLastSegment(view.Name),
				"google_logging_log_view",
				viewAttrs,
			))
			g.Resources = append(g.Resources, g.loadLogViewIam(ctx, service, view.Name, viewAttrs)...)
		}
		return nil
	})
}

// loadLogViewIam appends the google_logging_log_view_iam_policy/_binding/
// _member resources for one view, following the pattern used for GCS bucket
// IAM: the policy resource ID is the view's full resource name.
func (g *LoggingGenerator) loadLogViewIam(ctx context.Context, service *logging.Service, viewName string, viewAttrs map[string]string) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	policy, err := service.Projects.Locations.Buckets.Views.GetIamPolicy(viewName, &logging.GetIamPolicyRequest{}).Context(ctx).Do()
	if err != nil {
		log.Println(err)
		return resources
	}
	resources = append(resources, g.newLoggingResource(
		viewName,
		loggingLastSegment(viewName),
		"google_logging_log_view_iam_policy",
		viewAttrs,
	))
	for _, binding := range policy.Bindings {
		bindingAttrs := map[string]string{"role": binding.Role}
		for k, v := range viewAttrs {
			bindingAttrs[k] = v
		}
		resources = append(resources, g.newLoggingResource(
			viewName,
			loggingLastSegment(viewName),
			"google_logging_log_view_iam_binding",
			bindingAttrs,
		))
		for _, member := range binding.Members {
			memberAttrs := map[string]string{"role": binding.Role, "member": member}
			for k, v := range viewAttrs {
				memberAttrs[k] = v
			}
			resources = append(resources, g.newLoggingResource(
				viewName,
				loggingLastSegment(viewName),
				"google_logging_log_view_iam_member",
				memberAttrs,
			))
		}
	}
	return resources
}

func (g *LoggingGenerator) loadLogScopes(ctx context.Context, service *logging.Service, projectParent string) error {
	return service.Projects.Locations.LogScopes.List(projectParent+"/locations/-").Pages(ctx, func(page *logging.ListLogScopesResponse) error {
		for _, scope := range page.LogScopes {
			location, _ := loggingParentAttrs(scope.Name)
			g.Resources = append(g.Resources, g.newLoggingResource(
				scope.Name,
				loggingLastSegment(scope.Name),
				"google_logging_log_scope",
				map[string]string{
					"name":     loggingLastSegment(scope.Name),
					"location": location,
				},
			))
		}
		return nil
	})
}

func (g *LoggingGenerator) loadSavedQueries(ctx context.Context, service *logging.Service, projectParent string) error {
	return service.Projects.Locations.SavedQueries.List(projectParent+"/locations/-").Pages(ctx, func(page *logging.ListSavedQueriesResponse) error {
		for _, query := range page.SavedQueries {
			location, _ := loggingParentAttrs(query.Name)
			g.Resources = append(g.Resources, g.newLoggingResource(
				query.Name,
				loggingLastSegment(query.Name),
				"google_logging_saved_query",
				map[string]string{
					"name":     loggingLastSegment(query.Name),
					"location": location,
				},
			))
		}
		return nil
	})
}
