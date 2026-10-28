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
	"testing"

	"google.golang.org/api/compute/v1"
	contactcenterinsightsv1 "google.golang.org/api/contactcenterinsights/v1"
	"google.golang.org/api/option"
)

// TestContactCenterInsightsGenerator feeds canned Contact Center
// Insights responses and verifies all eight types of the family: a
// view, a QA scorecard with its revision and question, an analysis
// rule, an assessment rule, an auto-labeling rule (both raw REST)
// and the encryption spec singleton.
func TestContactCenterInsightsGenerator(t *testing.T) {
	parent := "projects/test-project/locations/us-central1"
	scorecard := parent + "/qaScorecards/sc1"
	revision := scorecard + "/revisions/rev1"
	transport := &cannedTransport{routes: []cannedRoute{
		{"/locations/us-central1/views", `{"views":[{"name":"` + parent + `/views/view1"}]}`},
		{"/qaScorecards/sc1/revisions/rev1/qaQuestions", `{"qaQuestions":[{"name":"` + revision + `/qaQuestions/q1"}]}`},
		{"/qaScorecards/sc1/revisions", `{"qaScorecardRevisions":[{"name":"` + revision + `"}]}`},
		{"/locations/us-central1/qaScorecards", `{"qaScorecards":[{"name":"` + scorecard + `"}]}`},
		{"/locations/us-central1/analysisRules", `{"analysisRules":[{"name":"` + parent + `/analysisRules/ar1"}]}`},
		{"/locations/us-central1/assessmentRules", `{"assessmentRules":[{"name":"` + parent + `/assessmentRules/asr1"}]}`},
		{"/locations/us-central1/autoLabelingRules", `{"autoLabelingRules":[{"name":"` + parent + `/autoLabelingRules/alr1"}]}`},
		{"/locations/us-central1/encryptionSpec", `{"name":"` + parent + `/encryptionSpec"}`},
	}}
	ctx := context.Background()
	service, err := contactcenterinsightsv1.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("contactcenterinsights NewService: %v", err)
	}
	g := &ContactCenterInsightsGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	g.initResources(ctx, service, transport.client())
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_contact_center_insights_view":                  {parent + "/views/view1"},
		"google_contact_center_insights_qa_scorecard":          {scorecard},
		"google_contact_center_insights_qa_scorecard_revision": {revision},
		"google_contact_center_insights_qa_question":           {revision + "/qaQuestions/q1"},
		"google_contact_center_insights_analysis_rule":         {parent + "/analysisRules/ar1"},
		"google_contact_center_insights_assessment_rule":       {parent + "/assessmentRules/asr1"},
		"google_contact_center_insights_auto_labeling_rule":    {parent + "/autoLabelingRules/alr1"},
		"google_contact_center_insights_encryption_spec":       {parent + "/encryptionSpec"},
	})
}
