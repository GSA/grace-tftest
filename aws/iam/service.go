package iam

import (
	"github.com/GSA/grace-tftest/aws/iam/policy"
	"github.com/GSA/grace-tftest/aws/iam/role"
	"github.com/aws/aws-sdk-go/aws/client"
)

// Service contains all the supported types for IAM
type Service struct {
	Policy *policy.Policy
	Role   *role.Role
}

// New returns a new *Service
func New(client client.ConfigProvider) *Service {
	return &Service{
		Policy: policy.New(client),
		Role:   role.New(client),
	}
}
