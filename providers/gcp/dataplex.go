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
	dataplexv1 "google.golang.org/api/dataplex/v1"
)

var dataplexAllowEmptyValues = []string{""}

var dataplexAdditionalFields = map[string]interface{}{}

const dataplexBaseURL = "https://dataplex.googleapis.com/v1/"

type DataplexGenerator struct {
	GCPService
}

// InitResources enumerates Dataplex resources for the configured
// project and region: lakes with their zones, the assets in each zone
// and the tasks on each lake; data scans; aspect types, entry types,
// entry groups and the entries in each group — each with the IAM
// policy, bindings and members expanded from its live IAM policy where
// the google provider exposes a consolidated *_iam resource for it —
// plus, through raw REST, glossaries with their categories and terms,
// data products with their data assets, and metadata feeds.
//
// The raw REST sections cover collections the pinned
// google.golang.org/api client (v0.230.0) cannot reach: it has no data
// product, metadata feed or entry link services at all, and its
// glossary services expose only IAM methods (no list). Their failures
// are logged rather than returned, since they are a newer API surface
// than the generated-client collections above.
//
// Deliberately not generated:
//   - google_dataplex_data_asset: a deprecated alias of
//     google_dataplex_data_product_data_asset over the same API
//     objects and ID format (its own provider documentation says so);
//     emitting both would double-manage every data product asset.
//   - google_dataplex_entry_link: the Dataplex API has no entry link
//     list method (create/get/patch/delete only; lookupEntryLinks is a
//     per-entry reverse lookup, not an enumeration), so entry links
//     cannot be discovered.
func (g *DataplexGenerator) InitResources() error {
	ctx := context.Background()
	dataplexService, err := dataplexv1.NewService(ctx)
	if err != nil {
		return err
	}
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, dataplexService, restClient)
}

func (g *DataplexGenerator) initResources(ctx context.Context, dataplexService *dataplexv1.Service, restClient *http.Client) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region)
	locationParent := "projects/" + project + "/locations/" + region.Name
	baseAttrs := map[string]string{"project": project, "location": region.Name}

	if err := g.loadLakes(ctx, dataplexService, locationParent, baseAttrs); err != nil {
		return err
	}
	if err := g.loadDataScans(ctx, dataplexService, locationParent, baseAttrs); err != nil {
		return err
	}
	if err := g.loadAspectTypes(ctx, dataplexService, locationParent, baseAttrs); err != nil {
		return err
	}
	if err := g.loadEntryTypes(ctx, dataplexService, locationParent, baseAttrs); err != nil {
		return err
	}
	if err := g.loadEntryGroups(ctx, dataplexService, locationParent, baseAttrs); err != nil {
		return err
	}
	g.loadGlossaries(ctx, restClient, locationParent, baseAttrs)
	g.loadDataProducts(ctx, restClient, locationParent, baseAttrs)
	g.loadMetadataFeeds(ctx, restClient, locationParent, baseAttrs)
	return nil
}

func (g *DataplexGenerator) newDataplexResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
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
		dataplexAllowEmptyValues,
		dataplexAdditionalFields,
	)
}

// loadDataplexIam expands the IAM policy of one generated-client
// resource into its _iam_policy/_iam_binding/_iam_member resources.
// parentKey is the provider argument that references the parent in the
// IAM resources (lake, zone, aspect_type, ...).
func (g *DataplexGenerator) loadDataplexIam(getPolicy func() (*dataplexv1.GoogleIamV1Policy, error), name, resourcePrefix, parentKey string, baseAttrs map[string]string) {
	policy, err := getPolicy()
	if err != nil {
		log.Println(err)
		return
	}
	bindings := []iamBinding{}
	for _, binding := range policy.Bindings {
		bindings = append(bindings, iamBinding{Role: binding.Role, Members: binding.Members})
	}
	parentAttrs := map[string]string{parentKey: name}
	for k, v := range baseAttrs {
		parentAttrs[k] = v
	}
	g.Resources = append(g.Resources, expandIamPolicy(
		bindings,
		name,
		resourcePrefix+"_iam_policy",
		resourcePrefix+"_iam_binding",
		resourcePrefix+"_iam_member",
		parentAttrs,
		g.newDataplexResource,
	)...)
}

func (g *DataplexGenerator) loadLakes(ctx context.Context, dataplexService *dataplexv1.Service, locationParent string, baseAttrs map[string]string) error {
	return dataplexService.Projects.Locations.Lakes.List(locationParent).Pages(ctx, func(page *dataplexv1.GoogleCloudDataplexV1ListLakesResponse) error {
		for _, lake := range page.Lakes {
			g.Resources = append(g.Resources, g.newDataplexResource(lake.Name, "google_dataplex_lake", baseAttrs))
			g.loadDataplexIam(func() (*dataplexv1.GoogleIamV1Policy, error) {
				return dataplexService.Projects.Locations.Lakes.GetIamPolicy(lake.Name).Context(ctx).Do()
			}, lake.Name, "google_dataplex_lake", "lake", baseAttrs)
			if err := g.loadZones(ctx, dataplexService, lake.Name, baseAttrs); err != nil {
				return err
			}
			if err := g.loadTasks(ctx, dataplexService, lake.Name, baseAttrs); err != nil {
				return err
			}
		}
		return nil
	})
}

func (g *DataplexGenerator) loadZones(ctx context.Context, dataplexService *dataplexv1.Service, lakeName string, baseAttrs map[string]string) error {
	return dataplexService.Projects.Locations.Lakes.Zones.List(lakeName).Pages(ctx, func(page *dataplexv1.GoogleCloudDataplexV1ListZonesResponse) error {
		for _, zone := range page.Zones {
			g.Resources = append(g.Resources, g.newDataplexResource(zone.Name, "google_dataplex_zone", baseAttrs))
			g.loadDataplexIam(func() (*dataplexv1.GoogleIamV1Policy, error) {
				return dataplexService.Projects.Locations.Lakes.Zones.GetIamPolicy(zone.Name).Context(ctx).Do()
			}, zone.Name, "google_dataplex_zone", "zone", baseAttrs)
			err := dataplexService.Projects.Locations.Lakes.Zones.Assets.List(zone.Name).Pages(ctx, func(page *dataplexv1.GoogleCloudDataplexV1ListAssetsResponse) error {
				for _, asset := range page.Assets {
					g.Resources = append(g.Resources, g.newDataplexResource(asset.Name, "google_dataplex_asset", baseAttrs))
					g.loadDataplexIam(func() (*dataplexv1.GoogleIamV1Policy, error) {
						return dataplexService.Projects.Locations.Lakes.Zones.Assets.GetIamPolicy(asset.Name).Context(ctx).Do()
					}, asset.Name, "google_dataplex_asset", "asset", baseAttrs)
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (g *DataplexGenerator) loadTasks(ctx context.Context, dataplexService *dataplexv1.Service, lakeName string, baseAttrs map[string]string) error {
	return dataplexService.Projects.Locations.Lakes.Tasks.List(lakeName).Pages(ctx, func(page *dataplexv1.GoogleCloudDataplexV1ListTasksResponse) error {
		for _, task := range page.Tasks {
			g.Resources = append(g.Resources, g.newDataplexResource(task.Name, "google_dataplex_task", baseAttrs))
			g.loadDataplexIam(func() (*dataplexv1.GoogleIamV1Policy, error) {
				return dataplexService.Projects.Locations.Lakes.Tasks.GetIamPolicy(task.Name).Context(ctx).Do()
			}, task.Name, "google_dataplex_task", "task", baseAttrs)
		}
		return nil
	})
}

func (g *DataplexGenerator) loadDataScans(ctx context.Context, dataplexService *dataplexv1.Service, locationParent string, baseAttrs map[string]string) error {
	return dataplexService.Projects.Locations.DataScans.List(locationParent).Pages(ctx, func(page *dataplexv1.GoogleCloudDataplexV1ListDataScansResponse) error {
		for _, scan := range page.DataScans {
			g.Resources = append(g.Resources, g.newDataplexResource(scan.Name, "google_dataplex_datascan", baseAttrs))
			g.loadDataplexIam(func() (*dataplexv1.GoogleIamV1Policy, error) {
				return dataplexService.Projects.Locations.DataScans.GetIamPolicy(scan.Name).Context(ctx).Do()
			}, scan.Name, "google_dataplex_datascan", "datascan", baseAttrs)
		}
		return nil
	})
}

func (g *DataplexGenerator) loadAspectTypes(ctx context.Context, dataplexService *dataplexv1.Service, locationParent string, baseAttrs map[string]string) error {
	return dataplexService.Projects.Locations.AspectTypes.List(locationParent).Pages(ctx, func(page *dataplexv1.GoogleCloudDataplexV1ListAspectTypesResponse) error {
		for _, aspectType := range page.AspectTypes {
			g.Resources = append(g.Resources, g.newDataplexResource(aspectType.Name, "google_dataplex_aspect_type", baseAttrs))
			g.loadDataplexIam(func() (*dataplexv1.GoogleIamV1Policy, error) {
				return dataplexService.Projects.Locations.AspectTypes.GetIamPolicy(aspectType.Name).Context(ctx).Do()
			}, aspectType.Name, "google_dataplex_aspect_type", "aspect_type", baseAttrs)
		}
		return nil
	})
}

func (g *DataplexGenerator) loadEntryTypes(ctx context.Context, dataplexService *dataplexv1.Service, locationParent string, baseAttrs map[string]string) error {
	return dataplexService.Projects.Locations.EntryTypes.List(locationParent).Pages(ctx, func(page *dataplexv1.GoogleCloudDataplexV1ListEntryTypesResponse) error {
		for _, entryType := range page.EntryTypes {
			g.Resources = append(g.Resources, g.newDataplexResource(entryType.Name, "google_dataplex_entry_type", baseAttrs))
			g.loadDataplexIam(func() (*dataplexv1.GoogleIamV1Policy, error) {
				return dataplexService.Projects.Locations.EntryTypes.GetIamPolicy(entryType.Name).Context(ctx).Do()
			}, entryType.Name, "google_dataplex_entry_type", "entry_type", baseAttrs)
		}
		return nil
	})
}

func (g *DataplexGenerator) loadEntryGroups(ctx context.Context, dataplexService *dataplexv1.Service, locationParent string, baseAttrs map[string]string) error {
	return dataplexService.Projects.Locations.EntryGroups.List(locationParent).Pages(ctx, func(page *dataplexv1.GoogleCloudDataplexV1ListEntryGroupsResponse) error {
		for _, group := range page.EntryGroups {
			g.Resources = append(g.Resources, g.newDataplexResource(group.Name, "google_dataplex_entry_group", baseAttrs))
			g.loadDataplexIam(func() (*dataplexv1.GoogleIamV1Policy, error) {
				return dataplexService.Projects.Locations.EntryGroups.GetIamPolicy(group.Name).Context(ctx).Do()
			}, group.Name, "google_dataplex_entry_group", "entry_group", baseAttrs)
			err := dataplexService.Projects.Locations.EntryGroups.Entries.List(group.Name).Pages(ctx, func(page *dataplexv1.GoogleCloudDataplexV1ListEntriesResponse) error {
				for _, entry := range page.Entries {
					g.Resources = append(g.Resources, g.newDataplexResource(entry.Name, "google_dataplex_entry", baseAttrs))
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// loadRestIam expands the IAM policy of one raw-REST resource through
// its :getIamPolicy endpoint.
func (g *DataplexGenerator) loadRestIam(ctx context.Context, restClient *http.Client, name, resourcePrefix, parentKey string, baseAttrs map[string]string) {
	bindings, err := restGetIamPolicy(ctx, restClient, dataplexBaseURL+name+":getIamPolicy")
	if err != nil {
		log.Println(err)
		return
	}
	parentAttrs := map[string]string{parentKey: name}
	for k, v := range baseAttrs {
		parentAttrs[k] = v
	}
	g.Resources = append(g.Resources, expandIamPolicy(
		bindings,
		name,
		resourcePrefix+"_iam_policy",
		resourcePrefix+"_iam_binding",
		resourcePrefix+"_iam_member",
		parentAttrs,
		g.newDataplexResource,
	)...)
}

func (g *DataplexGenerator) loadGlossaries(ctx context.Context, restClient *http.Client, locationParent string, baseAttrs map[string]string) {
	glossaries, err := restListResources(ctx, restClient, dataplexBaseURL+locationParent+"/glossaries")
	if err != nil {
		log.Println(err)
		return
	}
	for _, glossary := range glossaries {
		g.Resources = append(g.Resources, g.newDataplexResource(glossary.Name, "google_dataplex_glossary", baseAttrs))
		g.loadRestIam(ctx, restClient, glossary.Name, "google_dataplex_glossary", "glossary", baseAttrs)
		for _, collection := range []struct {
			suffix       string
			resourceType string
		}{
			{"/categories", "google_dataplex_glossary_category"},
			{"/terms", "google_dataplex_glossary_term"},
		} {
			items, err := restListResources(ctx, restClient, dataplexBaseURL+glossary.Name+collection.suffix)
			if err != nil {
				log.Println(err)
				continue
			}
			for _, item := range items {
				g.Resources = append(g.Resources, g.newDataplexResource(item.Name, collection.resourceType, baseAttrs))
			}
		}
	}
}

func (g *DataplexGenerator) loadDataProducts(ctx context.Context, restClient *http.Client, locationParent string, baseAttrs map[string]string) {
	products, err := restListResources(ctx, restClient, dataplexBaseURL+locationParent+"/dataProducts")
	if err != nil {
		log.Println(err)
		return
	}
	for _, product := range products {
		g.Resources = append(g.Resources, g.newDataplexResource(product.Name, "google_dataplex_data_product", baseAttrs))
		g.loadRestIam(ctx, restClient, product.Name, "google_dataplex_data_product", "data_product", baseAttrs)
		assets, err := restListResources(ctx, restClient, dataplexBaseURL+product.Name+"/dataAssets")
		if err != nil {
			log.Println(err)
			continue
		}
		for _, asset := range assets {
			g.Resources = append(g.Resources, g.newDataplexResource(asset.Name, "google_dataplex_data_product_data_asset", baseAttrs))
		}
	}
}

func (g *DataplexGenerator) loadMetadataFeeds(ctx context.Context, restClient *http.Client, locationParent string, baseAttrs map[string]string) {
	feeds, err := restListResources(ctx, restClient, dataplexBaseURL+locationParent+"/metadataFeeds")
	if err != nil {
		log.Println(err)
		return
	}
	for _, feed := range feeds {
		g.Resources = append(g.Resources, g.newDataplexResource(feed.Name, "google_dataplex_metadata_feed", baseAttrs))
	}
}
