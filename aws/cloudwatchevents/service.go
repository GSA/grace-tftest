package cloudwatchevents

import (
	"github.com/GSA/grace-tftest/aws/cloudwatchevents/rule"
	"github.com/aws/aws-sdk-go/aws/client"
)

// Service contains all the supported types for IAM
type Service struct {
	Rule *rule.Rule
}

// New returns a new *Service
func New(client client.ConfigProvider) *Service {
	return &Service{
		Rule: rule.New(client),
	}
}
