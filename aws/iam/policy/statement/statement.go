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

// Filter is an interface for filtering *PolicyStatement objects
type Filter func(*policy.Statement) bool

// Statement stores necessary objects for
// filtering *PolicyStatement objects
type Statement struct {
	filters   []Filter
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
	case l > 0:
		t.Error("more than one matching statement was found")
	default:
		s.statement = statements[0]
	}

	s.filters = []Filter{}
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

	s.filters = []Filter{}
	return s
}

// Action adds the Action filter to the filter list
// the Action filter: filters *Statement objects by 'Action' where 'arn' provided
// is the expected Action value
func (s *Statement) Action(action ...string) *Statement {
	s.filters = append(s.filters, func(statement *policy.Statement) bool {
		return statement != nil && shared.StringSliceEqual(action, statement.Action)
	})
	return s
}

// Effect adds the Effect filter to the filter list
// the Effect filter: filters *Statement objects by 'Effect' where 'effect' provided
// is the expected Effect value
func (s *Statement) Effect(effect string) *Statement {
	s.filters = append(s.filters, func(statement *policy.Statement) bool {
		return strings.EqualFold(effect, statement.Effect)
	})
	return s
}

// Filter adds the 'filter' provided to the filter list
func (s *Statement) Filter(filter Filter) *Statement {
	s.filters = append(s.filters, filter)
	return s
}

// Resource adds the Resource filter to the filter list
// the Resource filter: filters *Statement objects by 'Resource' where 'resource' provided
// is the expected Resource value
func (s *Statement) Resource(resource ...string) *Statement {
	s.filters = append(s.filters, func(statement *policy.Statement) bool {
		return shared.StringSliceEqual(resource, statement.Resource)
	})
	return s
}

// Sid adds the Sid filter to the filter list
// the Sid filter: filters *Statement objects by 'Sid' where 'sid' provided
// is the expected Sid value
func (s *Statement) Sid(sid string) *Statement {
	s.filters = append(s.filters, func(statement *policy.Statement) bool {
		return strings.EqualFold(sid, statement.Sid)
	})
	return s
}

// Principal adds the Principal filter to the filter list
// the Principal filter: filters *Statement objects by 'Principal' where
// 'typ, and values' provided are the expected Principal property values
func (s *Statement) Principal(typ string, values ...string) *Statement {
	s.filters = append(s.filters, func(statement *policy.Statement) bool {
		return statement.Principal != nil &&
			strings.EqualFold(typ, statement.Principal.Type) &&
			shared.StringSliceEqual(values, statement.Principal.Values)
	})
	return s
}

// Condition adds the Condition filter to the filter list
// the Condition filter: filters *Statement objects by 'Condition' where
// 'operator, property, and value' provided are the expected Condition property values
func (s *Statement) Condition(operator string, property string, value ...string) *Statement {
	s.filters = append(s.filters, func(statement *policy.Statement) bool {
		for _, c := range statement.Condition {
			if c.Operator == operator &&
				c.Property == property &&
				shared.StringSliceEqual(c.Value, value) {
				return true
			}
		}
		return false
	})
	return s
}

func (s *Statement) filter(statements []*policy.Statement) (result []*policy.Statement, err error) {
	if len(statements) == 0 {
		statements, err = s.statements()
		if err != nil {
			return
		}
	}
outer:
	for _, statement := range statements {
		for _, f := range s.filters {
			if !f(statement) {
				continue outer
			}
		}
		result = append(result, statement)
	}
	return
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
