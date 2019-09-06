package alias

import (
	"testing"

	"github.com/GSA/grace-tftest/aws/kms/policy"
	"github.com/GSA/grace-tftest/aws/shared"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/kms"
)

// Alias contains the necessary properties for testing *kms.AliasListEntry objects
type Alias struct {
	client  client.ConfigProvider
	alias   *kms.AliasListEntry
	filters []shared.Filter
}

// New returns a new *Alias
func New(client client.ConfigProvider) *Alias {
	return &Alias{client: client}
}

// Selected returns the currently selected *kms.AliasListEntry
func (a *Alias) Selected() *kms.AliasListEntry {
	return a.alias
}

// Policy returns a newly instantiated *policy.Policy
// using the TargetKeyId as the required keyID value
// requires a prior call to Assert or First to "select"
// the Alias whose TargetKeyId will be used
func (a *Alias) Policy(t *testing.T) *policy.Policy {
	if a.Selected() == nil {
		t.Errorf("failed to call Policy() before calling First() or Assert()")
		return nil
	}
	return policy.New(a.client, aws.StringValue(a.Selected().TargetKeyId))
}

// Assert applies all filters that have been called, resets the list of filters,
// fails the test if there is not exactly one match, and stores the matched alias
// if aliases is not provided, *kms.AliasListEntry objects will be retreived from AWS
func (a *Alias) Assert(t *testing.T, aliases ...*kms.AliasListEntry) *Alias {
	var err error
	aliases, err = a.filter(aliases)
	if err != nil {
		t.Error(err)
	}

	switch l := len(aliases); {
	case l == 0:
		t.Error("no matching alias was found")
	case l > 1:
		t.Error("more than one matching alias was found")
	default:
		a.alias = aliases[0]
	}

	a.filters = []shared.Filter{}
	return a
}

// First applies all filters that have been called, resets the list of filters,
// fails the test if there are no matches, and stores the first match
// if aliases is not provided, *kms.AliasListEntry objects will be retreived from AWS
func (a *Alias) First(t *testing.T, aliases ...*kms.AliasListEntry) *Alias {
	var err error
	aliases, err = a.filter(aliases)
	if err != nil {
		t.Error(err)
	}

	if len(aliases) == 0 {
		t.Error("no matching alias was found")
	} else {
		a.alias = aliases[0]
	}

	a.filters = []shared.Filter{}
	return a
}

// Arn adds the Arn filter to the filter list
// the Arn filter: filters aliases by Arn where 'arn' provided
// is the expected Arn value
func (a *Alias) Arn(arn string) *Alias {
	a.filters = append(a.filters, func(v interface{}) bool {
		alias := convert(v)
		if alias == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", arn, aws.StringValue(alias.AliasArn), arn == aws.StringValue(alias.AliasArn))
		return arn == aws.StringValue(alias.AliasArn)
	})
	return a
}

// Filter adds the 'filter' provided to the filter list
func (a *Alias) Filter(filter shared.Filter) *Alias {
	a.filters = append(a.filters, filter)
	return a
}

// ID adds the ID filter to the filter list
// the ID filter: filters aliases by ID where 'id' provided
// is the expected TargetKeyId value
func (a *Alias) ID(id string) *Alias {
	a.filters = append(a.filters, func(v interface{}) bool {
		alias := convert(v)
		if alias == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", id, aws.StringValue(alias.TargetKeyId), id == aws.StringValue(alias.TargetKeyId))
		return id == aws.StringValue(alias.TargetKeyId)
	})
	return a
}

// Name adds the Name filter to the filter list
// the Name filter: filters aliases by Name where 'name' provided
// is the expected PolicyName value
func (a *Alias) Name(name string) *Alias {
	a.filters = append(a.filters, func(v interface{}) bool {
		alias := convert(v)
		if alias == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", name, aws.StringValue(alias.AliasName), name == aws.StringValue(alias.AliasName))
		return name == aws.StringValue(alias.AliasName)
	})
	return a
}

func (a *Alias) filter(aliases []*kms.AliasListEntry) ([]*kms.AliasListEntry, error) {
	if len(aliases) == 0 {
		var err error
		aliases, err = a.aliases()
		if err != nil {
			return nil, err
		}
	}
	return fromIface(shared.GenericFilter(a.filters, toIface(aliases))), nil
}

func (a *Alias) aliases() ([]*kms.AliasListEntry, error) {
	svc := kms.New(a.client)
	var aliases []*kms.AliasListEntry
	err := svc.ListAliasesPages(&kms.ListAliasesInput{}, func(page *kms.ListAliasesOutput, lastPage bool) bool {
		aliases = append(aliases, page.Aliases...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	return aliases, nil
}

func convert(in interface{}) *kms.AliasListEntry {
	out, ok := in.(*kms.AliasListEntry)
	if !ok {
		shared.Debugf("object not convertible to *kms.AliasListEntry: ")
		shared.Dump(in)
		return nil
	}
	return out
}
func toIface(in []*kms.AliasListEntry) (out []interface{}) {
	for _, i := range in {
		out = append(out, i)
	}
	return
}
func fromIface(in []interface{}) (out []*kms.AliasListEntry) {
	for _, i := range in {
		v := convert(i)
		if v == nil {
			continue
		}
		out = append(out, v)
	}
	return
}
