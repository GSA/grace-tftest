// Package cloudtrail provides methods and filters to test AWS CloudTrail resources
package cloudtrail

import (
	"github.com/GSA/grace-tftest/aws/cloudtrail/trail"
	"github.com/aws/aws-sdk-go/aws/client"
)

// Service contains all the supported types for CloudTrail
type Service struct {
	Trail *trail.Trail
}

// New returns a new *Service
func New(client client.ConfigProvider) *Service {
	return &Service{
		Trail: trail.New(client),
	}
}
