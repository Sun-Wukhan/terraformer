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
	"github.com/aws/aws-sdk-go-v2/service/mwaa"
)

var mwaaAllowEmptyValues = []string{"tags."}

// MwaaGenerator covers Amazon MWAA. aws_mwaa_environment imports with the
// environment name (provider docs).
type MwaaGenerator struct {
	AWSService
}

func (g *MwaaGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := mwaa.NewFromConfig(config)

	var nextToken *string
	for {
		out, err := svc.ListEnvironments(context.TODO(), &mwaa.ListEnvironmentsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, name := range out.Environments {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name,
				name,
				"aws_mwaa_environment",
				"aws",
				mwaaAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
