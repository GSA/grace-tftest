package description

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

// Description contains the necessary properties for testing *kms.KeyMetadata objects
type Description struct {
	client      client.ConfigProvider
	description *kms.KeyMetadata
	filters     []shared.Filter
}

// New returns a new *Description
func New(client client.ConfigProvider) *Description {
	return &Description{client: client}
}

// Selected returns the currently selected *kms.KeyMetadata
func (a *Description) Selected() *kms.KeyMetadata {
	return a.description
}

// Key returns the currently selected Descriptions' targeted *kms.KeyMetadata
func (a *Description) Key(t *testing.T) *kms.KeyMetadata {
	if a.description == nil {
		t.Errorf("failed to call Key() before calling, call First() or Assert()")
		return nil
	}
	svc := kms.New(a.client)
	out, err := svc.DescribeKey(&kms.DescribeKeyInput{
		KeyId: a.description.KeyId,
	})
	if err != nil {
		t.Errorf("failed to DescribeKey for targetKeyId: %q -> %v",
			aws.StringValue(a.description.KeyId), err)
		return nil
	}
	return out.KeyMetadata
}

// Policy returns a newly instantiated *policy.Policy
// using the KeyId as the required keyID value
// requires a prior call to Assert or First to "select"
// the Description whose KeyId will be used
func (a *Description) Policy(t *testing.T) *policy.Policy {
	if a.Selected() == nil {
		t.Errorf("failed to call Policy() before calling, call First() or Assert()")
		return nil
	}
	return policy.New(a.client, aws.StringValue(a.Selected().KeyId))
}

// Assert applies all filters that have been called, resets the list of filters,
// fails the test if there is not exactly one match, and stores the matched description
// if descriptions is not provided, *kms.KeyMetadata objects will be retreived from AWS
func (a *Description) Assert(t *testing.T, descriptions ...*kms.KeyMetadata) *Description {
	var err error
	descriptions, err = a.filter(descriptions)
	if err != nil {
		t.Fatal(err)
	}

	switch l := len(descriptions); {
	case l == 0:
		t.Fatal("no matching description was found")
	case l > 1:
		t.Fatal("more than one matching description was found")
	default:
		a.description = descriptions[0]
	}

	a.filters = []shared.Filter{}
	return a
}

// First applies all filters that have been called, resets the list of filters,
// fails the test if there are no matches, and stores the first match
// if descriptions is not provided, *kms.KeyMetadata objects will be retreived from AWS
func (a *Description) First(t *testing.T, descriptions ...*kms.KeyMetadata) *Description {
	var err error
	descriptions, err = a.filter(descriptions)
	if err != nil {
		t.Fatal(err)
	}

	if len(descriptions) == 0 {
		t.Fatal("no matching description was found")
	} else {
		a.description = descriptions[0]
	}

	a.filters = []shared.Filter{}
	return a
}

// Arn adds the Arn filter to the filter list
// the Arn filter: filters descriptions by Arn where 'arn' provided
// is the expected Arn value
func (a *Description) Arn(arn string) *Description {
	a.filters = append(a.filters, func(v interface{}) bool {
		description := convert(v)
		if description == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", arn, aws.StringValue(description.Arn), arn == aws.StringValue(description.Arn))
		return arn == aws.StringValue(description.Arn)
	})
	return a
}

// Filter adds the 'filter' provided to the filter list
func (a *Description) Filter(filter shared.Filter) *Description {
	a.filters = append(a.filters, filter)
	return a
}

// ID adds the ID filter to the filter list
// the ID filter: filters descriptions by ID where 'id' provided
// is the expected KeyId value
func (a *Description) ID(id string) *Description {
	a.filters = append(a.filters, func(v interface{}) bool {
		description := convert(v)
		if description == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", id, aws.StringValue(description.KeyId), id == aws.StringValue(description.KeyId))
		return id == aws.StringValue(description.KeyId)
	})
	return a
}

// Name adds the Name filter to the filter list
// the Name filter: filters descriptions by Name where 'name' provided
// is the expected PolicyName value
func (a *Description) Name(name string) *Description {
	a.filters = append(a.filters, func(v interface{}) bool {
		description := convert(v)
		if description == nil {
			return false
		}
		shared.Debugf("%s like %s -> %t\n", name, aws.StringValue(description.Description), strings.EqualFold(name, aws.StringValue(description.Description)))
		return strings.EqualFold(name, aws.StringValue(description.Description))
	})
	return a
}

func (a *Description) filter(descriptions []*kms.KeyMetadata) ([]*kms.KeyMetadata, error) {
	if len(descriptions) == 0 {
		var err error
		descriptions, err = a.descriptions()
		if err != nil {
			return nil, err
		}
	}
	results := fromIface(shared.GenericFilter(a.filters, toIface(descriptions)))
	if len(results) == 0 {
		log.Println("aws.kms.description.filter had zero results: ")
		shared.Spew(os.Stdout, descriptions)
	}
	return results, nil
}

func (a *Description) descriptions() ([]*kms.KeyMetadata, error) {
	svc := kms.New(a.client)
	var descriptions []*kms.KeyMetadata
	err := svc.ListDescriptionsPages(&kms.ListDescriptionsInput{}, func(page *kms.ListDescriptionsOutput, lastPage bool) bool {
		descriptions = append(descriptions, page.Descriptions...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	return descriptions, nil
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
