// Package config provides access to filters related to AWS Config
package config

import (
	"github.com/GSA/grace-tftest/aws/config/deliverychannel"
	"github.com/GSA/grace-tftest/aws/config/recorder"
	"github.com/GSA/grace-tftest/aws/config/rule"
	"github.com/aws/aws-sdk-go/aws/client"
)

// Service contains all the supported types for AWS Config
type Service struct {
	DeliveryChannel *deliverychannel.DeliveryChannel
	Recorder        *recorder.Recorder
	Rule            *rule.Rule
}

// New returns a new *Service
func New(client client.ConfigProvider) *Service {
	return &Service{
		DeliveryChannel: deliverychannel.New(client),
		Recorder:        recorder.New(client),
		Rule:            rule.New(client),
	}
}
