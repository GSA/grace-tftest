package lifecycle

import (
	"testing"
	"time"

	"github.com/GSA/grace-tftest/aws/shared"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Lifecycle contains the necessary properties for filtering *s3.LifecycleRule objects
type Lifecycle struct {
	filters []shared.Filter
	client  client.ConfigProvider
	name    string
	rule    *s3.LifecycleRule
}

// New returns a new *Lifecycle
func New(client client.ConfigProvider, name string) *Lifecycle {
	return &Lifecycle{
		client: client,
		name:   name,
	}
}

// Selected returns the currently selected *s3.LifecycleRule
func (l *Lifecycle) Selected() *s3.LifecycleRule {
	return l.rule
}

// Assert applies all filters that have been called, resets the list of filters,
// fails the test if there is not exactly one match, and stores the matched rule
// if rules is not provided, *s3.LifecycleRule objects will be retreived from AWS
func (l *Lifecycle) Assert(t *testing.T, rules ...*s3.LifecycleRule) *Lifecycle {
	var err error
	rules, err = l.filter(rules)
	if err != nil {
		t.Error(err)
	}

	switch n := len(rules); {
	case n == 0:
		t.Error("no matching lifecycle rule was found")
	case n > 1:
		t.Error("more than one matching lifecycle rule was found")
	default:
		l.rule = rules[0]
	}

	l.filters = []shared.Filter{}
	return l
}

// First applies all filters that have been called, resets the list of filters,
// fails the test if there is not exactly one match, and stores the matched rule
// if rules is not provided, *s3.LifecycleRule objects will be retreived from AWS
func (l *Lifecycle) First(t *testing.T, rules ...*s3.LifecycleRule) *Lifecycle {
	var err error
	rules, err = l.filter(rules)
	if err != nil {
		t.Error(err)
	}

	if len(rules) == 0 {
		t.Error("no matching lifecycle rule was found")
	} else {
		l.rule = rules[0]
	}

	l.filters = []shared.Filter{}
	return l
}

// Filter adds the 'filter' provided to the filter list
func (l *Lifecycle) Filter(filter shared.Filter) *Lifecycle {
	l.filters = append(l.filters, filter)
	return l
}

// IsExp adds the IsExp filter to the filter list
// the IsExp filter: filters rules by whether they have
// an Expiration set
func (l *Lifecycle) IsExp() *Lifecycle {
	l.filters = append(l.filters, func(v interface{}) bool {
		rule := convert(v)
		if rule == nil {
			return false
		}
		shared.Debugf("%#v != nil -> %t\n", rule.Expiration, rule.Expiration != nil)
		return rule.Expiration != nil
	})
	return l
}

// FilterPrefix adds the FilterPrefix filter to the filter list
// the FilterPrefix filter: filters rules by Filter[Prefix] where 'value'
// provided is the expected Prefix value
func (l *Lifecycle) FilterPrefix(value string) *Lifecycle {
	l.filters = append(l.filters, func(v interface{}) bool {
		rule := convert(v)
		if rule == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", value, aws.StringValue(rule.Filter.Prefix), value == aws.StringValue(rule.Filter.Prefix))
		return value == aws.StringValue(rule.Filter.Prefix)
	})
	return l
}

// FilterTag adds the FilterTag filter to the filter list
// the FilterTag filter: filters rules by Filter[Tag] where 'key and value'
// provided is the expected Tag key and value
func (l *Lifecycle) FilterTag(key, value string) *Lifecycle {
	l.filters = append(l.filters, func(v interface{}) bool {
		rule := convert(v)
		if rule == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n%s == %s -> %t",
			key, aws.StringValue(rule.Filter.Tag.Key), key == aws.StringValue(rule.Filter.Tag.Key),
			value, aws.StringValue(rule.Filter.Tag.Value), value == aws.StringValue(rule.Filter.Tag.Value))

		return key == aws.StringValue(rule.Filter.Tag.Key) &&
			value == aws.StringValue(rule.Filter.Tag.Value)
	})
	return l
}

// FilterAnd adds the FilterAnd filter to the filter list
// the FilterAnd filter: filters rules by Filter[And] where 'key and value'
// provided is the expected Tag key and value
func (l *Lifecycle) FilterAnd(prefix string, tag ...*s3.Tag) *Lifecycle {
	l.filters = append(l.filters, func(v interface{}) bool {
		rule := convert(v)
		if rule == nil {
			return false
		}
		for _, t := range tag {
			var found bool
			for _, tt := range rule.Filter.And.Tags {
				if aws.StringValue(t.Key) == aws.StringValue(tt.Key) &&
					aws.StringValue(t.Value) == aws.StringValue(tt.Value) {
					found = true
				}
			}
			if !found {
				shared.Debugf("tag[Key: %s, Value: %s] not found\n", aws.StringValue(t.Key), aws.StringValue(t.Value))
				return false
			}
		}
		shared.Debugf("Prefix %s == %s -> %t\n", prefix, aws.StringValue(rule.Filter.And.Prefix), prefix == aws.StringValue(rule.Filter.And.Prefix))
		return prefix == aws.StringValue(rule.Filter.And.Prefix)
	})
	return l
}

// Status adds the Status filter to the filter list
// the Status filter: filters rules by Status where 'status'
// provided is the expected Status value
func (l *Lifecycle) Status(status string) *Lifecycle {
	l.filters = append(l.filters, func(v interface{}) bool {
		rule := convert(v)
		if rule == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", status, aws.StringValue(rule.Status), status == aws.StringValue(rule.Status))
		return status == aws.StringValue(rule.Status)
	})
	return l
}

// Method adds the Method filter to the filter list
// the Method filter: filters rules by Method where 'method'
// provided is the expected ID value
func (l *Lifecycle) Method(method string) *Lifecycle {
	l.filters = append(l.filters, func(v interface{}) bool {
		rule := convert(v)
		if rule == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", method, aws.StringValue(rule.ID), method == aws.StringValue(rule.ID))
		return method == aws.StringValue(rule.ID)
	})
	return l
}

// ExpDate adds the ExpDate filter to the filter list
// the ExpDate filter: filters rules by ExpDate where 'date'
// provided is the expected Expiration Date value
func (l *Lifecycle) ExpDate(date time.Time) *Lifecycle {
	l.filters = append(l.filters, func(v interface{}) bool {
		rule := convert(v)
		if rule == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", date, aws.TimeValue(rule.Expiration.Date), date == aws.TimeValue(rule.Expiration.Date))
		return date == aws.TimeValue(rule.Expiration.Date)
	})
	return l
}

// ExpDays adds the ExpDays filter to the filter list
// the ExpDays filter: filters rules by ExpDays where 'days'
// provided is the expected Expiration Days value
func (l *Lifecycle) ExpDays(days int64) *Lifecycle {
	l.filters = append(l.filters, func(v interface{}) bool {
		rule := convert(v)
		if rule == nil {
			return false
		}
		shared.Debugf("%d == %d -> %t\n", days, aws.Int64Value(rule.Expiration.Days), days == aws.Int64Value(rule.Expiration.Days))
		return days == aws.Int64Value(rule.Expiration.Days)
	})
	return l
}

func (l *Lifecycle) filter(rules []*s3.LifecycleRule) ([]*s3.LifecycleRule, error) {
	if len(rules) == 0 {
		var err error
		rules, err = l.rules()
		if err != nil {
			return nil, err
		}
	}
	return fromIface(shared.GenericFilter(l.filters, toIface(rules))), nil
}

func (l *Lifecycle) rules() ([]*s3.LifecycleRule, error) {
	svc := s3.New(l.client)
	out, err := svc.GetBucketLifecycleConfiguration(&s3.GetBucketLifecycleConfigurationInput{
		Bucket: &l.name,
	})
	if err != nil {
		return nil, err
	}
	return out.Rules, nil
}
func convert(in interface{}) *s3.LifecycleRule {
	out, ok := in.(*s3.LifecycleRule)
	if !ok {
		shared.Debugf("object not convertible to *s3.LifecycleRule: ")
		shared.Dump(in)
		return nil
	}
	return out
}
func toIface(in []*s3.LifecycleRule) (out []interface{}) {
	for _, i := range in {
		out = append(out, i)
	}
	return
}
func fromIface(in []interface{}) (out []*s3.LifecycleRule) {
	for _, i := range in {
		v := convert(i)
		if v == nil {
			continue
		}
		out = append(out, v)
	}
	return
}
