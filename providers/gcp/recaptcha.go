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
	recaptchav1 "google.golang.org/api/recaptchaenterprise/v1"
)

var recaptchaAllowEmptyValues = []string{"tags."}

var recaptchaAdditionalFields = map[string]interface{}{}

type RecaptchaEnterpriseGenerator struct {
	GCPService
}

// InitResources enumerates the project's reCAPTCHA Enterprise keys.
func (g *RecaptchaEnterpriseGenerator) InitResources() error {
	ctx := context.Background()
	service, err := recaptchav1.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, service)
}

func (g *RecaptchaEnterpriseGenerator) initResources(ctx context.Context, service *recaptchav1.Service) error {
	project := g.GetArgs()["project"].(string)

	err := service.Projects.Keys.List("projects/"+project).Pages(ctx, func(page *recaptchav1.GoogleCloudRecaptchaenterpriseV1ListKeysResponse) error {
		for _, key := range page.Keys {
			parts := strings.Split(key.Name, "/")
			g.Resources = append(g.Resources, terraformutils.NewResource(
				key.Name,
				parts[len(parts)-1],
				"google_recaptcha_enterprise_key",
				g.ProviderName,
				map[string]string{"project": project},
				recaptchaAllowEmptyValues,
				recaptchaAdditionalFields,
			))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	return nil
}
