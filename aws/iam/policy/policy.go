package policy

import (
	"testing"

	"github.com/GSA/grace-tftest/aws/iam/policy/statement"
	"github.com/GSA/grace-tftest/aws/shared"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/iam"
)

// Policy contains the necessary properties for testing *iam.Policy objects
type Policy struct {
	client  client.ConfigProvider
	policy  *iam.Policy
	filters []shared.Filter
}

// New returns a new *Policy
func New(client client.ConfigProvider) *Policy {
	return &Policy{client: client}
}

// Selected returns the currently selected *iam.Policy
func (p *Policy) Selected() *iam.Policy {
	return p.policy
}

// Statement returns a newly instantiated *statement.Statement object
// this is used for filtering by statements inside a policy
func (p *Policy) Statement(t *testing.T) *statement.Statement {
	return statement.New(p.client, p.policy)
}

// Assert applies all filters that have been called, resets the list of filters,
// fails the test if there is not exactly one match, and stores the matched policy
// if policies is not provided, *iam.Policy objects will be retreived from AWS
func (p *Policy) Assert(t *testing.T, policies ...*iam.Policy) *Policy {
	var err error
	policies, err = p.filter(policies)
	if err != nil {
		t.Error(err)
	}

	switch l := len(policies); {
	case l == 0:
		t.Error("no matching policy was found")
	case l > 1:
		t.Error("more than one matching policy was found")
	default:
		p.policy = policies[0]
	}

	p.filters = []shared.Filter{}
	return p
}

// First applies all filters that have been called, resets the list of filters,
// fails the test if there are no matches, and stores the first match
// if policies is not provided, *iam.Policy objects will be retreived from AWS
func (p *Policy) First(t *testing.T, policies ...*iam.Policy) *Policy {
	var err error
	policies, err = p.filter(policies)
	if err != nil {
		t.Error(err)
	}

	if len(policies) == 0 {
		t.Error("no matching policy was found")
	} else {
		p.policy = policies[0]
	}

	p.filters = []shared.Filter{}
	return p
}

// Arn adds the Arn filter to the filter list
// the Arn filter: filters policies by Arn where 'arn' provided
// is the expected Arn value
func (p *Policy) Arn(arn string) *Policy {
	p.filters = append(p.filters, func(v interface{}) bool {
		policy := convert(v)
		if policy == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", arn, aws.StringValue(policy.Arn), arn == aws.StringValue(policy.Arn))
		return arn == aws.StringValue(policy.Arn)
	})
	return p
}

// Filter adds the 'filter' provided to the filter list
func (p *Policy) Filter(filter shared.Filter) *Policy {
	p.filters = append(p.filters, filter)
	return p
}

// ID adds the ID filter to the filter list
// the ID filter: filters policies by ID where 'id' provided
// is the expected PolicyId value
func (p *Policy) ID(id string) *Policy {
	p.filters = append(p.filters, func(v interface{}) bool {
		policy := convert(v)
		if policy == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", id, aws.StringValue(policy.PolicyId), id == aws.StringValue(policy.PolicyId))
		return id == aws.StringValue(policy.PolicyId)
	})
	return p
}

// Name adds the Name filter to the filter list
// the Name filter: filters policies by Name where 'name' provided
// is the expected PolicyName value
func (p *Policy) Name(name string) *Policy {
	p.filters = append(p.filters, func(v interface{}) bool {
		policy := convert(v)
		if policy == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", name, aws.StringValue(policy.PolicyName), name == aws.StringValue(policy.PolicyName))
		return name == aws.StringValue(policy.PolicyName)
	})
	return p
}

func (p *Policy) filter(policies []*iam.Policy) ([]*iam.Policy, error) {
	if len(policies) == 0 {
		var err error
		policies, err = p.policies()
		if err != nil {
			return nil, err
		}
	}
	return fromIface(shared.GenericFilter(p.filters, toIface(policies))), nil
}

func (p *Policy) policies() ([]*iam.Policy, error) {
	svc := iam.New(p.client)
	var policies []*iam.Policy
	err := svc.ListPoliciesPages(&iam.ListPoliciesInput{}, func(out *iam.ListPoliciesOutput, lastPage bool) bool {
		policies = append(policies, out.Policies...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	return policies, nil
}

func convert(v interface{}) *iam.Policy {
	statement, ok := v.(*iam.Policy)
	if !ok {
		shared.Debugf("object not convertible to *iam.Policy: ")
		shared.Dump(v)
		return nil
	}
	return statement
}
func toIface(in []*iam.Policy) (out []interface{}) {
	for _, i := range in {
		out = append(out, i)
	}
	return
}
func fromIface(in []interface{}) (out []*iam.Policy) {
	for _, i := range in {
		v := convert(i)
		if v == nil {
			continue
		}
		out = append(out, v)
	}
	return
}
