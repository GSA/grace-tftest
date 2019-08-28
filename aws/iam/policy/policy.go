package policy

import (
	"testing"

	"github.com/GSA/grace-tftest/aws/iam/policy/statement"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/iam"
)

// Filter is an interface for filtering *iam.Policy objects
type Filter func(*iam.Policy) bool

// Policy contains the necessary properties for testing *iam.Policy objects
type Policy struct {
	client  client.ConfigProvider
	policy  *iam.Policy
	filters []Filter
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

	if len(policies) == 0 {
		t.Error("no matching policy was found")
	} else if len(policies) > 1 {
		t.Error("more than one matching policy was found")
	} else {
		p.policy = policies[0]
	}

	p.filters = []Filter{}
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

	p.filters = []Filter{}
	return p
}

// Arn adds the Arn filter to the filter list
// the Arn filter: filters policies by Arn where 'arn' provided
// is the expected Arn value
func (p *Policy) Arn(arn string) *Policy {
	p.filters = append(p.filters, func(policy *iam.Policy) bool {
		if arn == aws.StringValue(policy.Arn) {
			return true
		}
		return false
	})
	return p
}

// Filter adds the 'filter' provided to the filter list
func (p *Policy) Filter(filter Filter) *Policy {
	p.filters = append(p.filters, filter)
	return p
}

// ID adds the ID filter to the filter list
// the ID filter: filters policies by ID where 'id' provided
// is the expected PolicyId value
func (p *Policy) ID(id string) *Policy {
	p.filters = append(p.filters, func(policy *iam.Policy) bool {
		if id == aws.StringValue(policy.PolicyId) {
			return true
		}
		return false
	})
	return p
}

// Name adds the Name filter to the filter list
// the Name filter: filters policies by Name where 'name' provided
// is the expected PolicyName value
func (p *Policy) Name(name string) *Policy {
	p.filters = append(p.filters, func(policy *iam.Policy) bool {
		if name == aws.StringValue(policy.PolicyName) {
			return true
		}
		return false
	})
	return p
}

func (p *Policy) filter(policies []*iam.Policy) (result []*iam.Policy, err error) {
	if len(policies) == 0 {
		policies, err = p.policies()
		if err != nil {
			return
		}
	}
outer:
	for _, policy := range policies {
		for _, f := range p.filters {
			if !f(policy) {
				continue outer
			}
		}
		result = append(result, policy)
	}
	return
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
