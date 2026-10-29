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
	"github.com/aws/aws-sdk-go-v2/service/accessanalyzer"
)

// Access Analyzer depth pass (AWS full-coverage sweep, security
// cluster): aws_accessanalyzer_archive_rule for every analyzer,
// enumerated with ListArchiveRules. Provider importer ID format
// (verified): analyzer_name/rule_name.
func (g *AccessAnalyzerGenerator) addArchiveRules(svc *accessanalyzer.Client) {
	p := accessanalyzer.NewListAnalyzersPaginator(svc, &accessanalyzer.ListAnalyzersInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, analyzer := range page.Analyzers {
			analyzerName := StringValue(analyzer.Name)
			rp := accessanalyzer.NewListArchiveRulesPaginator(svc, &accessanalyzer.ListArchiveRulesInput{
				AnalyzerName: analyzer.Name,
			})
			for rp.HasMorePages() {
				rules, err := rp.NextPage(context.TODO())
				if err != nil {
					log.Println(err)
					break
				}
				for _, rule := range rules.ArchiveRules {
					ruleName := StringValue(rule.RuleName)
					g.Resources = append(g.Resources, terraformutils.NewResource(
						analyzerName+"/"+ruleName,
						analyzerName+"_"+ruleName,
						"aws_accessanalyzer_archive_rule",
						"aws",
						map[string]string{"analyzer_name": analyzerName},
						accessanalyzerAllowEmptyValues,
						map[string]interface{}{}))
				}
			}
		}
	}
}
