package iam

import (
	"github.com/GSA/grace-tftest/aws/iam/policy"
	"github.com/aws/aws-sdk-go/aws/client"
)

// Service contains all the supported types for IAM
type Service struct {
	Policy *policy.Policy
}

// New returns a new *Service
func New(client client.ConfigProvider) *Service {
	return &Service{
		Policy: policy.New(client),
	}
}
