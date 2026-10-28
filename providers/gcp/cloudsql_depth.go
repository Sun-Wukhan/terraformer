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
	"log"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
)

// This file extends CloudSQLGenerator (cloudsql.go) with the per
// instance depth collections: users and SSL certificates. Source
// representation instances are handled inside loadDBInstances in
// cloudsql.go: instances.list returns them alongside managed
// instances, distinguished by their onPremisesConfiguration, and they
// are emitted as google_sql_source_representation_instance instead of
// google_sql_database_instance (the provider's database instance
// resource cannot manage an external source representation).
//
// Deliberately not generated:
//   - google_sql_provision_script: a provision script is not a Cloud
//     SQL API object — the provider executes the script text against
//     the instance through the Data API at apply time and stores
//     nothing server-side, so there is no list/get surface (the
//     sqladmin API has no provision script resource) to enumerate.
//
// Provider ID formats (per the google provider import docs):
// users are {{project}}/{{instance}}/{{name}}/{{host}}, SSL certs are
// {{project}}/{{instance}}/{{common_name}}, and a source
// representation instance carries its instance name, matching how
// cloudsql.go emits google_sql_database_instance.
func (g *CloudSQLGenerator) loadDepth(svc *sqladmin.Service, project string) {
	instances, err := svc.Instances.List(project).Do()
	if err != nil {
		log.Println(err)
		return
	}
	for _, instance := range instances.Items {
		if instance.OnPremisesConfiguration != nil {
			continue
		}
		g.loadUsers(svc, project, instance.Name)
		g.loadSslCerts(svc, project, instance.Name)
	}
}

func (g *CloudSQLGenerator) loadUsers(svc *sqladmin.Service, project, instanceName string) {
	users, err := svc.Users.List(project, instanceName).Do()
	if err != nil {
		log.Println(err)
		return
	}
	for _, user := range users.Items {
		id := project + "/" + instanceName + "/" + user.Name + "/" + user.Host
		g.Resources = append(g.Resources, terraformutils.NewResource(
			id,
			instanceName+"-"+user.Name,
			"google_sql_user",
			g.ProviderName,
			map[string]string{
				"project":  project,
				"instance": instanceName,
				"name":     user.Name,
				"host":     user.Host,
			},
			cloudSQLAllowEmptyValues,
			cloudSQLAdditionalFields,
		))
	}
}

func (g *CloudSQLGenerator) loadSslCerts(svc *sqladmin.Service, project, instanceName string) {
	certs, err := svc.SslCerts.List(project, instanceName).Do()
	if err != nil {
		log.Println(err)
		return
	}
	for _, cert := range certs.Items {
		id := project + "/" + instanceName + "/" + cert.CommonName
		g.Resources = append(g.Resources, terraformutils.NewResource(
			id,
			instanceName+"-"+cert.CommonName,
			"google_sql_ssl_cert",
			g.ProviderName,
			map[string]string{
				"project":     project,
				"instance":    instanceName,
				"common_name": cert.CommonName,
			},
			cloudSQLAllowEmptyValues,
			cloudSQLAdditionalFields,
		))
	}
}
