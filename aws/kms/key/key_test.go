package key

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
)

func TestKey(t *testing.T) {
	keys := []*kms.KeyMetadata{
		{Arn: aws.String("a"), Description: aws.String("b"), KeyId: aws.String("c"), Enabled: aws.Bool(true)},
		{Arn: aws.String("d"), Description: aws.String("e"), KeyId: aws.String("f"), Enabled: aws.Bool(true)},
		{Arn: aws.String("g"), Description: aws.String("h"), KeyId: aws.String("c"), Enabled: aws.Bool(true)},
		{Arn: aws.String("g"), Description: aws.String("h"), KeyId: aws.String("c"), Enabled: aws.Bool(false)},
	}

	t.Run("assert multiple filters", func(t *testing.T) {
		key := New(nil).Arn("a").Description("b").ID("c").Assert(t, keys...).Selected()
		if key == nil {
			t.Errorf("kms/key failed. Key was nil")
		}
	})

	t.Run("assert unique id", func(t *testing.T) {
		key := New(nil).ID("f").Assert(t, keys...).Selected()
		if key == nil {
			t.Errorf("kms/key failed. Key was nil")
		}
	})
	t.Run("first", func(t *testing.T) {
		key := New(nil).ID("c").Enabled(true).First(t, keys...).Selected()
		if key == nil {
			t.Errorf("kms/key failed. Key was nil")
		}
	})
	t.Run("assert not enabled", func(t *testing.T) {
		key := New(nil).ID("c").Enabled(false).Assert(t, keys...).Selected()
		if key == nil {
			t.Errorf("kms/key failed. Key was nil")
		}
	})
}
