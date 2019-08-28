package encryption

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Filter is an interface for filtering *s3.ServerSideEncryptionRule objects
type Filter func(*s3.ServerSideEncryptionRule) bool

// Encryption contains the necessary properties for filtering *s3.ServerSideEncryptionRule objects
type Encryption struct {
	filters []Filter
	client  client.ConfigProvider
	rule    *s3.ServerSideEncryptionRule
	name    string
}

// New returns a new *Encryption
func New(client client.ConfigProvider, name string) *Encryption {
	return &Encryption{
		client: client,
		name:   name,
	}
}

// Selected returns the currently selected *s3.ServerSideEncryptionRule
func (e *Encryption) Selected() *s3.ServerSideEncryptionRule {
	return e.rule
}

// Assert applies all filters that have been called, resets the list of filters,
// fails the test if there is not exactly one match, and stores the matched rule
// if rule is not provided, *s3.ServerSideEncryptionRule objects will be retreived from AWS
func (e *Encryption) Assert(t *testing.T, rules ...*s3.ServerSideEncryptionRule) *Encryption {
	var err error
	rules, err = e.filter(rules)
	if err != nil {
		t.Error(err)
	}

	if len(rules) == 0 {
		t.Error("no matching lifecycle rule was found")
	} else if len(rules) > 1 {
		t.Error("more than one matching lifecycle rule was found")
	} else {
		e.rule = rules[0]
	}

	e.filters = []Filter{}
	return e
}

// First applies all filters that have been called, resets the list of filters,
// fails the test if there is not exactly one match, and stores the matched rule
// if rule is not provided, *s3.ServerSideEncryptionRule objects will be retreived from AWS
func (e *Encryption) First(t *testing.T, rules ...*s3.ServerSideEncryptionRule) *Encryption {
	var err error
	rules, err = e.filter(rules)
	if err != nil {
		t.Error(err)
	}

	if len(rules) == 0 {
		t.Error("no matching lifecycle rule was found")
	} else {
		e.rule = rules[0]
	}

	e.filters = []Filter{}
	return e
}

// Filter adds the 'filter' provided to the filter list
func (e *Encryption) Filter(filter Filter) *Encryption {
	e.filters = append(e.filters, filter)
	return e
}

// IsSSE adds the IsSSE filter to the filter list
// the IsSSE filter: filters rules by whether they have
// ApplyServerSideEncryptionByDefault set
func (e *Encryption) IsSSE() *Encryption {
	e.filters = append(e.filters, func(rule *s3.ServerSideEncryptionRule) bool {
		return rule.ApplyServerSideEncryptionByDefault != nil
	})
	return e
}

// Alg adds the Alg filter to the filter list
// the Alg filter: filters rules by SSEAlgorithm where 'alg'
// provided is the expected SSEAlgorithm value
func (e *Encryption) Alg(alg string) *Encryption {
	e.filters = append(e.filters, func(rule *s3.ServerSideEncryptionRule) bool {
		return alg == aws.StringValue(rule.ApplyServerSideEncryptionByDefault.SSEAlgorithm)
	})
	return e
}

// ID adds the ID filter to the filter list
// the ID filter: filters rules by KMSMasterKeyID where 'id'
// provided is the expected KMSMasterKeyID value
func (e *Encryption) ID(id string) *Encryption {
	e.filters = append(e.filters, func(rule *s3.ServerSideEncryptionRule) bool {
		return id == aws.StringValue(rule.ApplyServerSideEncryptionByDefault.KMSMasterKeyID)
	})
	return e
}

func (e *Encryption) filter(rules []*s3.ServerSideEncryptionRule) (result []*s3.ServerSideEncryptionRule, err error) {
	if len(rules) == 0 {
		var err error
		rules, err = e.rules()
		if err != nil {
			return nil, err
		}
	}
outer:
	for _, rule := range rules {
		for _, f := range e.filters {
			if !f(rule) {
				continue outer
			}
		}
		result = append(result, rule)
	}
	return
}

func (e *Encryption) rules() ([]*s3.ServerSideEncryptionRule, error) {
	svc := s3.New(e.client)
	out, err := svc.GetBucketEncryption(&s3.GetBucketEncryptionInput{
		Bucket: &e.name,
	})
	if err != nil {
		return nil, err
	}
	return out.ServerSideEncryptionConfiguration.Rules, nil
}
