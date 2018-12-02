package aws

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/lambda"
	terratest "github.com/gruntwork-io/terratest/modules/aws"
)

// LambdaFunctionConfigMatcher ... takes properties that are intended to be matched and returns
// a matcher function that verifies the properties passed are present and match in the configuration
// nolint: gocyclo
func LambdaFunctionConfigMatcher(roleArn string, handler string, keyArn, runtime string, timeout int64, environment map[string]string) func(*lambda.FunctionConfiguration) bool {
	return func(c *lambda.FunctionConfiguration) bool {
		if len(roleArn) > 0 && strings.ToLower(roleArn) != strings.ToLower(*c.Role) {
			return false
		}
		if len(handler) > 0 && strings.ToLower(handler) != strings.ToLower(*c.Handler) {
			return false
		}
		if len(keyArn) > 0 && strings.ToLower(keyArn) != strings.ToLower(*c.KMSKeyArn) {
			return false
		}
		if len(runtime) > 0 && strings.ToLower(runtime) != strings.ToLower(*c.Runtime) {
			return false
		}
		if timeout > 0 && timeout != *c.Timeout {
			return false
		}
		for k, v := range environment {
			var (
				ok  bool
				val *string
			)
			if val, ok = c.Environment.Variables[k]; !ok {
				return false
			}
			if strings.ToLower(v) != strings.ToLower(*val) {
				return false
			}
		}
		return true
	}
}

// MatchLambdaFunctionConfig ... uses the provided matcher to validate the configuration of the lambda function
// with the given name
func MatchLambdaFunctionConfig(t *testing.T, region string, functionName string, matcher func(*lambda.FunctionConfiguration) bool) *lambda.FunctionConfiguration {
	config, err := MatchLambdaFunctionConfigE(region, functionName, matcher)
	if err != nil {
		t.Fatal(err)
	}
	return config
}

// MatchLambdaFunctionConfigE ... uses the provided matcher to validate the configuration of the lambda function
// with the given name
func MatchLambdaFunctionConfigE(region string, functionName string, matcher func(*lambda.FunctionConfiguration) bool) (*lambda.FunctionConfiguration, error) {
	client, err := NewLambdaClientE(region)
	if err != nil {
		return nil, err
	}
	out, err := client.GetFunction(&lambda.GetFunctionInput{
		FunctionName: &functionName,
	})
	if err != nil {
		return nil, err
	}
	if matcher(out.Configuration) {
		return out.Configuration, nil
	}
	return nil, fmt.Errorf("failed to match lambda configuration for function %q", functionName)
}

// FindLambdaPolicy ... finds the matching policy statement for the given function name using the provided matcher
func FindLambdaPolicy(t *testing.T, region string, functionName string, matcher func(*PolicyStatement) bool) *PolicyStatement {
	statement, err := FindLambdaPolicyE(region, functionName, matcher)
	if err != nil {
		t.Fatal(err)
	}
	return statement
}

// FindLambdaPolicyE ... finds the matching policy statement for the given function name using the provided matcher
func FindLambdaPolicyE(region string, functionName string, matcher func(*PolicyStatement) bool) (*PolicyStatement, error) {
	doc, err := GetLambdaPolicyDocumentE(region, functionName)
	if err != nil {
		return nil, err
	}
	statement, err := doc.Find(matcher)
	if err != nil {
		return nil, err
	}
	return statement, nil
}

// GetLambdaPolicyDocument ... retrieves the policy document for the function with the given name
func GetLambdaPolicyDocument(t *testing.T, region, functionName string) *PolicyDocument {
	doc, err := GetLambdaPolicyDocumentE(region, functionName)
	if err != nil {
		t.Fatalf("GetLambdaPolicyDocument failed: %v", err)
	}
	return doc
}

// GetLambdaPolicyDocumentE ... retrieves the policy document for the function with the given name
func GetLambdaPolicyDocumentE(region, functionName string) (*PolicyDocument, error) {
	client, err := NewLambdaClientE(region)
	if err != nil {
		return nil, err
	}
	out, err := client.GetPolicy(&lambda.GetPolicyInput{
		FunctionName: &functionName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to GetPolicy with function name %q: %v", functionName, err)
	}
	doc, err := UnmarshalPolicy(*out.Policy)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal policy document with function name %q: %v", functionName, err)
	}
	return doc, nil
}

// NewLambdaClient ... creates an initiated lambda client
func NewLambdaClient(t *testing.T, region string) *lambda.Lambda {
	client, err := NewLambdaClientE(region)
	if err != nil {
		t.Fatal(err)
	}
	return client
}

// NewLambdaClientE ... creates an initiated lambda client
func NewLambdaClientE(region string) (*lambda.Lambda, error) {
	sess, err := terratest.NewAuthenticatedSession(region)
	if err != nil {
		return nil, err
	}
	return lambda.New(sess), nil
}
