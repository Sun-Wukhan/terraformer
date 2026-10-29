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
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cleanrooms"
	"github.com/aws/aws-sdk-go-v2/service/comprehend"
	"github.com/aws/aws-sdk-go-v2/service/evidently"
	"github.com/aws/aws-sdk-go-v2/service/kendra"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	"github.com/aws/aws-sdk-go-v2/service/transcribe"
)

var aiAllowEmptyValues = []string{"tags."}

// --- Comprehend ------------------------------------------------------------

type ComprehendGenerator struct {
	AWSService
}

// Comprehend document classifiers and entity recognizers are keyed by
// ARN (the provider importer ID); the list APIs return full property
// records including the ARN.
func (g *ComprehendGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := comprehend.NewFromConfig(config)
	classInput := &comprehend.ListDocumentClassifiersInput{}
	for {
		out, err := svc.ListDocumentClassifiers(context.TODO(), classInput)
		if err != nil {
			return err
		}
		for _, c := range out.DocumentClassifierPropertiesList {
			arn := StringValue(c.DocumentClassifierArn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_comprehend_document_classifier", "aws", aiAllowEmptyValues))
		}
		if out.NextToken == nil {
			break
		}
		classInput.NextToken = out.NextToken
	}
	recInput := &comprehend.ListEntityRecognizersInput{}
	for {
		out, err := svc.ListEntityRecognizers(context.TODO(), recInput)
		if err != nil {
			return err
		}
		for _, r := range out.EntityRecognizerPropertiesList {
			arn := StringValue(r.EntityRecognizerArn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_comprehend_entity_recognizer", "aws", aiAllowEmptyValues))
		}
		if out.NextToken == nil {
			return nil
		}
		recInput.NextToken = out.NextToken
	}
}

// --- Kendra ----------------------------------------------------------------

type KendraGenerator struct {
	AWSService
}

// Kendra indexes are keyed by index ID. The per-index resources
// (data sources, experiences, FAQs, query suggestions block lists and
// thesauri) are keyed by "<resource_id>/<index_id>" — the FAQ form was
// verified directly against the provider docs
// ("faq-123456780/idx-8012925589") and the sibling resources share the
// same two-part importer shape.
func (g *KendraGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := kendra.NewFromConfig(config)
	indexInput := &kendra.ListIndicesInput{}
	for {
		out, err := svc.ListIndices(context.TODO(), indexInput)
		if err != nil {
			return err
		}
		for _, index := range out.IndexConfigurationSummaryItems {
			indexID := StringValue(index.Id)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				indexID, StringValue(index.Name), "aws_kendra_index", "aws", aiAllowEmptyValues))
			if err := g.loadIndexChildren(svc, indexID); err != nil {
				return err
			}
		}
		if out.NextToken == nil {
			return nil
		}
		indexInput.NextToken = out.NextToken
	}
}

func (g *KendraGenerator) addKendraChild(id, indexID, name, resourceType string) {
	g.Resources = append(g.Resources, terraformutils.NewResource(
		id+"/"+indexID, name, resourceType, "aws",
		map[string]string{"index_id": indexID},
		aiAllowEmptyValues, map[string]interface{}{}))
}

func (g *KendraGenerator) loadIndexChildren(svc *kendra.Client, indexID string) error {
	dsInput := &kendra.ListDataSourcesInput{IndexId: aws.String(indexID)}
	for {
		out, err := svc.ListDataSources(context.TODO(), dsInput)
		if err != nil {
			return err
		}
		for _, ds := range out.SummaryItems {
			g.addKendraChild(StringValue(ds.Id), indexID, StringValue(ds.Name), "aws_kendra_data_source")
		}
		if out.NextToken == nil {
			break
		}
		dsInput.NextToken = out.NextToken
	}
	expInput := &kendra.ListExperiencesInput{IndexId: aws.String(indexID)}
	for {
		out, err := svc.ListExperiences(context.TODO(), expInput)
		if err != nil {
			return err
		}
		for _, exp := range out.SummaryItems {
			g.addKendraChild(StringValue(exp.Id), indexID, StringValue(exp.Name), "aws_kendra_experience")
		}
		if out.NextToken == nil {
			break
		}
		expInput.NextToken = out.NextToken
	}
	faqInput := &kendra.ListFaqsInput{IndexId: aws.String(indexID)}
	for {
		out, err := svc.ListFaqs(context.TODO(), faqInput)
		if err != nil {
			return err
		}
		for _, faq := range out.FaqSummaryItems {
			g.addKendraChild(StringValue(faq.Id), indexID, StringValue(faq.Name), "aws_kendra_faq")
		}
		if out.NextToken == nil {
			break
		}
		faqInput.NextToken = out.NextToken
	}
	blockInput := &kendra.ListQuerySuggestionsBlockListsInput{IndexId: aws.String(indexID)}
	for {
		out, err := svc.ListQuerySuggestionsBlockLists(context.TODO(), blockInput)
		if err != nil {
			return err
		}
		for _, b := range out.BlockListSummaryItems {
			g.addKendraChild(StringValue(b.Id), indexID, StringValue(b.Name), "aws_kendra_query_suggestions_block_list")
		}
		if out.NextToken == nil {
			break
		}
		blockInput.NextToken = out.NextToken
	}
	thesInput := &kendra.ListThesauriInput{IndexId: aws.String(indexID)}
	for {
		out, err := svc.ListThesauri(context.TODO(), thesInput)
		if err != nil {
			return err
		}
		for _, t := range out.ThesaurusSummaryItems {
			g.addKendraChild(StringValue(t.Id), indexID, StringValue(t.Name), "aws_kendra_thesaurus")
		}
		if out.NextToken == nil {
			break
		}
		thesInput.NextToken = out.NextToken
	}
	return nil
}

// --- Rekognition -----------------------------------------------------------

type RekognitionGenerator struct {
	AWSService
}

// Rekognition collections are keyed by collection ID and projects by
// ARN; stream processors are keyed by name (the ListStreamProcessors
// summary carries no ARN).
func (g *RekognitionGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := rekognition.NewFromConfig(config)
	collInput := &rekognition.ListCollectionsInput{}
	for {
		out, err := svc.ListCollections(context.TODO(), collInput)
		if err != nil {
			return err
		}
		for _, id := range out.CollectionIds {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_rekognition_collection", "aws", aiAllowEmptyValues))
		}
		if out.NextToken == nil {
			break
		}
		collInput.NextToken = out.NextToken
	}
	projInput := &rekognition.DescribeProjectsInput{}
	for {
		out, err := svc.DescribeProjects(context.TODO(), projInput)
		if err != nil {
			return err
		}
		for _, p := range out.ProjectDescriptions {
			arn := StringValue(p.ProjectArn)
			name := arn
			if parts := strings.Split(arn, "/"); len(parts) > 1 {
				name = parts[1]
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, name, "aws_rekognition_project", "aws", aiAllowEmptyValues))
		}
		if out.NextToken == nil {
			break
		}
		projInput.NextToken = out.NextToken
	}
	spInput := &rekognition.ListStreamProcessorsInput{}
	for {
		out, err := svc.ListStreamProcessors(context.TODO(), spInput)
		if err != nil {
			return err
		}
		for _, sp := range out.StreamProcessors {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(sp.Name), StringValue(sp.Name), "aws_rekognition_stream_processor", "aws", aiAllowEmptyValues))
		}
		if out.NextToken == nil {
			return nil
		}
		spInput.NextToken = out.NextToken
	}
}

// --- Transcribe ------------------------------------------------------------

type TranscribeGenerator struct {
	AWSService
}

// Transcribe custom language models, vocabularies, medical
// vocabularies and vocabulary filters are all keyed by name (the
// provider importer ID).
func (g *TranscribeGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := transcribe.NewFromConfig(config)
	lmInput := &transcribe.ListLanguageModelsInput{}
	for {
		out, err := svc.ListLanguageModels(context.TODO(), lmInput)
		if err != nil {
			return err
		}
		for _, m := range out.Models {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(m.ModelName), StringValue(m.ModelName), "aws_transcribe_language_model", "aws", aiAllowEmptyValues))
		}
		if out.NextToken == nil {
			break
		}
		lmInput.NextToken = out.NextToken
	}
	vocabInput := &transcribe.ListVocabulariesInput{}
	for {
		out, err := svc.ListVocabularies(context.TODO(), vocabInput)
		if err != nil {
			return err
		}
		for _, v := range out.Vocabularies {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(v.VocabularyName), StringValue(v.VocabularyName), "aws_transcribe_vocabulary", "aws", aiAllowEmptyValues))
		}
		if out.NextToken == nil {
			break
		}
		vocabInput.NextToken = out.NextToken
	}
	medInput := &transcribe.ListMedicalVocabulariesInput{}
	for {
		out, err := svc.ListMedicalVocabularies(context.TODO(), medInput)
		if err != nil {
			return err
		}
		for _, v := range out.Vocabularies {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(v.VocabularyName), StringValue(v.VocabularyName), "aws_transcribe_medical_vocabulary", "aws", aiAllowEmptyValues))
		}
		if out.NextToken == nil {
			break
		}
		medInput.NextToken = out.NextToken
	}
	filterInput := &transcribe.ListVocabularyFiltersInput{}
	for {
		out, err := svc.ListVocabularyFilters(context.TODO(), filterInput)
		if err != nil {
			return err
		}
		for _, v := range out.VocabularyFilters {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(v.VocabularyFilterName), StringValue(v.VocabularyFilterName), "aws_transcribe_vocabulary_filter", "aws", aiAllowEmptyValues))
		}
		if out.NextToken == nil {
			return nil
		}
		filterInput.NextToken = out.NextToken
	}
}

// --- Clean Rooms -----------------------------------------------------------

type CleanRoomsGenerator struct {
	AWSService
}

// Clean Rooms collaborations, configured tables and memberships are
// keyed by their IDs (the provider importer ID).
func (g *CleanRoomsGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := cleanrooms.NewFromConfig(config)
	collabInput := &cleanrooms.ListCollaborationsInput{}
	for {
		out, err := svc.ListCollaborations(context.TODO(), collabInput)
		if err != nil {
			return err
		}
		for _, c := range out.CollaborationList {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(c.Id), StringValue(c.Name), "aws_cleanrooms_collaboration", "aws", aiAllowEmptyValues))
		}
		if out.NextToken == nil {
			break
		}
		collabInput.NextToken = out.NextToken
	}
	tableInput := &cleanrooms.ListConfiguredTablesInput{}
	for {
		out, err := svc.ListConfiguredTables(context.TODO(), tableInput)
		if err != nil {
			return err
		}
		for _, t := range out.ConfiguredTableSummaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(t.Id), StringValue(t.Name), "aws_cleanrooms_configured_table", "aws", aiAllowEmptyValues))
		}
		if out.NextToken == nil {
			break
		}
		tableInput.NextToken = out.NextToken
	}
	memberInput := &cleanrooms.ListMembershipsInput{}
	for {
		out, err := svc.ListMemberships(context.TODO(), memberInput)
		if err != nil {
			return err
		}
		for _, m := range out.MembershipSummaries {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				StringValue(m.Id), StringValue(m.CollaborationName), "aws_cleanrooms_membership", "aws",
				map[string]string{"collaboration_id": StringValue(m.CollaborationId)},
				aiAllowEmptyValues, map[string]interface{}{}))
		}
		if out.NextToken == nil {
			return nil
		}
		memberInput.NextToken = out.NextToken
	}
}

// --- Evidently -------------------------------------------------------------

type EvidentlyGenerator struct {
	AWSService
}

// Evidently projects are keyed by project name and segments by ARN
// (both verified against the provider docs); features and launches
// are per-project and use the verified colon-separated importer IDs
// "feature_name:project_name" and "launch_name:project_name".
func (g *EvidentlyGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := evidently.NewFromConfig(config)
	projInput := &evidently.ListProjectsInput{}
	for {
		out, err := svc.ListProjects(context.TODO(), projInput)
		if err != nil {
			return err
		}
		for _, p := range out.Projects {
			project := StringValue(p.Name)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				project, project, "aws_evidently_project", "aws", aiAllowEmptyValues))
			if err := g.loadProjectChildren(svc, project); err != nil {
				return err
			}
		}
		if out.NextToken == nil {
			break
		}
		projInput.NextToken = out.NextToken
	}
	segInput := &evidently.ListSegmentsInput{}
	for {
		out, err := svc.ListSegments(context.TODO(), segInput)
		if err != nil {
			return err
		}
		for _, s := range out.Segments {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(s.Arn), StringValue(s.Name), "aws_evidently_segment", "aws", aiAllowEmptyValues))
		}
		if out.NextToken == nil {
			return nil
		}
		segInput.NextToken = out.NextToken
	}
}

func (g *EvidentlyGenerator) loadProjectChildren(svc *evidently.Client, project string) error {
	featInput := &evidently.ListFeaturesInput{Project: aws.String(project)}
	for {
		out, err := svc.ListFeatures(context.TODO(), featInput)
		if err != nil {
			return err
		}
		for _, f := range out.Features {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				StringValue(f.Name)+":"+project, StringValue(f.Name), "aws_evidently_feature", "aws",
				map[string]string{"project": project},
				aiAllowEmptyValues, map[string]interface{}{}))
		}
		if out.NextToken == nil {
			break
		}
		featInput.NextToken = out.NextToken
	}
	launchInput := &evidently.ListLaunchesInput{Project: aws.String(project)}
	for {
		out, err := svc.ListLaunches(context.TODO(), launchInput)
		if err != nil {
			return err
		}
		for _, l := range out.Launches {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				StringValue(l.Name)+":"+project, StringValue(l.Name), "aws_evidently_launch", "aws",
				map[string]string{"project": project},
				aiAllowEmptyValues, map[string]interface{}{}))
		}
		if out.NextToken == nil {
			return nil
		}
		launchInput.NextToken = out.NextToken
	}
}
