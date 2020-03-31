// Package policy  provides types and functions for breaking down
// SNS Buucket policies which allows the filter actions to take place.
package policy

import (
	"testing"

	"github.com/GSA/grace-tftest/aws/shared/policy"
	"github.com/GSA/grace-tftest/aws/shared/policy/statement"
)

// Policy contains the necessary properties for testing *policy.Document objects
type Policy struct {
	policy string
}

// New returns a new *Policy
func New(policy string) *Policy {
	return &Policy{policy: policy}
}

// Statement returns a newly instantiated *statement.Statement object
// this is used for filtering all of the statements in all of the policies
// related to the SNS topic. If doc is nil, the policies will be queried from AWS
func (p *Policy) Statement(t *testing.T, doc *policy.Document) *statement.Statement {
	if doc == nil {
		var err error
		doc, err = p.document()
		if err != nil {
			t.Errorf("failed to query statements: %v", err)
			return nil
		}
	}
	return statement.New(doc)
}

// document function that take a topic policy and parses it.  Its important to make sure
// that you understand that Policy only returns on Policy per topic.  If topic policy is nil
// Unmarshal will define a policy for the statement.
func (p *Policy) document() (*policy.Document, error) {
	doc, err := policy.Unmarshal(p.policy)
	if err != nil {
		return nil, err
	}
	return doc, nil
}
