// Package sns provides methods and filters to test AWS SNS resources
package sns

import (
	"github.com/GSA/grace-tftest/aws/sns/topic"
	"github.com/aws/aws-sdk-go/aws/client"
)

// Service contains all the supported types for SNS
type Service struct {
	Topic *topic.Topic
}

// New returns a new *Service
func New(client client.ConfigProvider) *Service {
	return &Service{
		Topic: topic.New(client),
	}
}
