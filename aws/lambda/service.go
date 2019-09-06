package lambda

import (
	"github.com/GSA/grace-tftest/aws/lambda/config"
	"github.com/GSA/grace-tftest/aws/lambda/policy"
	"github.com/aws/aws-sdk-go/aws/client"
)

// Service contains all the supported types for Lambda
type Service struct {
	Config *config.Config
	Policy *policy.Policy
}

// New returns a new *Service
func New(client client.ConfigProvider, functionName string) *Service {
	return &Service{
		Config: config.New(client, functionName),
		Policy: policy.New(client, functionName),
	}
}
