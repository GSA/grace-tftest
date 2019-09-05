package statement

import (
	"testing"

	"github.com/GSA/grace-tftest/aws/shared/policy"
)

const testpolicy = `
{
	"Version": "2012-10-17",
	"Statement": [
	  {
		"Sid": "a",
		"Action": [ "b" ],
		"Effect": "c",
		"Resource": [ "d", "e", "f" ]
	  },
	  {
		"Action": [ "a", "b", "c" ],
		"Effect": "d",
		"Resource": "e"
	  },
	  {
		"Action": [ "a" ],
		"Effect": "b",
		"Resource": "c",
		"Condition": {"d":{"e":["f"]}}
	  },
	  {
		"Action": [ "a" ],
		"Effect": "b",
		"Resource": "c",
		"Principal":{"d":"e"},
		"Condition": {"f":{"g":["h"]},"i":{"j":["k", "l", "m"]}}
	  }
	]
  }
`

func TestStatement(t *testing.T) {
	doc, err := policy.Unmarshal(testpolicy)
	if err != nil {
		t.Fatalf("failed to unmarshal test policy: %v", err)
	}

	New(doc).Sid("a").Action("b").Effect("c").Resource("d", "e", "f").Assert(t)
	New(doc).Action("a", "b", "c").Effect("d").Resource("e").Assert(t)
	New(doc).Action("a").Effect("b").Resource("c").Condition("d", "e", "f").Assert(t)
	New(doc).Action("a").Effect("b").Resource("c").Principal("d", "e").Condition("f", "g", "h").Condition("i", "j", "k", "l", "m").Assert(t)
}
