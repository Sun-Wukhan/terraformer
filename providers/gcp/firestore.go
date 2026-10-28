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
	firestorev1 "google.golang.org/api/firestore/v1"
)

var firestoreAllowEmptyValues = []string{""}

var firestoreAdditionalFields = map[string]interface{}{}

type FirestoreGenerator struct {
	GCPService
}

// InitResources enumerates Firestore resources for the configured
// project: databases, their backup schedules and user creds, and the
// fields and composite indexes of every collection group (listed
// through the "-" wildcard collection group, whose responses carry
// the real collection group in each resource name). Provider IDs are
// the full API resource names, the canonical import format for all of
// these (for fields and indexes the name is exactly the provider ID).
// The pinned Firestore Admin client's list calls expose no Pages
// helper, so single-page Do() calls are used.
//
// Deliberately not generated:
//   - google_firestore_document: documents are data-plane user
//     records, not infrastructure configuration — the document set of
//     a database is unbounded application data (and nested
//     subcollection documents sit under other documents), not a
//     configuration collection terraformer imports.
//   - google_firestore_change_stream: the Firestore Admin v1 API has
//     no change stream resource or collection (no changeStreams
//     list/get in the pinned generated client or its discovery
//     document), so there is nothing enumerable behind this type.
func (g *FirestoreGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := firestorev1.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, svc)
}

func (g *FirestoreGenerator) initResources(ctx context.Context, svc *firestorev1.Service) error {
	project := g.GetArgs()["project"].(string)
	baseAttrs := map[string]string{"project": project}
	page, err := svc.Projects.Databases.List("projects/" + project).Context(ctx).Do()
	if err != nil {
		log.Println(err)
		return nil
	}
	for _, database := range page.Databases {
		g.Resources = append(g.Resources, g.newFirestoreResource(database.Name, "google_firestore_database", baseAttrs))
		g.loadBackupSchedules(ctx, svc, database.Name, baseAttrs)
		g.loadUserCreds(ctx, svc, database.Name, baseAttrs)
		g.loadFieldsAndIndexes(ctx, svc, database.Name, baseAttrs)
	}
	return nil
}

func (g *FirestoreGenerator) newFirestoreResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
	if attrs == nil {
		attrs = map[string]string{}
	}
	parts := strings.Split(id, "/")
	return terraformutils.NewResource(
		id,
		parts[len(parts)-1],
		resourceType,
		g.ProviderName,
		attrs,
		firestoreAllowEmptyValues,
		firestoreAdditionalFields,
	)
}

func (g *FirestoreGenerator) loadBackupSchedules(ctx context.Context, svc *firestorev1.Service, databaseName string, baseAttrs map[string]string) {
	page, err := svc.Projects.Databases.BackupSchedules.List(databaseName).Context(ctx).Do()
	if err != nil {
		log.Println(err)
		return
	}
	for _, schedule := range page.BackupSchedules {
		g.Resources = append(g.Resources, g.newFirestoreResource(schedule.Name, "google_firestore_backup_schedule", baseAttrs))
	}
}

func (g *FirestoreGenerator) loadUserCreds(ctx context.Context, svc *firestorev1.Service, databaseName string, baseAttrs map[string]string) {
	page, err := svc.Projects.Databases.UserCreds.List(databaseName).Context(ctx).Do()
	if err != nil {
		log.Println(err)
		return
	}
	for _, creds := range page.UserCreds {
		g.Resources = append(g.Resources, g.newFirestoreResource(creds.Name, "google_firestore_user_creds", baseAttrs))
	}
}

func (g *FirestoreGenerator) loadFieldsAndIndexes(ctx context.Context, svc *firestorev1.Service, databaseName string, baseAttrs map[string]string) {
	wildcardGroup := databaseName + "/collectionGroups/-"
	page, err := svc.Projects.Databases.CollectionGroups.Fields.List(wildcardGroup).Context(ctx).Do()
	if err != nil {
		log.Println(err)
	} else {
		for _, field := range page.Fields {
			g.Resources = append(g.Resources, g.newFirestoreResource(field.Name, "google_firestore_field", baseAttrs))
		}
	}
	indexPage, err := svc.Projects.Databases.CollectionGroups.Indexes.List(wildcardGroup).Context(ctx).Do()
	if err != nil {
		log.Println(err)
		return
	}
	for _, index := range indexPage.Indexes {
		g.Resources = append(g.Resources, g.newFirestoreResource(index.Name, "google_firestore_index", baseAttrs))
	}
}
