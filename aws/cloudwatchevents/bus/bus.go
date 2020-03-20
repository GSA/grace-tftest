// Package bus provides testing for *cloudwatchevents.EventBus objects
package bus

import (
	"log"
	"os"
	"strings"
	"testing"

	"github.com/GSA/grace-tftest/aws/cloudwatchevents/bus/policy"
	"github.com/GSA/grace-tftest/aws/shared"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
)

// Bus contains the necessary properties for testing *cloudwatchevents.EventBus objects
type Bus struct {
	client  client.ConfigProvider
	bus     *cloudwatchevents.EventBus
	filters []shared.Filter
}

// New returns a new *Bus
func New(client client.ConfigProvider) *Bus {
	return &Bus{client: client}
}

// Selected returns the currently selected *cloudwatchevents.EventBus
func (r *Bus) Selected() *cloudwatchevents.EventBus {
	return r.bus
}

// Policy returns a new *policy.Policy
func (r *Bus) Policy() *policy.Policy {
	return policy.New(r.bus.Policy)
}

// Assert applies all filters that have been called, resets the list of filters,
// fails the test if there is not exactly one match, and stores the matched bus
// if buses is not provided, *cloudwatchevents.EventBus objects will be retreived from AWS
func (r *Bus) Assert(t *testing.T, buses ...*cloudwatchevents.EventBus) *Bus {
	var err error
	buses, err = r.filter(buses)
	if err != nil {
		t.Fatal(err)
	}

	switch l := len(buses); {
	case l == 0:
		t.Fatal("no matching bus was found")
	case l > 1:
		t.Fatal("more than one matching bus was found")
	default:
		r.bus = buses[0]
	}

	r.filters = []shared.Filter{}
	return r
}

// First applies all filters that have been called, resets the list of filters,
// fails the test if there are no matches, and stores the first match
// if buses is not provided, *cloudwatchevents.EventBus objects will be retreived from AWS
func (r *Bus) First(t *testing.T, buses ...*cloudwatchevents.EventBus) *Bus {
	var err error
	buses, err = r.filter(buses)
	if err != nil {
		t.Fatal(err)
	}

	if len(buses) == 0 {
		t.Fatal("no matching bus was found")
	} else {
		r.bus = buses[0]
	}

	r.filters = []shared.Filter{}
	return r
}

// Arn adds the Arn filter to the filter list
// the Arn filter: filters buses by Arn where 'arn' provided
// is the expected Arn value
func (r *Bus) Arn(arn string) *Bus {
	r.filters = append(r.filters, func(v interface{}) bool {
		bus := convert(v)
		if bus == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", arn, aws.StringValue(bus.Arn), arn == aws.StringValue(bus.Arn))
		return arn == aws.StringValue(bus.Arn)
	})
	return r
}

// Filter adds the 'filter' provided to the filter list
func (r *Bus) Filter(filter shared.Filter) *Bus {
	r.filters = append(r.filters, filter)
	return r
}

// Name adds the Name filter to the filter list
// the Name filter: filters buses by Name where 'name' provided
// is the expected Name value
func (r *Bus) Name(name string) *Bus {
	r.filters = append(r.filters, func(v interface{}) bool {
		bus := convert(v)
		if bus == nil {
			return false
		}
		shared.Debugf(
			"%s like %s -> %t\n",
			name,
			aws.StringValue(bus.Name),
			strings.EqualFold(name, aws.StringValue(bus.Name)),
		)
		return strings.EqualFold(name, aws.StringValue(bus.Name))
	})
	return r
}

func (r *Bus) filter(buses []*cloudwatchevents.EventBus) ([]*cloudwatchevents.EventBus, error) {
	if len(buses) == 0 {
		var err error
		buses, err = r.buses()
		if err != nil {
			return nil, err
		}
	}
	results := fromIface(shared.GenericFilter(r.filters, toIface(buses)))
	if len(results) == 0 {
		log.Println("aws.cloudwatchevents.bus.filter had zero results: ")
		shared.Spew(os.Stdout, buses)
	}
	return results, nil
}

func (r *Bus) buses() ([]*cloudwatchevents.EventBus, error) {
	svc := cloudwatchevents.New(r.client)
	input := &cloudwatchevents.ListEventBusesInput{}
	result, err := svc.ListEventBuses(input)
	if err != nil {
		return nil, err
	}
	buses := result.EventBuses
	token := aws.StringValue(result.NextToken)
	for token != "" {
		input.NextToken = &token
		result, err := svc.ListEventBuses(input)
		if err != nil {
			return nil, err
		}
		buses = append(buses, result.EventBuses...)
		token = aws.StringValue(result.NextToken)
	}
	return buses, nil
}

func convert(in interface{}) *cloudwatchevents.EventBus {
	out, ok := in.(*cloudwatchevents.EventBus)
	if !ok {
		shared.Debugf("object not convertible to *cloudwatchevents.EventBus: ")
		shared.Dump(in)
		return nil
	}
	return out
}
func toIface(in []*cloudwatchevents.EventBus) (out []interface{}) {
	for _, i := range in {
		out = append(out, i)
	}
	return
}
func fromIface(in []interface{}) (out []*cloudwatchevents.EventBus) {
	for _, i := range in {
		v := convert(i)
		if v == nil {
			continue
		}
		out = append(out, v)
	}
	return
}
