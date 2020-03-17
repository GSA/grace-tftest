// config provides access to filters related to AWS Config
package config

import (
	"github.com/GSA/grace-tftest/aws/config/recorder"
	"github.com/aws/aws-sdk-go/aws/client"
)

// Service contains all the supported types for AWS Config
type Service struct {
	Recorder *recorder.Recorder
}

// New returns a new *Service
func New(client client.ConfigProvider) *Service {
	return &Service{
		Recorder: recorder.New(client),
	}
}
