package deliverychannel

import (
	"testing"

	"github.com/GSA/grace-tftest/aws/shared"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/configservice"
)

// DeliveryChannel contains the necessary properties for testing *configservice.DeliveryChannel objects
type DeliveryChannel struct {
	client  client.ConfigProvider
	channel *configservice.DeliveryChannel
	filters []shared.Filter
}

// New returns a new *DeliveryChannel
func New(client client.ConfigProvider) *DeliveryChannel {
	return &DeliveryChannel{client: client}
}

// Selected returns the currently selected *configservice.DeliveryChannel
func (d *DeliveryChannel) Selected() *configservice.DeliveryChannel {
	return d.channel
}

// Assert applies all filters that have been called, resets the list of filters,
// fails the test if there is not exactly one match, and stores the matched channel
// if channels is not provided, *configservice.DeliveryChannel objects will be retreived from AWS
func (d *DeliveryChannel) Assert(t *testing.T, channels ...*configservice.DeliveryChannel) *DeliveryChannel {
	var err error
	channels, err = d.filter(channels)
	if err != nil {
		t.Error(err)
	}

	switch l := len(channels); {
	case l == 0:
		t.Error("no matching channel was found")
	case l > 1:
		t.Error("more than one matching channel was found")
	default:
		d.channel = channels[0]
	}

	d.filters = []shared.Filter{}
	return d
}

// First applies all filters that have been called, resets the list of filters,
// fails the test if there are no matches, and stores the first match
// if channels is not provided, *configservice.DeliveryChannel objects will be retreived from AWS
func (d *DeliveryChannel) First(t *testing.T, channels ...*configservice.DeliveryChannel) *DeliveryChannel {
	var err error
	channels, err = d.filter(channels)
	if err != nil {
		t.Error(err)
	}

	if len(channels) == 0 {
		t.Error("no matching channel was found")
	} else {
		d.channel = channels[0]
	}

	d.filters = []shared.Filter{}
	return d
}

// Filter adds the 'filter' provided to the filter list
func (d *DeliveryChannel) Filter(filter shared.Filter) *DeliveryChannel {
	d.filters = append(d.filters, filter)
	return d
}

// TopicArn adds the TopicArn filter to the filter list
// the TopicArn filter: filters channels by TopicArn where 'arn' provided
// is the expected TopicARN value
func (d *DeliveryChannel) TopicArn(arn string) *DeliveryChannel {
	d.filters = append(d.filters, func(v interface{}) bool {
		channel := convert(v)
		if channel == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n",
			arn,
			aws.StringValue(channel.SnsTopicARN),
			arn == aws.StringValue(channel.SnsTopicARN),
		)
		return arn == aws.StringValue(channel.SnsTopicARN)
	})
	return d
}

// BucketName adds the BucketName filter to the filter list
// the BucketName filter: filters channels by BucketName where 'name' provided
// is the expected S3BucketName value
func (d *DeliveryChannel) BucketName(name string) *DeliveryChannel {
	d.filters = append(d.filters, func(v interface{}) bool {
		channel := convert(v)
		if channel == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n",
			name, aws.StringValue(channel.S3BucketName),
			name == aws.StringValue(channel.S3BucketName),
		)
		return name == aws.StringValue(channel.S3BucketName)
	})
	return d
}

// BucketKeyPrefix adds the BucketKeyPrefix filter to the filter list
// the BucketKeyPrefix filter: filters channels by BucketKeyPrefix where 'prefix' provided
// is the expected S3KeyPrefix value
func (d *DeliveryChannel) BucketKeyPrefix(prefix string) *DeliveryChannel {
	d.filters = append(d.filters, func(v interface{}) bool {
		channel := convert(v)
		if channel == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n",
			prefix, aws.StringValue(channel.S3KeyPrefix),
			prefix == aws.StringValue(channel.S3KeyPrefix),
		)
		return prefix == aws.StringValue(channel.S3KeyPrefix)
	})
	return d
}

// Name adds the Name filter to the filter list
// the Name filter: filters channels by Name where 'name' provided
// is the expected Name value
func (d *DeliveryChannel) Name(name string) *DeliveryChannel {
	d.filters = append(d.filters, func(v interface{}) bool {
		channel := convert(v)
		if channel == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n",
			name,
			aws.StringValue(channel.Name),
			name == aws.StringValue(channel.Name),
		)
		return name == aws.StringValue(channel.Name)
	})
	return d
}

// Frequency adds the Frequency filter to the filter list
// the Frequency filter: filters channels by Frequency where 'freq' provided
// is the expected DeliveryFrequency value
func (d *DeliveryChannel) Frequency(freq string) *DeliveryChannel {
	d.filters = append(d.filters, func(v interface{}) bool {
		channel := convert(v)
		if channel == nil {
			return false
		}
		if channel.ConfigSnapshotDeliveryProperties == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n",
			freq,
			aws.StringValue(channel.ConfigSnapshotDeliveryProperties.DeliveryFrequency),
			freq == aws.StringValue(channel.ConfigSnapshotDeliveryProperties.DeliveryFrequency),
		)
		return freq == aws.StringValue(channel.ConfigSnapshotDeliveryProperties.DeliveryFrequency)
	})
	return d
}

func (d *DeliveryChannel) filter(channels []*configservice.DeliveryChannel) ([]*configservice.DeliveryChannel, error) {
	if len(channels) == 0 {
		var err error
		channels, err = d.channels()
		if err != nil {
			return nil, err
		}
	}
	return fromIface(shared.GenericFilter(d.filters, toIface(channels))), nil
}

func (d *DeliveryChannel) channels() ([]*configservice.DeliveryChannel, error) {
	svc := configservice.New(d.client)
	out, err := svc.DescribeDeliveryChannels(
		&configservice.DescribeDeliveryChannelsInput{},
	)
	if err != nil {
		return nil, err
	}
	return out.DeliveryChannels, nil
}

func convert(in interface{}) *configservice.DeliveryChannel {
	out, ok := in.(*configservice.DeliveryChannel)
	if !ok {
		shared.Debugf("object not convertible to *configservice.DeliveryChannel: ")
		shared.Dump(in)
		return nil
	}
	return out
}
func toIface(in []*configservice.DeliveryChannel) (out []interface{}) {
	for _, i := range in {
		out = append(out, i)
	}
	return
}
func fromIface(in []interface{}) (out []*configservice.DeliveryChannel) {
	for _, i := range in {
		v := convert(i)
		if v == nil {
			continue
		}
		out = append(out, v)
	}
	return
}
