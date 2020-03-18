package policy

import (
	"testing"

	"github.com/GSA/grace-tftest/aws/shared/policy"
	"github.com/aws/aws-sdk-go/aws"
)

func TestPolicy(t *testing.T) {
	doc := &policy.Document{
		Statement: []*policy.Statement{
			{Sid: "a", Effect: "b", Action: []string{"c"}, Resource: []string{"c"}, Principal: &policy.Principal{
				Type: "a", Values: []string{"d"}}},
			{Sid: "a", Effect: "b", Action: []string{"d"}, Resource: []string{"d"}, Principal: &policy.Principal{
				Type: "a", Values: []string{"e"}}},
			{Sid: "a", Effect: "b", Action: []string{"e"}, Resource: []string{"d"}, Principal: &policy.Principal{
				Type: "a", Values: []string{"e"}}},
			{Sid: "a", Effect: "b", Action: []string{"f"}, Resource: []string{"f"}, Principal: &policy.Principal{
				Type: "a", Values: []string{"g"}}},
			{Sid: "a", Effect: "b", Action: []string{"g"}, Resource: []string{"g"}, Principal: &policy.Principal{
				Type: "a", Values: []string{"h"}}},
		},
	}
	New(aws.String("")).Statement(t, doc).Sid("a").Effect("b").Action("c").Assert(t)
	s := New(aws.String("")).Statement(t, doc).Sid("a").Effect("b").Resource("d").First(t).Selected()
	if s == nil {
		t.Errorf("statement was nil")
	}
}
