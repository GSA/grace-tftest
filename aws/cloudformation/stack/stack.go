// Package stack contains the necessary properties for testing *cloudformation.Stack objects
package stack

import (
	"log"
	"os"
	"strings"
	"testing"

	"github.com/GSA/grace-tftest/aws/shared"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

// Stack contains the necessary properties for testing *cloudformation.Stack objects
type Stack struct {
	client  client.ConfigProvider
	stack   *cloudformation.Stack
	filters []shared.Filter
}

// New returns a new *Stack
func New(client client.ConfigProvider) *Stack {
	return &Stack{client: client}
}

// Selected returns the currently selected *cloudformation.Stack
func (r *Stack) Selected() *cloudformation.Stack {
	return r.stack
}

// Assert applies all filters that have been called, resets the list of filters,
// fails the test if there is not exactly one match, and stores the matched stack
// if stacks is not provided, *cloudformation.Stack objects will be retreived from AWS
func (r *Stack) Assert(t *testing.T, stacks ...*cloudformation.Stack) *Stack {
	var err error
	stacks, err = r.filter(stacks)
	if err != nil {
		t.Fatal(err)
	}

	switch l := len(stacks); {
	case l == 0:
		t.Fatal("no matching stack was found")
	case l > 1:
		t.Fatal("more than one matching stack was found")
	default:
		r.stack = stacks[0]
	}

	r.filters = []shared.Filter{}
	return r
}

// First applies all filters that have been called, resets the list of filters,
// fails the test if there are no matches, and stores the first match
// if stacks is not provided, *cloudformation.Stack objects will be retreived from AWS
func (r *Stack) First(t *testing.T, stacks ...*cloudformation.Stack) *Stack {
	var err error
	stacks, err = r.filter(stacks)
	if err != nil {
		t.Fatal(err)
	}

	if len(stacks) == 0 {
		t.Fatal("no matching stack was found")
	} else {
		r.stack = stacks[0]
	}

	r.filters = []shared.Filter{}
	return r
}

// StackID adds the StackID filter to the filter list
// the StackId filter: filters stacks by StackId where 'id' provided
// is the expected StackId value
func (r *Stack) StackID(id string) *Stack {
	r.filters = append(r.filters, func(v interface{}) bool {
		stack := convert(v)
		if stack == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", id, aws.StringValue(stack.StackId), id == aws.StringValue(stack.StackId))
		return id == aws.StringValue(stack.StackId)
	})
	return r
}

// Filter adds the 'filter' provided to the filter list
func (r *Stack) Filter(filter shared.Filter) *Stack {
	r.filters = append(r.filters, filter)
	return r
}

// Name adds the Name filter to the filter list
// the Name filter: filters stacks by Name where 'name' provided
// is the expected StackName value
func (r *Stack) Name(name string) *Stack {
	r.filters = append(r.filters, func(v interface{}) bool {
		stack := convert(v)
		if stack == nil {
			return false
		}
		shared.Debugf(
			"%s like %s -> %t\n",
			name,
			aws.StringValue(stack.StackName),
			strings.EqualFold(name, aws.StringValue(stack.StackName)),
		)
		return strings.EqualFold(name, aws.StringValue(stack.StackName))
	})
	return r
}

// ChangeSetID adds the ChangeSetID filter to the filter list
// the ChangeSetId filter: filters stacks by ChangeSetId where 'id' provided
// is the expected ChangeSetId value
func (r *Stack) ChangeSetID(id string) *Stack {
	r.filters = append(r.filters, func(v interface{}) bool {
		stack := convert(v)
		if stack == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", id, aws.StringValue(stack.ChangeSetId), id == aws.StringValue(stack.ChangeSetId))
		return id == aws.StringValue(stack.ChangeSetId)
	})
	return r
}

// ParentID adds the ParentID filter to the filter list
// the ParentId filter: filters stacks by ParentId where 'id' provided
// is the expected ParentId value
func (r *Stack) ParentID(id string) *Stack {
	r.filters = append(r.filters, func(v interface{}) bool {
		stack := convert(v)
		if stack == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", id, aws.StringValue(stack.ParentId), id == aws.StringValue(stack.ParentId))
		return id == aws.StringValue(stack.ParentId)
	})
	return r
}

// RootID adds the RootID filter to the filter list
// the RootId filter: filters stacks by RootId where 'id' provided
// is the expected RootId value
func (r *Stack) RootID(id string) *Stack {
	r.filters = append(r.filters, func(v interface{}) bool {
		stack := convert(v)
		if stack == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", id, aws.StringValue(stack.RootId), id == aws.StringValue(stack.RootId))
		return id == aws.StringValue(stack.RootId)
	})
	return r
}

// RoleARN adds the RoleARN filter to the filter list
// the RoleARN filter: filters stacks by RoleARN where 'arn' provided
// is the expected RoleARN value
func (r *Stack) RoleARN(arn string) *Stack {
	r.filters = append(r.filters, func(v interface{}) bool {
		stack := convert(v)
		if stack == nil {
			return false
		}
		shared.Debugf("%s like %s -> %t\n", arn, aws.StringValue(stack.RoleARN),
			strings.EqualFold(arn, aws.StringValue(stack.RoleARN)))
		return strings.EqualFold(arn, aws.StringValue(stack.RoleARN))
	})
	return r
}

// Description adds the Description filter to the filter list
// the Description filter: filters stacks by Description where 'desc' provided
// is the expected Description value
func (r *Stack) Description(desc string) *Stack {
	r.filters = append(r.filters, func(v interface{}) bool {
		stack := convert(v)
		if stack == nil {
			return false
		}
		shared.Debugf("%s like %s -> %t\n", desc, aws.StringValue(stack.Description),
			strings.EqualFold(desc, aws.StringValue(stack.Description)))
		return strings.EqualFold(desc, aws.StringValue(stack.Description))
	})
	return r
}

// StackStatus adds the StackStatus filter to the filter list
// the StackStatus filter: filters stacks by StackStatus where 'status' provided
// is the expected StackStatus value
func (r *Stack) StackStatus(status string) *Stack {
	r.filters = append(r.filters, func(v interface{}) bool {
		stack := convert(v)
		if stack == nil {
			return false
		}
		shared.Debugf(
			"%s like %s -> %t\n",
			status,
			aws.StringValue(stack.StackStatus),
			strings.EqualFold(status, aws.StringValue(stack.StackStatus)),
		)
		return strings.EqualFold(status, aws.StringValue(stack.StackStatus))
	})
	return r
}

func (r *Stack) filter(stacks []*cloudformation.Stack) ([]*cloudformation.Stack, error) {
	if len(stacks) == 0 {
		var err error
		stacks, err = r.stacks()
		if err != nil {
			return nil, err
		}
	}
	results := fromIface(shared.GenericFilter(r.filters, toIface(stacks)))
	if len(results) == 0 {
		log.Println("aws.cloudformation.stack.filter had zero results: ")
		shared.Spew(os.Stdout, stacks)
	}
	return results, nil
}

func (r *Stack) stacks() ([]*cloudformation.Stack, error) {
	svc := cloudformation.New(r.client)
	input := &cloudformation.DescribeStacksInput{}
	result, err := svc.DescribeStacks(input)
	if err != nil {
		return nil, err
	}
	stacks := result.Stacks
	token := aws.StringValue(result.NextToken)
	for token != "" {
		input.NextToken = &token
		result, err := svc.DescribeStacks(input)
		if err != nil {
			return nil, err
		}
		stacks = append(stacks, result.Stacks...)
		token = aws.StringValue(result.NextToken)
	}
	return stacks, nil
}

func convert(in interface{}) *cloudformation.Stack {
	out, ok := in.(*cloudformation.Stack)
	if !ok {
		shared.Debugf("object not convertible to *cloudformation.Stack: ")
		shared.Dump(in)
		return nil
	}
	return out
}
func toIface(in []*cloudformation.Stack) (out []interface{}) {
	for _, i := range in {
		out = append(out, i)
	}
	return
}
func fromIface(in []interface{}) (out []*cloudformation.Stack) {
	for _, i := range in {
		v := convert(i)
		if v == nil {
			continue
		}
		out = append(out, v)
	}
	return
}
