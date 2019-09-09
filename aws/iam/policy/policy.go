package policy

import (
	"strings"
	"testing"

	"github.com/GSA/grace-tftest/aws/shared"
	"github.com/GSA/grace-tftest/aws/shared/policy"
	"github.com/GSA/grace-tftest/aws/shared/policy/statement"
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
// this is used for filtering by statements inside a policy. If doc is nil
// the default policy document will be retrieved from AWS
func (p *Policy) Statement(t *testing.T, doc *policy.Document) *statement.Statement {
	if doc == nil {
		doc = p.Document(t, "")
	}
	return statement.New(doc)
}

// Assert applies all filters that have been called, resets the list of filters,
// fails the test if there is not exactly one match, and stores the matched policy
// if policies is not provided, *iam.Policy objects will be retreived from AWS
func (p *Policy) Assert(t *testing.T, policies ...*iam.Policy) *Policy {
	var err error
	policies, err = p.filter(policies)
	if err != nil {
		t.Fatal(err)
	}

	switch l := len(policies); {
	case l == 0:
		t.Fatal("no matching policy was found")
	case l > 1:
		t.Fatal("more than one matching policy was found")
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
		t.Fatal(err)
	}

	if len(policies) == 0 {
		t.Fatal("no matching policy was found")
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
		shared.Debugf("%s like %s -> %t\n", name, aws.StringValue(policy.PolicyName), strings.EqualFold(name, aws.StringValue(policy.PolicyName)))
		return strings.EqualFold(name, aws.StringValue(policy.PolicyName))
	})
	return p
}

// Document returns the unmarshaled policy document
// if versionID is empty, will return the default version
func (p *Policy) Document(t *testing.T, versionID string) *policy.Document {
	if p.policy == nil {
		t.Errorf("policy was nil")
		return nil
	}
	input := &iam.GetPolicyVersionInput{
		PolicyArn: p.policy.Arn,
		VersionId: p.policy.DefaultVersionId,
	}
	if len(versionID) > 0 {
		input.VersionId = &versionID
	}
	svc := iam.New(p.client)
	result, err := svc.GetPolicyVersion(input)
	if err != nil {
		t.Errorf("failed to locate policy version with id: %q, for arn: %q", aws.StringValue(input.VersionId), aws.StringValue(input.PolicyArn))
		return nil
	}
	doc, err := policy.Unmarshal(aws.StringValue(result.PolicyVersion.Document))
	if err != nil {
		t.Errorf("failed to unmarshal policy document: %v", err)
		return nil
	}
	return doc
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

func convert(in interface{}) *iam.Policy {
	out, ok := in.(*iam.Policy)
	if !ok {
		shared.Debugf("object not convertible to *iam.Policy: ")
		shared.Dump(in)
		return nil
	}
	return out
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
