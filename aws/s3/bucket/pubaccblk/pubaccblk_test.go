package pubaccblk

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

func TestPublicAccessBlock(t *testing.T) {
	configs := []*s3.PublicAccessBlockConfiguration{{
		BlockPublicAcls:       aws.Bool(false),
		BlockPublicPolicy:     aws.Bool(false),
		IgnorePublicAcls:      aws.Bool(false),
		RestrictPublicBuckets: aws.Bool(false),
	}}

	New(nil, "").BlockPublicAcls(false).BlockPublicPolicy(false).IgnorePublicAcls(false).RestrictPublicBuckets(false).Assert(t, configs...)
}
