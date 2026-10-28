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
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	biglakev1 "google.golang.org/api/biglake/v1"
	"google.golang.org/api/compute/v1"
)

// BiglakeGenerator enumerates the BigLake (Lakehouse) metastore
// surface of one project across its three catalog flavours:
//
//   - Classic metastore catalogs (generated v1 client) in the
//     configured region and the US/EU multi-regions, with their
//     databases and tables. IDs are full API names.
//   - Hive catalogs (raw REST, hive/v1 — no generated client):
//     project-level catalogs with databases, tables and the IAM
//     trio of each. IDs are full API names.
//   - Iceberg catalogs (raw REST): catalogs are listed through the
//     Iceberg REST catalog extensions API and carry the provider ID
//     iceberg/v1/restcatalog/extensions/projects/{project}/catalogs/{name};
//     namespaces and tables are listed through the Iceberg REST
//     catalog proper (its list replies use Iceberg-spec shapes —
//     namespace arrays and table identifiers — which are converted
//     to the provider's projects/{project}/catalogs/{catalog}/
//     namespaces/{namespace}[/tables/{table}] IDs). Each carries
//     its IAM trio, read from the v1 :getIamPolicy endpoint, which
//     this API serves over GET rather than POST.
type BiglakeGenerator struct {
	GCPService
}

const biglakeBaseURL = "https://biglake.googleapis.com/"

func (g *BiglakeGenerator) InitResources() error {
	ctx := context.Background()
	service, err := biglakev1.NewService(ctx)
	if err != nil {
		return err
	}
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		log.Println(err)
		return nil
	}
	g.initResources(ctx, service, restClient)
	return nil
}

func (g *BiglakeGenerator) initResources(ctx context.Context, service *biglakev1.Service, restClient *http.Client) {
	project := g.GetArgs()["project"].(string)
	g.biglakeClassic(ctx, service, project)
	g.biglakeHive(ctx, restClient, project)
	g.biglakeIceberg(ctx, restClient, project)
}

func (g *BiglakeGenerator) newBiglakeResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
	parts := strings.Split(id, "/")
	merged := map[string]string{"project": g.GetArgs()["project"].(string)}
	for k, v := range attrs {
		merged[k] = v
	}
	return terraformutils.NewResource(
		id,
		parts[len(parts)-1],
		resourceType,
		g.ProviderName,
		merged,
		[]string{},
		map[string]interface{}{},
	)
}

// biglakeClassic walks classic metastore catalogs in the configured
// region and the US/EU multi-regions.
func (g *BiglakeGenerator) biglakeClassic(ctx context.Context, service *biglakev1.Service, project string) {
	region := g.GetArgs()["region"].(compute.Region).Name
	locations := []string{}
	for _, location := range []string{region, "US", "EU"} {
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
		catalogs, err := cxListNames(func(pageToken string) ([]string, string, error) {
			resp, err := service.Projects.Locations.Catalogs.List(parent).PageToken(pageToken).Context(ctx).Do()
			if err != nil {
				return nil, "", err
			}
			names := []string{}
			for _, catalog := range resp.Catalogs {
				names = append(names, catalog.Name)
			}
			return names, resp.NextPageToken, nil
		})
		if err != nil {
			log.Println(err)
			continue
		}
		for _, catalog := range catalogs {
			g.Resources = append(g.Resources, g.newBiglakeResource(catalog, "google_biglake_catalog", map[string]string{}))
			databases, err := cxListNames(func(pageToken string) ([]string, string, error) {
				resp, err := service.Projects.Locations.Catalogs.Databases.List(catalog).PageToken(pageToken).Context(ctx).Do()
				if err != nil {
					return nil, "", err
				}
				names := []string{}
				for _, database := range resp.Databases {
					names = append(names, database.Name)
				}
				return names, resp.NextPageToken, nil
			})
			if err != nil {
				log.Println(err)
				continue
			}
			for _, database := range databases {
				g.Resources = append(g.Resources, g.newBiglakeResource(database, "google_biglake_database", map[string]string{}))
				tables, err := cxListNames(func(pageToken string) ([]string, string, error) {
					resp, err := service.Projects.Locations.Catalogs.Databases.Tables.List(database).PageToken(pageToken).Context(ctx).Do()
					if err != nil {
						return nil, "", err
					}
					names := []string{}
					for _, table := range resp.Tables {
						names = append(names, table.Name)
					}
					return names, resp.NextPageToken, nil
				})
				if err != nil {
					log.Println(err)
					continue
				}
				for _, table := range tables {
					g.Resources = append(g.Resources, g.newBiglakeResource(table, "google_biglake_table", map[string]string{}))
				}
			}
		}
	}
}

// biglakeGetIamPolicy reads a catalog-family IAM policy; unlike
// most Google APIs, BigLake serves :getIamPolicy over GET.
func biglakeGetIamPolicy(ctx context.Context, client *http.Client, resource string) ([]iamBinding, error) {
	body, err := restDo(ctx, client, http.MethodGet, biglakeBaseURL+"v1/"+resource+":getIamPolicy", nil)
	if err != nil {
		return nil, err
	}
	var policy struct {
		Bindings []iamBinding `json:"bindings"`
	}
	if err := json.Unmarshal(body, &policy); err != nil {
		return nil, err
	}
	return policy.Bindings, nil
}

// biglakeIam appends one IAM trio (literal type names) for a
// BigLake parent. apiResource is the short projects/... name used
// against the API; emittedID is the ID the provider expects for the
// IAM resources (identical to apiResource except for Iceberg
// catalogs, whose canonical ID carries the extensions prefix).
func (g *BiglakeGenerator) biglakeIam(ctx context.Context, restClient *http.Client, apiResource, emittedID, policyType, bindingType, memberType string, parentAttrs map[string]string) {
	bindings, err := biglakeGetIamPolicy(ctx, restClient, apiResource)
	if err != nil {
		log.Println(err)
		return
	}
	g.Resources = append(g.Resources, expandIamPolicy(bindings, emittedID, policyType, bindingType, memberType, parentAttrs,
		func(id, resourceType string, attrs map[string]string) terraformutils.Resource {
			return g.newBiglakeResource(id, resourceType, attrs)
		})...)
}

func (g *BiglakeGenerator) biglakeHive(ctx context.Context, restClient *http.Client, project string) {
	catalogs, err := restListResources(ctx, restClient, biglakeBaseURL+"hive/v1/projects/"+project+"/catalogs")
	if err != nil {
		log.Println(err)
		return
	}
	for _, catalog := range catalogs {
		g.Resources = append(g.Resources, g.newBiglakeResource(catalog.Name, "google_biglake_hive_catalog", map[string]string{}))
		g.biglakeIam(ctx, restClient, catalog.Name, catalog.Name,
			"google_biglake_hive_catalog_iam_policy",
			"google_biglake_hive_catalog_iam_binding",
			"google_biglake_hive_catalog_iam_member",
			map[string]string{"catalog": catalog.Name})
		databases, err := restListResources(ctx, restClient, biglakeBaseURL+"hive/v1/"+catalog.Name+"/databases")
		if err != nil {
			log.Println(err)
			continue
		}
		for _, database := range databases {
			g.Resources = append(g.Resources, g.newBiglakeResource(database.Name, "google_biglake_hive_database", map[string]string{}))
			g.biglakeIam(ctx, restClient, database.Name, database.Name,
				"google_biglake_hive_database_iam_policy",
				"google_biglake_hive_database_iam_binding",
				"google_biglake_hive_database_iam_member",
				map[string]string{"database": database.Name})
			tables, err := restListResources(ctx, restClient, biglakeBaseURL+"hive/v1/"+database.Name+"/tables")
			if err != nil {
				log.Println(err)
				continue
			}
			for _, table := range tables {
				g.Resources = append(g.Resources, g.newBiglakeResource(table.Name, "google_biglake_hive_table", map[string]string{}))
				g.biglakeIam(ctx, restClient, table.Name, table.Name,
					"google_biglake_hive_table_iam_policy",
					"google_biglake_hive_table_iam_binding",
					"google_biglake_hive_table_iam_member",
					map[string]string{"table": table.Name})
			}
		}
	}
}

func (g *BiglakeGenerator) biglakeIceberg(ctx context.Context, restClient *http.Client, project string) {
	catalogs, err := restListResources(ctx, restClient, biglakeBaseURL+"iceberg/v1/restcatalog/extensions/projects/"+project+"/catalogs")
	if err != nil {
		log.Println(err)
		return
	}
	for _, catalog := range catalogs {
		// catalog.Name is projects/{project}/catalogs/{catalog}; the
		// provider's canonical catalog ID carries the extensions
		// prefix.
		catalogID := "iceberg/v1/restcatalog/extensions/" + catalog.Name
		g.Resources = append(g.Resources, g.newBiglakeResource(catalogID, "google_biglake_iceberg_catalog", map[string]string{}))
		g.biglakeIam(ctx, restClient, catalog.Name, catalogID,
			"google_biglake_iceberg_catalog_iam_policy",
			"google_biglake_iceberg_catalog_iam_binding",
			"google_biglake_iceberg_catalog_iam_member",
			map[string]string{"catalog": catalog.Name})
		namespaces, err := g.icebergListNamespaces(ctx, restClient, catalog.Name)
		if err != nil {
			log.Println(err)
			continue
		}
		for _, namespace := range namespaces {
			namespaceID := catalog.Name + "/namespaces/" + namespace
			g.Resources = append(g.Resources, g.newBiglakeResource(namespaceID, "google_biglake_iceberg_namespace", map[string]string{}))
			g.biglakeIam(ctx, restClient, namespaceID, namespaceID,
				"google_biglake_iceberg_namespace_iam_policy",
				"google_biglake_iceberg_namespace_iam_binding",
				"google_biglake_iceberg_namespace_iam_member",
				map[string]string{"namespace": namespaceID})
			tables, err := g.icebergListTables(ctx, restClient, catalog.Name, namespace)
			if err != nil {
				log.Println(err)
				continue
			}
			for _, table := range tables {
				tableID := namespaceID + "/tables/" + table
				g.Resources = append(g.Resources, g.newBiglakeResource(tableID, "google_biglake_iceberg_table", map[string]string{}))
				g.biglakeIam(ctx, restClient, tableID, tableID,
					"google_biglake_iceberg_table_iam_policy",
					"google_biglake_iceberg_table_iam_binding",
					"google_biglake_iceberg_table_iam_member",
					map[string]string{"table": tableID})
			}
		}
	}
}

// icebergListNamespaces lists Iceberg namespaces under a catalog.
// The Iceberg REST reply is {"namespaces": [["a"], ["a","b"]]}; the
// API supports one level of nesting, and multi-part namespaces are
// joined with ".".
func (g *BiglakeGenerator) icebergListNamespaces(ctx context.Context, restClient *http.Client, catalog string) ([]string, error) {
	body, err := restDo(ctx, restClient, http.MethodGet, biglakeBaseURL+"iceberg/v1/restcatalog/v1/"+catalog+"/namespaces", nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Namespaces [][]string `json:"namespaces"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	namespaces := []string{}
	for _, namespace := range resp.Namespaces {
		namespaces = append(namespaces, strings.Join(namespace, "."))
	}
	return namespaces, nil
}

// icebergListTables lists Iceberg table names in one namespace. The
// Iceberg REST reply is {"identifiers": [{"namespace": [...],
// "name": "..."}]}.
func (g *BiglakeGenerator) icebergListTables(ctx context.Context, restClient *http.Client, catalog, namespace string) ([]string, error) {
	namespacePath := strings.ReplaceAll(namespace, ".", "%1F")
	body, err := restDo(ctx, restClient, http.MethodGet, biglakeBaseURL+"iceberg/v1/restcatalog/v1/"+catalog+"/namespaces/"+namespacePath+"/tables", nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Identifiers []struct {
			Name string `json:"name"`
		} `json:"identifiers"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	tables := []string{}
	for _, identifier := range resp.Identifiers {
		tables = append(tables, identifier.Name)
	}
	return tables, nil
}
