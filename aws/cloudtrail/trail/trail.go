package trail

import (
	"log"
	"os"
	"testing"

	"github.com/GSA/grace-tftest/aws/shared"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/cloudtrail"
)

// Trail contains the necessary properties for testing *cloudtrail.Trail objects
type Trail struct {
	client  client.ConfigProvider
	trail   *cloudtrail.TrailInfo
	filters []shared.Filter
}

func New(client client.ConfigProvider) *Trail {
	return &Trail{client: client}
}

// Selected returns the currently selected *cloudtrail.TrailInfo
func (t *Trail) Selected() *cloudtrail.TrailInfo {
	return t.trail
}

// Assert applies all filters that have been called, resets the list of filters,
// fails the test if there is not exactly one match, and stores the matched trail
// if trails is not provided, *cloudtrail.TrailInfo objects will be retreived from AWS
func (t *Trail) Assert(tt *testing.T, trails ...*cloudtrail.TrailInfo) *Trail {
	var err error
	trails, err = t.filter(trails)
	if err != nil {
		tt.Fatal(err)
	}

	switch l := len(trails); {
	case l == 0:
		tt.Fatal("no matching trail was found")
	case l > 1:
		tt.Fatal("more than one matching trail was found")
	default:
		t.trail = trails[0]
	}

	t.filters = []shared.Filter{}
	return t
}

// First applies all filters that have been called, resets the list of filters,
// fails the test if there are no matches, and stores the first match
// if trails is not provided, *cloudtrail.TrailInfo objects will be retreived from AWS
func (t *Trail) First(tt *testing.T, trails ...*cloudtrail.TrailInfo) *Trail {
	var err error
	trails, err = t.filter(trails)
	if err != nil {
		tt.Fatal(err)
	}

	if len(trails) == 0 {
		tt.Fatal("no matching trail was found")
	} else {
		t.trail = trails[0]
	}

	t.filters = []shared.Filter{}
	return t
}

// Arn adds the Arn filter to the filter list
// the Arn filter: filters trails by TrailArn where 'arn' provided
// is the expected Arn value
func (t *Trail) Arn(arn string) *Trail {
	t.filters = append(t.filters, func(v interface{}) bool {
		trail := convert(v)
		if trail == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", arn, aws.StringValue(trail.TrailARN), arn == aws.StringValue(trail.TrailARN))
		return arn == aws.StringValue(trail.TrailARN)
	})
	return t
}

// Name adds the Name filter to the filter list
// the Name filter: filters trails by Name where 'name' provided
// is the expected trail Name value
func (t *Trail) Name(name string) *Trail {
	t.filters = append(t.filters, func(v interface{}) bool {
		trail := convert(v)
		if trail == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", name, aws.StringValue(trail.Name), name == aws.StringValue(trail.Name))
		return name == aws.StringValue(trail.Name)
	})
	return t
}

// Region adds the Region filter to the filter list
// the Region filter: filters trails by HomeRegion where 'region' provided
// is the expected trail HomeRegion value
func (t *Trail) Region(region string) *Trail {
	t.filters = append(t.filters, func(v interface{}) bool {
		trail := convert(v)
		if trail == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", region, aws.StringValue(trail.HomeRegion), region == aws.StringValue(trail.HomeRegion))
		return region == aws.StringValue(trail.HomeRegion)
	})
	return t
}

func (t *Trail) filter(trails []*cloudtrail.TrailInfo) ([]*cloudtrail.TrailInfo, error) {
	if len(trails) == 0 {
		var err error
		trails, err = t.trails()
		if err != nil {
			return nil, err
		}
	}

	results := fromIface(shared.GenericFilter(t.filters, toIface(trails)))
	if len(results) == 0 {
		log.Println("aws.cloudtrail.trail.filter had zero results: ")
		shared.Spew(os.Stdout, trails)
	}
	return results, nil
}

func (t *Trail) trails() ([]*cloudtrail.TrailInfo, error) {
	svc := cloudtrail.New(t.client)
	var trails []*cloudtrail.TrailInfo
	err := svc.ListTrailsPages(&cloudtrail.ListTrailsInput{}, func(out *cloudtrail.ListTrailsOutput, lastPage bool) bool {
		trails = append(trails, out.Trails...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	return trails, nil
}

func convert(in interface{}) *cloudtrail.TrailInfo {
	out, ok := in.(*cloudtrail.TrailInfo)
	if !ok {
		shared.Debugf("object not convertible to *cloudtrail.TrailInfo: ")
		shared.Dump(in)
		return nil
	}
	return out
}
func toIface(in []*cloudtrail.TrailInfo) (out []interface{}) {
	for _, i := range in {
		out = append(out, i)
	}
	return
}
func fromIface(in []interface{}) (out []*cloudtrail.TrailInfo) {
	for _, i := range in {
		v := convert(i)
		if v == nil {
			continue
		}
		out = append(out, v)
	}
	return
}
