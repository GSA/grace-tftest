// Package policy provides testing for Cloudwatch Event Bus permissions policies
package policy

import (
	"testing"

	"github.com/GSA/grace-tftest/aws/shared/policy"
	"github.com/GSA/grace-tftest/aws/shared/policy/statement"
	"github.com/aws/aws-sdk-go/aws"
)

// Policy contains the necessary properties for testing *policy.Document objects
type Policy struct {
	policy *string
}

// New returns a new *Policy
func New(p *string) *Policy {
	return &Policy{policy: p}
}

// Statement returns a newly instantiated *statement.Statement object
// this is used for filtering all of the statements in all of the policies
// related to the cloudwatchevents key. If doc is nil, the policies will be queried from AWS
func (p *Policy) Statement(t *testing.T, doc *policy.Document) *statement.Statement {
	if doc == nil {
		statements, err := p.statements()
		if err != nil {
			t.Errorf("failed to query statements: %v", err)
			return nil
		}
		doc = &policy.Document{Statement: statements}
	}
	return statement.New(doc)
}

func (p *Policy) statements() ([]*policy.Statement, error) {
	var statements []*policy.Statement
	doc, err := policy.Unmarshal(aws.StringValue(p.policy))
	if err != nil {
		return nil, err
	}
	statements = append(statements, doc.Statement...)
	return statements, nil
}
