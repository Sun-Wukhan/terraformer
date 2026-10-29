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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicequotas"
)

var serviceQuotasAllowEmptyValues = []string{"tags."}

type ServiceQuotasGenerator struct {
	AWSService
}

// InitResources enumerates Service Quotas for the configured region:
// every service (ListServices) and, per service, its quotas
// (ListServiceQuotas). The provider importer ID of
// aws_servicequotas_service_quota is service_code/quota_code.
//
// Intentionally not generated:
//   - aws_servicequotas_auto_management is the singleton
//     account-level auto-management configuration
//     (GetAutoManagementConfiguration).
//   - aws_servicequotas_template is the singleton per-region quota
//     request template (GetServiceQuotaTemplate).
//   - aws_servicequotas_template_association is the singleton
//     association state of that template
//     (GetAssociationForServiceQuotaTemplate).
func (g *ServiceQuotasGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := servicequotas.NewFromConfig(config)

	p := servicequotas.NewListServicesPaginator(svc, &servicequotas.ListServicesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, service := range page.Services {
			g.addQuotasForService(svc, StringValue(service.ServiceCode))
		}
	}
	return nil
}

func (g *ServiceQuotasGenerator) addQuotasForService(svc *servicequotas.Client, serviceCode string) {
	p := servicequotas.NewListServiceQuotasPaginator(svc, &servicequotas.ListServiceQuotasInput{
		ServiceCode: aws.String(serviceCode),
	})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, quota := range page.Quotas {
			quotaCode := StringValue(quota.QuotaCode)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				serviceCode+"/"+quotaCode,
				StringValue(quota.QuotaName),
				"aws_servicequotas_service_quota",
				"aws",
				map[string]string{"service_code": serviceCode},
				serviceQuotasAllowEmptyValues,
				map[string]interface{}{}))
		}
	}
}
