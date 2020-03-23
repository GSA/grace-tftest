package kms

import (
	"github.com/GSA/grace-tftest/aws/kms/alias"
	"github.com/GSA/grace-tftest/aws/kms/key"
	"github.com/aws/aws-sdk-go/aws/client"
)

// Service contains all the supported types for S3
type Service struct {
	Alias *alias.Alias
	Key   *key.Key
}

// New returns a new *Service
func New(client client.ConfigProvider) *Service {
	return &Service{
		Alias: alias.New(client),
		Key:   key.New(client),
	}
}
