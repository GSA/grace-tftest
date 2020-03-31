// Package group contains the necessary properties for testing *cloudwatchlogs.group objects
package group

import (
	"log"
	"os"
	"strings"
	"testing"

	"github.com/GSA/grace-tftest/aws/shared"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

// Group contains the necessary properties for testing *cloudwatchlogs.LogGroup objects
type Group struct {
	client  client.ConfigProvider
	group   *cloudwatchlogs.LogGroup
	filters []shared.Filter
}

// New returns a new *Group
func New(client client.ConfigProvider) *Group {
	return &Group{client: client}
}

// Selected returns the currently selected *cloudwatchlogs.LogGroup
func (r *Group) Selected() *cloudwatchlogs.LogGroup {
	return r.group
}

// Assert applies all filters that have been called, resets the list of filters,
// fails the test if there is not exactly one match, and stores the matched group
// if groups is not provided, *cloudwatchlogs.LogGroup objects will be retreived from AWS
func (r *Group) Assert(t *testing.T, groups ...*cloudwatchlogs.LogGroup) *Group {
	var err error
	groups, err = r.filter(groups)
	if err != nil {
		t.Fatal(err)
	}

	switch l := len(groups); {
	case l == 0:
		t.Fatal("no matching group was found")
	case l > 1:
		t.Fatal("more than one matching group was found")
	default:
		r.group = groups[0]
	}

	r.filters = []shared.Filter{}
	return r
}

// First applies all filters that have been called, resets the list of filters,
// fails the test if there are no matches, and stores the first match
// if groups is not provided, *cloudwatchlogs.LogGroup objects will be retreived from AWS
func (r *Group) First(t *testing.T, groups ...*cloudwatchlogs.LogGroup) *Group {
	var err error
	groups, err = r.filter(groups)
	if err != nil {
		t.Fatal(err)
	}

	if len(groups) == 0 {
		t.Fatal("no matching group was found")
	} else {
		r.group = groups[0]
	}

	r.filters = []shared.Filter{}
	return r
}

// Filter adds the 'filter' provided to the filter list
func (r *Group) Filter(filter shared.Filter) *Group {
	r.filters = append(r.filters, filter)
	return r
}

// Arn adds the Arn filter to the filter list
// the Arn filter: filters groups by Arn where 'arn' provided
// is the expected Arn value
func (r *Group) Arn(arn string) *Group {
	r.filters = append(r.filters, func(v interface{}) bool {
		group := convert(v)
		if group == nil {
			return false
		}
		shared.Debugf(
			"%s like %s -> %t\n",
			arn,
			aws.StringValue(group.Arn),
			strings.EqualFold(arn, aws.StringValue(group.Arn)),
		)
		return strings.EqualFold(arn, aws.StringValue(group.Arn))
	})
	return r
}

// LogGroupName adds the LogGroupName filter to the filter list
// the LogGroupName filter: filters groups by LogGroupName where 'name' provided
// is the expected LogGroupName value
func (r *Group) LogGroupName(name string) *Group {
	r.filters = append(r.filters, func(v interface{}) bool {
		group := convert(v)
		if group == nil {
			return false
		}
		shared.Debugf(
			"%s like %s -> %t\n",
			name,
			aws.StringValue(group.LogGroupName),
			strings.EqualFold(name, aws.StringValue(group.LogGroupName)),
		)
		return strings.EqualFold(name, aws.StringValue(group.LogGroupName))
	})
	return r
}

// Name adds the Name filter as an alias to the LogGroupName filter
func (r *Group) Name(name string) *Group {
	return r.LogGroupName(name)
}

// KmsKeyID adds the KmsKeyID filter to the filter list
// the KmsKeyID filter: filters groups by KmsKeyId where 'id' provided
// is the expected KmsKeyId value
func (r *Group) KmsKeyID(id string) *Group {
	r.filters = append(r.filters, func(v interface{}) bool {
		group := convert(v)
		if group == nil {
			return false
		}
		shared.Debugf(
			"%s like %s -> %t\n",
			id,
			aws.StringValue(group.KmsKeyId),
			strings.EqualFold(id, aws.StringValue(group.KmsKeyId)),
		)
		return strings.EqualFold(id, aws.StringValue(group.KmsKeyId))
	})
	return r
}

// RetentionInDays adds the RetentionInDays filter to the filter list
// the RetentionInDays filter: filters groups by RetentionInDays where 'name' provided
// is the expected RetentionInDays value
func (r *Group) RetentionInDays(days int64) *Group {
	r.filters = append(r.filters, func(v interface{}) bool {
		group := convert(v)
		if group == nil {
			return false
		}
		shared.Debugf(
			"%d like %d -> %t\n",
			days,
			aws.Int64Value(group.RetentionInDays),
			days == aws.Int64Value(group.RetentionInDays),
		)
		return days == aws.Int64Value(group.RetentionInDays)
	})
	return r
}

func (r *Group) filter(groups []*cloudwatchlogs.LogGroup) ([]*cloudwatchlogs.LogGroup, error) {
	if len(groups) == 0 {
		var err error
		groups, err = r.groups()
		if err != nil {
			return nil, err
		}
	}
	results := fromIface(shared.GenericFilter(r.filters, toIface(groups)))
	if len(results) == 0 {
		log.Println("aws.cloudwatchlogs.group.filter had zero results: ")
		shared.Spew(os.Stdout, groups)
	}
	return results, nil
}

func (r *Group) groups() ([]*cloudwatchlogs.LogGroup, error) {
	svc := cloudwatchlogs.New(r.client)
	var groups []*cloudwatchlogs.LogGroup
	err := svc.DescribeLogGroupsPages(&cloudwatchlogs.DescribeLogGroupsInput{}, func(page *cloudwatchlogs.DescribeLogGroupsOutput, lastPage bool) bool {
		groups = append(groups, page.LogGroups...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	return groups, nil
}

func convert(in interface{}) *cloudwatchlogs.LogGroup {
	out, ok := in.(*cloudwatchlogs.LogGroup)
	if !ok {
		shared.Debugf("object not convertible to *cloudwatchlogs.LogGroup: ")
		shared.Dump(in)
		return nil
	}
	return out
}
func toIface(in []*cloudwatchlogs.LogGroup) (out []interface{}) {
	for _, i := range in {
		out = append(out, i)
	}
	return
}
func fromIface(in []interface{}) (out []*cloudwatchlogs.LogGroup) {
	for _, i := range in {
		v := convert(i)
		if v == nil {
			continue
		}
		out = append(out, v)
	}
	return
}
