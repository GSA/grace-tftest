package lifecycle

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

func TestLifecycle(t *testing.T) {
	rules := []*s3.LifecycleRule{
		{
			ID: aws.String("a"),
			Filter: &s3.LifecycleRuleFilter{
				Prefix: aws.String("b"),
			},
			Expiration: &s3.LifecycleExpiration{
				Days: aws.Int64(int64(1)),
			},
			Status: aws.String("c"),
		},
		{
			ID: aws.String("a"),
			Filter: &s3.LifecycleRuleFilter{
				And: &s3.LifecycleRuleAndOperator{
					Prefix: aws.String("b"),
					Tags: []*s3.Tag{
						{Key: aws.String("c"), Value: aws.String("d")},
						{Key: aws.String("e"), Value: aws.String("f")},
					},
				},
			},
			Expiration: &s3.LifecycleExpiration{
				Days: aws.Int64(int64(1)),
			},
			Status: aws.String("g"),
		},
		{
			ID: aws.String("a"),
			Filter: &s3.LifecycleRuleFilter{
				Tag: &s3.Tag{
					Key:   aws.String("b"),
					Value: aws.String("c"),
				},
			},
			Expiration: &s3.LifecycleExpiration{
				Days: aws.Int64(int64(1)),
			},
			Status: aws.String("d"),
		},
	}

	New(nil, "").Method("a").IsExp().ExpDays(1).Status("c").FilterPrefix("b").Assert(t, rules...)
	New(nil, "").Method("a").ExpDays(1).Status("d").FilterTag("b", "c").Assert(t, rules...)

	New(nil, "").Method("a").ExpDays(1).Status("g").
		FilterAnd("b",
			&s3.Tag{Key: aws.String("c"), Value: aws.String("d")},
			&s3.Tag{Key: aws.String("e"), Value: aws.String("f")}).
		Assert(t, rules...)
}
