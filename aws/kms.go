package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/kms"
	terratest "github.com/gruntwork-io/terratest/modules/aws"
)

// FindKmsKeyPolicy ... finds matching KMS Key Policy for key with given ID and matcher
func FindKmsKeyPolicy(t *testing.T, region string, keyID string, matcher func(*PolicyStatement) bool) *PolicyStatement {
	s, err := FindKmsKeyPolicyE(region, keyID, matcher)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// FindKmsKeyPolicyE ... finds matching KMS Key Policy for key with given ID and matcher
func FindKmsKeyPolicyE(region string, keyID string, matcher func(*PolicyStatement) bool) (*PolicyStatement, error) {
	documents, err := GetKmsKeyPolicyDocumentsE(region, keyID)
	if err != nil {
		return nil, err
	}
	for _, doc := range documents {
		s, err := doc.Find(matcher)
		if err == nil {
			return s, nil
		}
	}
	return nil, fmt.Errorf("failed to locate a matching policy")
}

// GetKmsKeyPolicyDocuments ... retrieves policy document for key with given ID
func GetKmsKeyPolicyDocuments(t *testing.T, region string, keyID string) []*PolicyDocument {
	documents, err := GetKmsKeyPolicyDocumentsE(region, keyID)
	if err != nil {
		t.Fatalf("GetKeyPolicyDocuments failed: %v", err)
	}
	return documents
}

// GetKmsKeyPolicyDocumentsE ... retrieves policy document for key with given ID
func GetKmsKeyPolicyDocumentsE(region string, keyID string) ([]*PolicyDocument, error) {
	var (
		documents []*PolicyDocument
	)
	policies, err := GetKmsKeyPoliciesE(region, keyID)
	if err != nil {
		return nil, err
	}
	for _, p := range policies {
		doc, err := UnmarshalPolicy(p)
		if err != nil {
			return nil, fmt.Errorf("failed to Unmarshal Policy document: %v", err)
		}
		documents = append(documents, doc)
	}
	return documents, nil
}

// GetKmsKeyPolicies ... retrieves policy raw policy documents for key matching the given ID
func GetKmsKeyPolicies(t *testing.T, region string, keyID string) []string {
	policies, err := GetKmsKeyPoliciesE(region, keyID)
	if err != nil {
		t.Fatalf("GetKeyPolicies failed: %v", err)
	}
	return policies
}

// GetKmsKeyPoliciesE ... retrieves policy raw policy documents for key matching the given ID
func GetKmsKeyPoliciesE(region string, keyID string) ([]string, error) {
	var (
		names    []*string
		out      *kms.GetKeyPolicyOutput
		policies []string
		err      error
		svc      *kms.KMS
	)
	svc, err = terratest.NewKmsClientE(nil, region)
	if err != nil {
		return nil, err
	}
	names, err = ListAllKmsKeyPoliciesE(region, keyID)
	if err != nil {
		return nil, err
	}
	for _, n := range names {
		out, err = svc.GetKeyPolicy(&kms.GetKeyPolicyInput{
			KeyId:      &keyID,
			PolicyName: n,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to GetKeyPolicy with name %q: %v", *n, err)
		}
		policies = append(policies, *out.Policy)
	}
	return policies, nil
}

// ListAllKmsKeyPolicies ... retrieves all policy names for key with given ID
func ListAllKmsKeyPolicies(t *testing.T, region string, keyID string) []*string {
	names, err := ListAllKmsKeyPoliciesE(region, keyID)
	if err != nil {
		t.Fatalf("ListAllKeyPolicies failed: %v", err)
	}
	return names
}

// ListAllKmsKeyPoliciesE ... retrieves all policy names for key with given ID
func ListAllKmsKeyPoliciesE(region string, keyID string) ([]*string, error) {
	var (
		err    error
		marker *string
		names  []*string
	)
	more := true
	for more {
		var n []*string
		n, marker, err = ListKmsKeyPoliciesE(region, keyID, marker)
		if err != nil {
			return nil, err
		}
		names = append(names, n...)
		if marker == nil {
			more = false
		}
	}
	return names, nil
}

// FindKmsAlias ... finds the kms key alias with the given name
func FindKmsAlias(t *testing.T, region string, aliasName string) *kms.AliasListEntry {
	entry, err := FindKmsAliasE(region, aliasName)
	if err != nil {
		t.Fatal(err)
	}
	return entry
}

// FindKmsAliasE ... finds the kms key alias with the given name
func FindKmsAliasE(region string, aliasName string) (*kms.AliasListEntry, error) {
	var (
		err    error
		marker *string
	)
	more := true
	for more {
		var aliases []*kms.AliasListEntry
		aliases, marker, err = ListKmsAliasesE(region, marker)
		if err != nil {
			return nil, err
		}
		for _, a := range aliases {
			if *a.AliasName == aliasName {
				return a, nil
			}
		}
		if marker == nil {
			more = false
		}
	}

	return nil, fmt.Errorf("failed to locate a KMS alias matching %s", aliasName)
}

// ListKmsKeyPolicies ... retrieves a batch of key policy names, where marker is the index token
// on the initial call marker should be nil
func ListKmsKeyPolicies(t *testing.T, region string, keyID string, marker *string) ([]*string, *string) {
	names, next, err := ListKmsKeyPoliciesE(region, keyID, marker)
	if err != nil {
		t.Fatal(err)
	}
	return names, next
}

// ListKmsKeyPoliciesE ... retrieves a batch of key policy names, where marker is the index token
// on the initial call marker should be nil
func ListKmsKeyPoliciesE(region string, keyID string, marker *string) ([]*string, *string, error) {
	client, err := terratest.NewKmsClientE(nil, region)
	if err != nil {
		return nil, nil, err
	}
	result, err := client.ListKeyPolicies(&kms.ListKeyPoliciesInput{
		KeyId:  &keyID,
		Marker: marker,
	})
	if err != nil {
		return nil, nil, err
	}

	return result.PolicyNames, result.NextMarker, nil
}

// ListKmsAliases ... retrieves a batch of key aliases, where marker is the index token
// on the initial call marker should be nil
func ListKmsAliases(t *testing.T, region string, marker *string) ([]*kms.AliasListEntry, *string) {
	aliases, next, err := ListKmsAliasesE(region, marker)
	if err != nil {
		t.Fatalf("ListKMSAliases failed: %v", err)
	}
	return aliases, next
}

// ListKmsAliasesE ... retrieves a batch of key aliases, where marker is the index token
//on the initial call marker should be nil
func ListKmsAliasesE(region string, marker *string) ([]*kms.AliasListEntry, *string, error) {
	client, err := terratest.NewKmsClientE(nil, region)
	if err != nil {
		return nil, nil, err
	}
	result, err := client.ListAliases(&kms.ListAliasesInput{
		Marker: marker,
	})
	if err != nil {
		return nil, nil, err
	}
	return result.Aliases, result.NextMarker, nil
}
