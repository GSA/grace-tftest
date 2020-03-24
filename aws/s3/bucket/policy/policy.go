// Package Policy  provides types and functions for breaking down
// S3 Buucket policies which allows the filter actions to take place.

package policy

import (
	"testing"

	"github.com/GSA/grace-tftest/aws/shared/policy"
	"github.com/GSA/grace-tftest/aws/shared/policy/statement"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Policy contains the necessary properties for testing *policy.Document objects
type Policy struct {
	client client.ConfigProvider
	name   string
}

// New returns a new *Policy
func New(client client.ConfigProvider, name string) *Policy {
	return &Policy{client: client, name: name}
}

// Statement returns a newly instantiated *statement.Statement object
// this is used for filtering all of the statements in all of the policies
// related to the kms key. If doc is nil, the policies will be queried from AWS
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

// document function that take a bucket policy and parses it.  Its important to make sure
// that you understand that Policy only returns on Policy per bucket.  If bucket policy is nil
// Unmarshal will define a policy for the statement.
func (p *Policy) document() (*policy.Document, error) {
	svc := s3.New(p.client)
	out, err := svc.GetBucketPolicy(&s3.GetBucketPolicyInput{Bucket: aws.String(p.name)})
	if err != nil {
		return nil, err
	}
	doc, err := policy.Unmarshal(aws.StringValue(out.Policy))
	if err != nil {
		return nil, err
	}
	return doc, nil
}
