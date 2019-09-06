package target

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
)

func TestTarget(t *testing.T) {
	targets := []*cloudwatchevents.Target{
		{
			Arn:     aws.String("a"),
			Id:      aws.String("b"),
			RoleArn: aws.String("c"),
		},
	}
	New(nil, nil).Arn("a").ID("b").RoleArn("c").Assert(t, targets...)
}
