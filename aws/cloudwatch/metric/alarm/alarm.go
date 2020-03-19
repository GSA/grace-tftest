// Package alarm the necessary properties for testing *cloudwatch.MetricAlarm objects
package alarm

import (
	"log"
	"os"
	"testing"

	"github.com/GSA/grace-tftest/aws/shared"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

// Alarm contains the necessary properties for testing *cloudwatch.MetricAlarm objects
type Alarm struct {
	client  client.ConfigProvider
	alarm   *cloudwatch.MetricAlarm
	metric  *cloudwatch.Metric
	filters []shared.Filter
}

// New returns a new *Alarm
func New(client client.ConfigProvider, metric *cloudwatch.Metric) *Alarm {
	return &Alarm{client: client, metric: metric}
}

// Selected returns the currently selected *cloudwatch.MetricAlarm
func (a *Alarm) Selected() *cloudwatch.MetricAlarm {
	return a.alarm
}

// Assert applies all filters that have been called, resets the list of filters,
// fails the test if there is not exactly one match, and stores the matched alarm
// if alarms is not provided, *cloudwatch.MetricAlarm objects will be retreived from AWS
func (a *Alarm) Assert(t *testing.T, alarms ...*cloudwatch.MetricAlarm) *Alarm {
	var err error
	alarms, err = a.filter(alarms)
	if err != nil {
		t.Fatal(err)
	}

	switch l := len(alarms); {
	case l == 0:
		t.Fatal("no matching alarm was found")
	case l > 1:
		t.Fatal("more than one matching alarm was found")
	default:
		a.alarm = alarms[0]
	}

	a.filters = []shared.Filter{}
	return a
}

// First applies all filters that have been called, resets the list of filters,
// fails the test if there are no matches, and stores the first match
// if alarms is not provided, *cloudwatch.MetricAlarm objects will be retreived from AWS
func (a *Alarm) First(t *testing.T, alarms ...*cloudwatch.MetricAlarm) *Alarm {
	var err error
	alarms, err = a.filter(alarms)
	if err != nil {
		t.Fatal(err)
	}

	if len(alarms) == 0 {
		t.Fatal("no matching metric was found")
	} else {
		a.alarm = alarms[0]
	}

	a.filters = []shared.Filter{}
	return a
}

// AlarmArn adds the AlarmArn filter to the filter list
// the AlarmArn filter: filters alarms by AlarmArn where 'arn' provided
// is the expected AlarmArn value
func (a *Alarm) AlarmArn(arn string) *Alarm {
	a.filters = append(a.filters, func(v interface{}) bool {
		alarm := convert(v)
		if alarm == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", arn, aws.StringValue(alarm.AlarmArn), arn == aws.StringValue(alarm.AlarmArn))
		return arn == aws.StringValue(alarm.AlarmArn)
	})
	return a
}

// Arn adds the Arn filter as an alias to the AlarmArn filter
func (a *Alarm) Arn(arn string) *Alarm {
	return a.AlarmArn(arn)
}

// Filter adds the 'filter' provided to the filter list
func (a *Alarm) Filter(filter shared.Filter) *Alarm {
	a.filters = append(a.filters, filter)
	return a
}

// AlarmName adds the AlarmName filter to the filter list
// the AlarmName filter: filters alarms by AlarmName where 'str' provided
// is the expected AlarmName value
func (a *Alarm) AlarmName(str string) *Alarm {
	a.filters = append(a.filters, func(v interface{}) bool {
		alarm := convert(v)
		if alarm == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", str, aws.StringValue(alarm.AlarmName), str == aws.StringValue(alarm.AlarmName))
		return str == aws.StringValue(alarm.AlarmName)
	})
	return a
}

// Name adds the Name filter as an alias to the AlarmName filter
func (a *Alarm) Name(arn string) *Alarm {
	return a.AlarmName(arn)
}

// AlarmDescription adds the AlarmDescription filter to the filter list
// the AlarmDescription filter: filters alarms by AlarmDescription where 'str' provided
// is the expected AlarmDescription value
func (a *Alarm) AlarmDescription(str string) *Alarm {
	a.filters = append(a.filters, func(v interface{}) bool {
		alarm := convert(v)
		if alarm == nil {
			return false
		}
		shared.Debugf(
			"%s == %s -> %t\n",
			str,
			aws.StringValue(alarm.AlarmDescription),
			str == aws.StringValue(alarm.AlarmDescription),
		)
		return str == aws.StringValue(alarm.AlarmDescription)
	})
	return a
}

// Description adds the Description filter as an alias to the AlarmDescription filter
func (a *Alarm) Description(arn string) *Alarm {
	return a.AlarmDescription(arn)
}

// ComparisonOperator adds the ComparisonOperator filter to the filter list
// the ComparisonOperator filter: filters alarms by ComparisonOperator where 'str' provided
// is the expected ComparisonOperator value
func (a *Alarm) ComparisonOperator(str string) *Alarm {
	a.filters = append(a.filters, func(v interface{}) bool {
		alarm := convert(v)
		if alarm == nil {
			return false
		}
		shared.Debugf(
			"%s == %s -> %t\n",
			str,
			aws.StringValue(alarm.ComparisonOperator),
			str == aws.StringValue(alarm.ComparisonOperator),
		)
		return str == aws.StringValue(alarm.ComparisonOperator)
	})
	return a
}

// EvaluationPeriods adds the EvaluationPeriods filter to the filter list
// the EvaluationPeriods filter: filters alarms by EvaluationPeriods where 'i' provided
// is the expected EvaluationPeriods value
func (a *Alarm) EvaluationPeriods(i int64) *Alarm {
	a.filters = append(a.filters, func(v interface{}) bool {
		alarm := convert(v)
		if alarm == nil {
			return false
		}
		shared.Debugf(
			"%d == %d -> %t\n",
			i,
			aws.Int64Value(alarm.EvaluationPeriods),
			i == aws.Int64Value(alarm.EvaluationPeriods),
		)
		return i == aws.Int64Value(alarm.EvaluationPeriods)
	})
	return a
}

// Period adds the Period filter to the filter list
// the Period filter: filters alarms by Period where 'i' provided
// is the expected Period value
func (a *Alarm) Period(i int64) *Alarm {
	a.filters = append(a.filters, func(v interface{}) bool {
		alarm := convert(v)
		if alarm == nil {
			return false
		}
		shared.Debugf("%d == %d -> %t\n", i, aws.Int64Value(alarm.Period), i == aws.Int64Value(alarm.Period))
		return i == aws.Int64Value(alarm.Period)
	})
	return a
}

// StateValue adds the StateValue filter to the filter list
// the StateValue filter: filters alarms by StateValue where 'str' provided
// is the expected StateValue value
func (a *Alarm) StateValue(str string) *Alarm {
	a.filters = append(a.filters, func(v interface{}) bool {
		alarm := convert(v)
		if alarm == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", str, aws.StringValue(alarm.StateValue), str == aws.StringValue(alarm.StateValue))
		return str == aws.StringValue(alarm.StateValue)
	})
	return a
}

// State adds the State filter as an alias to the StateValue filter
func (a *Alarm) State(arn string) *Alarm {
	return a.StateValue(arn)
}

// Statistic adds the Statistic filter to the filter list
// the Statistic filter: filters alarms by Statistic where 'str' provided
// is the expected Statistic value
func (a *Alarm) Statistic(str string) *Alarm {
	a.filters = append(a.filters, func(v interface{}) bool {
		alarm := convert(v)
		if alarm == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", str, aws.StringValue(alarm.Statistic), str == aws.StringValue(alarm.Statistic))
		return str == aws.StringValue(alarm.Statistic)
	})
	return a
}

// Threshold adds the Threshold filter to the filter list
// the Threshold filter: filters alarms by Threshold where 'i' provided
// is the expected Threshold value
func (a *Alarm) Threshold(f float64) *Alarm {
	a.filters = append(a.filters, func(v interface{}) bool {
		alarm := convert(v)
		if alarm == nil {
			return false
		}
		shared.Debugf("%f == %f -> %t\n", f, aws.Float64Value(alarm.Threshold), f == aws.Float64Value(alarm.Threshold))
		return f == aws.Float64Value(alarm.Threshold)
	})
	return a
}

// TreatMissingData adds the TreatMissingData filter to the filter list
// the TreatMissingData filter: filters alarms by TreatMissingData where 'str' provided
// is the expected TreatMissingData value
func (a *Alarm) TreatMissingData(str string) *Alarm {
	a.filters = append(a.filters, func(v interface{}) bool {
		alarm := convert(v)
		if alarm == nil {
			return false
		}
		shared.Debugf(
			"%s == %s -> %t\n",
			str,
			aws.StringValue(alarm.TreatMissingData),
			str == aws.StringValue(alarm.TreatMissingData),
		)
		return str == aws.StringValue(alarm.TreatMissingData)
	})
	return a
}

// Unit adds the Unit filter to the filter list
// the Unit filter: filters alarms by Unit where 'unit' provunited
// is the expected Unit value
func (a *Alarm) Unit(unit string) *Alarm {
	a.filters = append(a.filters, func(v interface{}) bool {
		alarm := convert(v)
		if alarm == nil {
			return false
		}
		shared.Debugf("%s == %s -> %t\n", unit, aws.StringValue(alarm.Unit), unit == aws.StringValue(alarm.Unit))
		return unit == aws.StringValue(alarm.Unit)
	})
	return a
}

func (a *Alarm) filter(alarms []*cloudwatch.MetricAlarm) ([]*cloudwatch.MetricAlarm, error) {
	if len(alarms) == 0 {
		var err error
		alarms, err = a.alarms()
		if err != nil {
			return nil, err
		}
	}
	results := fromIface(shared.GenericFilter(a.filters, toIface(alarms)))
	if len(results) == 0 {
		log.Println("aws.cloudwatch.metric.alarm.filter had zero results: ")
		shared.Spew(os.Stdout, alarms)
	}
	return results, nil
}

func (a *Alarm) alarms() ([]*cloudwatch.MetricAlarm, error) {
	svc := cloudwatch.New(a.client)
	input := &cloudwatch.DescribeAlarmsForMetricInput{
		MetricName: a.metric.MetricName,
		Namespace:  a.metric.Namespace,
	}
	result, err := svc.DescribeAlarmsForMetric(input)
	if err != nil {
		return nil, err
	}
	return result.MetricAlarms, nil
}

func convert(in interface{}) *cloudwatch.MetricAlarm {
	out, ok := in.(*cloudwatch.MetricAlarm)
	if !ok {
		shared.Debugf("object not convertible to *cloudwatch.MetricAlarm: ")
		shared.Dump(in)
		return nil
	}
	return out
}
func toIface(in []*cloudwatch.MetricAlarm) (out []interface{}) {
	for _, i := range in {
		out = append(out, i)
	}
	return
}
func fromIface(in []interface{}) (out []*cloudwatch.MetricAlarm) {
	for _, i := range in {
		v := convert(i)
		if v == nil {
			continue
		}
		out = append(out, v)
	}
	return
}
