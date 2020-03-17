// Package key provides filtering of KMS keys by Description
package key

import (
	"log"
	"os"
	"strings"
	"testing"

	"github.com/GSA/grace-tftest/aws/kms/policy"
	"github.com/GSA/grace-tftest/aws/shared"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/kms"
)

// Key contains the necessary properties for testing *kms.KeyMetadata objects
type Key struct {
	client  client.ConfigProvider
	key     *kms.KeyMetadata
	filters []shared.Filter
}

// New returns a new *Key
func New(client client.ConfigProvider) *Key {
	return &Key{client: client}
}

// Selected returns the currently selected *kms.KeyMetadata
func (a *Key) Selected() *kms.KeyMetadata {
	return a.key
}

// Key returns the currently selected Keys' targeted *kms.KeyMetadata
func (a *Key) Key(t *testing.T) *kms.KeyMetadata {
	if a.key == nil {
		t.Errorf("failed to call Key() before calling, call First() or Assert()")
		return nil
	}
	svc := kms.New(a.client)
	out, err := svc.DescribeKey(&kms.DescribeKeyInput{
		KeyId: a.key.KeyId,
	})
	if err != nil {
		t.Errorf("failed to DescribeKey for targetKeyId: %q -> %v",
			aws.StringValue(a.key.KeyId), err)
		return nil
	}
	return out.KeyMetadata
}

// Policy returns a newly instantiated *policy.Policy
// using the KeyId as the required keyID value
// requires a prior call to Assert or First to "select"
// the Key whose KeyId will be used
func (a *Key) Policy(t *testing.T) *policy.Policy {
	if a.key == nil {
		t.Errorf("failed to call Policy() before calling, call First() or Assert()")
		return nil
	}
	return policy.New(a.client, aws.StringValue(a.Selected().KeyId))
}

// Assert applies all filters that have been called, resets the list of filters,
// fails the test if there is not exactly one match, and stores the matched key
// if keys is not provided, *kms.KeyMetadata objects will be retreived from AWS
func (a *Key) Assert(t *testing.T, keys ...*kms.KeyMetadata) *Key {
	var err error
	keys, err = a.filter(keys)
	if err != nil {
		t.Fatal(err)
	}

	switch l := len(keys); {
	case l == 0:
		t.Fatal("no matching key was found")
	case l > 1:
		t.Fatal("more than one matching key was found")
	default:
		a.key = keys[0]
	}

	a.filters = []shared.Filter{}
	return a
}

// First applies all filters that have been called, resets the list of filters,
// fails the test if there are no matches, and stores the first match
// if keys is not provided, *kms.KeyMetadata objects will be retreived from AWS
func (a *Key) First(t *testing.T, keys ...*kms.KeyMetadata) *Key {
	var err error
	keys, err = a.filter(keys)
	if err != nil {
		t.Fatal(err)
	}

	if len(keys) == 0 {
		t.Fatal("no matching key was found")
	} else {
		a.key = keys[0]
	}

	a.filters = []shared.Filter{}
	return a
}

// Arn adds the Arn filter to the filter list
// the Arn filter: filters keys by Arn where 'arn' provided
// is the expected Arn value
func (a *Key) Arn(arn string) *Key {
	a.filters = append(a.filters, func(v interface{}) bool {
		key := convert(v)
		if key == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", arn, aws.StringValue(key.Arn), arn == aws.StringValue(key.Arn))
		return arn == aws.StringValue(key.Arn)
	})
	return a
}

// Filter adds the 'filter' provided to the filter list
func (a *Key) Filter(filter shared.Filter) *Key {
	a.filters = append(a.filters, filter)
	return a
}

// ID adds the ID filter to the filter list
// the ID filter: filters keys by ID where 'id' provided
// is the expected KeyId value
func (a *Key) ID(id string) *Key {
	a.filters = append(a.filters, func(v interface{}) bool {
		key := convert(v)
		if key == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", id, aws.StringValue(key.KeyId), id == aws.StringValue(key.KeyId))
		return id == aws.StringValue(key.KeyId)
	})
	return a
}

// Description adds the Description filter to the filter list
// the Description filter: filters keys by Description where 'description' provided
// is the expected key description value
func (a *Key) Description(description string) *Key {
	a.filters = append(a.filters, func(v interface{}) bool {
		key := convert(v)
		if key == nil {
			return false
		}
		shared.Debugf(
			"%s like %s -> %t\n",
			description,
			aws.StringValue(key.Description),
			strings.EqualFold(description, aws.StringValue(key.Description)),
		)
		return strings.EqualFold(description, aws.StringValue(key.Description))
	})
	return a
}

func (a *Key) filter(keys []*kms.KeyMetadata) ([]*kms.KeyMetadata, error) {
	if len(keys) == 0 {
		var err error
		keys, err = a.keys()
		if err != nil {
			return nil, err
		}
	}
	results := fromIface(shared.GenericFilter(a.filters, toIface(keys)))
	if len(results) == 0 {
		log.Println("aws.kms.key.filter had zero results: ")
		shared.Spew(os.Stdout, keys)
	}
	return results, nil
}

func (a *Key) keys() ([]*kms.KeyMetadata, error) {
	svc := kms.New(a.client)
	var entries []*kms.KeyListEntry
	err := svc.ListKeysPages(&kms.ListKeysInput{}, func(page *kms.ListKeysOutput, lastPage bool) bool {
		entries = append(entries, page.Keys...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	keys := make([]*kms.KeyMetadata, 0, len(entries))
	for i, k := range entries {
		input := kms.DescribeKeyInput{KeyId: k.KeyId}
		resp, err := svc.DescribeKey(&input)
		if err != nil {
			return nil, err
		}
		keys[i] = resp.KeyMetadata
	}
	return keys, nil
}

func convert(in interface{}) *kms.KeyMetadata {
	out, ok := in.(*kms.KeyMetadata)
	if !ok {
		shared.Debugf("object not convertible to *kms.KeyMetadata: ")
		shared.Dump(in)
		return nil
	}
	return out
}
func toIface(in []*kms.KeyMetadata) (out []interface{}) {
	for _, i := range in {
		out = append(out, i)
	}
	return
}
func fromIface(in []interface{}) (out []*kms.KeyMetadata) {
	for _, i := range in {
		v := convert(i)
		if v == nil {
			continue
		}
		out = append(out, v)
	}
	return
}
