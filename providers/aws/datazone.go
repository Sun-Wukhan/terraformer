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

// Amazon DataZone coverage (AWS sweep, database cluster). Importer ID
// formats differ per resource and were verified against the provider docs:
//   - domain: "<domain ID>"
//   - project: "<domain ID>:<project ID>"
//   - environment: "<domain ID>,<environment ID>"
//   - environment_profile: "<profile ID>,<domain ID>"
//   - environment_blueprint_configuration: "<domain ID>/<blueprint ID>"
//   - asset_type / form_type: "<domain ID>,<name>" (enumerated via SearchTypes)
//   - glossary: "<domain ID>,<glossary ID>,<owning project ID>" (via Search)
//   - glossary_term: "<domain ID>,<term ID>,<glossary ID>" (via Search)
//   - user_profile: "<user identifier>,<domain ID>,<IAM|SSO>"
//
// Skipped in this family, with reasons:
//   - aws_datazone_policy_grant: grants are enumerable only per entity via
//     ListPolicyGrants (entity type + entity ID required) and there is no
//     domain-wide grant listing; grant association glue.

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datazone"
	datazonetypes "github.com/aws/aws-sdk-go-v2/service/datazone/types"
)

var dataZoneAllowEmptyValues = []string{"tags."}

type DataZoneGenerator struct {
	AWSService
}

func (g *DataZoneGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := datazone.NewFromConfig(config)

	dp := datazone.NewListDomainsPaginator(svc, &datazone.ListDomainsInput{Status: datazonetypes.DomainStatusAvailable})
	for dp.HasMorePages() {
		page, err := dp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, domain := range page.Items {
			domainID := StringValue(domain.Id)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				domainID, StringValue(domain.Name), "aws_datazone_domain", "aws", dataZoneAllowEmptyValues))
			if err := g.loadDataZoneDomain(svc, domainID); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *DataZoneGenerator) loadDataZoneDomain(svc *datazone.Client, domainID string) error {
	// aws_datazone_environment_blueprint_configuration,
	// ID = "<domain ID>/<blueprint ID>".
	bp := datazone.NewListEnvironmentBlueprintConfigurationsPaginator(svc, &datazone.ListEnvironmentBlueprintConfigurationsInput{DomainIdentifier: aws.String(domainID)})
	for bp.HasMorePages() {
		page, err := bp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, item := range page.Items {
			blueprintID := StringValue(item.EnvironmentBlueprintId)
			id := domainID + "/" + blueprintID
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, blueprintID, "aws_datazone_environment_blueprint_configuration", "aws", dataZoneAllowEmptyValues))
		}
	}

	// aws_datazone_environment_profile, ID = "<profile ID>,<domain ID>".
	prp := datazone.NewListEnvironmentProfilesPaginator(svc, &datazone.ListEnvironmentProfilesInput{DomainIdentifier: aws.String(domainID)})
	for prp.HasMorePages() {
		page, err := prp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, profile := range page.Items {
			profileID := StringValue(profile.Id)
			id := profileID + "," + domainID
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(profile.Name), "aws_datazone_environment_profile", "aws", dataZoneAllowEmptyValues))
		}
	}

	// Projects, plus environments per project.
	pp := datazone.NewListProjectsPaginator(svc, &datazone.ListProjectsInput{DomainIdentifier: aws.String(domainID)})
	for pp.HasMorePages() {
		page, err := pp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, project := range page.Items {
			projectID := StringValue(project.Id)
			// aws_datazone_project, ID = "<domain ID>:<project ID>".
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				domainID+":"+projectID, StringValue(project.Name), "aws_datazone_project", "aws", dataZoneAllowEmptyValues))

			// aws_datazone_environment, ID = "<domain ID>,<environment ID>".
			ep := datazone.NewListEnvironmentsPaginator(svc, &datazone.ListEnvironmentsInput{
				DomainIdentifier:  aws.String(domainID),
				ProjectIdentifier: aws.String(projectID),
			})
			for ep.HasMorePages() {
				epage, err := ep.NextPage(context.TODO())
				if err != nil {
					return err
				}
				for _, env := range epage.Items {
					envID := StringValue(env.Id)
					id := domainID + "," + envID
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						id, StringValue(env.Name), "aws_datazone_environment", "aws", dataZoneAllowEmptyValues))
				}
			}
		}
	}

	// aws_datazone_asset_type and aws_datazone_form_type via SearchTypes,
	// ID = "<domain ID>,<name>".
	for _, scope := range []struct {
		scope        datazonetypes.TypesSearchScope
		resourceType string
	}{
		{datazonetypes.TypesSearchScopeAssetType, "aws_datazone_asset_type"},
		{datazonetypes.TypesSearchScopeFormType, "aws_datazone_form_type"},
	} {
		tp := datazone.NewSearchTypesPaginator(svc, &datazone.SearchTypesInput{
			DomainIdentifier: aws.String(domainID),
			SearchScope:      scope.scope,
			Managed:          aws.Bool(false), // custom (non-managed) types only
		})
		for tp.HasMorePages() {
			page, err := tp.NextPage(context.TODO())
			if err != nil {
				return err
			}
			for _, item := range page.Items {
				var name string
				switch v := item.(type) {
				case *datazonetypes.SearchTypesResultItemMemberAssetTypeItem:
					name = StringValue(v.Value.Name)
				case *datazonetypes.SearchTypesResultItemMemberFormTypeItem:
					name = StringValue(v.Value.Name)
				}
				if name == "" {
					continue
				}
				id := domainID + "," + name
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, name, scope.resourceType, "aws", dataZoneAllowEmptyValues))
			}
		}
	}

	// Glossaries and glossary terms via Search.
	gp := datazone.NewSearchPaginator(svc, &datazone.SearchInput{
		DomainIdentifier: aws.String(domainID),
		SearchScope:      datazonetypes.InventorySearchScopeGlossary,
	})
	for gp.HasMorePages() {
		page, err := gp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, item := range page.Items {
			member, ok := item.(*datazonetypes.SearchInventoryResultItemMemberGlossaryItem)
			if !ok {
				continue
			}
			glossary := member.Value
			// aws_datazone_glossary,
			// ID = "<domain ID>,<glossary ID>,<owning project ID>".
			id := domainID + "," + StringValue(glossary.Id) + "," + StringValue(glossary.OwningProjectId)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(glossary.Name), "aws_datazone_glossary", "aws", dataZoneAllowEmptyValues))
		}
	}

	tp2 := datazone.NewSearchPaginator(svc, &datazone.SearchInput{
		DomainIdentifier: aws.String(domainID),
		SearchScope:      datazonetypes.InventorySearchScopeGlossaryTerm,
	})
	for tp2.HasMorePages() {
		page, err := tp2.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, item := range page.Items {
			member, ok := item.(*datazonetypes.SearchInventoryResultItemMemberGlossaryTermItem)
			if !ok {
				continue
			}
			term := member.Value
			// aws_datazone_glossary_term,
			// ID = "<domain ID>,<term ID>,<glossary ID>".
			id := domainID + "," + StringValue(term.Id) + "," + StringValue(term.GlossaryId)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(term.Name), "aws_datazone_glossary_term", "aws", dataZoneAllowEmptyValues))
		}
	}

	// aws_datazone_user_profile via SearchUserProfiles,
	// ID = "<user identifier>,<domain ID>,<IAM|SSO>".
	{
		up := datazone.NewSearchUserProfilesPaginator(svc, &datazone.SearchUserProfilesInput{
			DomainIdentifier: aws.String(domainID),
			UserType:         datazonetypes.UserSearchTypeDatazoneIamUser,
		})
		for up.HasMorePages() {
			page, err := up.NextPage(context.TODO())
			if err != nil {
				return err
			}
			for _, profile := range page.Items {
				userIdentifier, typeToken := dataZoneUserProfileIDParts(profile)
				if userIdentifier == "" {
					continue
				}
				id := userIdentifier + "," + domainID + "," + typeToken
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, userIdentifier, "aws_datazone_user_profile", "aws", dataZoneAllowEmptyValues))
			}
		}
	}

	return nil
}

func dataZoneUserProfileIDParts(profile datazonetypes.UserProfileSummary) (string, string) {
	// Only IAM profiles can be enumerated into importer IDs: their user
	// identifier is the IAM ARN, which SearchUserProfiles returns. SSO
	// profiles are skipped because the search summary carries only the
	// user's name/username, not the Identity Center user ID the importer
	// ID requires.
	if iam, ok := profile.Details.(*datazonetypes.UserProfileDetailsMemberIam); ok {
		return StringValue(iam.Value.Arn), "IAM"
	}
	return "", ""
}
