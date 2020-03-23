// Package trail provides functions and filters to test AWS CloudTrail Trails
package trail

import (
	"log"
	"os"
	"strings"
	"testing"

	"github.com/GSA/grace-tftest/aws/shared"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/cloudtrail"
)

// Trail contains the necessary properties for testing *cloudtrail.Trail objects
type Trail struct {
	client  client.ConfigProvider
	trail   *cloudtrail.Trail
	filters []shared.Filter
}

// New returns a new *Trail
func New(client client.ConfigProvider) *Trail {
	return &Trail{client: client}
}

// Selected returns the currently selected *cloudtrail.Trail
func (r *Trail) Selected() *cloudtrail.Trail {
	return r.trail
}

// Assert applies all filters that have been called, resets the list of filters,
// fails the test if there is not exactly one match, and stores the matched trail
// if trails is not provided, *cloudtrail.Trail objects will be retreived from AWS
func (r *Trail) Assert(t *testing.T, trails ...*cloudtrail.Trail) *Trail {
	var err error
	trails, err = r.filter(trails)
	if err != nil {
		t.Fatal(err)
	}

	switch l := len(trails); {
	case l == 0:
		t.Fatal("no matching trail was found")
	case l > 1:
		t.Fatal("more than one matching trail was found")
	default:
		r.trail = trails[0]
	}

	r.filters = []shared.Filter{}
	return r
}

// First applies all filters that have been called, resets the list of filters,
// fails the test if there are no matches, and stores the first match
// if trails is not provided, *cloudtrail.Trail objects will be retreived from AWS
func (r *Trail) First(t *testing.T, trails ...*cloudtrail.Trail) *Trail {
	var err error
	trails, err = r.filter(trails)
	if err != nil {
		t.Fatal(err)
	}

	if len(trails) == 0 {
		t.Fatal("no matching trail was found")
	} else {
		r.trail = trails[0]
	}

	r.filters = []shared.Filter{}
	return r
}

// Filter adds the 'filter' provided to the filter list
func (r *Trail) Filter(filter shared.Filter) *Trail {
	r.filters = append(r.filters, filter)
	return r
}

// TrailARN adds the TrailARN filter to the filter list
// the TrailARN filter: filters trails by TrailARN where 'arn' provided
// is the expected TrailARN value
func (r *Trail) TrailARN(arn string) *Trail {
	r.filters = append(r.filters, func(v interface{}) bool {
		trail := convert(v)
		if trail == nil {
			return false
		}
		shared.Debugf(
			"%s like %s -> %t\n",
			arn,
			aws.StringValue(trail.TrailARN),
			strings.EqualFold(arn, aws.StringValue(trail.TrailARN)),
		)
		return strings.EqualFold(arn, aws.StringValue(trail.TrailARN))
	})
	return r
}

// ARN adds the ARN filter as an alias to the TrailARN filter
func (r *Trail) ARN(arn string) *Trail {
	return r.TrailARN(arn)
}

// Name adds the Name filter to the filter list
// the Name filter: filters trails by Name where 'name' provided
// is the expected Name value
func (r *Trail) Name(name string) *Trail {
	r.filters = append(r.filters, func(v interface{}) bool {
		trail := convert(v)
		if trail == nil {
			return false
		}
		shared.Debugf(
			"%s like %s -> %t\n",
			name,
			aws.StringValue(trail.Name),
			strings.EqualFold(name, aws.StringValue(trail.Name)),
		)
		return strings.EqualFold(name, aws.StringValue(trail.Name))
	})
	return r
}

// S3BucketName adds the S3BucketName filter to the filter list
// the S3BucketName filter: filters trails by S3BucketName where 'name' provided
// is the expected S3BucketName value
func (r *Trail) S3BucketName(name string) *Trail {
	r.filters = append(r.filters, func(v interface{}) bool {
		trail := convert(v)
		if trail == nil {
			return false
		}
		shared.Debugf(
			"%s like %s -> %t\n",
			name,
			aws.StringValue(trail.S3BucketName),
			strings.EqualFold(name, aws.StringValue(trail.S3BucketName)),
		)
		return strings.EqualFold(name, aws.StringValue(trail.S3BucketName))
	})
	return r
}

// S3KeyPrefix adds the S3KeyPrefix filter to the filter list
// the S3KeyPrefix filter: filters trails by S3KeyPrefix where 'str' provided
// is the expected S3KeyPrefix value
func (r *Trail) S3KeyPrefix(str string) *Trail {
	r.filters = append(r.filters, func(v interface{}) bool {
		trail := convert(v)
		if trail == nil {
			return false
		}
		shared.Debugf(
			"%s like %s -> %t\n",
			str,
			aws.StringValue(trail.S3KeyPrefix),
			strings.EqualFold(str, aws.StringValue(trail.S3KeyPrefix)),
		)
		return strings.EqualFold(str, aws.StringValue(trail.S3KeyPrefix))
	})
	return r
}

// CloudWatchLogsLogGroupArn adds the CloudWatchLogsLogGroupArn filter to the filter list
// the CloudWatchLogsLogGroupArn filter: filters trails by CloudWatchLogsLogGroupArn where 'arn' provided
// is the expected CloudWatchLogsLogGroupArn value
func (r *Trail) CloudWatchLogsLogGroupArn(arn string) *Trail {
	r.filters = append(r.filters, func(v interface{}) bool {
		trail := convert(v)
		if trail == nil {
			return false
		}
		shared.Debugf(
			"%s like %s -> %t\n",
			arn,
			aws.StringValue(trail.CloudWatchLogsLogGroupArn),
			strings.EqualFold(arn, aws.StringValue(trail.CloudWatchLogsLogGroupArn)),
		)
		return strings.EqualFold(arn, aws.StringValue(trail.CloudWatchLogsLogGroupArn))
	})
	return r
}

// CloudWatchLogsRoleArn adds the CloudWatchLogsRoleArn filter to the filter list
// the CloudWatchLogsRoleArn filter: filters trails by CloudWatchLogsRoleArn where 'arn' provided
// is the expected CloudWatchLogsRoleArn value
func (r *Trail) CloudWatchLogsRoleArn(arn string) *Trail {
	r.filters = append(r.filters, func(v interface{}) bool {
		trail := convert(v)
		if trail == nil {
			return false
		}
		shared.Debugf(
			"%s like %s -> %t\n",
			arn,
			aws.StringValue(trail.CloudWatchLogsRoleArn),
			strings.EqualFold(arn, aws.StringValue(trail.CloudWatchLogsRoleArn)),
		)
		return strings.EqualFold(arn, aws.StringValue(trail.CloudWatchLogsRoleArn))
	})
	return r
}

// SnsTopicARN adds the SnsTopicARN filter to the filter list
// the SnsTopicARN filter: filters trails by SnsTopicARN where 'arn' provided
// is the expected SnsTopicARN value
func (r *Trail) SnsTopicARN(arn string) *Trail {
	r.filters = append(r.filters, func(v interface{}) bool {
		trail := convert(v)
		if trail == nil {
			return false
		}
		shared.Debugf(
			"%s like %s -> %t\n",
			arn,
			aws.StringValue(trail.SnsTopicARN),
			strings.EqualFold(arn, aws.StringValue(trail.SnsTopicARN)),
		)
		return strings.EqualFold(arn, aws.StringValue(trail.SnsTopicARN))
	})
	return r
}

// SnsTopicName adds the SnsTopicName filter to the filter list
// the SnsTopicName filter: filters trails by SnsTopicName where 'name' provided
// is the expected SnsTopicName value
func (r *Trail) SnsTopicName(name string) *Trail {
	r.filters = append(r.filters, func(v interface{}) bool {
		trail := convert(v)
		if trail == nil {
			return false
		}
		shared.Debugf(
			"%s like %s -> %t\n",
			name,
			aws.StringValue(trail.SnsTopicName),
			strings.EqualFold(name, aws.StringValue(trail.SnsTopicName)),
		)
		return strings.EqualFold(name, aws.StringValue(trail.SnsTopicName))
	})
	return r
}

// KmsKeyID adds the KmsKeyID filter to the filter list
// the KmsKeyID filter: filters trails by KmsKeyId where 'id' provided
// is the expected KmsKeyId value
func (r *Trail) KmsKeyID(id string) *Trail {
	r.filters = append(r.filters, func(v interface{}) bool {
		trail := convert(v)
		if trail == nil {
			return false
		}
		shared.Debugf(
			"%s like %s -> %t\n",
			id,
			aws.StringValue(trail.KmsKeyId),
			strings.EqualFold(id, aws.StringValue(trail.KmsKeyId)),
		)
		return strings.EqualFold(id, aws.StringValue(trail.KmsKeyId))
	})
	return r
}

// HomeRegion adds the HomeRegion filter to the filter list
// the HomeRegion filter: filters trails by HomeRegion where 'str' provided
// is the expected HomeRegion value
func (r *Trail) HomeRegion(str string) *Trail {
	r.filters = append(r.filters, func(v interface{}) bool {
		trail := convert(v)
		if trail == nil {
			return false
		}
		shared.Debugf(
			"%s like %s -> %t\n",
			str,
			aws.StringValue(trail.HomeRegion),
			strings.EqualFold(str, aws.StringValue(trail.HomeRegion)),
		)
		return strings.EqualFold(str, aws.StringValue(trail.HomeRegion))
	})
	return r
}

// HasCustomEventSelectors adds the HasCustomEventSelectors filter to the filter list
// the HasCustomEventSelectors filter: filters trails by HasCustomEventSelectors where 'b' provided
// is the expected HasCustomEventSelectors value
func (r *Trail) HasCustomEventSelectors(b bool) *Trail {
	r.filters = append(r.filters, func(v interface{}) bool {
		trail := convert(v)
		if trail == nil {
			return false
		}
		shared.Debugf(
			"%t like %t -> %t\n",
			b,
			aws.BoolValue(trail.HasCustomEventSelectors),
			b == aws.BoolValue(trail.HasCustomEventSelectors),
		)
		return b == aws.BoolValue(trail.HasCustomEventSelectors)
	})
	return r
}

// HasInsightSelectors adds the HasInsightSelectors filter to the filter list
// the HasInsightSelectors filter: filters trails by HasInsightSelectors where 'b' provided
// is the expected HasInsightSelectors value
func (r *Trail) HasInsightSelectors(b bool) *Trail {
	r.filters = append(r.filters, func(v interface{}) bool {
		trail := convert(v)
		if trail == nil {
			return false
		}
		shared.Debugf(
			"%t like %t -> %t\n",
			b,
			aws.BoolValue(trail.HasInsightSelectors),
			b == aws.BoolValue(trail.HasInsightSelectors),
		)
		return b == aws.BoolValue(trail.HasInsightSelectors)
	})
	return r
}

// IncludeGlobalServiceEvents adds the IncludeGlobalServiceEvents filter to the filter list
// the IncludeGlobalServiceEvents filter: filters trails by IncludeGlobalServiceEvents where 'b' provided
// is the expected IncludeGlobalServiceEvents value
func (r *Trail) IncludeGlobalServiceEvents(b bool) *Trail {
	r.filters = append(r.filters, func(v interface{}) bool {
		trail := convert(v)
		if trail == nil {
			return false
		}
		shared.Debugf(
			"%t like %t -> %t\n",
			b,
			aws.BoolValue(trail.IncludeGlobalServiceEvents),
			b == aws.BoolValue(trail.IncludeGlobalServiceEvents),
		)
		return b == aws.BoolValue(trail.IncludeGlobalServiceEvents)
	})
	return r
}

// IsMultiRegionTrail adds the IsMultiRegionTrail filter to the filter list
// the IsMultiRegionTrail filter: filters trails by IsMultiRegionTrail where 'b' provided
// is the expected IsMultiRegionTrail value
func (r *Trail) IsMultiRegionTrail(b bool) *Trail {
	r.filters = append(r.filters, func(v interface{}) bool {
		trail := convert(v)
		if trail == nil {
			return false
		}
		shared.Debugf(
			"%t like %t -> %t\n",
			b,
			aws.BoolValue(trail.IsMultiRegionTrail),
			b == aws.BoolValue(trail.IsMultiRegionTrail),
		)
		return b == aws.BoolValue(trail.IsMultiRegionTrail)
	})
	return r
}

// IsOrganizationTrail adds the IsOrganizationTrail filter to the filter list
// the IsOrganizationTrail filter: filters trails by IsOrganizationTrail where 'b' provided
// is the expected IsOrganizationTrail value
func (r *Trail) IsOrganizationTrail(b bool) *Trail {
	r.filters = append(r.filters, func(v interface{}) bool {
		trail := convert(v)
		if trail == nil {
			return false
		}
		shared.Debugf(
			"%t like %t -> %t\n",
			b,
			aws.BoolValue(trail.IsOrganizationTrail),
			b == aws.BoolValue(trail.IsOrganizationTrail),
		)
		return b == aws.BoolValue(trail.IsOrganizationTrail)
	})
	return r
}

// LogFileValidationEnabled adds the LogFileValidationEnabled filter to the filter list
// the LogFileValidationEnabled filter: filters trails by LogFileValidationEnabled where 'b' provided
// is the expected LogFileValidationEnabled value
func (r *Trail) LogFileValidationEnabled(b bool) *Trail {
	r.filters = append(r.filters, func(v interface{}) bool {
		trail := convert(v)
		if trail == nil {
			return false
		}
		shared.Debugf(
			"%t like %t -> %t\n",
			b,
			aws.BoolValue(trail.LogFileValidationEnabled),
			b == aws.BoolValue(trail.LogFileValidationEnabled),
		)
		return b == aws.BoolValue(trail.LogFileValidationEnabled)
	})
	return r
}

func (r *Trail) filter(trails []*cloudtrail.Trail) ([]*cloudtrail.Trail, error) {
	if len(trails) == 0 {
		var err error
		trails, err = r.trails()
		if err != nil {
			return nil, err
		}
	}
	results := fromIface(shared.GenericFilter(r.filters, toIface(trails)))
	if len(results) == 0 {
		log.Println("aws.cloudtrail.trail.filter had zero results: ")
		shared.Spew(os.Stdout, trails)
	}
	return results, nil
}

func (r *Trail) trails() ([]*cloudtrail.Trail, error) {
	svc := cloudtrail.New(r.client)
	resp, err := svc.DescribeTrails(&cloudtrail.DescribeTrailsInput{})
	if err != nil {
		return nil, err
	}

	return resp.TrailList, nil
}

func convert(in interface{}) *cloudtrail.Trail {
	out, ok := in.(*cloudtrail.Trail)
	if !ok {
		shared.Debugf("object not convertible to *cloudtrail.Trail: ")
		shared.Dump(in)
		return nil
	}
	return out
}
func toIface(in []*cloudtrail.Trail) (out []interface{}) {
	for _, i := range in {
		out = append(out, i)
	}
	return
}
func fromIface(in []interface{}) (out []*cloudtrail.Trail) {
	for _, i := range in {
		v := convert(i)
		if v == nil {
			continue
		}
		out = append(out, v)
	}
	return
}
