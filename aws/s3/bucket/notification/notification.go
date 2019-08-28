package notification

import (
	"testing"

	"github.com/GSA/grace-tftest/aws/shared"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Filter is an interface for filtering *Configuration objects
type Filter func(*Configuration) bool

// Notification contains the necessary properties for filtering *Configuration objects
type Notification struct {
	filters []Filter
	client  client.ConfigProvider
	config  *Configuration
	name    string
}

// New returns a new *Notification
func New(client client.ConfigProvider, name string) *Notification {
	return &Notification{
		client: client,
		name:   name,
	}
}

// Selected returns the currently selected *Configuration
func (n *Notification) Selected() *Configuration {
	return n.config
}

// Assert applies all filters that have been called, resets the list of filters,
// fails the test if there is not exactly one match, and stores the matched config
// if configs is not provided, *Configuration objects will be retreived from AWS
func (n *Notification) Assert(t *testing.T, configs ...*Configuration) *Notification {
	var err error
	configs, err = n.filter(configs)
	if err != nil {
		t.Error(err)
	}

	if len(configs) == 0 {
		t.Error("no matching configuration was found")
	} else if len(configs) > 1 {
		t.Error("more than one matching configuration was found")
	} else {
		n.config = configs[0]
	}

	n.filters = []Filter{}
	return n
}

// First applies all filters that have been called, resets the list of filters,
// fails the test if there is not a match, and stores the first matched config
// if configs is not provided, *Configuration objects will be retreived from AWS
func (n *Notification) First(t *testing.T, configs ...*Configuration) *Notification {
	var err error
	configs, err = n.filter(configs)
	if err != nil {
		t.Error(err)
	}

	if len(configs) == 0 {
		t.Error("no matching configuration was found")
	} else {
		n.config = configs[0]
	}

	n.filters = []Filter{}
	return n
}

// Filter adds the 'filter' provided to the filter list
func (n *Notification) Filter(filter Filter) *Notification {
	n.filters = append(n.filters, filter)
	return n
}

// Arn adds the Arn filter to the filter list
// the Arn filter: filters configs by Arn where 'arn' provided
// is the expected Arn value
func (n *Notification) Arn(arn string) *Notification {
	n.filters = append(n.filters, func(c *Configuration) bool {
		return arn == c.Arn
	})
	return n
}

// ID adds the ID filter to the filter list
// the ID filter: filters configs by ID where 'id' provided
// is the expected ID value
func (n *Notification) ID(id string) *Notification {
	n.filters = append(n.filters, func(c *Configuration) bool {
		return id == c.ID
	})
	return n
}

// Events adds the Events filter to the filter list
// the Events filter: filters configs by Events where 'event' provided
// is the expected Events value
func (n *Notification) Events(event ...string) *Notification {
	n.filters = append(n.filters, func(c *Configuration) bool {
		return shared.StringSliceEqual(c.Events, event)
	})
	return n
}

// Rule adds the Rule filter to the filter list
// the Rule filter: filters configs by FilterRule where 'name and value' provided
// is the expected FilterRule name and value
func (n *Notification) Rule(name, value string) *Notification {
	n.filters = append(n.filters, func(c *Configuration) bool {
		for _, f := range c.Filter {
			if f.Name == name && f.Value == value {
				return true
			}
		}
		return false
	})
	return n
}

const strBucketPrefix = "Prefix"
const strBucketSuffix = "Suffix"

// Prefix adds the Prefix filter to the filter list
// the Prefix filter: filters policies by Filter[Name:Prefix] where 'value' provided
// is the expected prefix value
func (n *Notification) Prefix(value string) *Notification {
	return n.Rule(strBucketPrefix, value)
}

// Suffix adds the Suffix filter to the filter list
// the Suffix filter: filters policies by Filter[Name:Suffix] where 'value' provided
// is the expected suffix value
func (n *Notification) Suffix(value string) *Notification {
	return n.Rule(strBucketSuffix, value)
}

func (n *Notification) filter(configs []*Configuration) (result []*Configuration, err error) {
	if len(configs) == 0 {
		configs, err = n.configs()
		if err != nil {
			return nil, err
		}
	}
outer:
	for _, config := range configs {
		for _, f := range n.filters {
			if !f(config) {
				continue outer
			}
		}
		result = append(result, config)
	}
	return
}

// Configuration ... is a generic structure used for normalizing
// the three different types of S3 notifications.
// types: LambdaFunctionConfiguration, QueueConfiguration, and TopicConfiguration
type Configuration struct {
	// Events is a required field
	Events []string

	// A container for object key name filtering rules. For information about key
	// name filtering, see Configuring Event Notifications (http://docs.aws.amazon.com/AmazonS3/latest/dev/NotificationHowTo.html)
	// in the Amazon Simple Storage Service Developer Guide.
	Filter []*FilterRule

	// An optional unique identifier for configurations in a notification configuration.
	// If you don't provide one, Amazon S3 will assign an ID.
	ID string

	// The Amazon Resource Name (ARN) of the resource triggered
	//
	// Arn is a required field
	Arn string
}

// FilterRule ... is a normalized version of s3.FilterRule without *string
type FilterRule struct {
	Name  string
	Value string
}

func (n *Notification) configs() ([]*Configuration, error) {
	svc := s3.New(n.client)
	out, err := svc.GetBucketNotificationConfiguration(&s3.GetBucketNotificationConfigurationRequest{
		Bucket: &n.name,
	})
	if err != nil {
		return nil, err
	}
	var configs []*Configuration
	configs = append(configs, convertLambdaConfigs(out.LambdaFunctionConfigurations)...)
	configs = append(configs, convertQueueConfigs(out.QueueConfigurations)...)
	configs = append(configs, convertTopicConfigs(out.TopicConfigurations)...)
	return configs, nil
}

func convertLambdaConfigs(in []*s3.LambdaFunctionConfiguration) (out []*Configuration) {
	for _, i := range in {
		out = append(out, &Configuration{
			Events: aws.StringValueSlice(i.Events),
			Filter: convertFilterRules(i.Filter.Key.FilterRules),
			ID:     aws.StringValue(i.Id),
			Arn:    aws.StringValue(i.LambdaFunctionArn),
		})
	}
	return
}
func convertQueueConfigs(in []*s3.QueueConfiguration) (out []*Configuration) {
	for _, i := range in {
		out = append(out, &Configuration{
			Events: aws.StringValueSlice(i.Events),
			Filter: convertFilterRules(i.Filter.Key.FilterRules),
			ID:     aws.StringValue(i.Id),
			Arn:    aws.StringValue(i.QueueArn),
		})
	}
	return
}
func convertTopicConfigs(in []*s3.TopicConfiguration) (out []*Configuration) {
	for _, i := range in {
		out = append(out, &Configuration{
			Events: aws.StringValueSlice(i.Events),
			Filter: convertFilterRules(i.Filter.Key.FilterRules),
			ID:     aws.StringValue(i.Id),
			Arn:    aws.StringValue(i.TopicArn),
		})
	}
	return
}

func convertFilterRules(in []*s3.FilterRule) (out []*FilterRule) {
	for _, i := range in {
		out = append(out, &FilterRule{
			Name:  aws.StringValue(i.Name),
			Value: aws.StringValue(i.Value),
		})
	}
	return
}
