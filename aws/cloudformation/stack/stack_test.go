package stack

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

func TestStack(t *testing.T) {
	stacks := []*cloudformation.Stack{
		{
			StackId:     aws.String("a"),
			ChangeSetId: aws.String("b"),
			StackName:   aws.String("c"),
			RoleARN:     aws.String("d"),
			Description: aws.String("e"),
			ParentId:    aws.String("f"),
			RootId:      aws.String("g"),
			StackStatus: aws.String("h"),
		},
	}
	target := New(nil).StackID("a").ChangeSetID("b").Name("c").RoleARN("d").
		Description("e").ParentID("f").RootID("g").StackStatus("h").Assert(t, stacks...)
	if target == nil {
		t.Error("target should not be nil")
	}
}
