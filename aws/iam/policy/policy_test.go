package policy

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
)

func TestPolicy(t *testing.T) {
	policies := []*iam.Policy{
		{Arn: aws.String("a"), PolicyId: aws.String("b"), PolicyName: aws.String("c")},
		{Arn: aws.String("a"), PolicyId: aws.String("b"), PolicyName: aws.String("d")},
		{Arn: aws.String("a"), PolicyId: aws.String("b"), PolicyName: aws.String("e")},
	}

	expected := "c"
	result := New(nil).Arn("a").ID("b").Name("c").Assert(t, policies...).Selected()
	if result == nil {
		t.Fatal("policy was nil")
	}
	if aws.StringValue(result.PolicyName) != expected {
		t.Errorf("policyName invalid, expected: %q, got: %q", expected, aws.StringValue(result.PolicyName))
	}
	result = New(nil).Arn("a").ID("b").First(t, policies...).Selected()
	if result == nil {
		t.Fatal("policy was nil")
	}
	if aws.StringValue(result.PolicyName) != expected {
		t.Errorf("policyName invalid, expected: %q, got: %q", expected, aws.StringValue(result.PolicyName))
	}
}
