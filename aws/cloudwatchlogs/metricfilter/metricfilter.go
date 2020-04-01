// Package metricfilter contains the necessary properties for testing *cloudwatchlogs.MetricFilter objects
package metricfilter

import (
	"testing"

	"github.com/GSA/grace-tftest/aws/shared"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

// MetricFilter contains the necessary properties for testing *cloudwatchlogs.MetricFilter objects
type MetricFilter struct {
	client     client.ConfigProvider
	selected   *cloudwatchlogs.MetricFilter
	filterList []shared.Filter
}

// New returns a new *MetricFilter
func New(client client.ConfigProvider) *MetricFilter {
	return &MetricFilter{client: client}
}

// Selected returns the currently selected *cloudwatchlogs.MetricFilter
func (m *MetricFilter) Selected() *cloudwatchlogs.MetricFilter {
	return m.selected
}

// Assert applies all filters that have been called, resets the list of filters,
// fails the test if there is not exactly one match, and stores the matched filter
// if filters is not provided, *cloudwatchlogs.MetricFilter objects will be retreived from AWS
func (m *MetricFilter) Assert(t *testing.T, filters ...*cloudwatchlogs.MetricFilter) *MetricFilter {
	var err error
	filters, err = m.filter(filters)
	if err != nil {
		t.Error(err)
	}

	switch l := len(filters); {
	case l == 0:
		t.Error("no matching filter was found")
	case l > 1:
		t.Error("more than one matching filter was found")
	default:
		m.selected = filters[0]
	}

	m.filterList = []shared.Filter{}
	return m
}

// First applies all filters that have been called, resets the list of filters,
// fails the test if there are no matches, and stores the first match
// if filters is not provided, *cloudwatchlogs.MetricFilter objects will be retreived from AWS
func (m *MetricFilter) First(t *testing.T, filters ...*cloudwatchlogs.MetricFilter) *MetricFilter {
	var err error
	filters, err = m.filter(filters)
	if err != nil {
		t.Error(err)
	}

	if len(filters) == 0 {
		t.Error("no matching filter was found")
	} else {
		m.selected = filters[0]
	}

	m.filterList = []shared.Filter{}
	return m
}

// Filter adds the 'filter' provided to the filter list
func (m *MetricFilter) Filter(filter shared.Filter) *MetricFilter {
	m.filterList = append(m.filterList, filter)
	return m
}

// Name is an alias to FilterName filter which filters by
// the FilterName field
func (m *MetricFilter) Name(name string) *MetricFilter {
	return m.FilterName(name)
}

// FilterName adds the FilterName filter to the filter list
// the FilterName filter: filters filters by FilterName where 'name' provided
// is the expected FilterName value
func (m *MetricFilter) FilterName(name string) *MetricFilter {
	m.filterList = append(m.filterList, func(v interface{}) bool {
		filter := convert(v)
		if filter == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n",
			name,
			aws.StringValue(filter.FilterName),
			name == aws.StringValue(filter.FilterName),
		)
		return name == aws.StringValue(filter.FilterName)
	})
	return m
}

// LogGroupName adds the LogGroupName filter to the filter list
// the LogGroupName filter: filters filters by LogGroupName where 'name' provided
// is the expected LogGroupName value
func (m *MetricFilter) LogGroupName(name string) *MetricFilter {
	m.filterList = append(m.filterList, func(v interface{}) bool {
		filter := convert(v)
		if filter == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n",
			name,
			aws.StringValue(filter.LogGroupName),
			name == aws.StringValue(filter.LogGroupName),
		)
		return name == aws.StringValue(filter.LogGroupName)
	})
	return m
}

// TName is an alias to TransformationName
func (m *MetricFilter) TName(name string) *MetricFilter {
	return m.TransformationName(name)
}

// TransformationName adds the TransformationName filter to the filter list
// the TransformationName filter: filters filters by TransformationName where 'name' provided
// is the expected MetricTransformations[0].MetricName value
func (m *MetricFilter) TransformationName(name string) *MetricFilter {
	m.filterList = append(m.filterList, func(v interface{}) bool {
		filter := convert(v)
		if filter == nil ||
			len(filter.MetricTransformations) == 0 {
			return false
		}
		shared.Debugf("%s == %s -> %t\n",
			name,
			aws.StringValue(filter.MetricTransformations[0].MetricName),
			name == aws.StringValue(filter.MetricTransformations[0].MetricName),
		)
		return name == aws.StringValue(filter.MetricTransformations[0].MetricName)
	})
	return m
}

// TNamespace is an alias to TransformationNamespace
func (m *MetricFilter) TNamespace(namespace string) *MetricFilter {
	return m.TransformationNamespace(namespace)
}

// TransformationNamespace adds the TransformationNamespace filter to the filter list
// the TransformationNamespace filter: filters filters by TransformationNamespace where
// 'namespace' provided is the expected MetricTransformations[0].MetricNamespace value
func (m *MetricFilter) TransformationNamespace(namespace string) *MetricFilter {
	m.filterList = append(m.filterList, func(v interface{}) bool {
		filter := convert(v)
		if filter == nil ||
			len(filter.MetricTransformations) == 0 {
			return false
		}
		shared.Debugf("%s == %s -> %t\n",
			namespace,
			aws.StringValue(filter.MetricTransformations[0].MetricNamespace),
			namespace == aws.StringValue(filter.MetricTransformations[0].MetricNamespace),
		)
		return namespace == aws.StringValue(filter.MetricTransformations[0].MetricNamespace)
	})
	return m
}

// TValue is an alias to TransformationMetricValue
func (m *MetricFilter) TValue(value string) *MetricFilter {
	return m.TransformationMetricValue(value)
}

// TransformationMetricValue adds the TransformationMetricValue filter to the filter list
// the TransformationMetricValue filter: filters filters by TransformationMetricValue where
// 'value' provided is the expected MetricTransformations[0].MetricValue value
func (m *MetricFilter) TransformationMetricValue(value string) *MetricFilter {
	m.filterList = append(m.filterList, func(v interface{}) bool {
		filter := convert(v)
		if filter == nil ||
			len(filter.MetricTransformations) == 0 {
			return false
		}
		shared.Debugf("%s == %s -> %t\n",
			value,
			aws.StringValue(filter.MetricTransformations[0].MetricValue),
			value == aws.StringValue(filter.MetricTransformations[0].MetricValue),
		)
		return value == aws.StringValue(filter.MetricTransformations[0].MetricValue)
	})
	return m
}

// TDefault is an alias to TransformationDefaultValue
func (m *MetricFilter) TDefault(value float64) *MetricFilter {
	return m.TransformationDefaultValue(value)
}

// TransformationDefaultValue adds the TransformationDefaultValue filter to the filter list
// the TransformationDefaultValue filter: filters filters by TransformationDefaultValue where
// 'value' provided is the expected MetricTransformations[0].DefaultValue value
func (m *MetricFilter) TransformationDefaultValue(value float64) *MetricFilter {
	m.filterList = append(m.filterList, func(v interface{}) bool {
		filter := convert(v)
		if filter == nil ||
			len(filter.MetricTransformations) == 0 {
			return false
		}
		shared.Debugf("%f == %f -> %t\n",
			value,
			aws.Float64Value(filter.MetricTransformations[0].DefaultValue),
			value == aws.Float64Value(filter.MetricTransformations[0].DefaultValue),
		)
		return value == aws.Float64Value(filter.MetricTransformations[0].DefaultValue)
	})
	return m
}

func (m *MetricFilter) filter(filters []*cloudwatchlogs.MetricFilter) ([]*cloudwatchlogs.MetricFilter, error) {
	if len(filters) == 0 {
		var err error
		filters, err = m.filters()
		if err != nil {
			return nil, err
		}
	}
	return fromIface(shared.GenericFilter(m.filterList, toIface(filters))), nil
}

func (m *MetricFilter) filters() ([]*cloudwatchlogs.MetricFilter, error) {
	svc := cloudwatchlogs.New(m.client)
	var filters []*cloudwatchlogs.MetricFilter
	err := svc.DescribeMetricFiltersPages(
		&cloudwatchlogs.DescribeMetricFiltersInput{},
		func(out *cloudwatchlogs.DescribeMetricFiltersOutput, lastPage bool) bool {
			filters = append(filters, out.MetricFilters...)
			return !lastPage
		})
	if err != nil {
		return nil, err
	}
	return filters, nil
}

func convert(in interface{}) *cloudwatchlogs.MetricFilter {
	out, ok := in.(*cloudwatchlogs.MetricFilter)
	if !ok {
		shared.Debugf("object not convertible to *cloudwatchlogs.MetricFilter: ")
		shared.Dump(in)
		return nil
	}
	return out
}
func toIface(in []*cloudwatchlogs.MetricFilter) (out []interface{}) {
	for _, i := range in {
		out = append(out, i)
	}
	return
}
func fromIface(in []interface{}) (out []*cloudwatchlogs.MetricFilter) {
	for _, i := range in {
		v := convert(i)
		if v == nil {
			continue
		}
		out = append(out, v)
	}
	return
}
