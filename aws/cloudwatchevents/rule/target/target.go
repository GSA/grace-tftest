package target

import (
	"testing"

	"github.com/GSA/grace-tftest/aws/shared"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
)

// Target contains the necessary properties for testing *cloudwatchevents.Target objects
type Target struct {
	client  client.ConfigProvider
	target  *cloudwatchevents.Target
	rule    *cloudwatchevents.Rule
	filters []shared.Filter
}

// New returns a new *Target
func New(client client.ConfigProvider, rule *cloudwatchevents.Rule) *Target {
	return &Target{client: client, rule: rule}
}

// Selected returns the currently selected *cloudwatchevents.Target
func (g *Target) Selected() *cloudwatchevents.Target {
	return g.target
}

// Assert applies all filters that have been called, resets the list of filters,
// fails the test if there is not exactly one match, and stores the matched target
// if targets is not provided, *cloudwatchevents.Target objects will be retreived from AWS
func (g *Target) Assert(t *testing.T, targets ...*cloudwatchevents.Target) *Target {
	var err error
	targets, err = g.filter(targets)
	if err != nil {
		t.Fatal(err)
	}

	switch l := len(targets); {
	case l == 0:
		t.Fatal("no matching target was found")
	case l > 1:
		t.Fatal("more than one matching target was found")
	default:
		g.target = targets[0]
	}

	g.filters = []shared.Filter{}
	return g
}

// First applies all filters that have been called, resets the list of filters,
// fails the test if there are no matches, and stores the first match
// if targets is not provided, *cloudwatchevents.Target objects will be retreived from AWS
func (g *Target) First(t *testing.T, targets ...*cloudwatchevents.Target) *Target {
	var err error
	targets, err = g.filter(targets)
	if err != nil {
		t.Fatal(err)
	}

	if len(targets) == 0 {
		t.Fatal("no matching rule was found")
	} else {
		g.target = targets[0]
	}

	g.filters = []shared.Filter{}
	return g
}

// Arn adds the Arn filter to the filter list
// the Arn filter: filters targets by Arn where 'arn' provided
// is the expected Arn value
func (g *Target) Arn(arn string) *Target {
	g.filters = append(g.filters, func(v interface{}) bool {
		target := convert(v)
		if target == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", arn, aws.StringValue(target.Arn), arn == aws.StringValue(target.Arn))
		return arn == aws.StringValue(target.Arn)
	})
	return g
}

// Filter adds the 'filter' provided to the filter list
func (g *Target) Filter(filter shared.Filter) *Target {
	g.filters = append(g.filters, filter)
	return g
}

// ID adds the ID filter to the filter list
// the ID filter: filters targets by ID where 'id' provided
// is the expected Id value
func (g *Target) ID(id string) *Target {
	g.filters = append(g.filters, func(v interface{}) bool {
		target := convert(v)
		if target == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", id, aws.StringValue(target.Id), id == aws.StringValue(target.Id))
		return id == aws.StringValue(target.Id)
	})
	return g
}

// RoleArn adds the RoleArn filter to the filter list
// the RoleArn filter: filters targets by RoleArn where 'arn' provided
// is the expected RoleArn value
func (g *Target) RoleArn(arn string) *Target {
	g.filters = append(g.filters, func(v interface{}) bool {
		target := convert(v)
		if target == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", arn, aws.StringValue(target.RoleArn), arn == aws.StringValue(target.RoleArn))
		return arn == aws.StringValue(target.RoleArn)
	})
	return g
}

func (g *Target) filter(targets []*cloudwatchevents.Target) ([]*cloudwatchevents.Target, error) {
	if len(targets) == 0 {
		var err error
		targets, err = g.targets()
		if err != nil {
			return nil, err
		}
	}
	return fromIface(shared.GenericFilter(g.filters, toIface(targets))), nil
}

func (g *Target) targets() ([]*cloudwatchevents.Target, error) {
	svc := cloudwatchevents.New(g.client)
	input := &cloudwatchevents.ListTargetsByRuleInput{Rule: g.rule.Name}
	result, err := svc.ListTargetsByRule(input)
	if err != nil {
		return nil, err
	}
	targets := result.Targets
	token := aws.StringValue(result.NextToken)
	for token != "" {
		input.NextToken = &token
		result, err := svc.ListTargetsByRule(input)
		if err != nil {
			return nil, err
		}
		targets = append(targets, result.Targets...)
		token = aws.StringValue(result.NextToken)
	}
	return targets, nil
}

func convert(in interface{}) *cloudwatchevents.Target {
	out, ok := in.(*cloudwatchevents.Target)
	if !ok {
		shared.Debugf("object not convertible to *cloudwatchevents.Target: ")
		shared.Dump(in)
		return nil
	}
	return out
}
func toIface(in []*cloudwatchevents.Target) (out []interface{}) {
	for _, i := range in {
		out = append(out, i)
	}
	return
}
func fromIface(in []interface{}) (out []*cloudwatchevents.Target) {
	for _, i := range in {
		v := convert(i)
		if v == nil {
			continue
		}
		out = append(out, v)
	}
	return
}
