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

package aws

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudsearch"
	"github.com/aws/aws-sdk-go-v2/service/location"
	"github.com/aws/aws-sdk-go-v2/service/pipes"
	"github.com/aws/aws-sdk-go-v2/service/schemas"
)

var searchAllowEmptyValues = []string{"tags."}

// --- CloudSearch -----------------------------------------------------------

type CloudSearchGenerator struct {
	AWSService
}

// CloudSearch domains are keyed by domain name (the provider importer
// ID); the domain service access policy subresource is keyed by the
// same domain name and exists for every domain.
func (g *CloudSearchGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := cloudsearch.NewFromConfig(config)
	out, err := svc.ListDomainNames(context.TODO(), &cloudsearch.ListDomainNamesInput{})
	if err != nil {
		return err
	}
	for name := range out.DomainNames {
		g.Resources = append(g.Resources, terraformutils.NewResource(
			name, name, "aws_cloudsearch_domain", "aws",
			map[string]string{"domain_name": name},
			searchAllowEmptyValues, map[string]interface{}{}))
		g.Resources = append(g.Resources, terraformutils.NewResource(
			name, name, "aws_cloudsearch_domain_service_access_policy", "aws",
			map[string]string{"domain_name": name},
			searchAllowEmptyValues, map[string]interface{}{}))
	}
	return nil
}

// --- Location Service ------------------------------------------------------

type LocationGenerator struct {
	AWSService
}

// Location resources are keyed by name (the provider importer ID):
// geofence collections, maps, place indexes, route calculators and
// trackers.
//
// Deliberately not generated: aws_location_tracker_association — the
// (tracker, consumer ARN) association is glue, discoverable only per
// tracker via ListTrackerConsumers, not a standalone resource.
func (g *LocationGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := location.NewFromConfig(config)

	geoInput := &location.ListGeofenceCollectionsInput{}
	for {
		out, err := svc.ListGeofenceCollections(context.TODO(), geoInput)
		if err != nil {
			return err
		}
		for _, entry := range out.Entries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(entry.CollectionName), StringValue(entry.CollectionName),
				"aws_location_geofence_collection", "aws", searchAllowEmptyValues))
		}
		if out.NextToken == nil {
			break
		}
		geoInput.NextToken = out.NextToken
	}
	mapInput := &location.ListMapsInput{}
	for {
		out, err := svc.ListMaps(context.TODO(), mapInput)
		if err != nil {
			return err
		}
		for _, entry := range out.Entries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(entry.MapName), StringValue(entry.MapName),
				"aws_location_map", "aws", searchAllowEmptyValues))
		}
		if out.NextToken == nil {
			break
		}
		mapInput.NextToken = out.NextToken
	}
	indexInput := &location.ListPlaceIndexesInput{}
	for {
		out, err := svc.ListPlaceIndexes(context.TODO(), indexInput)
		if err != nil {
			return err
		}
		for _, entry := range out.Entries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(entry.IndexName), StringValue(entry.IndexName),
				"aws_location_place_index", "aws", searchAllowEmptyValues))
		}
		if out.NextToken == nil {
			break
		}
		indexInput.NextToken = out.NextToken
	}
	calcInput := &location.ListRouteCalculatorsInput{}
	for {
		out, err := svc.ListRouteCalculators(context.TODO(), calcInput)
		if err != nil {
			return err
		}
		for _, entry := range out.Entries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(entry.CalculatorName), StringValue(entry.CalculatorName),
				"aws_location_route_calculator", "aws", searchAllowEmptyValues))
		}
		if out.NextToken == nil {
			break
		}
		calcInput.NextToken = out.NextToken
	}
	trackerInput := &location.ListTrackersInput{}
	for {
		out, err := svc.ListTrackers(context.TODO(), trackerInput)
		if err != nil {
			return err
		}
		for _, entry := range out.Entries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(entry.TrackerName), StringValue(entry.TrackerName),
				"aws_location_tracker", "aws", searchAllowEmptyValues))
		}
		if out.NextToken == nil {
			return nil
		}
		trackerInput.NextToken = out.NextToken
	}
}

// --- EventBridge Schemas ---------------------------------------------------

type SchemasGenerator struct {
	AWSService
}

// Schemas: discoverers by discoverer ID, registries by registry name,
// schemas by "schema_name/registry_name" (the verified provider
// importer order: name first, then registry) and a registry's resource
// policy by registry name (emitted only when GetResourcePolicy finds a
// policy for that registry).
func (g *SchemasGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := schemas.NewFromConfig(config)

	discInput := &schemas.ListDiscoverersInput{}
	for {
		out, err := svc.ListDiscoverers(context.TODO(), discInput)
		if err != nil {
			return err
		}
		for _, d := range out.Discoverers {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(d.DiscovererId), StringValue(d.DiscovererId),
				"aws_schemas_discoverer", "aws", searchAllowEmptyValues))
		}
		if out.NextToken == nil {
			break
		}
		discInput.NextToken = out.NextToken
	}

	regInput := &schemas.ListRegistriesInput{}
	for {
		out, err := svc.ListRegistries(context.TODO(), regInput)
		if err != nil {
			return err
		}
		for _, r := range out.Registries {
			registryName := StringValue(r.RegistryName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				registryName, registryName, "aws_schemas_registry", "aws", searchAllowEmptyValues))
			if err := g.loadSchemas(svc, registryName); err != nil {
				return err
			}
			policy, err := svc.GetResourcePolicy(context.TODO(), &schemas.GetResourcePolicyInput{RegistryName: aws.String(registryName)})
			if err == nil && policy.Policy != nil && *policy.Policy != "" {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					registryName, registryName, "aws_schemas_registry_policy", "aws", searchAllowEmptyValues))
			}
		}
		if out.NextToken == nil {
			return nil
		}
		regInput.NextToken = out.NextToken
	}
}

func (g *SchemasGenerator) loadSchemas(svc *schemas.Client, registryName string) error {
	input := &schemas.ListSchemasInput{RegistryName: aws.String(registryName)}
	for {
		out, err := svc.ListSchemas(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, s := range out.Schemas {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				StringValue(s.SchemaName)+"/"+registryName, StringValue(s.SchemaName),
				"aws_schemas_schema", "aws",
				map[string]string{"registry_name": registryName},
				searchAllowEmptyValues, map[string]interface{}{}))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

// --- EventBridge Pipes -----------------------------------------------------

type PipesGenerator struct {
	AWSService
}

// Pipes are keyed by pipe name (the provider importer ID).
func (g *PipesGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := pipes.NewFromConfig(config)
	input := &pipes.ListPipesInput{}
	for {
		out, err := svc.ListPipes(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, p := range out.Pipes {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(p.Name), StringValue(p.Name),
				"aws_pipes_pipe", "aws", searchAllowEmptyValues))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}
