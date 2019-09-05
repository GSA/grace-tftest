package role

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
)

var doc1 = `
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

func TestRole(t *testing.T) {
	roles := []*iam.Role{
		{Arn: aws.String("a"), RoleId: aws.String("b"), RoleName: aws.String("c")},
		{Arn: aws.String("d"), RoleId: aws.String("e"), RoleName: aws.String("f"), AssumeRolePolicyDocument: aws.String(doc1)},
	}
	New(nil).Arn("a").ID("b").Name("c").Assert(t, roles...)
	New(nil).Arn("d").ID("e").Name("f").Assert(t, roles...).
		Statement(t).Action("a", "b", "c").Effect("d").Resource("e").Assert(t)
}
