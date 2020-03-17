package deliverychannel

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
)

func TestDeliveryChannel(t *testing.T) {
	channels := []*configservice.DeliveryChannel{
		{
			Name:         aws.String("a"),
			S3BucketName: aws.String("b"),
			S3KeyPrefix:  aws.String("c"),
			SnsTopicARN:  aws.String("d"),
			ConfigSnapshotDeliveryProperties: &configservice.ConfigSnapshotDeliveryProperties{
				DeliveryFrequency: aws.String("e"),
			},
		},
		{
			Name:         aws.String("e"),
			S3BucketName: aws.String("d"),
			S3KeyPrefix:  aws.String("c"),
			SnsTopicARN:  aws.String("b"),
			ConfigSnapshotDeliveryProperties: &configservice.ConfigSnapshotDeliveryProperties{
				DeliveryFrequency: aws.String("a"),
			},
		},
	}

	svc := New(nil)

	c1 := svc.
		Frequency("e").
		Assert(t, channels...).
		Selected()
	if aws.StringValue(c1.Name) != "a" {
		t.Fatalf("failed to match channel frequency, expected: %s, got: %s",
			"a",
			aws.StringValue(c1.Name),
		)
	}

	c2 := svc.
		Name("e").
		BucketName("d").
		BucketKeyPrefix("c").
		TopicArn("b").
		Frequency("a").
		Assert(t, channels...).
		Selected()
	if aws.StringValue(c2.Name) != "e" {
		t.Fatalf("failed to match channel frequency, expected: %s, got: %s",
			"e",
			aws.StringValue(c2.Name),
		)
	}
}
