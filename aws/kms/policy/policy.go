package policy

import (
	"testing"

	"github.com/GSA/grace-tftest/aws/shared/policy"
	"github.com/GSA/grace-tftest/aws/shared/policy/statement"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/kms"
)

// Policy contains the necessary properties for testing *policy.Document objects
type Policy struct {
	client client.ConfigProvider
	keyID  string
}

// New returns a new *Policy
func New(client client.ConfigProvider, keyID string) *Policy {
	return &Policy{client: client, keyID: keyID}
}

// Statement returns a newly instantiated *statement.Statement object
// this is used for filtering all of the statements in all of the policies
// related to the kms key. If doc is nil, the policies will be queried from AWS
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
	svc := kms.New(p.client)

	var names []string
	err := svc.ListKeyPoliciesPages(&kms.ListKeyPoliciesInput{
		KeyId: aws.String(p.keyID),
	}, func(out *kms.ListKeyPoliciesOutput, lastPage bool) bool {
		names = append(names, aws.StringValueSlice(out.PolicyNames)...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}

	var statements []*policy.Statement
	for _, name := range names {
		if len(name) == 0 {
			continue
		}
		out, err := svc.GetKeyPolicy(&kms.GetKeyPolicyInput{
			KeyId:      aws.String(p.keyID),
			PolicyName: aws.String(name),
		})
		if err != nil {
			return nil, err
		}
		doc, err := policy.Unmarshal(aws.StringValue(out.Policy))
		if err != nil {
			return nil, err
		}
		statements = append(statements, doc.Statement...)
	}
	return statements, nil
}
