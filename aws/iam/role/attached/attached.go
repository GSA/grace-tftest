package attached

import (
	"testing"

	"github.com/GSA/grace-tftest/aws/shared"
	"github.com/GSA/grace-tftest/aws/shared/policy"
	"github.com/GSA/grace-tftest/aws/shared/policy/statement"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/iam"
)

// Attached contains the necessary properties for testing *iam.AttachedPolicy objects
type Attached struct {
	client   client.ConfigProvider
	attached *iam.AttachedPolicy
	roleName string
	filters  []shared.Filter
}

// New returns a new *Attached
func New(client client.ConfigProvider, roleName string) *Attached {
	return &Attached{client: client, roleName: roleName}
}

// Selected returns the currently selected *iam.AttachedPolicy
func (a *Attached) Selected() *iam.AttachedPolicy {
	return a.attached
}

// Statement returns a newly instantiated *statement.Statement object
// this is used for filtering by statements inside a policy. If doc is nil
// the default policy document will be retrieved from AWS
func (a *Attached) Statement(t *testing.T, doc *policy.Document) *statement.Statement {
	if doc == nil {
		doc = a.Document(t, "")
	}
	return statement.New(doc)
}

// Assert applies all filters that have been called, resets the list of filters,
// fails the test if there is not exactly one match, and stores the matched policy
// if policies is not provided, *iam.AttachedPolicy objects will be retreived from AWS
func (a *Attached) Assert(t *testing.T, policies ...*iam.AttachedPolicy) *Attached {
	var err error
	policies, err = a.filter(policies)
	if err != nil {
		t.Fatal(err)
	}

	switch l := len(policies); {
	case l == 0:
		t.Fatal("no matching attached policy was found")
	case l > 1:
		t.Fatal("more than one matching attached policy was found")
	default:
		a.attached = policies[0]
	}

	a.filters = []shared.Filter{}
	return a
}

// First applies all filters that have been called, resets the list of filters,
// fails the test if there are no matches, and stores the first match
// if policies is not provided, *iam.AttachedPolicy objects will be retreived from AWS
func (a *Attached) First(t *testing.T, policies ...*iam.AttachedPolicy) *Attached {
	var err error
	policies, err = a.filter(policies)
	if err != nil {
		t.Fatal(err)
	}

	if len(policies) == 0 {
		t.Fatal("no matching attached policy was found")
	} else {
		a.attached = policies[0]
	}

	a.filters = []shared.Filter{}
	return a
}

// Arn adds the Arn filter to the filter list
// the Arn filter: filters policies by Arn where 'arn' provided
// is the expected Arn value
func (a *Attached) Arn(arn string) *Attached {
	a.filters = append(a.filters, func(v interface{}) bool {
		policy := convert(v)
		if policy == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", arn, aws.StringValue(policy.PolicyArn), arn == aws.StringValue(policy.PolicyArn))
		return arn == aws.StringValue(policy.PolicyArn)
	})
	return a
}

// Filter adds the 'filter' provided to the filter list
func (a *Attached) Filter(filter shared.Filter) *Attached {
	a.filters = append(a.filters, filter)
	return a
}

// Name adds the Name filter to the filter list
// the Name filter: filters policies by Name where 'name' provided
// is the expected PolicyName value
func (a *Attached) Name(name string) *Attached {
	a.filters = append(a.filters, func(v interface{}) bool {
		policy := convert(v)
		if policy == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", name, aws.StringValue(policy.PolicyName), name == aws.StringValue(policy.PolicyName))
		return name == aws.StringValue(policy.PolicyName)
	})
	return a
}

// Document returns the unmarshaled policy document
// if versionID is empty, will return the default version
func (a *Attached) Document(t *testing.T, versionID string) *policy.Document {
	if a.attached == nil {
		t.Errorf("attached policy was nil")
		return nil
	}
	svc := iam.New(a.client)
	output, err := svc.GetPolicy(&iam.GetPolicyInput{
		PolicyArn: a.attached.PolicyArn,
	})
	if err != nil {
		t.Errorf("failed to get policy with Arn: %s -> %v", aws.StringValue(a.attached.PolicyArn), err)
		return nil
	}
	input := &iam.GetPolicyVersionInput{
		PolicyArn: output.Policy.Arn,
		VersionId: output.Policy.DefaultVersionId,
	}
	if len(versionID) > 0 {
		input.VersionId = &versionID
	}
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

func (a *Attached) filter(policies []*iam.AttachedPolicy) ([]*iam.AttachedPolicy, error) {
	if len(policies) == 0 {
		var err error
		policies, err = a.policies()
		if err != nil {
			return nil, err
		}
	}
	return fromIface(shared.GenericFilter(a.filters, toIface(policies))), nil
}

func (a *Attached) policies() ([]*iam.AttachedPolicy, error) {
	svc := iam.New(a.client)
	var policies []*iam.AttachedPolicy
	err := svc.ListAttachedRolePoliciesPages(&iam.ListAttachedRolePoliciesInput{
		RoleName: aws.String(a.roleName),
	}, func(out *iam.ListAttachedRolePoliciesOutput, lastPage bool) bool {
		policies = append(policies, out.AttachedPolicies...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	return policies, nil
}

func convert(in interface{}) *iam.AttachedPolicy {
	out, ok := in.(*iam.AttachedPolicy)
	if !ok {
		shared.Debugf("object not convertible to *iam.AttachedPolicy: ")
		shared.Dump(in)
		return nil
	}
	return out
}
func toIface(in []*iam.AttachedPolicy) (out []interface{}) {
	for _, i := range in {
		out = append(out, i)
	}
	return
}
func fromIface(in []interface{}) (out []*iam.AttachedPolicy) {
	for _, i := range in {
		v := convert(i)
		if v == nil {
			continue
		}
		out = append(out, v)
	}
	return
}
