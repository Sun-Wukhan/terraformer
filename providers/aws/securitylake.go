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
	"log"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/securitylake"
	sltypes "github.com/aws/aws-sdk-go-v2/service/securitylake/types"
)

var securityLakeAllowEmptyValues = []string{"tags."}

type SecurityLakeGenerator struct {
	AWSService
}

// InitResources enumerates Security Lake data lakes (ListDataLakes;
// the provider importer ID of aws_securitylake_data_lake is the
// region), subscribers (ListSubscribers; provider ID = subscriber
// ID) and custom log sources (the distinct custom source names seen
// in ListLogSources; provider ID = the custom log source name).
//
// Intentionally not generated:
//   - aws_securitylake_aws_log_source is per-account/region
//     enablement configuration for a named AWS log source — the same
//     ListLogSources entries are source configurations, not
//     independent objects with their own IDs.
//   - aws_securitylake_subscriber_notification: the SDK exposes no
//     separate get/list for a subscriber's notification
//     configuration; it is part of the subscriber object itself.
func (g *SecurityLakeGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := securitylake.NewFromConfig(config)
	g.addDataLakes(svc)
	g.addSubscribers(svc)
	g.addCustomLogSources(svc)
	return nil
}

func (g *SecurityLakeGenerator) addDataLakes(svc *securitylake.Client) {
	out, err := svc.ListDataLakes(context.TODO(), &securitylake.ListDataLakesInput{})
	if err != nil {
		log.Println(err)
		return
	}
	for _, lake := range out.DataLakes {
		region := StringValue(lake.Region)
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			region,
			region,
			"aws_securitylake_data_lake",
			"aws",
			securityLakeAllowEmptyValues))
	}
}

func (g *SecurityLakeGenerator) addSubscribers(svc *securitylake.Client) {
	p := securitylake.NewListSubscribersPaginator(svc, &securitylake.ListSubscribersInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, subscriber := range page.Subscribers {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(subscriber.SubscriberId),
				StringValue(subscriber.SubscriberName),
				"aws_securitylake_subscriber",
				"aws",
				securityLakeAllowEmptyValues))
		}
	}
}

func (g *SecurityLakeGenerator) addCustomLogSources(svc *securitylake.Client) {
	out, err := svc.ListLogSources(context.TODO(), &securitylake.ListLogSourcesInput{})
	if err != nil {
		log.Println(err)
		return
	}
	seen := map[string]bool{}
	for _, logSource := range out.Sources {
		for _, source := range logSource.Sources {
			custom, ok := source.(*sltypes.LogSourceResourceMemberCustomLogSource)
			if !ok || custom.Value.SourceName == nil {
				continue
			}
			name := *custom.Value.SourceName
			if seen[name] {
				continue
			}
			seen[name] = true
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name,
				name,
				"aws_securitylake_custom_log_source",
				"aws",
				securityLakeAllowEmptyValues))
		}
	}
}
