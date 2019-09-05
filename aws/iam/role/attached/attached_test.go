package attached

import (
	"testing"

	"github.com/GSA/grace-tftest/aws/shared/policy"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
)

const testpolicy = `
{
	"Version": "2012-10-17",
	"Statement": [
	  {
		"Sid": "a",
		"Action": [ "b" ],
		"Effect": "c",
		"Resource": [ "d", "e", "f" ]
	  },
	  {
		"Action": [ "a", "b", "c" ],
		"Effect": "d",
		"Resource": "e"
	  }
	]
  }
`

func TestAttached(t *testing.T) {
	policies := []*iam.AttachedPolicy{
		{PolicyArn: aws.String("a"), PolicyName: aws.String("b")},
		{PolicyArn: aws.String("c"), PolicyName: aws.String("d")},
	}
	doc, err := policy.Unmarshal(testpolicy)
	if err != nil {
		t.Fatalf("failed to unmarshal test policy: %v", err)
	}

	New(nil, "").Arn("a").Name("b").Assert(t, policies...)
	stmt := New(nil, "").Arn("c").Name("d").Assert(t, policies...).
		Statement(t, doc)

	stmt.Sid("a").Action("b").Effect("c").Resource("d", "e", "f").Assert(t)
	stmt.Action("a", "b", "c").Effect("d").Resource("e").Assert(t)
}
