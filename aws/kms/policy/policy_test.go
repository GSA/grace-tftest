package policy

import (
	"testing"

	"github.com/GSA/grace-tftest/aws/shared/policy"
)

func TestPolicy(t *testing.T) {
	doc := &policy.Document{
		Statement: []*policy.Statement{
			{Sid: "a", Effect: "b", Action: []string{"c"}, Resource: []string{"c"}},
			{Sid: "a", Effect: "b", Action: []string{"d"}, Resource: []string{"d"}},
			{Sid: "a", Effect: "b", Action: []string{"e"}, Resource: []string{"d"}},
			{Sid: "a", Effect: "b", Action: []string{"f"}, Resource: []string{"f"}},
			{Sid: "a", Effect: "b", Action: []string{"g"}, Resource: []string{"g"}},
		},
	}
	New(nil, "").Statement(t, doc).Sid("a").Effect("b").Action("c").Assert(t)
	s := New(nil, "").Statement(t, doc).Sid("a").Effect("b").Resource("d").First(t).Selected()
	if s == nil {
		t.Errorf("statement was nil")
	}
}
