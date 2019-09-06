package policy

import (
	"fmt"
	"testing"

	"github.com/GSA/grace-tftest/aws/shared"
)

const unmarshaltest1 = `
{
	"Version": "a",
	"Statement": [
	  {
		"Sid": "a",
		"Action": [ "b", "c" ],
		"Effect": "d",
		"Principal":{ "e": "f" },
		"Resource": [ "g", "h" ],
		"Condition": {"i":{"j":["k"]}, "l":{"m":["n", "o", "p"]}}
	  }
	]
  }
`

// nolint: gocyclo
func TestUnmarshal(t *testing.T) {
	got, err := Unmarshal(unmarshaltest1)
	if err != nil {
		t.Fatalf("failed to unmarshal policy: %v", err)
	}

	exp := &Document{
		Version: "a",
		Statement: []*Statement{
			{
				Sid:       "a",
				Action:    []string{"b", "c"},
				Effect:    "d",
				Principal: &Principal{Type: "e", Values: []string{"f"}},
				Resource:  []string{"g", "h"},
				Condition: []*Condition{
					{Operator: "i", Property: "j", Value: []string{"k"}},
					{Operator: "l", Property: "m", Value: []string{"n", "o", "p"}},
				},
			},
		},
	}

	if got.Version != exp.Version {
		t.Errorf("policy versions do no match, expected: %s, got: %s", exp.Version, got.Version)
	}
	if len(got.Statement) != len(exp.Statement) {
		t.Errorf("statement lengths do not match, expected: %d, got: %d", len(exp.Statement), len(got.Statement))
	}
	for i, g := range got.Statement {
		if g.Sid != exp.Statement[i].Sid {
			t.Errorf("policy.statement[%d].sid does not match, expected: %s, got: %s", i, exp.Statement[i].Sid, g.Sid)
		}
		if !shared.StringSliceEqual(g.Action, exp.Statement[i].Action) {
			t.Errorf("policy.statement[%d].action does not match, expected: %v, got: %v", i, exp.Statement[i].Action, g.Action)
		}
		if g.Effect != exp.Statement[i].Effect {
			t.Errorf("policy.statement[%d].effect does not match, expected: %s, got: %s", i, exp.Statement[i].Effect, g.Effect)
		}
		if g.Principal.Type != exp.Statement[i].Principal.Type {
			t.Errorf("policy.statement[%d].principal.type does not match, expected: %s, got: %s", i, exp.Statement[i].Principal.Type, g.Principal.Type)
		}
		if !shared.StringSliceEqual(g.Principal.Values, exp.Statement[i].Principal.Values) {
			t.Errorf("policy.statement[%d].principal.values do not match, expected: %v, got: %v", i, exp.Statement[i].Principal.Values, g.Principal.Values)
		}
		if !shared.StringSliceEqual(g.Resource, exp.Statement[i].Resource) {
			t.Errorf("policy.statement[%d].resource does not match, expected: %v, got: %v", i, exp.Statement[i].Resource, g.Resource)
		}
		if len(g.Condition) != len(exp.Statement[i].Condition) {
			t.Errorf("policy.statement[%d].condition lengths do not match, expected: %d, got: %d", i, len(exp.Statement[i].Condition), len(g.Condition))
		}
		for _, c := range exp.Statement[i].Condition {
			err := hasCondition(g, c.Operator, c.Property, c.Value...)
			if err != nil {
				t.Error(err)
			}
		}
	}
}

func hasCondition(statement *Statement, operator string, property string, value ...string) error {
	for _, c := range statement.Condition {
		if c.Operator == operator &&
			c.Property == property &&
			shared.StringSliceEqual(c.Value, value) {
			return nil
		}
	}
	return fmt.Errorf("condition not found matching: {operator: %s, property: %s, value: %v}",
		operator, property, value)
}
