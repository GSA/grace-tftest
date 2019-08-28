package policy

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
)

func TestPolicy(t *testing.T) {
	policies := []*iam.Policy{
		{Arn: aws.String("a"), PolicyName: aws.String("b"), PolicyId: aws.String("c")},
		{Arn: aws.String("a"), PolicyName: aws.String("b"), PolicyId: aws.String("d")},
		{Arn: aws.String("a"), PolicyName: aws.String("b"), PolicyId: aws.String("e")},
	}

	expected := "c"
	result := New(nil).Arn("a").Name("b").ID("c").Assert(t, policies...).Selected()
	if result == nil {
		t.Fatal("policy was nil")
	}
	if aws.StringValue(result.PolicyId) != expected {
		t.Errorf("policyID invalid, expected: %q, got: %q", expected, aws.StringValue(result.PolicyId))
	}
	result = New(nil).Arn("a").Name("b").First(t, policies...).Selected()
	if result == nil {
		t.Fatal("policy was nil")
	}
	if aws.StringValue(result.PolicyId) != expected {
		t.Errorf("policyID invalid, expected: %q, got: %q", expected, aws.StringValue(result.PolicyId))
	}
}
