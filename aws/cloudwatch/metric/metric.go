// Package metric provides functions and filters to test AWS CloudWatch Metrics
package metric

import (
	"log"
	"os"
	"strings"
	"testing"

	"github.com/GSA/grace-tftest/aws/cloudwatch/metric/alarm"
	"github.com/GSA/grace-tftest/aws/shared"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

// Metric contains the necessary properties for testing *cloudwatch.Metric objects
type Metric struct {
	client  client.ConfigProvider
	metric  *cloudwatch.Metric
	filters []shared.Filter
}

// New returns a new *Metric
func New(client client.ConfigProvider) *Metric {
	return &Metric{client: client}
}

// Selected returns the currently selected *cloudwatch.Metric
func (r *Metric) Selected() *cloudwatch.Metric {
	return r.metric
}

// Alarm returns a new *alarm.Alarm
func (r *Metric) Alarm() *alarm.Alarm {
	return alarm.New(r.client, r.metric)
}

// Assert applies all filters that have been called, resets the list of filters,
// fails the test if there is not exactly one match, and stores the matched metric
// if metrics is not provided, *cloudwatch.Metric objects will be retreived from AWS
func (r *Metric) Assert(t *testing.T, metrics ...*cloudwatch.Metric) *Metric {
	var err error
	metrics, err = r.filter(metrics)
	if err != nil {
		t.Fatal(err)
	}

	switch l := len(metrics); {
	case l == 0:
		t.Fatal("no matching metric was found")
	case l > 1:
		t.Fatal("more than one matching metric was found")
	default:
		r.metric = metrics[0]
	}

	r.filters = []shared.Filter{}
	return r
}

// First applies all filters that have been called, resets the list of filters,
// fails the test if there are no matches, and stores the first match
// if metrics is not provided, *cloudwatch.Metric objects will be retreived from AWS
func (r *Metric) First(t *testing.T, metrics ...*cloudwatch.Metric) *Metric {
	var err error
	metrics, err = r.filter(metrics)
	if err != nil {
		t.Fatal(err)
	}

	if len(metrics) == 0 {
		t.Fatal("no matching metric was found")
	} else {
		r.metric = metrics[0]
	}

	r.filters = []shared.Filter{}
	return r
}

// Filter adds the 'filter' provided to the filter list
func (r *Metric) Filter(filter shared.Filter) *Metric {
	r.filters = append(r.filters, filter)
	return r
}

// Namespace adds the Namespace filter to the filter list
// the Namespace filter: filters metrics by Namespace where 'name' provided
// is the expected Namespace value
func (r *Metric) Namespace(name string) *Metric {
	r.filters = append(r.filters, func(v interface{}) bool {
		metric := convert(v)
		if metric == nil {
			return false
		}
		shared.Debugf(
			"%s like %s -> %t\n",
			name,
			aws.StringValue(metric.Namespace),
			strings.EqualFold(name, aws.StringValue(metric.Namespace)),
		)
		return strings.EqualFold(name, aws.StringValue(metric.Namespace))
	})
	return r
}

// MetricName adds the MetricName filter to the filter list
// the MetricName filter: filters metrics by MetricName where 'name' provided
// is the expected MetricName value
func (r *Metric) MetricName(name string) *Metric {
	r.filters = append(r.filters, func(v interface{}) bool {
		metric := convert(v)
		if metric == nil {
			return false
		}
		shared.Debugf(
			"%s like %s -> %t\n",
			name,
			aws.StringValue(metric.MetricName),
			strings.EqualFold(name, aws.StringValue(metric.MetricName)),
		)
		return strings.EqualFold(name, aws.StringValue(metric.MetricName))
	})
	return r
}

// Name adds the Name filter as an alias to the MetricName filter
func (r *Metric) Name(name string) *Metric {
	return r.MetricName(name)
}

func (r *Metric) filter(metrics []*cloudwatch.Metric) ([]*cloudwatch.Metric, error) {
	if len(metrics) == 0 {
		var err error
		metrics, err = r.metrics()
		if err != nil {
			return nil, err
		}
	}
	results := fromIface(shared.GenericFilter(r.filters, toIface(metrics)))
	if len(results) == 0 {
		log.Println("aws.cloudwatch.metric.filter had zero results: ")
		shared.Spew(os.Stdout, metrics)
	}
	return results, nil
}

func (r *Metric) metrics() ([]*cloudwatch.Metric, error) {
	svc := cloudwatch.New(r.client)
	input := &cloudwatch.ListMetricsInput{}
	result, err := svc.ListMetrics(input)
	if err != nil {
		return nil, err
	}
	metrics := result.Metrics
	token := aws.StringValue(result.NextToken)
	for token != "" {
		input.NextToken = &token
		result, err := svc.ListMetrics(input)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, result.Metrics...)
		token = aws.StringValue(result.NextToken)
	}
	return metrics, nil
}

func convert(in interface{}) *cloudwatch.Metric {
	out, ok := in.(*cloudwatch.Metric)
	if !ok {
		shared.Debugf("object not convertible to *cloudwatch.Metric: ")
		shared.Dump(in)
		return nil
	}
	return out
}
func toIface(in []*cloudwatch.Metric) (out []interface{}) {
	for _, i := range in {
		out = append(out, i)
	}
	return
}
func fromIface(in []interface{}) (out []*cloudwatch.Metric) {
	for _, i := range in {
		v := convert(i)
		if v == nil {
			continue
		}
		out = append(out, v)
	}
	return
}
