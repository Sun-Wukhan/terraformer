package vault

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	vault "github.com/hashicorp/vault/api"
)

// fakeVault serves the subset of the Vault API that generic_secret
// discovery uses: sys/mounts plus LIST on kv paths. mounts maps a mount
// path to its kv version; listings maps a full list path to its keys.
func fakeVault(t *testing.T, mounts map[string]string, listings map[string][]string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/sys/mounts", func(w http.ResponseWriter, r *http.Request) {
		entries := ""
		for path, version := range mounts {
			options := "null"
			if version != "" {
				options = fmt.Sprintf(`{"version":%q}`, version)
			}
			if entries != "" {
				entries += ","
			}
			entries += fmt.Sprintf(`%q:{"type":"kv","options":%s,"description":""}`, path, options)
		}
		// Real Vault returns mounts at the top level and mirrored under
		// "data"; the api client iterates the top level.
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{%s,"data":{%s}}`, entries, entries)
	})
	mux.HandleFunc("/v1/", func(w http.ResponseWriter, r *http.Request) {
		keys, ok := listings[strings.TrimSuffix(r.URL.Path[len("/v1/"):], "/")]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"errors":[]}`)
			return
		}
		keyJSON := ""
		for i, k := range keys {
			if i > 0 {
				keyJSON += ","
			}
			keyJSON += fmt.Sprintf("%q", k)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"data":{"keys":[%s]}}`, keyJSON)
	})
	return httptest.NewServer(mux)
}

func TestGenericSecretDiscoveryKvV1AndV2(t *testing.T) {
	server := fakeVault(t,
		map[string]string{
			"apps/":   "2",
			"secret/": "2", // enabled but empty
			"legacy":  "1",
			"old":     "", // no version option: kv-v1
		},
		map[string][]string{
			"apps/metadata":      {"prod/", "dev/"},
			"apps/metadata/prod": {"config", "db"},
			"apps/metadata/dev":  {"config"},
			"legacy":             {"api", "team/"},
			"legacy/team":        {"app"},
			"old":                {"flat"},
			"secret/metadata":    {},
		})
	defer server.Close()

	client, err := vault.NewClient(&vault.Config{Address: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	g := &ServiceGenerator{client: client, resource: "generic_secret", mountType: "kv"}
	if err := g.createGenericSecretResources(); err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, r := range g.Resources {
		ids = append(ids, r.InstanceState.ID)
	}
	sort.Strings(ids)
	want := []string{
		"apps/dev/config",
		"apps/prod/config",
		"apps/prod/db",
		"legacy/api",
		"legacy/team/app",
		"old/flat",
	}
	if fmt.Sprint(ids) != fmt.Sprint(want) {
		t.Fatalf("discovered IDs:\n got %v\nwant %v", ids, want)
	}
}

func TestGenericSecretDiscoverySkipsNonKvMounts(t *testing.T) {
	server := fakeVault(t,
		map[string]string{"apps/": "2"},
		map[string][]string{"apps/metadata": {"a"}})
	defer server.Close()

	client, err := vault.NewClient(&vault.Config{Address: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	g := &ServiceGenerator{client: client, resource: "generic_secret", mountType: "kv"}
	mounts, err := g.kvMounts()
	if err != nil {
		t.Fatal(err)
	}
	if len(mounts) != 1 || mounts["apps"] != "2" {
		t.Fatalf("kvMounts = %v, want only apps=2", mounts)
	}
}
