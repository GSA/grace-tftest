package notification

import (
	"testing"

	"github.com/GSA/grace-tftest/aws/shared"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Notification contains the necessary properties for filtering *Configuration objects
type Notification struct {
	filters []shared.Filter
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
		t.Fatal(err)
	}

	switch l := len(configs); {
	case l == 0:
		t.Fatal("no matching configuration was found")
	case l > 1:
		t.Fatal("more than one matching configuration was found")
	default:
		n.config = configs[0]
	}

	n.filters = []shared.Filter{}
	return n
}

// First applies all filters that have been called, resets the list of filters,
// fails the test if there is not a match, and stores the first matched config
// if configs is not provided, *Configuration objects will be retreived from AWS
func (n *Notification) First(t *testing.T, configs ...*Configuration) *Notification {
	var err error
	configs, err = n.filter(configs)
	if err != nil {
		t.Fatal(err)
	}

	if len(configs) == 0 {
		t.Fatal("no matching configuration was found")
	} else {
		n.config = configs[0]
	}

	n.filters = []shared.Filter{}
	return n
}

// Filter adds the 'filter' provided to the filter list
func (n *Notification) Filter(filter shared.Filter) *Notification {
	n.filters = append(n.filters, filter)
	return n
}

// Arn adds the Arn filter to the filter list
// the Arn filter: filters configs by Arn where 'arn' provided
// is the expected Arn value
func (n *Notification) Arn(arn string) *Notification {
	n.filters = append(n.filters, func(v interface{}) bool {
		c := convert(v)
		if c == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", arn, c.Arn, arn == c.Arn)
		return arn == c.Arn
	})
	return n
}

// ID adds the ID filter to the filter list
// the ID filter: filters configs by ID where 'id' provided
// is the expected ID value
func (n *Notification) ID(id string) *Notification {
	n.filters = append(n.filters, func(v interface{}) bool {
		c := convert(v)
		if c == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", id, c.ID, id == c.ID)
		return id == c.ID
	})
	return n
}

// Events adds the Events filter to the filter list
// the Events filter: filters configs by Events where 'event' provided
// is the expected Events value
func (n *Notification) Events(event ...string) *Notification {
	n.filters = append(n.filters, func(v interface{}) bool {
		c := convert(v)
		if c == nil {
			return false
		}
		shared.Debugf("%v == %v\n", event, c.Events)
		return shared.StringSliceEqual(event, c.Events)
	})
	return n
}

// Rule adds the Rule filter to the filter list
// the Rule filter: filters configs by FilterRule where 'name and value' provided
// is the expected FilterRule name and value
func (n *Notification) Rule(name, value string) *Notification {
	n.filters = append(n.filters, func(v interface{}) bool {
		c := convert(v)
		if c == nil {
			return false
		}
		shared.Debugf("len(c.Filter) = %d", len(c.Filter))
		for _, f := range c.Filter {
			shared.Debugf("Name: %s == %s -> %t, Value: %s == %s -> %t\n",
				name, f.Name, name == f.Name,
				value, f.Value, value == f.Value)
			if name == f.Name && value == f.Value {
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

func (n *Notification) filter(configs []*Configuration) ([]*Configuration, error) {
	if len(configs) == 0 {
		var err error
		configs, err = n.configs()
		if err != nil {
			return nil, err
		}
	}
	return fromIface(shared.GenericFilter(n.filters, toIface(configs))), nil
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

func convert(v interface{}) *Configuration {
	statement, ok := v.(*Configuration)
	if !ok {
		shared.Debugf("object not convertible to *Configuration: ")
		shared.Dump(v)
		return nil
	}
	return statement
}
func toIface(in []*Configuration) (out []interface{}) {
	for _, i := range in {
		out = append(out, i)
	}
	return
}
func fromIface(in []interface{}) (out []*Configuration) {
	for _, i := range in {
		v := convert(i)
		if v == nil {
			continue
		}
		out = append(out, v)
	}
	return
}
