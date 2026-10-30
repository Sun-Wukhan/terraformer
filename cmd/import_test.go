package cmd

import (
	"errors"
	"testing"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils/providerwrapper"
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
