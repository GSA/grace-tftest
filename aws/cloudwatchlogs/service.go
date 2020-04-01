// Package cloudwatchlogs provides methods and filters to test AWS CloudWatchlogs resources
package cloudwatchlogs

import (
	"github.com/GSA/grace-tftest/aws/cloudwatchlogs/group"
	"github.com/GSA/grace-tftest/aws/cloudwatchlogs/metricfilter"
	"github.com/aws/aws-sdk-go/aws/client"
)

// Service contains all the supported types for CloudWatchLogs
type Service struct {
	Group        *group.Group
	MetricFilter *metricfilter.MetricFilter
}

// New returns a new *Service
func New(client client.ConfigProvider) *Service {
	return &Service{
		Group:        group.New(client),
		MetricFilter: metricfilter.New(client),
	}
}
