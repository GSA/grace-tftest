package alarm

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

func TestAlarm(t *testing.T) {
	alarms := []*cloudwatch.MetricAlarm{
		{
			AlarmArn:           aws.String("a"),
			AlarmDescription:   aws.String("b"),
			AlarmName:          aws.String("c"),
			ComparisonOperator: aws.String("d"),
		},
	}
	New(nil, nil).AlarmArn("a").AlarmDescription("b").AlarmName("c").ComparisonOperator("d").Assert(t, alarms...)
}
