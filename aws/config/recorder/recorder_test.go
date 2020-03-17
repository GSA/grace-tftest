package recorder

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
)

func TestRecorder(t *testing.T) {
	recorders := []*configservice.ConfigurationRecorder{
		{
			Name:    aws.String("a"),
			RoleARN: aws.String("b"),
			RecordingGroup: &configservice.RecordingGroup{
				AllSupported:               aws.Bool(true),
				IncludeGlobalResourceTypes: aws.Bool(true),
				ResourceTypes: aws.StringSlice([]string{
					"c",
					"d",
					"e",
				}),
			},
		},
		{
			Name:    aws.String("b"),
			RoleARN: aws.String("c"),
			RecordingGroup: &configservice.RecordingGroup{
				AllSupported:               aws.Bool(true),
				IncludeGlobalResourceTypes: aws.Bool(true),
				ResourceTypes: aws.StringSlice([]string{
					"f",
					"g",
					"h",
				}),
			},
		},
	}
	statuses := []*configservice.ConfigurationRecorderStatus{
		{Name: aws.String("a"), Recording: aws.Bool(true)},
		{Name: aws.String("b"), Recording: aws.Bool(false)},
	}

	svc := New(nil)
	r := svc.ResourceTypes("f", "h", "g").Assert(t, recorders...).Selected()
	if aws.StringValue(r.Name) != "b" {
		t.Fatalf("failed to match resource name, expected: %s, got: %s", "b", aws.StringValue(r.Name))
	}
	if svc.Recording(t, statuses...) {
		t.Fatalf("failed to match recording status, expected: %t, got: %t", false, true)
	}
}
