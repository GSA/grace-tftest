// Package pubaccblk provides methods and filters to test AWS S3 PublicAccessBlocks
package pubaccblk

import (
	"log"
	"os"
	"testing"

	"github.com/GSA/grace-tftest/aws/shared"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/s3"
)

// PublicAccessBlock contains the necessary properties for filtering *s3.PublicAccessBlockConfiguration objects
type PublicAccessBlock struct {
	filters []shared.Filter
	client  client.ConfigProvider
	config  *s3.PublicAccessBlockConfiguration
	name    string
}

// New returns a new *PublicAccessBlock
func New(client client.ConfigProvider, name string) *PublicAccessBlock {
	return &PublicAccessBlock{
		client: client,
		name:   name,
	}
}

// Selected returns the currently selected *s3.PublicAccessBlockConfiguration
func (e *PublicAccessBlock) Selected() *s3.PublicAccessBlockConfiguration {
	return e.config
}

// Assert applies all filters that have been called, resets the list of filters,
// fails the test if there is not exactly one match, and stores the matched config
// if config is not provided, *s3.PublicAccessBlockConfiguration objects will be retreived from AWS
func (e *PublicAccessBlock) Assert(t *testing.T, configs ...*s3.PublicAccessBlockConfiguration) *PublicAccessBlock {
	var err error
	configs, err = e.filter(configs)
	if err != nil {
		t.Fatal(err)
	}

	switch l := len(configs); {
	case l == 0:
		t.Fatal("no matching public access block configuration was found")
	case l > 1:
		t.Fatal("more than one matching public access block configuration was found")
	default:
		e.config = configs[0]
	}

	e.filters = []shared.Filter{}
	return e
}

// First applies all filters that have been called, resets the list of filters,
// fails the test if there is not exactly one match, and stores the matched config
// if config is not provided, *s3.PublicAccessBlockConfiguration objects will be retreived from AWS
func (e *PublicAccessBlock) First(t *testing.T, configs ...*s3.PublicAccessBlockConfiguration) *PublicAccessBlock {
	var err error
	configs, err = e.filter(configs)
	if err != nil {
		t.Fatal(err)
	}

	if len(configs) == 0 {
		t.Fatal("no matching public access block configuration was found")
	} else {
		e.config = configs[0]
	}

	e.filters = []shared.Filter{}
	return e
}

// Filter adds the 'filter' provided to the filter list
func (e *PublicAccessBlock) Filter(filter shared.Filter) *PublicAccessBlock {
	e.filters = append(e.filters, filter)
	return e
}

// BlockPublicAcls adds the BlockPublicAcls filter to the filter list
// the BlockPublicAcls filter: filters configs by whether they have
// BlockPublicAcls set to the provided boolean value
func (e *PublicAccessBlock) BlockPublicAcls(t bool) *PublicAccessBlock {
	e.filters = append(e.filters, func(v interface{}) bool {
		config := convert(v)
		if config == nil {
			return false
		}
		shared.Debugf("%t == %t -> %t\n", aws.BoolValue(config.BlockPublicAcls), t, aws.BoolValue(config.BlockPublicAcls) == t)
		return aws.BoolValue(config.BlockPublicAcls) == t
	})
	return e
}

// BlockPublicPolicy adds the BlockPublicPolicy filter to the filter list
// the BlockPublicPolicy filter: filters configs by whether they have
// BlockPublicPolicy set to the provided boolean value
func (e *PublicAccessBlock) BlockPublicPolicy(t bool) *PublicAccessBlock {
	e.filters = append(e.filters, func(v interface{}) bool {
		config := convert(v)
		if config == nil {
			return false
		}
		shared.Debugf("%t == %t -> %t\n", aws.BoolValue(config.BlockPublicPolicy), t, aws.BoolValue(config.BlockPublicPolicy) == t)
		return aws.BoolValue(config.BlockPublicPolicy) == t
	})
	return e
}

// IgnorePublicAcls adds the IgnorePublicAcls filter to the filter list
// the IgnorePublicAcls filter: filters configs by whether they have
// IgnorePublicAcls set to the provided boolean value
func (e *PublicAccessBlock) IgnorePublicAcls(t bool) *PublicAccessBlock {
	e.filters = append(e.filters, func(v interface{}) bool {
		config := convert(v)
		if config == nil {
			return false
		}
		shared.Debugf("%t == %t -> %t\n", aws.BoolValue(config.IgnorePublicAcls), t, aws.BoolValue(config.IgnorePublicAcls) == t)
		return aws.BoolValue(config.IgnorePublicAcls) == t
	})
	return e
}

// RestrictPublicBuckets adds the RestrictPublicBuckets filter to the filter list
// the RestrictPublicBuckets filter: filters configs by whether they have
// RestrictPublicBuckets set to the provided boolean value
func (e *PublicAccessBlock) RestrictPublicBuckets(t bool) *PublicAccessBlock {
	e.filters = append(e.filters, func(v interface{}) bool {
		config := convert(v)
		if config == nil {
			return false
		}
		shared.Debugf("%t == %t -> %t\n", aws.BoolValue(config.RestrictPublicBuckets), t, aws.BoolValue(config.RestrictPublicBuckets) == t)
		return aws.BoolValue(config.RestrictPublicBuckets) == t
	})
	return e
}

func (e *PublicAccessBlock) filter(configs []*s3.PublicAccessBlockConfiguration) ([]*s3.PublicAccessBlockConfiguration, error) {
	if len(configs) == 0 {
		var err error
		configs, err = e.configs()
		if err != nil {
			return nil, err
		}
	}
	results := fromIface(shared.GenericFilter(e.filters, toIface(configs)))
	if len(results) == 0 {
		log.Println("aws.s3.bucket.pubaccblk.filter had zero results: ")
		shared.Spew(os.Stdout, configs)
	}
	return results, nil
}

func (e *PublicAccessBlock) configs() ([]*s3.PublicAccessBlockConfiguration, error) {
	svc := s3.New(e.client)
	resp, err := svc.GetPublicAccessBlock(&s3.GetPublicAccessBlockInput{
		Bucket: &e.name,
	})
	if err != nil {
		return nil, err
	}
	s := make([]*s3.PublicAccessBlockConfiguration, 1)
	s[1] = resp.PublicAccessBlockConfiguration
	return s, nil
}

func convert(in interface{}) *s3.PublicAccessBlockConfiguration {
	out, ok := in.(*s3.PublicAccessBlockConfiguration)
	if !ok {
		shared.Debugf("object not convertible to *s3.PublicAccessBlockConfiguration: ")
		shared.Dump(in)
		return nil
	}
	return out
}
func toIface(in []*s3.PublicAccessBlockConfiguration) (out []interface{}) {
	for _, i := range in {
		out = append(out, i)
	}
	return
}
func fromIface(in []interface{}) (out []*s3.PublicAccessBlockConfiguration) {
	for _, i := range in {
		v := convert(i)
		if v == nil {
			continue
		}
		out = append(out, v)
	}
	return
}
