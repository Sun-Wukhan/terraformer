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
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"google.golang.org/api/cloudfunctions/v1"
	cfv2 "google.golang.org/api/cloudfunctions/v2"
	"google.golang.org/api/compute/v1"
)

var cloudFunctionsIamAllowEmptyValues = []string{""}

var cloudFunctionsIamAdditionalFields = map[string]interface{}{}

type CloudFunctionsIamGenerator struct {
	GCPService
}

// InitResources is the IAM attachment pass for Cloud Functions: it
// lists functions with the v2 API (which returns both generations),
// and expands each function's live IAM policy into the google
// provider's _iam_policy, _iam_binding and _iam_member resources —
// google_cloudfunctions_function_iam for GEN_1 functions (parent ID
// in the gen1 provider format {project}/{location}/{name}) and
// google_cloudfunctions2_function_iam for GEN_2 functions (parent ID
// is the full resource name). The function resources themselves are
// emitted by the cloudFunctions service.
func (g *CloudFunctionsIamGenerator) InitResources() error {
	ctx := context.Background()
	v1Svc, err := cloudfunctions.NewService(ctx)
	if err != nil {
		return err
	}
	v2Svc, err := cfv2.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, v1Svc, v2Svc)
}

func (g *CloudFunctionsIamGenerator) initResources(ctx context.Context, v1Svc *cloudfunctions.Service, v2Svc *cfv2.Service) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + region
	return v2Svc.Projects.Locations.Functions.List(parent).Pages(ctx, func(page *cfv2.ListFunctionsResponse) error {
		for _, fn := range page.Functions {
			parts := strings.Split(fn.Name, "/")
			name := parts[len(parts)-1]
			if fn.Environment == "GEN_1" {
				policy, err := v1Svc.Projects.Locations.Functions.GetIamPolicy(fn.Name).Context(ctx).Do()
				if err != nil {
					log.Println(err)
					continue
				}
				bindings := []iamBinding{}
				for _, b := range policy.Bindings {
					bindings = append(bindings, iamBinding{Role: b.Role, Members: b.Members})
				}
				parentID := project + "/" + region + "/" + name
				g.Resources = append(g.Resources, expandIamPolicy(bindings, parentID,
					"google_cloudfunctions_function_iam_policy",
					"google_cloudfunctions_function_iam_binding",
					"google_cloudfunctions_function_iam_member",
					map[string]string{}, g.newResource)...)
			} else {
				policy, err := v2Svc.Projects.Locations.Functions.GetIamPolicy(fn.Name).Context(ctx).Do()
				if err != nil {
					log.Println(err)
					continue
				}
				bindings := []iamBinding{}
				for _, b := range policy.Bindings {
					bindings = append(bindings, iamBinding{Role: b.Role, Members: b.Members})
				}
				g.Resources = append(g.Resources, expandIamPolicy(bindings, fn.Name,
					"google_cloudfunctions2_function_iam_policy",
					"google_cloudfunctions2_function_iam_binding",
					"google_cloudfunctions2_function_iam_member",
					map[string]string{}, g.newResource)...)
			}
		}
		return nil
	})
}

func (g *CloudFunctionsIamGenerator) newResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
	parts := strings.Split(id, "/")
	if attrs == nil {
		attrs = map[string]string{}
	}
	return terraformutils.NewResource(id, parts[len(parts)-1], resourceType, g.ProviderName, attrs,
		cloudFunctionsIamAllowEmptyValues, cloudFunctionsIamAdditionalFields)
}
