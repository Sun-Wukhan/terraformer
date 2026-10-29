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
	"errors"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/appsync"
	"github.com/aws/aws-sdk-go-v2/service/appsync/types"
)

// Depth coverage for AppSync. Importer ID formats verified against the
// Terraform AWS provider docs.
//
// aws_appsync_api_key and aws_appsync_function are generated in
// assembly_followup.go; their importer ID formats (api_id:key_id and
// api_id-function_id) were verified against the provider docs at
// assembly time after the cluster run's doc fetches were rate-limited.

var appSyncAllowEmptyValues = []string{"tags."}

func (g *AppSyncGenerator) loadAppSyncDepth(svc *appsync.Client) error {
	if err := g.loadAppSyncEventAPIs(svc); err != nil {
		return err
	}
	if err := g.loadAppSyncGraphQLDepth(svc); err != nil {
		return err
	}
	if err := g.loadAppSyncFunctionsAndKeys(svc); err != nil {
		return err
	}
	if err := g.loadAppSyncDomains(svc); err != nil {
		return err
	}
	return nil
}

func (g *AppSyncGenerator) loadAppSyncEventAPIs(svc *appsync.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListApis(context.TODO(), &appsync.ListApisInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, api := range out.Apis {
			apiID := StringValue(api.ApiId)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				apiID,
				StringValue(api.Name),
				"aws_appsync_api",
				"aws",
				appSyncAllowEmptyValues))
			if err := g.loadChannelNamespaces(svc, apiID); err != nil {
				return err
			}
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

func (g *AppSyncGenerator) loadChannelNamespaces(svc *appsync.Client, apiID string) error {
	var nextToken *string
	for {
		out, err := svc.ListChannelNamespaces(context.TODO(), &appsync.ListChannelNamespacesInput{
			ApiId:     &apiID,
			NextToken: nextToken,
		})
		if err != nil {
			return err
		}
		for _, ns := range out.ChannelNamespaces {
			// Import ID is api_id,name.
			id := apiID + "," + StringValue(ns.Name)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				StringValue(ns.Name),
				"aws_appsync_channel_namespace",
				"aws",
				appSyncAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

func (g *AppSyncGenerator) loadAppSyncGraphQLDepth(svc *appsync.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListGraphqlApis(context.TODO(), &appsync.ListGraphqlApisInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, api := range out.GraphqlApis {
			apiID := StringValue(api.ApiId)
			if err := g.loadAppSyncAPICache(svc, apiID); err != nil {
				return err
			}
			if err := g.loadAppSyncDataSources(svc, apiID); err != nil {
				return err
			}
			if err := g.loadAppSyncTypesAndResolvers(svc, apiID); err != nil {
				return err
			}
			if err := g.loadSourceAPIassociations(svc, apiID); err != nil {
				return err
			}
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

func (g *AppSyncGenerator) loadAppSyncAPICache(svc *appsync.Client, apiID string) error {
	out, err := svc.GetApiCache(context.TODO(), &appsync.GetApiCacheInput{ApiId: &apiID})
	if err != nil {
		var notFound *types.NotFoundException
		if errors.As(err, &notFound) {
			return nil // no cache configured for this API
		}
		return err
	}
	if out.ApiCache == nil || out.ApiCache.Type == "" {
		return nil
	}
	// Import ID is the API ID.
	g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
		apiID,
		apiID,
		"aws_appsync_api_cache",
		"aws",
		appSyncAllowEmptyValues))
	return nil
}

func (g *AppSyncGenerator) loadAppSyncDataSources(svc *appsync.Client, apiID string) error {
	var nextToken *string
	for {
		out, err := svc.ListDataSources(context.TODO(), &appsync.ListDataSourcesInput{
			ApiId:     &apiID,
			NextToken: nextToken,
		})
		if err != nil {
			return err
		}
		for _, ds := range out.DataSources {
			// Import ID is api_id-name.
			id := apiID + "-" + StringValue(ds.Name)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				StringValue(ds.Name),
				"aws_appsync_datasource",
				"aws",
				appSyncAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

func (g *AppSyncGenerator) loadAppSyncTypesAndResolvers(svc *appsync.Client, apiID string) error {
	var nextToken *string
	for {
		out, err := svc.ListTypes(context.TODO(), &appsync.ListTypesInput{
			ApiId:     &apiID,
			Format:    types.TypeDefinitionFormatSdl,
			NextToken: nextToken,
		})
		if err != nil {
			return err
		}
		for _, t := range out.Types {
			typeName := StringValue(t.Name)
			// Import ID is api_id:format:name.
			id := apiID + ":" + string(t.Format) + ":" + typeName
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				typeName,
				"aws_appsync_type",
				"aws",
				appSyncAllowEmptyValues))
			if err := g.loadAppSyncResolvers(svc, apiID, typeName); err != nil {
				return err
			}
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

func (g *AppSyncGenerator) loadAppSyncResolvers(svc *appsync.Client, apiID, typeName string) error {
	var nextToken *string
	for {
		out, err := svc.ListResolvers(context.TODO(), &appsync.ListResolversInput{
			ApiId:     &apiID,
			TypeName:  &typeName,
			NextToken: nextToken,
		})
		if err != nil {
			return err
		}
		for _, resolver := range out.Resolvers {
			// Import ID is api_id-type-field.
			id := apiID + "-" + typeName + "-" + StringValue(resolver.FieldName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				StringValue(resolver.FieldName),
				"aws_appsync_resolver",
				"aws",
				appSyncAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

func (g *AppSyncGenerator) loadSourceAPIassociations(svc *appsync.Client, mergedAPIID string) error {
	var nextToken *string
	for {
		out, err := svc.ListSourceApiAssociations(context.TODO(), &appsync.ListSourceApiAssociationsInput{
			ApiId:     &mergedAPIID,
			NextToken: nextToken,
		})
		if err != nil {
			return err
		}
		for _, assoc := range out.SourceApiAssociationSummaries {
			// Import ID is merged_api_id,association_id.
			id := mergedAPIID + "," + StringValue(assoc.AssociationId)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				id,
				"aws_appsync_source_api_association",
				"aws",
				appSyncAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

func (g *AppSyncGenerator) loadAppSyncDomains(svc *appsync.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListDomainNames(context.TODO(), &appsync.ListDomainNamesInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, domain := range out.DomainNameConfigs {
			domainName := StringValue(domain.DomainName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				domainName,
				domainName,
				"aws_appsync_domain_name",
				"aws",
				appSyncAllowEmptyValues))
			assoc, err := svc.GetApiAssociation(context.TODO(), &appsync.GetApiAssociationInput{DomainName: &domainName})
			if err != nil {
				var notFound *types.NotFoundException
				if errors.As(err, &notFound) {
					continue
				}
				return err
			}
			if assoc.ApiAssociation != nil && assoc.ApiAssociation.ApiId != nil {
				// Import ID is the domain name.
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					domainName,
					domainName,
					"aws_appsync_domain_name_api_association",
					"aws",
					appSyncAllowEmptyValues))
			}
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
