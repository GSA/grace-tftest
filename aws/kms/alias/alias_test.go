package alias

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
)

func TestAlias(t *testing.T) {
	aliases := []*kms.AliasListEntry{
		{AliasArn: aws.String("a"), AliasName: aws.String("b"), TargetKeyId: aws.String("c")},
		{AliasArn: aws.String("d"), AliasName: aws.String("e"), TargetKeyId: aws.String("f")},
		{AliasArn: aws.String("g"), AliasName: aws.String("h"), TargetKeyId: aws.String("c")},
	}

	New(nil).Arn("a").Name("b").ID("c").Assert(t, aliases...)
	New(nil).ID("f").Assert(t, aliases...)
	alias := New(nil).ID("c").First(t, aliases...).Selected()
	if alias == nil {
		t.Errorf("alias was nil")
	}
}
