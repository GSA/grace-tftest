// Package cloudformation provides types and functions for filtering AWS
// CloudFormation stacks
package cloudformation

import (
	"github.com/GSA/grace-tftest/aws/cloudformation/stack"
	"github.com/aws/aws-sdk-go/aws/client"
)

// Service contains all the supported types for CloudFormation
type Service struct {
	Stack *stack.Stack
}

// New returns a new *Service
func New(client client.ConfigProvider) *Service {
	return &Service{
		Stack: stack.New(client),
	}
}
