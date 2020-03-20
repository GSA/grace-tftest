package topic

import (
	"testing"
)

func TestTopic(t *testing.T) {
	topics := []*Attributes{
		{
			TopicArn: "a",
		},
	}
	topic := New(nil).TopicArn("a").Assert(t, topics...)
	if topic == nil {
		t.Error("topic should not be nil")
	}
}
