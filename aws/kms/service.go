package kms

import (
	"github.com/GSA/grace-tftest/aws/kms/alias"
	"github.com/aws/aws-sdk-go/aws/client"
)

// Service contains all the supported types for S3
type Service struct {
	Alias *alias.Alias
}

// New returns a new *Service
func New(client client.ConfigProvider) *Service {
	return &Service{
		Alias: alias.New(client),
	}
}
