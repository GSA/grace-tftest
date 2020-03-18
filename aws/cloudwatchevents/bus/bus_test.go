package bus

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
)

func TestBus(t *testing.T) {
	buses := []*cloudwatchevents.EventBus{
		{
			Arn:    aws.String("a"),
			Name:   aws.String("c"),
			Policy: aws.String("{}"),
		},
	}
	target := New(nil).Arn("a").Name("c").Assert(t, buses...).Policy()
	if target == nil {
		t.Error("policy should not be nil")
	}
}
