package key

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
)

func TestKey(t *testing.T) {
	keys := []*kms.KeyMetadata{
		{Arn: aws.String("a"), Description: aws.String("b"), KeyId: aws.String("c")},
		{Arn: aws.String("d"), Description: aws.String("e"), KeyId: aws.String("f")},
		{Arn: aws.String("g"), Description: aws.String("h"), KeyId: aws.String("c")},
	}

	New(nil).Arn("a").Description("b").ID("c").Assert(t, keys...)
	New(nil).ID("f").Assert(t, keys...)
	key := New(nil).ID("c").First(t, keys...).Selected()
	if key == nil {
		t.Errorf("key was nil")
	}
}
