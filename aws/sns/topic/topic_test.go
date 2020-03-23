package topic

import (
	"testing"
)

func TestTopic(t *testing.T) {
	topics := []*Attributes{
		{
			TopicArn:       "a",
			DisplayName:    "b",
			Owner:          "c",
			KmsMasterKeyID: "d",
		},
	}
	topic := New(nil).TopicArn("a").Arn("a").DisplayName("b").Name("b").Owner("c").KmsMasterKeyID("d").Assert(t, topics...)
	if topic == nil {
		t.Error("topic should not be nil")
	}
}
