package grafana

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// fakeGrafana serves the search + dashboard-by-uid endpoints the
// DashboardGenerator uses: one foldered dashboard (with a folder UID) and
// one root dashboard.
func fakeGrafana(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/search", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[
			{"id": 3, "uid": "node-overview", "title": "Node Overview", "folderId": 1, "folderUid": "infra-f", "folderTitle": "Infra"},
			{"id": 4, "uid": "root-dash", "title": "Standalone Root", "folderId": 0, "folderUid": ""}
		]`)
	})
	mux.HandleFunc("/api/dashboards/uid/", func(w http.ResponseWriter, r *http.Request) {
		uid := strings.TrimPrefix(r.URL.Path, "/api/dashboards/uid/")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"dashboard": {"uid": %q, "title": "Dash %s", "schemaVersion": 39, "panels": []},
			"meta": {"slug": %q, "folderId": 1}}`, uid, uid, uid)
	})
	return httptest.NewServer(mux)
}

func TestDashboardResourcesUseFolderUIDAndInlineJSON(t *testing.T) {
	server := fakeGrafana(t)
	defer server.Close()

	g := &DashboardGenerator{}
	g.SetArgs(map[string]interface{}{
		"org_id":               1,
		"url":                  server.URL,
		"auth":                 "admin:admin",
		"tls_key":              "",
		"tls_cert":             "",
		"ca_cert":              "",
		"insecure_skip_verify": false,
	})
	if err := g.InitResources(); err != nil {
		t.Fatal(err)
	}
	if len(g.Resources) != 2 {
		t.Fatalf("got %d resources, want 2", len(g.Resources))
	}
	byUID := map[string]map[string]interface{}{}
	for _, r := range g.Resources {
		byUID[r.InstanceState.ID] = r.AdditionalFields
	}
	foldered := byUID["node-overview"]
	if foldered == nil {
		t.Fatalf("node-overview resource missing; got %v", byUID)
	}
	if got := foldered["folder"]; got != "infra-f" {
		t.Fatalf("foldered dashboard folder = %v (%T), want folder UID %q", got, got, "infra-f")
	}
	root := byUID["root-dash"]
	if root == nil {
		t.Fatalf("root-dash resource missing; got %v", byUID)
	}
	if got := root["folder"]; got != uint(0) {
		t.Fatalf("root dashboard folder = %v (%T), want 0 (general folder)", got, got)
	}
	for uid, item := range byUID {
		cj, ok := item["config_json"].(string)
		if !ok {
			t.Fatalf("%s: config_json is %T, want string", uid, item["config_json"])
		}
		if strings.HasPrefix(cj, "file(") || strings.Contains(cj, "file(\"data/") {
			t.Fatalf("%s: config_json still a file() reference: %.40q", uid, cj)
		}
		if !strings.HasPrefix(cj, "<<JSON") || !strings.Contains(cj, "\"title\"") {
			t.Fatalf("%s: config_json is not an inline JSON heredoc: %.40q", uid, cj)
		}
	}
}

func TestProviderDataOmitsAuth(t *testing.T) {
	p := GrafanaProvider{orgID: 1, url: "http://localhost:3000", auth: "admin:secret"}
	data := p.GetProviderData()
	grafanaData, ok := data["provider"].(map[string]interface{})["grafana"].(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected provider data shape: %v", data)
	}
	if _, present := grafanaData["auth"]; present {
		t.Fatal("generated provider data must not contain the auth credential")
	}
}
