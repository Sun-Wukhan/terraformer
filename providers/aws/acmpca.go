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
	"github.com/aws/aws-sdk-go-v2/service/acmpca"
)

var acmpcaAllowEmptyValues = []string{"tags."}

type AcmpcaGenerator struct {
	AWSService
}

// InitResources enumerates ACM Private CA certificate authorities in
// the configured region (ListCertificateAuthorities) and, per CA, its
// permissions (ListPermissions) and its resource policy (GetPolicy,
// only when one is set).
//
// Provider importer ID formats (terraform-provider-aws docs):
//   - aws_acmpca_certificate_authority: the CA ARN.
//   - aws_acmpca_permission: certificate_authority_arn:principal.
//   - aws_acmpca_policy: the CA ARN.
//
// Intentionally not generated:
//   - aws_acmpca_certificate: issued certificates cannot be
//     enumerated — the SDK has no list API for them (GetCertificate
//     requires the CA ARN and the certificate serial number), and an
//     issuance is a derived artifact of an issuing call.
//   - aws_acmpca_certificate_authority_certificate is the
//     singleton-per-CA installation of the CA's own certificate
//     chain; that certificate is an attribute of the certificate
//     authority itself (GetCertificateAuthorityCertificate), not an
//     independent object.
func (g *AcmpcaGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := acmpca.NewFromConfig(config)

	p := acmpca.NewListCertificateAuthoritiesPaginator(svc, &acmpca.ListCertificateAuthoritiesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, ca := range page.CertificateAuthorities {
			arn := StringValue(ca.Arn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn,
				arnResourceName(arn),
				"aws_acmpca_certificate_authority",
				"aws",
				acmpcaAllowEmptyValues))
			g.addPermissions(svc, arn)
			g.addPolicy(svc, arn)
		}
	}
	return nil
}

func (g *AcmpcaGenerator) addPermissions(svc *acmpca.Client, caArn string) {
	p := acmpca.NewListPermissionsPaginator(svc, &acmpca.ListPermissionsInput{
		CertificateAuthorityArn: &caArn,
	})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, permission := range page.Permissions {
			principal := StringValue(permission.Principal)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				caArn+":"+principal,
				principal,
				"aws_acmpca_permission",
				"aws",
				map[string]string{"certificate_authority_arn": caArn},
				acmpcaAllowEmptyValues,
				map[string]interface{}{}))
		}
	}
}

func (g *AcmpcaGenerator) addPolicy(svc *acmpca.Client, caArn string) {
	out, err := svc.GetPolicy(context.TODO(), &acmpca.GetPolicyInput{
		ResourceArn: &caArn,
	})
	if err != nil || out.Policy == nil || *out.Policy == "" {
		return // no resource policy set on this CA
	}
	g.Resources = append(g.Resources, terraformutils.NewResource(
		caArn,
		arnResourceName(caArn),
		"aws_acmpca_policy",
		"aws",
		map[string]string{"certificate_authority_arn": caArn},
		acmpcaAllowEmptyValues,
		map[string]interface{}{}))
}
