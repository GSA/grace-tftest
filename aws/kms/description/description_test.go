package description

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
)

func TestDescription(t *testing.T) {
	descriptions := []*kms.KeyMetadata{
		{Arn: aws.String("a"), Description: aws.String("b"), KeyId: aws.String("c")},
		{Arn: aws.String("d"), Description: aws.String("e"), KeyId: aws.String("f")},
		{Arn: aws.String("g"), Description: aws.String("h"), KeyId: aws.String("c")},
	}

	New(nil).Arn("a").Name("b").ID("c").Assert(t, descriptions...)
	New(nil).ID("f").Assert(t, descriptions...)
	description := New(nil).ID("c").First(t, descriptions...).Selected()
	if description == nil {
		t.Errorf("description was nil")
	}
}
