package notification

import "testing"

func TestNotification(t *testing.T) {
	configs := []*Configuration{
		{
			Arn:    "a",
			ID:     "b",
			Events: []string{"c"},
			Filter: []*FilterRule{
				{Name: "d", Value: "e"},
			},
		},
		{
			Arn:    "a",
			ID:     "b",
			Events: []string{"c", "d"},
			Filter: []*FilterRule{
				{Name: "e", Value: "f"},
				{Name: "g", Value: "h"},
			},
		},
	}

	New(nil, "").Arn("a").ID("b").Events("c").Rule("d", "e").Assert(t, configs...)
	New(nil, "").Arn("a").ID("b").Events("c", "d").Rule("e", "f").Rule("g", "h").Assert(t, configs...)
}
