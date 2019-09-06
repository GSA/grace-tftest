package policy

import (
	"testing"

	"github.com/GSA/grace-tftest/aws/shared/policy"
	"github.com/GSA/grace-tftest/aws/shared/policy/statement"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/lambda"
)

// Policy contains the necessary properties for testing *policy.Document objects
type Policy struct {
	client       client.ConfigProvider
	functionName string
}

// New returns a new *Policy
func New(client client.ConfigProvider, functionName string) *Policy {
	return &Policy{client: client, functionName: functionName}
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

func (p *Policy) document() (*policy.Document, error) {
	svc := lambda.New(p.client)
	out, err := svc.GetPolicy(&lambda.GetPolicyInput{FunctionName: aws.String(p.functionName)})
	if err != nil {
		return nil, err
	}
	doc, err := policy.Unmarshal(aws.StringValue(out.Policy))
	if err != nil {
		return nil, err
	}
	return doc, nil
}
