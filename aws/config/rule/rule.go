// Package rule provides the ability to filter *configservice.ConfigRule objects
package rule

import (
	"testing"

	"github.com/GSA/grace-tftest/aws/shared"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/configservice"
)

// Rule contains the necessary properties for testing *configservice.ConfigRule objects
type Rule struct {
	client  client.ConfigProvider
	rule    *configservice.ConfigRule
	filters []shared.Filter
}

// New returns a new *Rule
func New(client client.ConfigProvider) *Rule {
	return &Rule{client: client}
}

// Selected returns the currently selected *configservice.ConfigRule
func (r *Rule) Selected() *configservice.ConfigRule {
	return r.rule
}

// Assert applies all filters that have been called, resets the list of filters,
// fails the test if there is not exactly one match, and stores the matched rule
// if rules is not provided, *configservice.ConfigRule objects will be retreived from AWS
func (r *Rule) Assert(t *testing.T, rules ...*configservice.ConfigRule) *Rule {
	var err error
	rules, err = r.filter(rules)
	if err != nil {
		t.Error(err)
	}

	switch l := len(rules); {
	case l == 0:
		t.Error("no matching rule was found")
	case l > 1:
		t.Error("more than one matching rule was found")
	default:
		r.rule = rules[0]
	}

	r.filters = []shared.Filter{}
	return r
}

// First applies all filters that have been called, resets the list of filters,
// fails the test if there are no matches, and stores the first match
// if rules is not provided, *configservice.ConfigRule objects will be retreived from AWS
func (r *Rule) First(t *testing.T, rules ...*configservice.ConfigRule) *Rule {
	var err error
	rules, err = r.filter(rules)
	if err != nil {
		t.Error(err)
	}

	if len(rules) == 0 {
		t.Error("no matching rule was found")
	} else {
		r.rule = rules[0]
	}

	r.filters = []shared.Filter{}
	return r
}

// Filter adds the 'filter' provided to the filter list
func (r *Rule) Filter(filter shared.Filter) *Rule {
	r.filters = append(r.filters, filter)
	return r
}

// Arn adds the Arn filter to the filter list
// the Arn filter: filters rules by Arn where 'arn' provided
// is the expected ConfigRuleArn value
func (r *Rule) Arn(arn string) *Rule {
	r.filters = append(r.filters, func(v interface{}) bool {
		rule := convert(v)
		if rule == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n",
			arn,
			aws.StringValue(rule.ConfigRuleArn),
			arn == aws.StringValue(rule.ConfigRuleArn),
		)
		return arn == aws.StringValue(rule.ConfigRuleArn)
	})
	return r
}

// ID adds the ID filter to the filter list
// the ID filter: filters rules by ID where 'id' provided
// is the expected ConfigRuleId value
func (r *Rule) ID(id string) *Rule {
	r.filters = append(r.filters, func(v interface{}) bool {
		rule := convert(v)
		if rule == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n",
			id,
			aws.StringValue(rule.ConfigRuleId),
			id == aws.StringValue(rule.ConfigRuleId),
		)
		return id == aws.StringValue(rule.ConfigRuleId)
	})
	return r
}

// Name adds the Name filter to the filter list
// the Name filter: filters rules by Name where 'name' provided
// is the expected ConfigRuleName value
func (r *Rule) Name(name string) *Rule {
	r.filters = append(r.filters, func(v interface{}) bool {
		rule := convert(v)
		if rule == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n",
			name,
			aws.StringValue(rule.ConfigRuleName),
			name == aws.StringValue(rule.ConfigRuleName),
		)
		return name == aws.StringValue(rule.ConfigRuleName)
	})
	return r
}

// State adds the State filter to the filter list
// the State filter: filters rules by State where 'state' provided
// is the expected ConfigRuleState value
func (r *Rule) State(state string) *Rule {
	r.filters = append(r.filters, func(v interface{}) bool {
		rule := convert(v)
		if rule == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n",
			state,
			aws.StringValue(rule.ConfigRuleState),
			state == aws.StringValue(rule.ConfigRuleState),
		)
		return state == aws.StringValue(rule.ConfigRuleState)
	})
	return r
}

// CreatedBy adds the CreatedBy filter to the filter list
// the CreatedBy filter: filters rules by CreatedBy where 'service' provided
// is the expected CreatedBy value
func (r *Rule) CreatedBy(service string) *Rule {
	r.filters = append(r.filters, func(v interface{}) bool {
		rule := convert(v)
		if rule == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n",
			service,
			aws.StringValue(rule.CreatedBy),
			service == aws.StringValue(rule.CreatedBy),
		)
		return service == aws.StringValue(rule.CreatedBy)
	})
	return r
}

// Description adds the Description filter to the filter list
// the Description filter: filters rules by Description where 'desc' provided
// is the expected Description value
func (r *Rule) Description(desc string) *Rule {
	r.filters = append(r.filters, func(v interface{}) bool {
		rule := convert(v)
		if rule == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n",
			desc,
			aws.StringValue(rule.Description),
			desc == aws.StringValue(rule.Description),
		)
		return desc == aws.StringValue(rule.Description)
	})
	return r
}

// Frequency adds the Frequency filter to the filter list
// the Frequency filter: filters rules by Frequency where 'freq' provided
// is the expected MaximumExecutionFrequency value
func (r *Rule) Frequency(freq string) *Rule {
	r.filters = append(r.filters, func(v interface{}) bool {
		rule := convert(v)
		if rule == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n",
			freq,
			aws.StringValue(rule.MaximumExecutionFrequency),
			freq == aws.StringValue(rule.MaximumExecutionFrequency),
		)
		return freq == aws.StringValue(rule.MaximumExecutionFrequency)
	})
	return r
}

// ScopeID adds the ScopeID filter to the filter list
// the ScopeID filter: filters rules by ScopeID where 'id' provided
// is the expected ComplianceResourceId value
func (r *Rule) ScopeID(id string) *Rule {
	r.filters = append(r.filters, func(v interface{}) bool {
		rule := convert(v)
		if rule == nil {
			return false
		}
		if rule.Scope == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n",
			id,
			aws.StringValue(rule.Scope.ComplianceResourceId),
			id == aws.StringValue(rule.Scope.ComplianceResourceId),
		)
		return id == aws.StringValue(rule.Scope.ComplianceResourceId)
	})
	return r
}

// ScopeTypes adds the ScopeTypes filter to the filter list
// the ScopeTypes filter: filters rules by ScopeTypes where
// 'types' provided is the expected ComplianceResourceTypes value
func (r *Rule) ScopeTypes(types ...string) *Rule {
	r.filters = append(r.filters, func(v interface{}) bool {
		rule := convert(v)
		if rule == nil {
			return false
		}
		if rule.Scope == nil {
			return false
		}
		return shared.StringSliceEqual(
			types,
			aws.StringValueSlice(rule.Scope.ComplianceResourceTypes),
		)
	})
	return r
}

// ScopeTagKey adds the ScopeTagKey filter to the filter list
// the ScopeTagKey filter: filters rules by ScopeTagKey where 'key' provided
// is the expected TagKey value
func (r *Rule) ScopeTagKey(key string) *Rule {
	r.filters = append(r.filters, func(v interface{}) bool {
		rule := convert(v)
		if rule == nil {
			return false
		}
		if rule.Scope == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n",
			key,
			aws.StringValue(rule.Scope.TagKey),
			key == aws.StringValue(rule.Scope.TagKey),
		)
		return key == aws.StringValue(rule.Scope.TagKey)
	})
	return r
}

// ScopeTagValue adds the ScopeTagValue filter to the filter list
// the ScopeTagValue filter: filters rules by ScopeTagValue where
// 'value' provided is the expected TagValue value
func (r *Rule) ScopeTagValue(value string) *Rule {
	r.filters = append(r.filters, func(v interface{}) bool {
		rule := convert(v)
		if rule == nil {
			return false
		}
		if rule.Scope == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n",
			value,
			aws.StringValue(rule.Scope.TagValue),
			value == aws.StringValue(rule.Scope.TagValue),
		)
		return value == aws.StringValue(rule.Scope.TagValue)
	})
	return r
}

// SourceOwner adds the SourceOwner filter to the filter list
// the SourceOwner filter: filters rules by SourceOwner where
// 'owner' provided is the expected Owner value
func (r *Rule) SourceOwner(owner string) *Rule {
	r.filters = append(r.filters, func(v interface{}) bool {
		rule := convert(v)
		if rule == nil {
			return false
		}
		if rule.Scope == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n",
			owner,
			aws.StringValue(rule.Source.Owner),
			owner == aws.StringValue(rule.Source.Owner),
		)
		return owner == aws.StringValue(rule.Source.Owner)
	})
	return r
}

// SourceID adds the SourceID filter to the filter list
// the SourceID filter: filters rules by SourceID where
// 'id' provided is the expected SourceIdentifier value
func (r *Rule) SourceID(id string) *Rule {
	r.filters = append(r.filters, func(v interface{}) bool {
		rule := convert(v)
		if rule == nil {
			return false
		}
		if rule.Source == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n",
			id,
			aws.StringValue(rule.Source.SourceIdentifier),
			id == aws.StringValue(rule.Source.SourceIdentifier),
		)
		return id == aws.StringValue(rule.Source.SourceIdentifier)
	})
	return r
}

// SourceDetailEventSource adds the SourceDetailEventSource filter to the filter list
// the SourceDetailEventSource filter: filters rules by SourceDetailEventSource where
// 'source' provided is the expected EventSource value
func (r *Rule) SourceDetailEventSource(source string) *Rule {
	r.filters = append(r.filters, func(v interface{}) bool {
		rule := convert(v)
		if rule == nil {
			return false
		}

		return detailEquals(
			rule,
			source,
			func(d *configservice.SourceDetail) string {
				return aws.StringValue(d.EventSource)
			},
		)
	})
	return r
}

// SourceDetailFrequency adds the SourceDetailFrequency filter to the filter list
// the SourceDetailFrequency filter: filters rules by SourceDetailFrequency where
// 'freq' provided is the expected MaximumExecutionFrequency value
func (r *Rule) SourceDetailFrequency(freq string) *Rule {
	r.filters = append(r.filters, func(v interface{}) bool {
		rule := convert(v)
		if rule == nil {
			return false
		}

		return detailEquals(
			rule,
			freq,
			func(d *configservice.SourceDetail) string {
				return aws.StringValue(d.MaximumExecutionFrequency)
			},
		)
	})
	return r
}

// SourceDetailMessageType adds the SourceDetailMessageType filter to the filter list
// the SourceDetailMessageType filter: filters rules by SourceDetailMessageType where
// 'typ' provided is the expected MessageType value
func (r *Rule) SourceDetailMessageType(typ string) *Rule {
	r.filters = append(r.filters, func(v interface{}) bool {
		rule := convert(v)
		if rule == nil {
			return false
		}

		return detailEquals(
			rule,
			typ,
			func(d *configservice.SourceDetail) string {
				return aws.StringValue(d.MessageType)
			},
		)
	})
	return r
}

func detailEquals(rule *configservice.ConfigRule, expected string,
	actual func(*configservice.SourceDetail) string) bool {
	if rule.Source == nil ||
		rule.Source.SourceDetails == nil {
		return false
	}
	for _, d := range rule.Source.SourceDetails {
		if expected == actual(d) {
			shared.Debugf("%s == %s -> %t\n",
				expected,
				actual(d),
				expected == actual(d),
			)
			return true
		}
	}
	return false
}

func (r *Rule) filter(rules []*configservice.ConfigRule) ([]*configservice.ConfigRule, error) {
	if len(rules) == 0 {
		var err error
		rules, err = r.rules()
		if err != nil {
			return nil, err
		}
	}
	return fromIface(shared.GenericFilter(r.filters, toIface(rules))), nil
}

func (r *Rule) rules() ([]*configservice.ConfigRule, error) {
	svc := configservice.New(r.client)
	in := &configservice.DescribeConfigRulesInput{}
	out, err := svc.DescribeConfigRules(in)
	if err != nil {
		return nil, err
	}
	rules := out.ConfigRules
	for out.NextToken != nil {
		in.NextToken = out.NextToken
		out, err = svc.DescribeConfigRules(in)
		if err != nil {
			return nil, err
		}
		rules = append(rules, out.ConfigRules...)
	}
	return rules, nil
}

func convert(in interface{}) *configservice.ConfigRule {
	out, ok := in.(*configservice.ConfigRule)
	if !ok {
		shared.Debugf("object not convertible to *configservice.ConfigRule: ")
		shared.Dump(in)
		return nil
	}
	return out
}
func toIface(in []*configservice.ConfigRule) (out []interface{}) {
	for _, i := range in {
		out = append(out, i)
	}
	return
}
func fromIface(in []interface{}) (out []*configservice.ConfigRule) {
	for _, i := range in {
		v := convert(i)
		if v == nil {
			continue
		}
		out = append(out, v)
	}
	return
}
