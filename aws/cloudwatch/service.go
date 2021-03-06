// Package cloudwatch provides methods and filters to test AWS CloudWatch resources
package cloudwatch

import (
	"github.com/GSA/grace-tftest/aws/cloudwatch/metric"
	"github.com/aws/aws-sdk-go/aws/client"
)

// Service contains all the supported types for IAM
type Service struct {
	Metric *metric.Metric
}

// New returns a new *Service
func New(client client.ConfigProvider) *Service {
	return &Service{
		Metric: metric.New(client),
	}
}
