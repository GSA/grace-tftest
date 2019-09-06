package s3

import (
	"github.com/GSA/grace-tftest/aws/s3/bucket"
	"github.com/aws/aws-sdk-go/aws/client"
)

// Service contains all the supported types for S3
type Service struct {
	Bucket *bucket.Bucket
}

// New returns a new *Service
func New(client client.ConfigProvider) *Service {
	return &Service{
		Bucket: bucket.New(client),
	}
}
