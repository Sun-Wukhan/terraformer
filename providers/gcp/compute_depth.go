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
	"google.golang.org/api/compute/v1"
)

var computeDepthAllowEmptyValues = []string{""}

var computeDepthAdditionalFields = map[string]interface{}{}

// computeZoneNames extracts zone names from the configured region's
// zone URLs. Shared by the Batch 5 compute generators.
func computeZoneNames(region compute.Region) []string {
	zones := []string{}
	for _, zone := range region.Zones {
		zones = append(zones, zone[strings.LastIndex(zone, "/")+1:])
	}
	return zones
}

// computeRestItem is one item of a raw REST compute list response. The
// collections terraformer reads this way are single-level; only the
// fields the generators need are decoded.
type computeRestItem struct {
	Name             string `json:"name"`
	ActivationStatus string `json:"activationStatus"`
}

// listComputeRest pages through a compute REST list endpoint (the
// collections the pinned generated client does not expose) and returns
// its items. Endpoints are the GA ones from the compute discovery
// document; a nil client or a failed call yields an error the caller
// logs, mirroring the log-and-continue style of bigquery_depth.go.
func listComputeRest(ctx context.Context, client *http.Client, url string) ([]computeRestItem, error) {
	items := []computeRestItem{}
	pageToken := ""
	for {
		pageURL := url
		if pageToken != "" {
			separator := "?"
			if strings.Contains(pageURL, "?") {
				separator = "&"
			}
			pageURL += separator + "pageToken=" + pageToken
		}
		body, err := restDo(ctx, client, http.MethodGet, pageURL, nil)
		if err != nil {
			return nil, err
		}
		var page struct {
			Items         []computeRestItem `json:"items"`
			NextPageToken string            `json:"nextPageToken"`
		}
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		items = append(items, page.Items...)
		if page.NextPageToken == "" {
			return items, nil
		}
		pageToken = page.NextPageToken
	}
}

type ComputeDepthGenerator struct {
	GCPService
}

// InitResources enumerates the Compute Engine collections the code
// generator never covered: interconnects, HA VPN gateways, storage
// pools, instant snapshots, region commitments, region instance
// templates, region SSL policies, region target TCP proxies, target
// gRPC proxies, service and network attachments, public prefixes,
// network edge security services, region network endpoint groups,
// future reservations, region backend buckets, cross-site networks and
// their wire groups, interconnect groups, VM extension policies,
// preview features, rollout plans and the region health aggregation
// family.
//
// Most collections come from the generated compute client. The ones
// the pinned client has no service for (interconnect groups,
// interconnect attachment groups, cross-site networks, wire groups,
// future reservations, region backend buckets, VM extension policies,
// preview features, rollout plans, region health aggregation policies,
// region health sources, region composite health checks) are read over
// raw REST against the GA endpoints in the compute discovery document,
// per the accepted method for client gaps (see gcp_rest.go).
//
// google_compute_ha_vpn_gateway is enumerated from the vpnGateways
// collection, which is distinct from the targetVpnGateways collection
// behind google_compute_vpn_gateway: HA gateways were in fact not
// covered before this batch (the note in routerSubresources.go has
// been corrected).
//
// A failure in any one collection is logged and never aborts the rest:
// every collection here is additive depth on the covered estate.
//
// Deliberately not generated:
//   - google_compute_instance_from_template and
//     google_compute_instance_from_machine_image: creation-time
//     wrappers; the resulting object is a google_compute_instance,
//     which the instances generator already emits.
//   - google_compute_disk_async_replication: the provider resource
//     does not support import, and the replication pairing is
//     expressed through fields (asyncPrimaryDisk / asyncSecondaryDisks)
//     on the two google_compute_disk resources, both already emitted.
//   - google_compute_backend_bucket_signed_url_key and
//     google_compute_backend_service_signed_url_key: the provider
//     documents that these resources do not support import; the key
//     material is secret and returned only at creation.
//   - google_compute_global_network_endpoint,
//     google_compute_network_endpoint, google_compute_network_endpoints
//     and google_compute_region_network_endpoint: endpoint members are
//     identified by IP/port rather than by a stable name (the position
//     taken for global endpoints in Batch 1); the enumerable parents
//     (the network endpoint groups) are generated.
func (g *ComputeDepthGenerator) InitResources() error {
	ctx := context.Background()
	computeService, err := compute.NewService(ctx)
	if err != nil {
		return err
	}
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, computeService, restClient)
}

func (g *ComputeDepthGenerator) initResources(ctx context.Context, computeService *compute.Service, restClient *http.Client) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region)
	regionName := region.Name
	zones := computeZoneNames(region)

	g.loadGlobalCollections(ctx, computeService, project)
	g.loadRegionalCollections(ctx, computeService, project, regionName)
	for _, zone := range zones {
		g.loadZonalCollections(ctx, computeService, project, zone)
	}
	g.loadRestCollections(ctx, restClient, project, regionName, zones)
	return nil
}

func (g *ComputeDepthGenerator) newComputeDepthResource(id, name, resourceType string, attrs map[string]string) terraformutils.Resource {
	if attrs == nil {
		attrs = map[string]string{}
	}
	return terraformutils.NewResource(
		id,
		name,
		resourceType,
		g.ProviderName,
		attrs,
		computeDepthAllowEmptyValues,
		computeDepthAdditionalFields,
	)
}

func (g *ComputeDepthGenerator) loadGlobalCollections(ctx context.Context, computeService *compute.Service, project string) {
	err := computeService.Interconnects.List(project).Pages(ctx, func(page *compute.InterconnectList) error {
		for _, item := range page.Items {
			g.Resources = append(g.Resources, g.newComputeDepthResource(
				item.Name, item.Name, "google_compute_interconnect",
				map[string]string{"name": item.Name, "project": project}))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = computeService.TargetGrpcProxies.List(project).Pages(ctx, func(page *compute.TargetGrpcProxyList) error {
		for _, item := range page.Items {
			g.Resources = append(g.Resources, g.newComputeDepthResource(
				item.Name, item.Name, "google_compute_target_grpc_proxy",
				map[string]string{"name": item.Name, "project": project}))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = computeService.PublicAdvertisedPrefixes.List(project).Pages(ctx, func(page *compute.PublicAdvertisedPrefixList) error {
		for _, item := range page.Items {
			g.Resources = append(g.Resources, g.newComputeDepthResource(
				"projects/"+project+"/global/publicAdvertisedPrefixes/"+item.Name, item.Name, "google_compute_public_advertised_prefix",
				map[string]string{"name": item.Name, "project": project}))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (g *ComputeDepthGenerator) loadRegionalCollections(ctx context.Context, computeService *compute.Service, project, region string) {
	attrs := func(name string) map[string]string {
		return map[string]string{"name": name, "project": project, "region": region}
	}
	err := computeService.VpnGateways.List(project, region).Pages(ctx, func(page *compute.VpnGatewayList) error {
		for _, item := range page.Items {
			g.Resources = append(g.Resources, g.newComputeDepthResource(
				item.Name, item.Name, "google_compute_ha_vpn_gateway", attrs(item.Name)))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = computeService.RegionCommitments.List(project, region).Pages(ctx, func(page *compute.CommitmentList) error {
		for _, item := range page.Items {
			g.Resources = append(g.Resources, g.newComputeDepthResource(
				item.Name, item.Name, "google_compute_region_commitment", attrs(item.Name)))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = computeService.RegionInstanceTemplates.List(project, region).Pages(ctx, func(page *compute.InstanceTemplateList) error {
		for _, item := range page.Items {
			g.Resources = append(g.Resources, g.newComputeDepthResource(
				item.Name, item.Name, "google_compute_region_instance_template", attrs(item.Name)))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = computeService.RegionSslPolicies.List(project, region).Pages(ctx, func(page *compute.SslPoliciesList) error {
		for _, item := range page.Items {
			g.Resources = append(g.Resources, g.newComputeDepthResource(
				item.Name, item.Name, "google_compute_region_ssl_policy", attrs(item.Name)))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = computeService.RegionTargetTcpProxies.List(project, region).Pages(ctx, func(page *compute.TargetTcpProxyList) error {
		for _, item := range page.Items {
			g.Resources = append(g.Resources, g.newComputeDepthResource(
				item.Name, item.Name, "google_compute_region_target_tcp_proxy", attrs(item.Name)))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = computeService.ServiceAttachments.List(project, region).Pages(ctx, func(page *compute.ServiceAttachmentList) error {
		for _, item := range page.Items {
			g.Resources = append(g.Resources, g.newComputeDepthResource(
				item.Name, item.Name, "google_compute_service_attachment", attrs(item.Name)))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = computeService.NetworkAttachments.List(project, region).Pages(ctx, func(page *compute.NetworkAttachmentList) error {
		for _, item := range page.Items {
			g.Resources = append(g.Resources, g.newComputeDepthResource(
				item.Name, item.Name, "google_compute_network_attachment", attrs(item.Name)))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = computeService.PublicDelegatedPrefixes.List(project, region).Pages(ctx, func(page *compute.PublicDelegatedPrefixList) error {
		for _, item := range page.Items {
			g.Resources = append(g.Resources, g.newComputeDepthResource(
				item.Name, item.Name, "google_compute_public_delegated_prefix", attrs(item.Name)))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	// Network edge security services have no per-region list method in
	// the client, only the aggregated list; keep the configured region.
	err = computeService.NetworkEdgeSecurityServices.AggregatedList(project).Pages(ctx, func(page *compute.NetworkEdgeSecurityServiceAggregatedList) error {
		for scope, scopedList := range page.Items {
			if scope != "regions/"+region {
				continue
			}
			for _, item := range scopedList.NetworkEdgeSecurityServices {
				g.Resources = append(g.Resources, g.newComputeDepthResource(
					item.Name, item.Name, "google_compute_network_edge_security_service", attrs(item.Name)))
			}
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = computeService.RegionNetworkEndpointGroups.List(project, region).Pages(ctx, func(page *compute.NetworkEndpointGroupList) error {
		for _, item := range page.Items {
			g.Resources = append(g.Resources, g.newComputeDepthResource(
				item.Name, item.Name, "google_compute_region_network_endpoint_group", attrs(item.Name)))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = computeService.RegionInstantSnapshots.List(project, region).Pages(ctx, func(page *compute.InstantSnapshotList) error {
		for _, item := range page.Items {
			g.Resources = append(g.Resources, g.newComputeDepthResource(
				item.Name, item.Name, "google_compute_region_instant_snapshot", attrs(item.Name)))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (g *ComputeDepthGenerator) loadZonalCollections(ctx context.Context, computeService *compute.Service, project, zone string) {
	attrs := func(name string) map[string]string {
		return map[string]string{"name": name, "project": project, "zone": zone}
	}
	err := computeService.StoragePools.List(project, zone).Pages(ctx, func(page *compute.StoragePoolList) error {
		for _, item := range page.Items {
			g.Resources = append(g.Resources, g.newComputeDepthResource(
				zone+"/"+item.Name, item.Name, "google_compute_storage_pool", attrs(item.Name)))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = computeService.InstantSnapshots.List(project, zone).Pages(ctx, func(page *compute.InstantSnapshotList) error {
		for _, item := range page.Items {
			g.Resources = append(g.Resources, g.newComputeDepthResource(
				zone+"/"+item.Name, item.Name, "google_compute_instant_snapshot", attrs(item.Name)))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

// loadRestCollections reads the collections that exist in the GA
// compute API but not in the pinned generated client, over raw REST.
func (g *ComputeDepthGenerator) loadRestCollections(ctx context.Context, restClient *http.Client, project, region string, zones []string) {
	base := "https://compute.googleapis.com/compute/v1/projects/" + project
	emit := func(items []computeRestItem, idPrefix, resourceType string, attrs map[string]string) {
		for _, item := range items {
			resourceAttrs := map[string]string{"name": item.Name, "project": project}
			for k, v := range attrs {
				resourceAttrs[k] = v
			}
			g.Resources = append(g.Resources, g.newComputeDepthResource(
				idPrefix+item.Name, item.Name, resourceType, resourceAttrs))
		}
	}
	load := func(url, idPrefix, resourceType string, attrs map[string]string) []computeRestItem {
		items, err := listComputeRest(ctx, restClient, url)
		if err != nil {
			log.Println(err)
			return nil
		}
		emit(items, idPrefix, resourceType, attrs)
		return items
	}

	load(base+"/global/interconnectGroups", "", "google_compute_interconnect_group", nil)
	load(base+"/global/interconnectAttachmentGroups", "", "google_compute_interconnect_attachment_group", nil)
	load(base+"/global/vmExtensionPolicies", "", "google_compute_global_vm_extension_policy", nil)
	load(base+"/global/rolloutPlans", "", "google_compute_rollout_plan", nil)

	// Preview features are project enrollments: only the features the
	// project has actually enabled are configuration.
	if items, err := listComputeRest(ctx, restClient, base+"/global/previewFeatures"); err != nil {
		log.Println(err)
	} else {
		for _, item := range items {
			if item.ActivationStatus != "ENABLED" {
				continue
			}
			name := item.Name[strings.LastIndex(item.Name, "/")+1:]
			g.Resources = append(g.Resources, g.newComputeDepthResource(
				name, name, "google_compute_preview_feature",
				map[string]string{"name": name, "project": project}))
		}
	}

	// Wire groups live under their cross-site network.
	for _, network := range load(base+"/global/crossSiteNetworks", "", "google_compute_cross_site_network", nil) {
		load(base+"/global/crossSiteNetworks/"+network.Name+"/wireGroups",
			network.Name+"/", "google_compute_wire_group",
			map[string]string{"cross_site_network": network.Name})
	}

	regionAttrs := map[string]string{"region": region}
	load(base+"/regions/"+region+"/backendBuckets", "", "google_compute_region_backend_bucket", regionAttrs)
	load(base+"/regions/"+region+"/healthAggregationPolicies", "", "google_compute_region_health_aggregation_policy", regionAttrs)
	load(base+"/regions/"+region+"/healthSources", "", "google_compute_region_health_source", regionAttrs)
	load(base+"/regions/"+region+"/compositeHealthChecks", "", "google_compute_region_composite_health_check", regionAttrs)

	for _, zone := range zones {
		zoneAttrs := map[string]string{"zone": zone}
		load(base+"/zones/"+zone+"/futureReservations", zone+"/", "google_compute_future_reservation", zoneAttrs)
		load(base+"/zones/"+zone+"/vmExtensionPolicies", zone+"/", "google_compute_zone_vm_extension_policy", zoneAttrs)
	}
}
