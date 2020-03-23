package trail

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudtrail"
)

func TestTrail(t *testing.T) {
	trails := []*cloudtrail.Trail{
		{
			TrailARN:                   aws.String("a"),
			Name:                       aws.String("b"),
			KmsKeyId:                   aws.String("c"),
			CloudWatchLogsLogGroupArn:  aws.String("d"),
			CloudWatchLogsRoleArn:      aws.String("e"),
			HasCustomEventSelectors:    aws.Bool(false),
			HasInsightSelectors:        aws.Bool(false),
			HomeRegion:                 aws.String("f"),
			IncludeGlobalServiceEvents: aws.Bool(false),
			IsMultiRegionTrail:         aws.Bool(false),
			IsOrganizationTrail:        aws.Bool(false),
			LogFileValidationEnabled:   aws.Bool(false),
			S3BucketName:               aws.String("g"),
			S3KeyPrefix:                aws.String("h"),
			SnsTopicARN:                aws.String("i"),
		},
	}
	trail := New(nil).TrailARN("a").ARN("a").Name("b").KmsKeyID("c").
		CloudWatchLogsLogGroupArn("d").CloudWatchLogsRoleArn("e").
		HasCustomEventSelectors(false).HasInsightSelectors(false).HomeRegion("f").
		IncludeGlobalServiceEvents(false).IsMultiRegionTrail(false).
		IsOrganizationTrail(false).LogFileValidationEnabled(false).
		S3BucketName("g").S3KeyPrefix("h").SnsTopicARN("i").
		Assert(t, trails...)
	if trail == nil {
		t.Error("trail should not be nil")
	}
}
