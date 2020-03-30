package group

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

func TestGroup(t *testing.T) {
	groups := []*cloudwatchlogs.LogGroup{
		{
			LogGroupName:    aws.String("a"),
			Arn:             aws.String("b"),
			KmsKeyId:        aws.String("c"),
			RetentionInDays: aws.Int64(1),
		},
	}
	group := New(nil).LogGroupName("a").Arn("b").KmsKeyID("c").RetentionInDays(1).Assert(t, groups...)
	if group == nil {
		t.Error("group should not be nil")
	}
}
