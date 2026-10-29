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

// Redshift Serverless coverage (AWS sweep, database cluster).
//
// Skipped in this family, with reasons:
//   - aws_redshiftserverless_resource_policy: policies attach to individual
//     resources (namespaces, snapshots); the SDK exposes only per-resource
//     GetResourcePolicy with no list/enumeration API.

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshiftserverless"
)

var redshiftServerlessAllowEmptyValues = []string{"tags."}

type RedshiftServerlessGenerator struct {
	AWSService
}

func (g *RedshiftServerlessGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := redshiftserverless.NewFromConfig(config)

	// aws_redshiftserverless_namespace, ID = namespace name.
	np := redshiftserverless.NewListNamespacesPaginator(svc, &redshiftserverless.ListNamespacesInput{})
	for np.HasMorePages() {
		page, err := np.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, ns := range page.Namespaces {
			name := StringValue(ns.NamespaceName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_redshiftserverless_namespace", "aws", redshiftServerlessAllowEmptyValues))
		}
	}

	// aws_redshiftserverless_workgroup, ID = workgroup name; plus per-workgroup
	// endpoint access.
	wp := redshiftserverless.NewListWorkgroupsPaginator(svc, &redshiftserverless.ListWorkgroupsInput{})
	for wp.HasMorePages() {
		page, err := wp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, wg := range page.Workgroups {
			wgName := StringValue(wg.WorkgroupName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				wgName, wgName, "aws_redshiftserverless_workgroup", "aws", redshiftServerlessAllowEmptyValues))

			// aws_redshiftserverless_endpoint_access, ID = endpoint name.
			ep := redshiftserverless.NewListEndpointAccessPaginator(svc, &redshiftserverless.ListEndpointAccessInput{WorkgroupName: aws.String(wgName)})
			for ep.HasMorePages() {
				epage, err := ep.NextPage(context.TODO())
				if err != nil {
					return err
				}
				for _, endpoint := range epage.Endpoints {
					name := StringValue(endpoint.EndpointName)
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						name, name, "aws_redshiftserverless_endpoint_access", "aws", redshiftServerlessAllowEmptyValues))
				}
			}

		}
	}

	// aws_redshiftserverless_custom_domain_association,
	// ID = "<workgroup name>,<custom domain name>". The list API takes no
	// workgroup filter; each Association carries its workgroup name.
	dap := redshiftserverless.NewListCustomDomainAssociationsPaginator(svc, &redshiftserverless.ListCustomDomainAssociationsInput{})
	for dap.HasMorePages() {
		page, err := dap.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, assoc := range page.Associations {
			wgName := StringValue(assoc.WorkgroupName)
			domain := StringValue(assoc.CustomDomainName)
			id := wgName + "," + domain
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, wgName+"_"+domain, "aws_redshiftserverless_custom_domain_association", "aws", redshiftServerlessAllowEmptyValues))
		}
	}

	// aws_redshiftserverless_snapshot, ID = snapshot name.
	sp := redshiftserverless.NewListSnapshotsPaginator(svc, &redshiftserverless.ListSnapshotsInput{})
	for sp.HasMorePages() {
		page, err := sp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, snap := range page.Snapshots {
			name := StringValue(snap.SnapshotName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_redshiftserverless_snapshot", "aws", redshiftServerlessAllowEmptyValues))
		}
	}

	// aws_redshiftserverless_usage_limit, ID = usage limit ID.
	up := redshiftserverless.NewListUsageLimitsPaginator(svc, &redshiftserverless.ListUsageLimitsInput{})
	for up.HasMorePages() {
		page, err := up.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, limit := range page.UsageLimits {
			id := StringValue(limit.UsageLimitId)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_redshiftserverless_usage_limit", "aws", redshiftServerlessAllowEmptyValues))
		}
	}

	return nil
}
