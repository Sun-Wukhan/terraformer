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

	"google.golang.org/api/option"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
)

// TestCloudSQLDepthGenerator feeds canned sqladmin responses and
// verifies users and SSL certs are emitted with the provider's
// composite IDs, and that an on-premises instance from instances.list
// is emitted as a source representation instance, not a database
// instance.
func TestCloudSQLDepthGenerator(t *testing.T) {
	transport := &cannedTransport{routes: []cannedRoute{
		{"/instances/inst1/users", `{"items":[{"name":"root","host":"%"}]}`},
		{"/instances/inst1/sslCerts", `{"items":[{"commonName":"cert1"}]}`},
		{"/instances/inst1/databases", `{"items":[{"name":"db1"}]}`},
		{"/projects/test-project/instances", `{"items":[{"name":"inst1"},{"name":"src1","onPremisesConfiguration":{"hostPort":"10.0.0.1:3306"}}]}`},
	}}
	ctx := context.Background()
	svc, err := sqladmin.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("sqladmin.NewService: %v", err)
	}
	g := &CloudSQLGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project"})
	if err := g.loadDBInstances(svc, "test-project"); err != nil {
		t.Fatalf("loadDBInstances: %v", err)
	}
	g.loadDepth(svc, "test-project")
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_sql_database_instance":              {"inst1"},
		"google_sql_database":                       {"inst1:db1"},
		"google_sql_source_representation_instance": {"src1"},
		"google_sql_user":                           {"test-project/inst1/root/%"},
		"google_sql_ssl_cert":                       {"test-project/inst1/cert1"},
	})
}
