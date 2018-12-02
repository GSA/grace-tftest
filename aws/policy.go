package aws

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/iam"
	terratest "github.com/gruntwork-io/terratest/modules/aws"
)

// PolicyDocument ... is a generic structure that is used by UnmarshalPolicy
type PolicyDocument struct {
	Version   string
	Statement []*PolicyStatement
}

// Find ... finds a matching policy statement in the policy document using the matcher
func (d *PolicyDocument) Find(matcher func(*PolicyStatement) bool) (*PolicyStatement, error) {
	for _, statement := range d.Statement {
		if matcher(statement) {
			return statement, nil
		}
	}
	return nil, fmt.Errorf("failed to locate a matching statement")
}

// PolicyStatement ... is a generic structure to hold an AWS policy statement
type PolicyStatement struct {
	Sid       string
	Effect    string
	Action    []string
	Resource  []string
	Principal *PolicyPrincipal
	Condition *PolicyCondition
}

// PolicyPrincipal ... holds an AWS policy principal
type PolicyPrincipal struct {
	Type   string
	Values []string
}

// PolicyCondition ... holds an AWS policy condition
type PolicyCondition struct {
	Operator string
	Property string
	Value    []string
}

// PolicyStatementMatcher ... takes the expected policy statement and returns
// a matcher that can be given an actual policy statement to match against
// nolint: gocyclo
func PolicyStatementMatcher(s *PolicyStatement) func(*PolicyStatement) bool {
	return func(statement *PolicyStatement) bool {
		if len(s.Action) > 0 && !stringSliceEqual(s.Action, statement.Action) {
			return false
		}
		if len(s.Effect) > 0 && s.Effect != statement.Effect {
			return false
		}
		if len(s.Resource) > 0 && !stringSliceEqual(s.Resource, statement.Resource) {
			return false
		}
		if len(s.Sid) > 0 && strings.ToLower(s.Sid) != strings.ToLower(statement.Sid) {
			return false
		}
		if s.Principal != nil &&
			(strings.ToLower(s.Principal.Type) != strings.ToLower(statement.Principal.Type) ||
				!stringSliceEqual(s.Principal.Values, statement.Principal.Values)) {
			return false
		}
		if s.Condition != nil &&
			(strings.ToLower(s.Condition.Operator) != strings.ToLower(statement.Condition.Operator) ||
				strings.ToLower(s.Condition.Property) != strings.ToLower(statement.Condition.Property) ||
				!stringSliceEqual(s.Condition.Value, statement.Condition.Value)) {
			return false
		}
		return true
	}
}

// MatchIamPolicy ... retrieves policy with the given arn and version, then matches the policy statements with the given matcher
func MatchIamPolicy(t *testing.T, region string, arn string, version string, matcher func(*PolicyStatement) bool) *PolicyStatement {
	statement, err := MatchIamPolicyE(region, arn, version, matcher)
	if err != nil {
		t.Fatal(err)
	}
	return statement
}

// MatchIamPolicyE ... retrieves policy with the given arn and version, then matches the policy statements with the given matcher
func MatchIamPolicyE(region string, arn string, version string, matcher func(*PolicyStatement) bool) (*PolicyStatement, error) {
	doc, err := GetIamPolicyDocumentE(region, arn, version)
	if err != nil {
		return nil, err
	}
	statement, err := doc.Find(matcher)
	if err != nil {
		return nil, err
	}
	return statement, nil
}

// MatchIamPolicyByName ... retrieves policy with given name, then matches the policy statements with the given matcher
func MatchIamPolicyByName(t *testing.T, region string, name string, matcher func(*PolicyStatement) bool) *PolicyStatement {
	statement, err := MatchIamPolicyByNameE(region, name, matcher)
	if err != nil {
		t.Fatal(err)
	}
	return statement
}

// MatchIamPolicyByNameE ... retrieves policy with given name, then matches the policy statements with the given matcher
func MatchIamPolicyByNameE(region string, name string, matcher func(*PolicyStatement) bool) (*PolicyStatement, error) {
	policy, err := FindIamPolicyByNameE(region, name)
	if err != nil {
		return nil, err
	}
	doc, err := GetIamPolicyDocumentE(region, *policy.Arn, *policy.DefaultVersionId)
	if err != nil {
		return nil, err
	}
	statement, err := doc.Find(matcher)
	if err != nil {
		return nil, err
	}
	return statement, nil
}

// GetIamPolicyDocument ... retrieves the policy document matching the given arn and version
func GetIamPolicyDocument(t *testing.T, region string, arn string, version string) *PolicyDocument {
	doc, err := GetIamPolicyDocumentE(region, arn, version)
	if err != nil {
		t.Fatalf("GetIamPolicyDocument failed: %v", err)
	}
	return doc
}

// GetIamPolicyDocumentE ... retrieves the policy document matching the given arn and version
func GetIamPolicyDocumentE(region string, arn string, version string) (*PolicyDocument, error) {
	client, err := terratest.NewIamClientE(nil, region)
	if err != nil {
		return nil, err
	}
	result, err := client.GetPolicyVersion(&iam.GetPolicyVersionInput{
		PolicyArn: &arn,
		VersionId: &version,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to locate policy version with id: %q, for arn: %q", version, arn)
	}
	doc, err := UnmarshalPolicy(*result.PolicyVersion.Document)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

// FindIamPolicyByName ... finds policy with the given name
func FindIamPolicyByName(t *testing.T, region string, name string) *iam.Policy {
	policy, err := FindIamPolicyByNameE(region, name)
	if err != nil {
		t.Fatalf("FindIamPolicyByName failed: %v", err)
	}
	return policy
}

// FindIamPolicyByNameE ... finds policy with the given name
func FindIamPolicyByNameE(region string, name string) (*iam.Policy, error) {
	policy, err := FindIamPolicyE(region, func(p *iam.Policy) bool {
		return *p.PolicyName == name
	})
	if err != nil {
		return nil, err
	}
	return policy, nil
}

// FindIamPolicy ... finds policy with the given matcher
func FindIamPolicy(t *testing.T, region string, matcher func(*iam.Policy) bool) *iam.Policy {
	policy, err := FindIamPolicyE(region, matcher)
	if err != nil {
		t.Fatalf("FindIamPolicy failed: %v", err)
	}
	return policy
}

// FindIamPolicyE ... finds policy with the given matcher
func FindIamPolicyE(region string, matcher func(*iam.Policy) bool) (*iam.Policy, error) {
	var (
		err    error
		marker *string
	)

	more := true
	for more {
		var policies []*iam.Policy
		policies, marker, err = ListIamPoliciesE(region, marker)
		if err != nil {
			return nil, err
		}
		for _, p := range policies {
			if matcher(p) {
				return p, nil
			}
		}
		if marker == nil {
			more = false
		}
	}

	return nil, fmt.Errorf("failed to locate a matching policy")
}

// ListIamPolicies ... retrieves a batch of policies starting at the given marker
// marker should be nil on first call
func ListIamPolicies(t *testing.T, region string, marker *string) ([]*iam.Policy, *string) {
	policies, next, err := ListIamPoliciesE(region, marker)
	if err != nil {
		t.Fatalf("ListPolicies failed: %v", err)
	}
	return policies, next
}

// ListIamPoliciesE ... retrieves a batch of policies starting at the given marker
// marker should be nil on first call
func ListIamPoliciesE(region string, marker *string) ([]*iam.Policy, *string, error) {
	client, err := terratest.NewIamClientE(nil, region)
	if err != nil {
		return nil, nil, err
	}
	result, err := client.ListPolicies(&iam.ListPoliciesInput{
		Marker: marker,
	})
	if err != nil {
		return nil, nil, err
	}
	return result.Policies, result.Marker, nil
}

// stringSliceEqual ... validates that two slice of strings are equal
// where 'a' is the authoritative slice
func stringSliceEqual(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if strings.ToLower(a[i]) != strings.ToLower(b[i]) {
			return false
		}
	}
	return true
}

// UnmarshalPolicy ... unmarshals a raw policy document
func UnmarshalPolicy(raw string) (*PolicyDocument, error) {
	data, err := url.QueryUnescape(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to unescape Policy Document: %v", err)
	}
	pdoc, err := parsePolicy([]byte(data))
	if err != nil {
		return nil, fmt.Errorf("failed to parse policy document: %v", err)
	}
	return pdoc, nil
}

// setterFunc ... used by parseStatement to set each property
type setterFunc func(*PolicyStatement, interface{}) error

// parsePolicy ... takes a policy document in json format and returns a *PolicyDocument
func parsePolicy(data []byte) (*PolicyDocument, error) {
	var m map[string]interface{}
	err := json.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}
	var pdoc PolicyDocument
	items := m["Statement"].([]interface{})
	for i := 0; i < len(items); i++ {
		item := items[i].(map[string]interface{})
		err = parseStatement(&pdoc, item)
		if err != nil {
			return nil, err
		}
	}
	return &pdoc, nil
}

// parsePolicy ... takes a *PolicyDocument and the Statement value as map[string]interface{}
// then populates the *PolicyDocument with the parsed policy statements
func parseStatement(doc *PolicyDocument, m map[string]interface{}) error {
	statement := &PolicyStatement{}
	setters := map[string]setterFunc{
		"Sid":       setSid,
		"Effect":    setEffect,
		"Principal": setPrincipal,
		"Action":    setAction,
		"Resource":  setResource,
		"Condition": setCondition,
	}
	for field, fn := range setters {
		if val, ok := m[field]; ok {
			err := fn(statement, val)
			if err != nil {
				return fmt.Errorf("failed to set field %s: %v", field, err)
			}
		}
	}
	doc.Statement = append(doc.Statement, statement)
	return nil
}

// setSid ... converts and sets Sid in statement
func setSid(statement *PolicyStatement, m interface{}) error {
	statement.Sid = m.(string)
	return nil
}

// setEffect ... converts and sets Effect in statement
func setEffect(statement *PolicyStatement, m interface{}) error {
	statement.Effect = m.(string)
	return nil
}

// setPrincipal ... converts and sets Principal in statement
func setPrincipal(statement *PolicyStatement, m interface{}) error {
	statement.Principal = &PolicyPrincipal{}
	err := setPrincipalProperty(statement.Principal, m.(map[string]interface{}))
	return err
}

// setPrincipalProperty ... converts and sets Principal Type and Values
func setPrincipalProperty(principal *PolicyPrincipal, m map[string]interface{}) error {
	for k, v := range m {
		principal.Type = k
		switch val := v.(type) {
		case string:
			principal.Values = []string{val}
		case []string:
			principal.Values = val
		default:
			return fmt.Errorf("type not supported: %T", val)
		}
		break
	}
	return nil
}

// setAction ... converts and sets Action in statement
func setAction(statement *PolicyStatement, m interface{}) (err error) {
	statement.Action, err = interfaceToStringSlice(m)
	return err
}

// setResource ... converts and sets Resource in statement
func setResource(statement *PolicyStatement, m interface{}) (err error) {
	statement.Resource, err = interfaceToStringSlice(m)
	return err
}

// setCondition ... converts and sets Condition in statement
func setCondition(statement *PolicyStatement, m interface{}) (err error) {
	statement.Condition = &PolicyCondition{}
	mm := m.(map[string]interface{})
	for k, v := range mm {
		statement.Condition.Operator = k
		switch val := v.(type) {
		case map[string]interface{}:
			for k, v := range val {
				statement.Condition.Property = k
				statement.Condition.Value, err = interfaceToStringSlice(v)
				if err != nil {
					return
				}
				break
			}
		case map[string][]string:
			for k, v := range val {
				statement.Condition.Property = k
				statement.Condition.Value = v
				break
			}
		case map[string]string:
			for k, v := range val {
				statement.Condition.Property = k
				statement.Condition.Value = []string{v}
			}
		default:
			return fmt.Errorf("type not supported: %T", val)
		}
		break
	}
	return nil
}

// interfaceToStringSlice ... converts []interface{}, []string, string to []string
func interfaceToStringSlice(m interface{}) ([]string, error) {
	switch val := m.(type) {
	case []interface{}:
		var value []string
		for _, v := range val {
			value = append(value, v.(string))
		}
		return value, nil
	case []string:
		return val, nil
	case string:
		return []string{val}, nil
	default:
		return nil, fmt.Errorf("type not supported: %T", val)
	}
}

/*



Principal Field Types:

"Principal": "*"
"Principal" : { "AWS" : "*" }
"Principal": { "CanonicalUser": "79a59df900b949e55d96a1e698fbacedfd6e09d98eacf8f8d5218e7cd47ef2be" }
"Principal": { "AWS": "arn:aws:sts::AWS-account-ID:assumed-role/role-name/role-session-name" }
"Principal": { "Federated": "accounts.google.com" }
"Principal": {
  "AWS": [
    "arn:aws:iam::AWS-account-ID:user/user-name-1",
    "arn:aws:iam::AWS-account-ID:user/UserName2"
  ]
}
"Principal": {
  "Service": [
    "elasticmapreduce.amazonaws.com",
    "datapipeline.amazonaws.com"
  ]
}
&aws.PolicyStatement{
	Effect:[]string{"Allow"},
	Action:[]string{"sts:AssumeRole"},
	Resource:[]string(nil),
	Principal:(*aws.PolicyPrincipal)(0xc00054c840)
}, not found in
{
"Version":"2012-10-17",
"Statement":[
{"Sid":"",
"Effect":"Allow",
"Principal":{"Service":"lambda.amazonaws.com"},
"Action":"sts:AssumeRole"
}]}


"Condition":
	{"ArnLike":{"AWS:SourceArn":"arn:aws:events:us-east-1:422624340815:rule/grace-inventory-lambda"}}}
"Condition":
	{"DateGreaterThan":{"aws:CurrentTime":"2013-08-16T12:00:00Z"},"DateLessThan":{"aws:CurrentTime" : "2013-08-16T15:00:00Z"},"IpAddress":{"aws:SourceIp":["192.0.2.0/24","203.0.113.0/24"]}}

func:{key:value}
*/
