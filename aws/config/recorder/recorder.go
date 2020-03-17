// Package recorder provides the ability to filter *configservice.ConfigurationRecorder objects
package recorder

import (
	"testing"

	"github.com/GSA/grace-tftest/aws/shared"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/configservice"
)

// Recorder contains the necessary properties for testing *configservice.ConfigurationRecorder objects
type Recorder struct {
	client   client.ConfigProvider
	recorder *configservice.ConfigurationRecorder
	filters  []shared.Filter
}

// New returns a new *Recorder
func New(client client.ConfigProvider) *Recorder {
	return &Recorder{client: client}
}

// Selected returns the currently selected *configservice.ConfigurationRecorder
func (r *Recorder) Selected() *configservice.ConfigurationRecorder {
	return r.recorder
}

// Recording returns the RecorderStatus.Recording value of the selected recorder
func (r *Recorder) Recording(t *testing.T,
	statuses ...*configservice.ConfigurationRecorderStatus) bool {
	if r.recorder == nil {
		t.Fatal("failed to get recorder status, Selected is nil")
		return false
	}

	if statuses == nil {
		svc := configservice.New(r.client)
		out, err := svc.DescribeConfigurationRecorderStatus(
			&configservice.DescribeConfigurationRecorderStatusInput{
				ConfigurationRecorderNames: []*string{r.recorder.Name},
			},
		)
		if err != nil {
			t.Fatalf("failed get recorder status for recorder: %s -> %v",
				aws.StringValue(r.recorder.Name),
				err,
			)
			return false
		}
		statuses = out.ConfigurationRecordersStatus
	}
	for _, s := range statuses {
		if aws.StringValue(s.Name) == aws.StringValue(r.recorder.Name) {
			return aws.BoolValue(s.Recording)
		}
	}
	t.Fatalf("failed to locate status for %s", aws.StringValue(r.recorder.Name))
	return false
}

// Assert applies all filters that have been called, resets the list of filters,
// fails the test if there is not exactly one match, and stores the matched recorder
// if recorders is not provided, *configservice.ConfigurationRecorder objects will be retreived from AWS
func (r *Recorder) Assert(t *testing.T, recorders ...*configservice.ConfigurationRecorder) *Recorder {
	var err error
	recorders, err = r.filter(recorders)
	if err != nil {
		t.Error(err)
	}

	switch l := len(recorders); {
	case l == 0:
		t.Error("no matching recorder was found")
	case l > 1:
		t.Error("more than one matching recorder was found")
	default:
		r.recorder = recorders[0]
	}

	r.filters = []shared.Filter{}
	return r
}

// First applies all filters that have been called, resets the list of filters,
// fails the test if there are no matches, and stores the first match
// if recorders is not provided, *configservice.ConfigurationRecorder objects will be retreived from AWS
func (r *Recorder) First(t *testing.T, recorders ...*configservice.ConfigurationRecorder) *Recorder {
	var err error
	recorders, err = r.filter(recorders)
	if err != nil {
		t.Error(err)
	}

	if len(recorders) == 0 {
		t.Error("no matching recorder was found")
	} else {
		r.recorder = recorders[0]
	}

	r.filters = []shared.Filter{}
	return r
}

// RoleArn adds the RoleArn filter to the filter list
// the RoleArn filter: filters recorders by RoleArn where 'arn' provided
// is the expected RoleARN value
func (r *Recorder) RoleArn(arn string) *Recorder {
	r.filters = append(r.filters, func(v interface{}) bool {
		recorder := convert(v)
		if recorder == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", arn, aws.StringValue(recorder.RoleARN), arn == aws.StringValue(recorder.RoleARN))
		return arn == aws.StringValue(recorder.RoleARN)
	})
	return r
}

// Filter adds the 'filter' provided to the filter list
func (r *Recorder) Filter(filter shared.Filter) *Recorder {
	r.filters = append(r.filters, filter)
	return r
}

// AllSupported adds the AllSupported filter to the filter list
// the AllSupported filter: filters recorders by AllSupported where
// 'enabled' provided is the expected AllSupported value
func (r *Recorder) AllSupported(enabled bool) *Recorder {
	r.filters = append(r.filters, func(v interface{}) bool {
		recorder := convert(v)
		if recorder == nil {
			return false
		}
		shared.Debugf("%t == %t -> %t\n",
			enabled,
			aws.BoolValue(recorder.RecordingGroup.AllSupported),
			enabled == aws.BoolValue(recorder.RecordingGroup.AllSupported),
		)
		return enabled == aws.BoolValue(recorder.RecordingGroup.AllSupported)
	})
	return r
}

// IncludeGlobalResourceTypes adds the IncludeGlobalResourceTypes
// filter to the filter list the IncludeGlobalResourceTypes filter:
// filters recorders by IncludeGlobalResourceTypes where 'enabled'
// provided is the expected IncludeGlobalResourceTypes value
func (r *Recorder) IncludeGlobalResourceTypes(enabled bool) *Recorder {
	r.filters = append(r.filters, func(v interface{}) bool {
		recorder := convert(v)
		if recorder == nil {
			return false
		}
		shared.Debugf("%t == %t -> %t\n",
			enabled,
			aws.BoolValue(recorder.RecordingGroup.IncludeGlobalResourceTypes),
			enabled == aws.BoolValue(recorder.RecordingGroup.IncludeGlobalResourceTypes),
		)
		return enabled == aws.BoolValue(recorder.RecordingGroup.IncludeGlobalResourceTypes)
	})
	return r
}

// ResourceTypes adds the ResourceTypes filter to the filter
// list the ResourceTypes filter: filters recorders by
// ResourceTypes where 'types' provided is the expected
// ResourceTypes value slice
func (r *Recorder) ResourceTypes(types ...string) *Recorder {
	r.filters = append(r.filters, func(v interface{}) bool {
		recorder := convert(v)
		if recorder == nil {
			return false
		}
		return shared.StringSliceEqual(
			types,
			aws.StringValueSlice(recorder.RecordingGroup.ResourceTypes),
		)
	})
	return r
}

// Name adds the Name filter to the filter list the Name
// filter: filters recorders by Name where 'name' provided
// is the expected PolicyName value
func (r *Recorder) Name(name string) *Recorder {
	r.filters = append(r.filters, func(v interface{}) bool {
		recorder := convert(v)
		if recorder == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n",
			name,
			aws.StringValue(recorder.Name),
			name == aws.StringValue(recorder.Name),
		)
		return name == aws.StringValue(recorder.Name)
	})
	return r
}

func (r *Recorder) filter(recorders []*configservice.ConfigurationRecorder) (
	[]*configservice.ConfigurationRecorder, error) {
	if len(recorders) == 0 {
		var err error
		recorders, err = r.recorders()
		if err != nil {
			return nil, err
		}
	}
	return fromIface(shared.GenericFilter(r.filters, toIface(recorders))), nil
}

func (r *Recorder) recorders() ([]*configservice.ConfigurationRecorder, error) {
	svc := configservice.New(r.client)
	out, err := svc.DescribeConfigurationRecorders(
		&configservice.DescribeConfigurationRecordersInput{},
	)
	if err != nil {
		return nil, err
	}
	return out.ConfigurationRecorders, nil
}

func convert(in interface{}) *configservice.ConfigurationRecorder {
	out, ok := in.(*configservice.ConfigurationRecorder)
	if !ok {
		shared.Debugf("object not convertible to *configservice.ConfigurationRecorder: ")
		shared.Dump(in)
		return nil
	}
	return out
}
func toIface(in []*configservice.ConfigurationRecorder) (out []interface{}) {
	for _, i := range in {
		out = append(out, i)
	}
	return
}
func fromIface(in []interface{}) (out []*configservice.ConfigurationRecorder) {
	for _, i := range in {
		v := convert(i)
		if v == nil {
			continue
		}
		out = append(out, v)
	}
	return
}
