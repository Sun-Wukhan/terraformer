package terraformutils

import (
	"errors"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform/terraform"
)

var errTestFailure = errors.New("enumeration failed")

func TestEmptyFiltersParsing(t *testing.T) {
	service := Service{}
	service.ParseFilters([]string{})

	if !reflect.DeepEqual(service.Filter, []ResourceFilter{}) {
		t.Errorf("failed to parse, got %v", service.Filter)
	}
}

func TestIdFiltersParsing(t *testing.T) {
	service := Service{}
	service.ParseFilters([]string{"aws_vpc=myid"})

	if !reflect.DeepEqual(service.Filter, []ResourceFilter{
		{
			ServiceName:      "aws_vpc",
			FieldPath:        "id",
			AcceptableValues: []string{"myid"},
		}}) {
		t.Errorf("failed to parse, got %v", service.Filter)
	}
}

func TestComplexIdFiltersParsing(t *testing.T) {
	service := Service{}
	service.ParseFilters([]string{"resource=id1:'project:dataset_id'"})

	if !reflect.DeepEqual(service.Filter, []ResourceFilter{
		{
			ServiceName:      "resource",
			FieldPath:        "id",
			AcceptableValues: []string{"id1", "project:dataset_id"},
		}}) {
		t.Errorf("failed to parse, got %v", service.Filter)
	}
}

func TestEdgeIdFiltersParsing(t *testing.T) {
	service := Service{}
	service.ParseFilters([]string{"aws_vpc=:myid"})

	if !reflect.DeepEqual(service.Filter, []ResourceFilter{
		{
			ServiceName:      "aws_vpc",
			FieldPath:        "id",
			AcceptableValues: []string{"myid"},
		}}) {
		t.Errorf("failed to parse, got %v", service.Filter)
	}
}

func TestServiceIdCleanupWithFilter(t *testing.T) {
	service := Service{
		Resources: []Resource{{
			InstanceInfo: &terraform.InstanceInfo{
				Type: "type1",
			},
			InstanceState: &terraform.InstanceState{
				ID: "myid",
			}}, {
			InstanceInfo: &terraform.InstanceInfo{
				Type: "type2",
			},
			InstanceState: &terraform.InstanceState{
				ID: "myid",
			}}},
	}
	service.ParseFilters([]string{"type1=:otherId"})
	service.InitialCleanup()

	if !reflect.DeepEqual(len(service.Resources), 1) {
		t.Errorf("failed to cleanup")
	}
}

func TestServiceAttributeCleanupWithFilter(t *testing.T) {
	service := Service{
		Resources: []Resource{
			{
				InstanceInfo: &terraform.InstanceInfo{
					Type: "aws_vpc",
				},
				InstanceState: &terraform.InstanceState{
					ID: "vpc1",
				},
				Item: mapI("tags", mapI("Name", "some"))},
			{
				InstanceInfo: &terraform.InstanceInfo{
					Type: "aws_vpc",
				},
				InstanceState: &terraform.InstanceState{
					ID: "vpc2",
				},
				Item: mapI("tags", mapI("Name", "default"))}},
	}
	service.ParseFilters([]string{"Name=tags.Name;Value=default"})
	service.PostRefreshCleanup()

	if !reflect.DeepEqual(len(service.Resources), 1) {
		t.Errorf("failed to cleanup")
	}
}

func TestServiceAttributeNameOnlyCleanupWithFilter(t *testing.T) {
	service := Service{
		Resources: []Resource{
			{
				InstanceInfo: &terraform.InstanceInfo{
					Type: "aws_vpc",
				},
				InstanceState: &terraform.InstanceState{
					ID: "vpc1",
				},
				Item: mapI("tags", mapI("Abc", nil))},
			{
				InstanceInfo: &terraform.InstanceInfo{
					Type: "aws_vpc",
				},
				InstanceState: &terraform.InstanceState{
					ID: "vpc2",
				},
				Item: mapI("tags", mapI("Name", "default"))}},
	}
	service.ParseFilters([]string{"Name=tags.Abc"})
	service.PostRefreshCleanup()

	if !reflect.DeepEqual(len(service.Resources), 1) {
		t.Errorf("failed to cleanup")
	}
}

func TestServiceRecordFailure(t *testing.T) {
	// No report attached: must be a safe no-op (generators in tests and
	// other non-import callers never have one).
	bare := Service{Name: "iam"}
	bare.RecordFailure("aws_iam_user", errTestFailure)
	if bare.GetReport() != nil {
		t.Errorf("expected no report to be attached, got %v", bare.GetReport())
	}

	report := NewImportReport()
	service := Service{Name: "iam"}
	service.SetReport(report)
	if service.GetReport() != report {
		t.Fatalf("GetReport did not return the report given to SetReport")
	}

	// A nil error records nothing.
	service.RecordFailure("aws_iam_user", nil)
	if report.FailureCount() != 0 {
		t.Fatalf("nil error recorded a failure: %v", report.Failures())
	}

	service.RecordFailure("aws_iam_user", errTestFailure)
	failures := report.Failures()
	if len(failures) != 1 {
		t.Fatalf("expected 1 recorded failure, got %d", len(failures))
	}
	f := failures[0]
	if f.Stage != StageEnumerate || f.Service != "iam" || f.ResourceType != "aws_iam_user" || f.Error != errTestFailure.Error() {
		t.Errorf("unexpected recorded failure: %+v", f)
	}
}
