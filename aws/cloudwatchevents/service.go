// Package cloudwatchevents provides functions and filters to test AWS CloudWatch Event resources
package cloudwatchevents

import (
	"github.com/GSA/grace-tftest/aws/cloudwatchevents/bus"
	"github.com/GSA/grace-tftest/aws/cloudwatchevents/rule"
	"github.com/aws/aws-sdk-go/aws/client"
)

// Service contains all the supported types for IAM
type Service struct {
	Bus  *bus.Bus
	Rule *rule.Rule
}

// New returns a new *Service
func New(client client.ConfigProvider) *Service {
	return &Service{
		Bus:  bus.New(client),
		Rule: rule.New(client),
	}
}
