package metric

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

func TestMetric(t *testing.T) {
	metrics := []*cloudwatch.Metric{
		{
			MetricName: aws.String("a"),
			Namespace:  aws.String("b"),
		},
	}
	alarm := New(nil).MetricName("a").Namespace("b").Assert(t, metrics...).Alarm()
	if alarm == nil {
		t.Error("alarm should not be nil")
	}
}
