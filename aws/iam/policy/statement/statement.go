package statement

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/GSA/grace-tftest/aws/shared"
	"github.com/GSA/grace-tftest/aws/shared/policy"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/iam"
)

// Statement stores necessary objects for
// filtering *PolicyStatement objects
type Statement struct {
	filters   []shared.Filter
	client    client.ConfigProvider
	policy    *iam.Policy
	statement *policy.Statement
}

// New returns a new *Statement
func New(client client.ConfigProvider, policy *iam.Policy) *Statement {
	return &Statement{
		client: client,
		policy: policy,
	}
}

// Selected returns the currently selected *policy.Statement
func (s *Statement) Selected() *policy.Statement {
	return s.statement
}

// Assert executes the filter list against all *policy.Statement objects
// if none are provided, they will be gathered from the *iam.Policy provided to New()
// it will reset the filter list, fail the test if there is not exactly one match
// and store the match
func (s *Statement) Assert(t *testing.T, statements ...*policy.Statement) *Statement {
	var err error
	statements, err = s.filter(statements)
	if err != nil {
		t.Error(err)
	}

	switch l := len(statements); {
	case l == 0:
		t.Error("no matching statement was found")
	case l > 1:
		t.Error("more than one matching statement was found")
	default:
		s.statement = statements[0]
	}

	s.filters = []shared.Filter{}
	return s
}

// First executes the filter list against all *policy.Statement objects
// if none are provided, they will be gathered from the *iam.Policy provided to New()
// it will reset the filter list, fail the test if there no match, and store the first match
func (s *Statement) First(t *testing.T, statements ...*policy.Statement) *Statement {
	var err error
	statements, err = s.filter(statements)
	if err != nil {
		t.Error(err)
	}

	if len(statements) == 0 {
		t.Error("no matching statement was found")
	} else {
		s.statement = statements[0]
	}

	s.filters = []shared.Filter{}
	return s
}

// Action adds the Action filter to the filter list
// the Action filter: filters *Statement objects by 'Action' where 'arn' provided
// is the expected Action value
func (s *Statement) Action(action ...string) *Statement {
	s.filters = append(s.filters, func(v interface{}) bool {
		statement := convert(v)
		if statement == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", action, statement.Action, shared.StringSliceEqual(action, statement.Action))
		return shared.StringSliceEqual(action, statement.Action)
	})
	return s
}

// Effect adds the Effect filter to the filter list
// the Effect filter: filters *Statement objects by 'Effect' where 'effect' provided
// is the expected Effect value
func (s *Statement) Effect(effect string) *Statement {
	s.filters = append(s.filters, func(v interface{}) bool {
		statement := convert(v)
		if statement == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", effect, statement.Effect, strings.EqualFold(effect, statement.Effect))
		return strings.EqualFold(effect, statement.Effect)
	})
	return s
}

// Filter adds the 'filter' provided to the filter list
func (s *Statement) Filter(filter shared.Filter) *Statement {
	s.filters = append(s.filters, filter)
	return s
}

// Resource adds the Resource filter to the filter list
// the Resource filter: filters *Statement objects by 'Resource' where 'resource' provided
// is the expected Resource value
func (s *Statement) Resource(resource ...string) *Statement {
	s.filters = append(s.filters, func(v interface{}) bool {
		statement := convert(v)
		if statement == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", resource, statement.Resource, shared.StringSliceEqual(resource, statement.Resource))
		return shared.StringSliceEqual(resource, statement.Resource)
	})
	return s
}

// Sid adds the Sid filter to the filter list
// the Sid filter: filters *Statement objects by 'Sid' where 'sid' provided
// is the expected Sid value
func (s *Statement) Sid(sid string) *Statement {
	s.filters = append(s.filters, func(v interface{}) bool {
		statement := convert(v)
		if statement == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", sid, statement.Sid, strings.EqualFold(sid, statement.Sid))
		return strings.EqualFold(sid, statement.Sid)
	})
	return s
}

// Principal adds the Principal filter to the filter list
// the Principal filter: filters *Statement objects by 'Principal' where
// 'typ, and values' provided are the expected Principal property values
func (s *Statement) Principal(typ string, values ...string) *Statement {
	s.filters = append(s.filters, func(v interface{}) bool {
		statement := convert(v)
		if statement == nil || statement.Principal == nil {
			return false
		}
		shared.Debugf("principal.type: %s == %s -> %t\nprincipal.values: %v == %v",
			typ, statement.Principal.Type, strings.EqualFold(typ, statement.Principal.Type),
			values, statement.Principal.Values)
		return strings.EqualFold(typ, statement.Principal.Type) &&
			shared.StringSliceEqual(values, statement.Principal.Values)
	})
	return s
}

// Condition adds the Condition filter to the filter list
// the Condition filter: filters *Statement objects by 'Condition' where
// 'operator, property, and value' provided are the expected Condition property values
func (s *Statement) Condition(operator string, property string, value ...string) *Statement {
	s.filters = append(s.filters, func(v interface{}) bool {
		statement := convert(v)
		if statement == nil {
			return false
		}
		for _, c := range statement.Condition {
			shared.Debugf("operator: %s == %s -> %t\nproperty: %s == %s -> %t\nvalue: %v == %v\n",
				operator, c.Operator, operator == c.Operator,
				property, c.Property, property == c.Property,
				value, c.Value)

			if operator == c.Operator &&
				property == c.Property &&
				shared.StringSliceEqual(value, c.Value) {
				return true
			}
		}
		return false
	})
	return s
}

func (s *Statement) filter(statements []*policy.Statement) ([]*policy.Statement, error) {
	if len(statements) == 0 {
		var err error
		statements, err = s.statements()
		if err != nil {
			return nil, err
		}
	}
	return fromIface(shared.GenericFilter(s.filters, toIface(statements))), nil
}

func (s *Statement) statements() ([]*policy.Statement, error) {
	document, err := s.policyDocument(s.policy)
	if err != nil {
		return nil, err
	}
	return document.Statement, nil
}

// policyDocument ... retrieves the policy document matching the given arn and version
func (s *Statement) policyDocument(p *iam.Policy) (*policy.Document, error) {
	if p == nil {
		return nil, errors.New("policy was nil")
	}
	svc := iam.New(s.client)
	result, err := svc.GetPolicyVersion(&iam.GetPolicyVersionInput{
		PolicyArn: p.Arn,
		VersionId: p.DefaultVersionId,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to locate policy version with id: %q, for arn: %q", aws.StringValue(p.DefaultVersionId), aws.StringValue(p.Arn))
	}
	doc, err := policy.Unmarshal(aws.StringValue(result.PolicyVersion.Document))
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func convert(in interface{}) *policy.Statement {
	out, ok := in.(*policy.Statement)
	if !ok {
		shared.Debugf("object not convertible to *policy.Statement: ")
		shared.Dump(in)
		return nil
	}
	return out
}
func toIface(in []*policy.Statement) (out []interface{}) {
	for _, i := range in {
		out = append(out, i)
	}
	return
}
func fromIface(in []interface{}) (out []*policy.Statement) {
	for _, i := range in {
		v := convert(i)
		if v == nil {
			continue
		}
		out = append(out, v)
	}
	return
}
