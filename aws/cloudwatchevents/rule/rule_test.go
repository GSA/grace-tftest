package rule

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
)

func TestRule(t *testing.T) {
	rules := []*cloudwatchevents.Rule{
		{
			Arn:                aws.String("a"),
			ManagedBy:          aws.String("b"),
			Name:               aws.String("c"),
			RoleArn:            aws.String("d"),
			ScheduleExpression: aws.String("e"),
			State:              aws.String("f"),
		},
	}
	target := New(nil).Arn("a").ManagedBy("b").Name("c").RoleArn("d").SchedExpr("e").State("f").Assert(t, rules...).Target()
	if target == nil {
		t.Error("target should not be nil")
	}
}
