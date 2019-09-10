package encryption

import (
	"log"
	"os"
	"testing"

	"github.com/GSA/grace-tftest/aws/shared"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Encryption contains the necessary properties for filtering *s3.ServerSideEncryptionRule objects
type Encryption struct {
	filters []shared.Filter
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
		t.Fatal(err)
	}

	switch l := len(rules); {
	case l == 0:
		t.Fatal("no matching lifecycle rule was found")
	case l > 1:
		t.Fatal("more than one matching lifecycle rule was found")
	default:
		e.rule = rules[0]
	}

	e.filters = []shared.Filter{}
	return e
}

// First applies all filters that have been called, resets the list of filters,
// fails the test if there is not exactly one match, and stores the matched rule
// if rule is not provided, *s3.ServerSideEncryptionRule objects will be retreived from AWS
func (e *Encryption) First(t *testing.T, rules ...*s3.ServerSideEncryptionRule) *Encryption {
	var err error
	rules, err = e.filter(rules)
	if err != nil {
		t.Fatal(err)
	}

	if len(rules) == 0 {
		t.Fatal("no matching lifecycle rule was found")
	} else {
		e.rule = rules[0]
	}

	e.filters = []shared.Filter{}
	return e
}

// Filter adds the 'filter' provided to the filter list
func (e *Encryption) Filter(filter shared.Filter) *Encryption {
	e.filters = append(e.filters, filter)
	return e
}

// IsSSE adds the IsSSE filter to the filter list
// the IsSSE filter: filters rules by whether they have
// ApplyServerSideEncryptionByDefault set
func (e *Encryption) IsSSE() *Encryption {
	e.filters = append(e.filters, func(v interface{}) bool {
		rule := convert(v)
		if rule == nil {
			return false
		}
		shared.Debugf("%#v != nil -> %t\n", rule.ApplyServerSideEncryptionByDefault, rule.ApplyServerSideEncryptionByDefault != nil)
		return rule.ApplyServerSideEncryptionByDefault != nil
	})
	return e
}

// Alg adds the Alg filter to the filter list
// the Alg filter: filters rules by SSEAlgorithm where 'alg'
// provided is the expected SSEAlgorithm value
func (e *Encryption) Alg(alg string) *Encryption {
	e.filters = append(e.filters, func(v interface{}) bool {
		rule := convert(v)
		if rule == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n",
			alg, aws.StringValue(rule.ApplyServerSideEncryptionByDefault.SSEAlgorithm),
			alg == aws.StringValue(rule.ApplyServerSideEncryptionByDefault.SSEAlgorithm))
		return alg == aws.StringValue(rule.ApplyServerSideEncryptionByDefault.SSEAlgorithm)
	})
	return e
}

// ID adds the ID filter to the filter list
// the ID filter: filters rules by KMSMasterKeyID where 'id'
// provided is the expected KMSMasterKeyID value
func (e *Encryption) ID(id string) *Encryption {
	e.filters = append(e.filters, func(v interface{}) bool {
		rule := convert(v)
		if rule == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n",
			id, aws.StringValue(rule.ApplyServerSideEncryptionByDefault.KMSMasterKeyID),
			id == aws.StringValue(rule.ApplyServerSideEncryptionByDefault.KMSMasterKeyID))
		return id == aws.StringValue(rule.ApplyServerSideEncryptionByDefault.KMSMasterKeyID)
	})
	return e
}

func (e *Encryption) filter(rules []*s3.ServerSideEncryptionRule) ([]*s3.ServerSideEncryptionRule, error) {
	if len(rules) == 0 {
		var err error
		rules, err = e.rules()
		if err != nil {
			return nil, err
		}
	}
	results := fromIface(shared.GenericFilter(e.filters, toIface(rules)))
	if len(results) == 0 {
		log.Println("aws.s3.bucket.encryption.filter had zero results: ")
		shared.Spew(os.Stdout, rules)
	}
	return results, nil
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

func convert(in interface{}) *s3.ServerSideEncryptionRule {
	out, ok := in.(*s3.ServerSideEncryptionRule)
	if !ok {
		shared.Debugf("object not convertible to *s3.ServerSideEncryptionRule: ")
		shared.Dump(in)
		return nil
	}
	return out
}
func toIface(in []*s3.ServerSideEncryptionRule) (out []interface{}) {
	for _, i := range in {
		out = append(out, i)
	}
	return
}
func fromIface(in []interface{}) (out []*s3.ServerSideEncryptionRule) {
	for _, i := range in {
		v := convert(i)
		if v == nil {
			continue
		}
		out = append(out, v)
	}
	return
}
