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
	"net/http"
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"google.golang.org/api/compute/v1"
	contactcenterinsightsv1 "google.golang.org/api/contactcenterinsights/v1"
)

// ContactCenterInsightsGenerator enumerates the Contact Center
// Insights surface of one project/location: views, QA scorecards
// with their revisions and questions, analysis rules, assessment
// rules, auto-labeling rules and the per-location encryption spec
// singleton. Assessment rules, auto-labeling rules and the
// encryption spec have no list/get surface in the pinned generated
// client, so they run over raw REST against
// contactcenterinsights.googleapis.com/v1. Every provider ID is the
// resource's full API name.
type ContactCenterInsightsGenerator struct {
	GCPService
}

const contactCenterInsightsBaseURL = "https://contactcenterinsights.googleapis.com/v1/"

func (g *ContactCenterInsightsGenerator) InitResources() error {
	ctx := context.Background()
	service, err := contactcenterinsightsv1.NewService(ctx)
	if err != nil {
		return err
	}
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		log.Println(err)
		return nil
	}
	g.initResources(ctx, service, restClient)
	return nil
}

func (g *ContactCenterInsightsGenerator) cciLocation() string {
	region := g.GetArgs()["region"].(compute.Region).Name
	if region == "" {
		return "us-central1"
	}
	return region
}

func (g *ContactCenterInsightsGenerator) initResources(ctx context.Context, service *contactcenterinsightsv1.Service, restClient *http.Client) {
	project := g.GetArgs()["project"].(string)
	parent := "projects/" + project + "/locations/" + g.cciLocation()

	views, err := cxListNames(func(pageToken string) ([]string, string, error) {
		resp, err := service.Projects.Locations.Views.List(parent).PageToken(pageToken).Context(ctx).Do()
		if err != nil {
			return nil, "", err
		}
		names := []string{}
		for _, view := range resp.Views {
			names = append(names, view.Name)
		}
		return names, resp.NextPageToken, nil
	})
	g.appendCciList(views, err, "google_contact_center_insights_view")

	scorecards, err := cxListNames(func(pageToken string) ([]string, string, error) {
		resp, err := service.Projects.Locations.QaScorecards.List(parent).PageToken(pageToken).Context(ctx).Do()
		if err != nil {
			return nil, "", err
		}
		names := []string{}
		for _, scorecard := range resp.QaScorecards {
			names = append(names, scorecard.Name)
		}
		return names, resp.NextPageToken, nil
	})
	if err != nil {
		log.Println(err)
	} else {
		for _, scorecard := range scorecards {
			g.Resources = append(g.Resources, g.newCciResource(scorecard, "google_contact_center_insights_qa_scorecard"))
			g.cciScorecardChildren(ctx, service, scorecard)
		}
	}

	analysisRules, err := cxListNames(func(pageToken string) ([]string, string, error) {
		resp, err := service.Projects.Locations.AnalysisRules.List(parent).PageToken(pageToken).Context(ctx).Do()
		if err != nil {
			return nil, "", err
		}
		names := []string{}
		for _, rule := range resp.AnalysisRules {
			names = append(names, rule.Name)
		}
		return names, resp.NextPageToken, nil
	})
	g.appendCciList(analysisRules, err, "google_contact_center_insights_analysis_rule")

	for segment, resourceType := range map[string]string{
		"assessmentRules":   "google_contact_center_insights_assessment_rule",
		"autoLabelingRules": "google_contact_center_insights_auto_labeling_rule",
	} {
		rules, err := restListResources(ctx, restClient, contactCenterInsightsBaseURL+parent+"/"+segment)
		if err != nil {
			log.Println(err)
			continue
		}
		for _, rule := range rules {
			g.Resources = append(g.Resources, g.newCciResource(rule.Name, resourceType))
		}
	}

	// The encryption spec is a per-location singleton; it only
	// exists once CMEK has been initialized for the location.
	specName := parent + "/encryptionSpec"
	if name, err := aiRestSingletonName(ctx, restClient, contactCenterInsightsBaseURL+specName); err != nil {
		log.Println(err)
	} else if name != "" {
		g.Resources = append(g.Resources, g.newCciResource(name, "google_contact_center_insights_encryption_spec"))
	}
}

func (g *ContactCenterInsightsGenerator) cciScorecardChildren(ctx context.Context, service *contactcenterinsightsv1.Service, scorecard string) {
	revisions, err := cxListNames(func(pageToken string) ([]string, string, error) {
		resp, err := service.Projects.Locations.QaScorecards.Revisions.List(scorecard).PageToken(pageToken).Context(ctx).Do()
		if err != nil {
			return nil, "", err
		}
		names := []string{}
		for _, revision := range resp.QaScorecardRevisions {
			names = append(names, revision.Name)
		}
		return names, resp.NextPageToken, nil
	})
	if err != nil {
		log.Println(err)
		return
	}
	for _, revision := range revisions {
		g.Resources = append(g.Resources, g.newCciResource(revision, "google_contact_center_insights_qa_scorecard_revision"))
		questions, err := cxListNames(func(pageToken string) ([]string, string, error) {
			resp, err := service.Projects.Locations.QaScorecards.Revisions.QaQuestions.List(revision).PageToken(pageToken).Context(ctx).Do()
			if err != nil {
				return nil, "", err
			}
			names := []string{}
			for _, question := range resp.QaQuestions {
				names = append(names, question.Name)
			}
			return names, resp.NextPageToken, nil
		})
		g.appendCciList(questions, err, "google_contact_center_insights_qa_question")
	}
}

func (g *ContactCenterInsightsGenerator) newCciResource(id, resourceType string) terraformutils.Resource {
	parts := strings.Split(id, "/")
	return terraformutils.NewResource(
		id,
		parts[len(parts)-1],
		resourceType,
		g.ProviderName,
		map[string]string{
			"project":  g.GetArgs()["project"].(string),
			"location": g.cciLocation(),
		},
		[]string{},
		map[string]interface{}{},
	)
}

func (g *ContactCenterInsightsGenerator) appendCciList(names []string, err error, resourceType string) {
	if err != nil {
		log.Println(err)
		return
	}
	for _, name := range names {
		g.Resources = append(g.Resources, g.newCciResource(name, resourceType))
	}
}
