package grafana

import (
	"encoding/json"
	"fmt"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	gapi "github.com/grafana/grafana-api-golang-client"
)

type DashboardGenerator struct {
	GrafanaService
}

func (g *DashboardGenerator) InitResources() error {
	client, err := g.buildClient()
	if err != nil {
		return fmt.Errorf("unable to build grafana client: %v", err)
	}

	err = g.createDashboardResources(client)
	if err != nil {
		return err
	}

	return nil
}

func (g *DashboardGenerator) createDashboardResources(client *gapi.Client) error {
	dashboards, err := client.Dashboards()
	if err != nil {
		return fmt.Errorf("unable to list grafana dashboards: %v", err)
	}

	for _, dashboard := range dashboards {
		// search result doesn't include slug, so need to look up dashboard.
		dash, err := client.DashboardByUID(dashboard.UID)
		if err != nil {
			return fmt.Errorf("unable to read grafana dashboard %s: %v", dashboard.Title, err)
		}

		configJSON, err := json.MarshalIndent(dash.Model, "", "  ")
		if err != nil {
			return fmt.Errorf("unable to marshal configuration for grafana dashboard %s: %v", dashboard.Title, err)
		}

		// The folder attribute must carry the folder UID: that is what
		// the grafana provider stores in state. The numeric FolderID the
		// search API also returns plans as a spurious update against
		// every foldered dashboard (and would target the wrong folder
		// on apply). Root dashboards keep 0 (the general folder).
		folderRef := interface{}(dashboard.FolderID)
		if dashboard.FolderUID != "" {
			folderRef = dashboard.FolderUID
		}
		resource := terraformutils.NewResource(
			dashboard.UID,
			dashboard.Title,
			"grafana_dashboard",
			"grafana",
			map[string]string{},
			[]string{},
			map[string]interface{}{
				// Inline the dashboard JSON as a heredoc. Storing a
				// file("data/...") expression here emitted it as a
				// quoted string, so Terraform passed the literal text
				// file("...") to the provider as the config JSON and
				// every plan failed with a JSON parse error.
				"config_json": fmt.Sprintf("<<JSON\n%s\nJSON", configJSON),
				"folder":      folderRef,
			},
		)
		g.Resources = append(g.Resources, resource)
	}

	return nil
}
