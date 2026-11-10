package cmd

import (
	"errors"
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils/providerwrapper"
	"github.com/hashicorp/terraform/terraform"
	"github.com/zclconf/go-cty/cty"
)

// Regression tests for the nil-pointer panic introduced when the import
// report was attached to the service before InitService ran: GetService()
// returns nil until InitService succeeds, so an unsupported service name
// crashed the whole run instead of being recorded as a failure.

var reportSeenByService *terraformutils.ImportReport

type stubService struct {
	terraformutils.Service
}

func (s *stubService) InitResources() error {
	reportSeenByService = s.GetReport()
	return nil
}

// PopulateIgnoreKeys is a no-op here: the base implementation shells out to
// the provider plugin through the wrapper, which these tests do not run.
func (s *stubService) PopulateIgnoreKeys(*providerwrapper.ProviderWrapper) {}

type stubProvider struct {
	terraformutils.Provider
}

func (p *stubProvider) Init(args []string) error { return nil }

func (p *stubProvider) GetName() string { return "stub" }

func (p *stubProvider) GetConfig() cty.Value { return cty.EmptyObjectVal }

func (p *stubProvider) GetProviderData(arg ...string) map[string]interface{} { return nil }

func (p *stubProvider) GetResourceConnections() map[string]map[string][]string { return nil }

func (p *stubProvider) InitService(serviceName string, verbose bool) error {
	if serviceName != "known" {
		return errors.New("stub: " + serviceName + " not supported service")
	}
	p.Service = &stubService{}
	p.Service.SetName(serviceName)
	p.Service.SetVerbose(verbose)
	p.Service.SetProviderName(p.GetName())
	return nil
}

func runInitAll(t *testing.T, resources []string) *terraformutils.ImportReport {
	t.Helper()
	reportSeenByService = nil
	mapping := terraformutils.NewProvidersMapping(&stubProvider{})
	options := ImportOptions{Resources: resources}
	report := terraformutils.NewImportReport()
	if err := initAllServicesResources(mapping, options, nil, nil, report); err != nil {
		t.Fatalf("initAllServicesResources returned error: %v", err)
	}
	return report
}

func TestInitAllServicesResourcesUnsupportedService(t *testing.T) {
	report := runInitAll(t, []string{"bogus"})
	if report.FailureCount() != 1 {
		t.Fatalf("expected 1 recorded failure, got %d", report.FailureCount())
	}
	failure := report.Failures()[0]
	if failure.Service != "bogus" {
		t.Errorf("failure recorded for service %q, want %q", failure.Service, "bogus")
	}
	if failure.Stage != terraformutils.StageEnumerate {
		t.Errorf("failure stage = %q, want %q", failure.Stage, terraformutils.StageEnumerate)
	}
}

func TestInitAllServicesResourcesSupportedServiceGetsReport(t *testing.T) {
	report := runInitAll(t, []string{"known"})
	if report.FailureCount() != 0 {
		t.Fatalf("expected no failures, got %v", report.Failures())
	}
	if reportSeenByService != report {
		t.Errorf("service did not have the run report attached before InitResources")
	}
}

func TestDryRunSummary(t *testing.T) {
	mk := func(name, typ, id string) terraformutils.Resource {
		return terraformutils.Resource{
			ResourceName:  name,
			InstanceInfo:  &terraform.InstanceInfo{Type: typ},
			InstanceState: &terraform.InstanceState{ID: id},
		}
	}
	byService := map[string][]terraformutils.Resource{
		"mount": {
			mk("tfer--sys", "vault_mount", "sys"),
			mk("tfer--apps", "vault_mount", "apps"),
		},
		"generic_secret": {
			mk("tfer--apps_prod_config", "vault_generic_secret", "apps/prod/config"),
		},
	}
	failures := []terraformutils.ImportFailure{
		{Stage: terraformutils.StageEnumerate, Service: "policy", Error: "permission denied"},
	}
	out := dryRunSummary("vault", byService, failures)

	for _, want := range []string{
		`Dry run for provider "vault"`,
		"generic_secret (1 resource(s))",
		"mount (2 resource(s))",
		"vault_mount (2)",
		"tfer--apps (id: apps)",
		"tfer--apps_prod_config (id: apps/prod/config)",
		"Total: 3 resource(s) would be imported.",
		"1 failure(s) during discovery:",
		"service=policy",
		"permission denied",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("summary missing %q:\n%s", want, out)
		}
	}
	// Services and resources are sorted: generic_secret before mount,
	// tfer--apps before tfer--sys.
	if strings.Index(out, "generic_secret") > strings.Index(out, "mount (") {
		t.Errorf("services not sorted:\n%s", out)
	}
	if strings.Index(out, "tfer--apps (id") > strings.Index(out, "tfer--sys") {
		t.Errorf("resources not sorted:\n%s", out)
	}
	// No failures: no failure section.
	out = dryRunSummary("vault", byService, nil)
	if strings.Contains(out, "failure(s) during discovery") {
		t.Errorf("unexpected failure section:\n%s", out)
	}
}
