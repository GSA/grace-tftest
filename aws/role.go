package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/iam"
	terratest "github.com/gruntwork-io/terratest/modules/aws"
)

// MatchIamRolePolicy ... retrieves policy document for role with given name
// then finds a policy statement that matches using the given matcher
func MatchIamRolePolicy(t *testing.T, region string, name string, matcher func(*PolicyStatement) bool) *PolicyStatement {
	statement, err := MatchIamRolePolicyE(region, name, matcher)
	if err != nil {
		t.Fatal(err)
	}
	return statement
}

// MatchIamRolePolicyE ... retrieves policy document for role with given name
// then finds a policy statement that matches using the given matcher
func MatchIamRolePolicyE(region string, name string, matcher func(*PolicyStatement) bool) (*PolicyStatement, error) {
	client, err := terratest.NewIamClientE(nil, region)
	if err != nil {
		return nil, err
	}
	out, err := client.GetRole(&iam.GetRoleInput{
		RoleName: &name,
	})
	if err != nil {
		return nil, err
	}
	doc, err := UnmarshalPolicy(*out.Role.AssumeRolePolicyDocument)
	if err != nil {
		return nil, err
	}
	statement, err := doc.Find(matcher)
	if err != nil {
		return nil, err
	}
	return statement, nil
}

// FindIamAttachedRolePolicyByName ... finds an attached role policy with the given name
func FindIamAttachedRolePolicyByName(t *testing.T, region string, role string, policyName string) *iam.AttachedPolicy {
	policy, err := FindIamAttachedRolePolicyByNameE(region, role, policyName)
	if err != nil {
		t.Fatalf("FindIamAttachedRolePolicyByName failed: %v", err)
	}
	return policy
}

// FindIamAttachedRolePolicyByNameE ... finds an attached role policy with the given name
func FindIamAttachedRolePolicyByNameE(region string, role string, policyName string) (*iam.AttachedPolicy, error) {
	policy, err := FindIamAttachedRolePolicyE(region, role, func(p *iam.AttachedPolicy) bool {
		return *p.PolicyName == policyName
	})
	if err != nil {
		return nil, err
	}
	return policy, nil
}

// FindIamAttachedRolePolicy ... finds an attached role policy with the given matcher
func FindIamAttachedRolePolicy(t *testing.T, region string, role string, matcher func(*iam.AttachedPolicy) bool) *iam.AttachedPolicy {
	policy, err := FindIamAttachedRolePolicyE(region, role, matcher)
	if err != nil {
		t.Fatalf("FindIamAttachedRolePolicy failed: %v", err)
	}
	return policy
}

// FindIamAttachedRolePolicyE ... finds an attached role policy with the given matcher
func FindIamAttachedRolePolicyE(region string, role string, matcher func(*iam.AttachedPolicy) bool) (*iam.AttachedPolicy, error) {
	var (
		err    error
		marker *string
	)

	more := true
	for more {
		var policies []*iam.AttachedPolicy
		policies, marker, err = ListIamAttachedRolePoliciesE(region, role, marker)
		if err != nil {
			return nil, err
		}
		for _, p := range policies {
			if matcher(p) {
				return p, nil
			}
		}
		if marker == nil {
			more = false
		}
	}

	return nil, fmt.Errorf("failed to locate a matching attached role policy")
}

// ListIamAttachedRolePolicies ... retrieves a batch of attached policies with the given role
// starting at marker, on the first call marker should be nil
func ListIamAttachedRolePolicies(t *testing.T, region string, role string, marker *string) ([]*iam.AttachedPolicy, *string) {
	policies, next, err := ListIamAttachedRolePoliciesE(region, role, marker)
	if err != nil {
		t.Fatalf("ListIamAttachedRolePolicies failed: %v", err)
	}
	return policies, next
}

// ListIamAttachedRolePoliciesE ... retrieves a batch of attached policies with the given role
// starting at marker, on the first call marker should be nil
func ListIamAttachedRolePoliciesE(region string, role string, marker *string) ([]*iam.AttachedPolicy, *string, error) {
	client, err := terratest.NewIamClientE(nil, region)
	if err != nil {
		return nil, nil, err
	}
	result, err := client.ListAttachedRolePolicies(&iam.ListAttachedRolePoliciesInput{
		RoleName: &role,
		Marker:   marker,
	})
	if err != nil {
		return nil, nil, err
	}
	return result.AttachedPolicies, result.Marker, nil
}
