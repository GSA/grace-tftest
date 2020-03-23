package policy

import (
	 "testing"
	
	"github.com/GRACE/grace-tftest/aws/shared"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Policy: Contains the necessary properties for testing *s3.Policy objects
type Policy struct {
	filters []shared.Filter
	client  client.ConfigProvider
	name    string
	policy  *string
}

// New:  Returns a new *Policy type 
func New(client client.ConfigProvider, name string) *Policy {
	return &Policy{
		client: client,
		name:   name,
	}
}

// Selected returns the currently selected *s3.Bucket Policy 
func (p *Policy) Selected() *s3.GetBucketPolicy(&s3.GetBucketPolicyInput{
	Bucket: *Policy.name
}) 
{
	return p.policy
}

// Statement returns a newly instantiated *statement.Statement object
// this is used for filtering by statements inside a policy. If doc is nil
// the default policy document will be retrieved from AWS
func (p *Policy) Statement(t *testing.T, doc *policy.Document) *statement.Statement {
	if doc == nil {
		doc = p.Document(t, "")
	}
	return statement.New(doc)
}

// Assert applies all filters that have been called, resets the list of filters,
// fails the test if there is not exactly one match, and stores the matched policy
// if policies is not provided, *iam.Policy objects will be retreived from AWS
func (p *Policy) Assert(t *testing.T, policies ...*string) *Policy {
	var err error
	policies, err = p.filter(policies)
	if err != nil {
		t.Fatal(err)
	}

	switch l := len(policies); {
	case l == 0:
		t.Fatal("no matching policy was found")
	case l > 1:
		t.Fatal("more than one matching policy was found")
	default:
		p.policy = policies[0]
	}

	p.filters = []shared.Filter{}
	return p
}



