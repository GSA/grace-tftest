package rule

import (
	"strings"
	"testing"

	"github.com/GSA/grace-tftest/aws/cloudwatchevents/rule/target"
	"github.com/GSA/grace-tftest/aws/shared"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
)

// Rule contains the necessary properties for testing *cloudwatchevents.Rule objects
type Rule struct {
	client  client.ConfigProvider
	rule    *cloudwatchevents.Rule
	filters []shared.Filter
}

// New returns a new *Rule
func New(client client.ConfigProvider) *Rule {
	return &Rule{client: client}
}

// Selected returns the currently selected *cloudwatchevents.Rule
func (r *Rule) Selected() *cloudwatchevents.Rule {
	return r.rule
}

// Target returns a new *target.Target
func (r *Rule) Target() *target.Target {
	return target.New(r.client, r.rule)
}

// Assert applies all filters that have been called, resets the list of filters,
// fails the test if there is not exactly one match, and stores the matched rule
// if rules is not provided, *cloudwatchevents.Rule objects will be retreived from AWS
func (r *Rule) Assert(t *testing.T, rules ...*cloudwatchevents.Rule) *Rule {
	var err error
	rules, err = r.filter(rules)
	if err != nil {
		t.Fatal(err)
	}

	switch l := len(rules); {
	case l == 0:
		t.Fatal("no matching rule was found")
	case l > 1:
		t.Fatal("more than one matching rule was found")
	default:
		r.rule = rules[0]
	}

	r.filters = []shared.Filter{}
	return r
}

// First applies all filters that have been called, resets the list of filters,
// fails the test if there are no matches, and stores the first match
// if rules is not provided, *cloudwatchevents.Rule objects will be retreived from AWS
func (r *Rule) First(t *testing.T, rules ...*cloudwatchevents.Rule) *Rule {
	var err error
	rules, err = r.filter(rules)
	if err != nil {
		t.Fatal(err)
	}

	if len(rules) == 0 {
		t.Fatal("no matching rule was found")
	} else {
		r.rule = rules[0]
	}

	r.filters = []shared.Filter{}
	return r
}

// Arn adds the Arn filter to the filter list
// the Arn filter: filters rules by Arn where 'arn' provided
// is the expected Arn value
func (r *Rule) Arn(arn string) *Rule {
	r.filters = append(r.filters, func(v interface{}) bool {
		rule := convert(v)
		if rule == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", arn, aws.StringValue(rule.Arn), arn == aws.StringValue(rule.Arn))
		return arn == aws.StringValue(rule.Arn)
	})
	return r
}

// Filter adds the 'filter' provided to the filter list
func (r *Rule) Filter(filter shared.Filter) *Rule {
	r.filters = append(r.filters, filter)
	return r
}

// ManagedBy adds the ManagedBy filter to the filter list
// the ManagedBy filter: filters rules by ManagedBy where 'name' provided
// is the expected ManagedBy value
func (r *Rule) ManagedBy(name string) *Rule {
	r.filters = append(r.filters, func(v interface{}) bool {
		rule := convert(v)
		if rule == nil {
			return false
		}
		shared.Debugf("%s like %s -> %t\n", name, aws.StringValue(rule.ManagedBy), strings.EqualFold(name, aws.StringValue(rule.ManagedBy)))
		return strings.EqualFold(name, aws.StringValue(rule.ManagedBy))
	})
	return r
}

// Name adds the Name filter to the filter list
// the Name filter: filters rules by Name where 'name' provided
// is the expected Name value
func (r *Rule) Name(name string) *Rule {
	r.filters = append(r.filters, func(v interface{}) bool {
		rule := convert(v)
		if rule == nil {
			return false
		}
		shared.Debugf("%s like %s -> %t\n", name, aws.StringValue(rule.Name), strings.EqualFold(name, aws.StringValue(rule.Name)))
		return strings.EqualFold(name, aws.StringValue(rule.Name))
	})
	return r
}

// RoleArn adds the RoleArn filter to the filter list
// the RoleArn filter: filters rules by RoleArn where 'arn' provided
// is the expected RoleArn value
func (r *Rule) RoleArn(arn string) *Rule {
	r.filters = append(r.filters, func(v interface{}) bool {
		rule := convert(v)
		if rule == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", arn, aws.StringValue(rule.RoleArn), arn == aws.StringValue(rule.RoleArn))
		return arn == aws.StringValue(rule.RoleArn)
	})
	return r
}

// SchedExpr adds the SchedExpr filter to the filter list
// the SchedExpr filter: filters rules by SchedExpr where 'expr' provided
// is the expected ScheduleExpression value
func (r *Rule) SchedExpr(expr string) *Rule {
	r.filters = append(r.filters, func(v interface{}) bool {
		rule := convert(v)
		if rule == nil {
			return false
		}
		shared.Debugf("%s like %s -> %t\n", expr, aws.StringValue(rule.ScheduleExpression),
			strings.EqualFold(expr, aws.StringValue(rule.ScheduleExpression)))
		return strings.EqualFold(expr, aws.StringValue(rule.ScheduleExpression))
	})
	return r
}

// State adds the State filter to the filter list
// the State filter: filters rules by State where 'state' provided
// is the expected State value
func (r *Rule) State(state string) *Rule {
	r.filters = append(r.filters, func(v interface{}) bool {
		rule := convert(v)
		if rule == nil {
			return false
		}
		shared.Debugf("%s like %s -> %t\n", state, aws.StringValue(rule.State), strings.EqualFold(state, aws.StringValue(rule.State)))
		return strings.EqualFold(state, aws.StringValue(rule.State))
	})
	return r
}

func (r *Rule) filter(rules []*cloudwatchevents.Rule) ([]*cloudwatchevents.Rule, error) {
	if len(rules) == 0 {
		var err error
		rules, err = r.rules()
		if err != nil {
			return nil, err
		}
	}
	return fromIface(shared.GenericFilter(r.filters, toIface(rules))), nil
}

func (r *Rule) rules() ([]*cloudwatchevents.Rule, error) {
	svc := cloudwatchevents.New(r.client)
	input := &cloudwatchevents.ListRulesInput{}
	result, err := svc.ListRules(input)
	if err != nil {
		return nil, err
	}
	rules := result.Rules
	token := aws.StringValue(result.NextToken)
	for token != "" {
		input.NextToken = &token
		result, err := svc.ListRules(input)
		if err != nil {
			return nil, err
		}
		rules = append(rules, result.Rules...)
		token = aws.StringValue(result.NextToken)
	}
	return rules, nil
}

func convert(in interface{}) *cloudwatchevents.Rule {
	out, ok := in.(*cloudwatchevents.Rule)
	if !ok {
		shared.Debugf("object not convertible to *cloudwatchevents.Rule: ")
		shared.Dump(in)
		return nil
	}
	return out
}
func toIface(in []*cloudwatchevents.Rule) (out []interface{}) {
	for _, i := range in {
		out = append(out, i)
	}
	return
}
func fromIface(in []interface{}) (out []*cloudwatchevents.Rule) {
	for _, i := range in {
		v := convert(i)
		if v == nil {
			continue
		}
		out = append(out, v)
	}
	return
}
