package role

import (
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/GSA/grace-tftest/aws/iam/role/attached"
	"github.com/GSA/grace-tftest/aws/shared"
	"github.com/GSA/grace-tftest/aws/shared/policy"
	"github.com/GSA/grace-tftest/aws/shared/policy/statement"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/iam"
)

// Role contains the necessary properties for testing *iam.Role objects
type Role struct {
	client  client.ConfigProvider
	role    *iam.Role
	filters []shared.Filter
}

// New returns a new *Role
func New(client client.ConfigProvider) *Role {
	return &Role{client: client}
}

// Selected returns the currently selected *iam.Role
func (r *Role) Selected() *iam.Role {
	return r.role
}

// Statement returns a newly instantiated *statement.Statement object
// this is used for filtering by statements inside the AssumeRolePolicyDocument
func (r *Role) Statement(t *testing.T) *statement.Statement {
	doc := r.Document(t)
	return statement.New(doc)
}

// Attached returns a newly instantiated *attached.Attached object
// used for finding *iam.AttachedPolicy objects
func (r *Role) Attached() *attached.Attached {
	return attached.New(r.client, aws.StringValue(r.role.RoleName))
}

// Inlined returns a newly instantiated *statement.Statement object
// used for filtering inlined Role Policies
func (r *Role) Inlined(t *testing.T, doc *policy.Document) *statement.Statement {
	if r.role == nil {
		t.Errorf("failed to call Inlined() before calling, call First() or Assert()")
		return nil
	}
	if doc == nil {
		statements, err := r.inlined()
		if err != nil {
			t.Errorf("failed to query inlined policies: %v", err)
			return nil
		}
		doc = &policy.Document{Statement: statements}
	}
	return statement.New(doc)
}

// Assert applies all filters that have been called, resets the list of filters,
// fails the test if there is not exactly one match, and stores the matched role
// if roles is not provided, *iam.Role objects will be retreived from AWS
func (r *Role) Assert(t *testing.T, roles ...*iam.Role) *Role {
	var err error
	roles, err = r.filter(roles)
	if err != nil {
		t.Fatal(err)
	}

	switch l := len(roles); {
	case l == 0:
		t.Fatal("no matching role was found")
	case l > 1:
		t.Fatal("more than one matching role was found")
	default:
		r.role = roles[0]
	}

	r.filters = []shared.Filter{}
	return r
}

// First applies all filters that have been called, resets the list of filters,
// fails the test if there are no matches, and stores the first match
// if roles is not provided, *iam.Role objects will be retreived from AWS
func (r *Role) First(t *testing.T, roles ...*iam.Role) *Role {
	var err error
	roles, err = r.filter(roles)
	if err != nil {
		t.Fatal(err)
	}

	if len(roles) == 0 {
		t.Fatal("no matching role was found")
	} else {
		r.role = roles[0]
	}

	r.filters = []shared.Filter{}
	return r
}

// Arn adds the Arn filter to the filter list
// the Arn filter: filters roles by Arn where 'arn' provided
// is the expected Arn value
func (r *Role) Arn(arn string) *Role {
	r.filters = append(r.filters, func(v interface{}) bool {
		role := convert(v)
		if role == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", arn, aws.StringValue(role.Arn), arn == aws.StringValue(role.Arn))
		return arn == aws.StringValue(role.Arn)
	})
	return r
}

func (r *Role) Document(t *testing.T) *policy.Document {
	doc, err := policy.Unmarshal(aws.StringValue(r.role.AssumeRolePolicyDocument))
	if err != nil {
		t.Errorf("failed to unmarshal policy document: %v", err)
		return nil
	}
	return doc
}

// Filter adds the 'filter' provided to the filter list
func (r *Role) Filter(filter shared.Filter) *Role {
	r.filters = append(r.filters, filter)
	return r
}

// ID adds the ID filter to the filter list
// the ID filter: filters roles by ID where 'id' provided
// is the expected RoleId value
func (r *Role) ID(id string) *Role {
	r.filters = append(r.filters, func(v interface{}) bool {
		role := convert(v)
		if role == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", id, aws.StringValue(role.RoleId), id == aws.StringValue(role.RoleId))
		return id == aws.StringValue(role.RoleId)
	})
	return r
}

// Name adds the Name filter to the filter list
// the Name filter: filters roles by Name where 'name' provided
// is the expected RoleName value
func (r *Role) Name(name string) *Role {
	r.filters = append(r.filters, func(v interface{}) bool {
		role := convert(v)
		if role == nil {
			return false
		}
		shared.Debugf("%s like %s -> %t\n", name, aws.StringValue(role.RoleName), strings.EqualFold(name, aws.StringValue(role.RoleName)))
		return strings.EqualFold(name, aws.StringValue(role.RoleName))
	})
	return r
}

func (r *Role) filter(roles []*iam.Role) ([]*iam.Role, error) {
	if len(roles) == 0 {
		var err error
		roles, err = r.roles()
		if err != nil {
			return nil, err
		}
	}
	results := fromIface(shared.GenericFilter(r.filters, toIface(roles)))
	if len(results) == 0 {
		log.Println("aws.iam.role.filter had zero results: ")
		shared.Spew(os.Stdout, roles)
	}
	return results, nil
}

func (r *Role) roles() ([]*iam.Role, error) {
	svc := iam.New(r.client)
	var roles []*iam.Role
	err := svc.ListRolesPages(&iam.ListRolesInput{}, func(page *iam.ListRolesOutput, lastPage bool) bool {
		roles = append(roles, page.Roles...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *Role) inlined() ([]*policy.Statement, error) {
	svc := iam.New(r.client)
	var names []*string
	err := svc.ListRolePoliciesPages(&iam.ListRolePoliciesInput{
		RoleName: r.role.RoleName,
	}, func(page *iam.ListRolePoliciesOutput, lastPage bool) bool {
		names = append(names, page.PolicyNames...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	var statements []*policy.Statement
	for _, n := range names {
		out, err := svc.GetRolePolicy(&iam.GetRolePolicyInput{
			RoleName:   r.role.RoleName,
			PolicyName: n,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get inline policy for role: %s -> %v", aws.StringValue(r.role.RoleName), err)
		}
		doc, err := policy.Unmarshal(aws.StringValue(out.PolicyDocument))
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal inline policy document for role: %s -> %v", aws.StringValue(r.role.RoleName), err)
		}
		statements = append(statements, doc.Statement...)
	}
	return statements, nil
}

func convert(in interface{}) *iam.Role {
	out, ok := in.(*iam.Role)
	if !ok {
		shared.Debugf("object not convertible to *iam.Role: ")
		shared.Dump(in)
		return nil
	}
	return out
}
func toIface(in []*iam.Role) (out []interface{}) {
	for _, i := range in {
		out = append(out, i)
	}
	return
}
func fromIface(in []interface{}) (out []*iam.Role) {
	for _, i := range in {
		v := convert(i)
		if v == nil {
			continue
		}
		out = append(out, v)
	}
	return
}
