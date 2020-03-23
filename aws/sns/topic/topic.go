// Package topic provides functions and filters to test AWS SNS Topics
package topic

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/GSA/grace-tftest/aws/shared"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/sns"
)

// Topic contains the necessary properties for testing *sns.Topic objects
type Topic struct {
	client  client.ConfigProvider
	topic   *Attributes
	filters []shared.Filter
}

// Attributes A struct of the topic's attributes map.
type Attributes struct {
	Policy                  string
	DeliveryPolicy          string
	Owner                   string
	SubscriptionsPending    string
	TopicArn                string
	EffectiveDeliveryPolicy string
	SubscriptionsConfirmed  string
	DisplayName             string
	SubscriptionsDeleted    string
	KmsMasterKeyID          string `json:"KmsMasterKeyId"`
}

// New returns a new *Topic
func New(client client.ConfigProvider) *Topic {
	return &Topic{client: client}
}

// Selected returns the currently selected *Attributes
func (r *Topic) Selected() *Attributes {
	return r.topic
}

// Assert applies all filters that have been called, resets the list of filters,
// fails the test if there is not exactly one match, and stores the matched topic
// if topics is not provided, *Attributes objects will be retreived from AWS
func (r *Topic) Assert(t *testing.T, topics ...*Attributes) *Topic {
	var err error
	topics, err = r.filter(topics)
	if err != nil {
		t.Fatal(err)
	}

	switch l := len(topics); {
	case l == 0:
		t.Fatal("no matching topic was found")
	case l > 1:
		t.Fatal("more than one matching topic was found")
	default:
		r.topic = topics[0]
	}

	r.filters = []shared.Filter{}
	return r
}

// First applies all filters that have been called, resets the list of filters,
// fails the test if there are no matches, and stores the first match
// if topics is not provided, *Attributes objects will be retreived from AWS
func (r *Topic) First(t *testing.T, topics ...*Attributes) *Topic {
	var err error
	topics, err = r.filter(topics)
	if err != nil {
		t.Fatal(err)
	}

	if len(topics) == 0 {
		t.Fatal("no matching topic was found")
	} else {
		r.topic = topics[0]
	}

	r.filters = []shared.Filter{}
	return r
}

// Filter adds the 'filter' provided to the filter list
func (r *Topic) Filter(filter shared.Filter) *Topic {
	r.filters = append(r.filters, filter)
	return r
}

// TopicArn adds the TopicArn filter to the filter list
// the TopicArn filter: filters topics by TopicArn where 'arn' provided
// is the expected TopicArn value
func (r *Topic) TopicArn(arn string) *Topic {
	r.filters = append(r.filters, func(v interface{}) bool {
		topic := convert(v)
		if topic == nil {
			return false
		}
		shared.Debugf(
			"%s like %s -> %t\n",
			arn,
			topic.TopicArn,
			strings.EqualFold(arn, topic.TopicArn),
		)
		return strings.EqualFold(arn, topic.TopicArn)
	})
	return r
}

// Arn adds the Arn filter as an alias to the TopicArn filter
func (r *Topic) Arn(arn string) *Topic {
	return r.TopicArn(arn)
}

// DisplayName adds the DisplayName filter to the filter list
// the DisplayName filter: filters topics by DisplayName where 'name' provided
// is the expected DisplayName value
func (r *Topic) DisplayName(name string) *Topic {
	r.filters = append(r.filters, func(v interface{}) bool {
		topic := convert(v)
		if topic == nil {
			return false
		}
		shared.Debugf(
			"%s like %s -> %t\n",
			name,
			topic.DisplayName,
			strings.EqualFold(name, topic.DisplayName),
		)
		return strings.EqualFold(name, topic.DisplayName)
	})
	return r
}

// Name adds the Name filter as an alias to the DisplayName filter
func (r *Topic) Name(name string) *Topic {
	return r.DisplayName(name)
}

// Owner adds the Owner filter to the filter list
// the Owner filter: filters topics by Owner where 'str' provided
// is the expected Owner value
func (r *Topic) Owner(str string) *Topic {
	r.filters = append(r.filters, func(v interface{}) bool {
		topic := convert(v)
		if topic == nil {
			return false
		}
		shared.Debugf(
			"%s like %s -> %t\n",
			str,
			topic.Owner,
			strings.EqualFold(str, topic.Owner),
		)
		return strings.EqualFold(str, topic.Owner)
	})
	return r
}

// KmsMasterKeyID adds the KmsMasterKeyID filter to the filter list
// the KmsMasterKeyID filter: filters topics by KmsMasterKeyId where 'id' provided
// is the expected KmsMasterKeyId value
func (r *Topic) KmsMasterKeyID(id string) *Topic {
	r.filters = append(r.filters, func(v interface{}) bool {
		topic := convert(v)
		if topic == nil {
			return false
		}
		shared.Debugf(
			"%s like %s -> %t\n",
			id,
			topic.KmsMasterKeyID,
			strings.EqualFold(id, topic.KmsMasterKeyID),
		)
		return strings.EqualFold(id, topic.KmsMasterKeyID)
	})
	return r
}

func (r *Topic) filter(topics []*Attributes) ([]*Attributes, error) {
	if len(topics) == 0 {
		var err error
		topics, err = r.topics()
		if err != nil {
			return nil, err
		}
	}
	results := fromIface(shared.GenericFilter(r.filters, toIface(topics)))
	if len(results) == 0 {
		log.Println("aws.sns.topic.filter had zero results: ")
		shared.Spew(os.Stdout, topics)
	}
	return results, nil
}

func (r *Topic) topics() ([]*Attributes, error) {
	svc := sns.New(r.client)
	var topics []*sns.Topic
	err := svc.ListTopicsPages(&sns.ListTopicsInput{}, func(page *sns.ListTopicsOutput, lastPage bool) bool {
		topics = append(topics, page.Topics...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}

	attributes := make([]*Attributes, len(topics))
	for i, t := range topics {
		resp, err := svc.GetTopicAttributes(&sns.GetTopicAttributesInput{TopicArn: t.TopicArn})
		if err != nil {
			return nil, err
		}

		attributes[i], err = unmarshal(resp.Attributes)
		if err != nil {
			return nil, err
		}
	}

	return attributes, nil
}

func unmarshal(m map[string]*string) (*Attributes, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	var a *Attributes
	err = json.Unmarshal(b, &a)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func convert(in interface{}) *Attributes {
	out, ok := in.(*Attributes)
	if !ok {
		shared.Debugf("object not convertible to *Attributes: ")
		shared.Dump(in)
		return nil
	}
	return out
}
func toIface(in []*Attributes) (out []interface{}) {
	for _, i := range in {
		out = append(out, i)
	}
	return
}
func fromIface(in []interface{}) (out []*Attributes) {
	for _, i := range in {
		v := convert(i)
		if v == nil {
			continue
		}
		out = append(out, v)
	}
	return
}
