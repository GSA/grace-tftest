package metricfilter

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

func TestMetricFilter(t *testing.T) {
	filters := []*cloudwatchlogs.MetricFilter{
		{
			FilterName:   aws.String("a"),
			LogGroupName: aws.String("b"),
			MetricTransformations: []*cloudwatchlogs.MetricTransformation{
				{
					DefaultValue:    aws.Float64(1.0),
					MetricName:      aws.String("c"),
					MetricNamespace: aws.String("d"),
					MetricValue:     aws.String("e"),
				},
			},
		},
		{
			FilterName:   aws.String("f"),
			LogGroupName: aws.String("g"),
			MetricTransformations: []*cloudwatchlogs.MetricTransformation{
				{
					DefaultValue:    aws.Float64(2.0),
					MetricName:      aws.String("h"),
					MetricNamespace: aws.String("i"),
					MetricValue:     aws.String("j"),
				},
			},
		},
		{
			FilterName:   aws.String("k"),
			LogGroupName: aws.String("l"),
			MetricTransformations: []*cloudwatchlogs.MetricTransformation{
				{
					DefaultValue:    aws.Float64(3.0),
					MetricName:      aws.String("m"),
					MetricNamespace: aws.String("n"),
					MetricValue:     aws.String("o"),
				},
			},
		},
	}

	New(nil).
		Name("a").
		LogGroupName("b").
		TDefault(1.0).
		TName("c").
		TNamespace("d").
		TValue("e").
		Assert(t, filters...)

	New(nil).
		FilterName("f").
		LogGroupName("g").
		TransformationDefaultValue(2.0).
		TransformationName("h").
		TransformationNamespace("i").
		TransformationMetricValue("j").
		Assert(t, filters...)
}
