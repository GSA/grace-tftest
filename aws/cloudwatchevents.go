package aws

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
	terratest "github.com/gruntwork-io/terratest/modules/aws"
)

// CloudwatchEventsRuleMatcher ... takes properties expected to be in the rule, returns
// a matcher function that validates a *cloudwatchevents.DescribeRuleOutput
func CloudwatchEventsRuleMatcher(arn string, state string, scheduleExpression string, description string) func(*cloudwatchevents.DescribeRuleOutput) bool {
	return func(r *cloudwatchevents.DescribeRuleOutput) bool {
		if len(arn) > 0 && !strings.EqualFold(arn, aws.StringValue(r.Arn)) {
			return false
		}
		if len(state) > 0 && !strings.EqualFold(state, aws.StringValue(r.State)) {
			return false
		}
		if len(scheduleExpression) > 0 && !strings.EqualFold(scheduleExpression, aws.StringValue(r.ScheduleExpression)) {
			return false
		}
		if len(description) > 0 && !strings.EqualFold(description, aws.StringValue(r.Description)) {
			return false
		}
		return true
	}
}

// MatchCloudwatchEventsRule ... uses matcher to validate cloudwatchevents rule with given name
func MatchCloudwatchEventsRule(t *testing.T, region string, rule string, matcher func(*cloudwatchevents.DescribeRuleOutput) bool) *cloudwatchevents.DescribeRuleOutput {
	out, err := MatchCloudwatchEventsRuleE(region, rule, matcher)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// MatchCloudwatchEventsRuleE ... uses matcher to validate cloudwatchevents rule with given name
func MatchCloudwatchEventsRuleE(region string, ruleName string, matcher func(*cloudwatchevents.DescribeRuleOutput) bool) (*cloudwatchevents.DescribeRuleOutput, error) {
	client, err := NewCloudWatchEventsClientE(region)
	if err != nil {
		return nil, err
	}
	out, err := client.DescribeRule(&cloudwatchevents.DescribeRuleInput{
		Name: &ruleName,
	})
	if err != nil {
		return nil, err
	}
	if !matcher(out) {
		return nil, fmt.Errorf("failed to match rule")
	}
	return out, nil
}

// FindTargetArnByRule ... finds target matching the arn on the rule with the given name
func FindTargetArnByRule(t *testing.T, region string, rule string, arn string) *cloudwatchevents.Target {
	target, err := FindTargetArnByRuleE(region, rule, arn)
	if err != nil {
		t.Fatalf("FindTargetArnByRule failed: %v", err)
	}
	return target
}

// FindTargetArnByRuleE ... finds target matching the arn on the rule with the given name
func FindTargetArnByRuleE(region string, rule string, arn string) (*cloudwatchevents.Target, error) {
	target, err := FindTargetByRuleE(region, rule, func(t *cloudwatchevents.Target) bool {
		return *t.Arn == arn
	})
	if err != nil {
		return nil, err
	}
	return target, nil
}

// FindTargetByRule ... finds target matching the matcher on the rule with the given name
func FindTargetByRule(t *testing.T, region string, rule string, matcher func(*cloudwatchevents.Target) bool) *cloudwatchevents.Target {
	target, err := FindTargetByRuleE(region, rule, matcher)
	if err != nil {
		t.Fatalf("FindTargetByRule failed: %v", err)
	}
	return target
}

// FindTargetByRuleE ... finds target matching the matcher on the rule with the given name
func FindTargetByRuleE(region string, rule string, matcher func(*cloudwatchevents.Target) bool) (*cloudwatchevents.Target, error) {
	var (
		err    error
		marker *string
	)
	more := true
	for more {
		var targets []*cloudwatchevents.Target
		targets, marker, err = ListTargetsByRuleE(region, rule, marker)
		if err != nil {
			return nil, err
		}
		for _, t := range targets {
			if matcher(t) {
				return t, nil
			}
		}
		if marker == nil {
			more = false
		}
	}

	return nil, fmt.Errorf("failed to locate a matching target")
}

// ListTargetsByRule ... returns a batch of targets for the given rule
// where marker is the index token should be nil on first call
func ListTargetsByRule(t *testing.T, region string, rule string, marker *string) ([]*cloudwatchevents.Target, *string) {
	targets, next, err := ListTargetsByRuleE(region, rule, marker)
	if err != nil {
		t.Fatalf("ListTargetsByRule failed: %v", err)
	}
	return targets, next
}

// ListTargetsByRuleE ... returns a batch of targets for the given rule
// where marker is the index token should be nil on first call
func ListTargetsByRuleE(region string, rule string, marker *string) ([]*cloudwatchevents.Target, *string, error) {
	client, err := NewCloudWatchEventsClientE(region)
	if err != nil {
		return nil, nil, err
	}
	result, err := client.ListTargetsByRule(&cloudwatchevents.ListTargetsByRuleInput{
		Rule:      &rule,
		NextToken: marker,
	})
	if err != nil {
		return nil, nil, err
	}
	return result.Targets, result.NextToken, nil
}

// NewCloudWatchEventsClient ... returns an initialized cloudwatchevents client in the given region
func NewCloudWatchEventsClient(t *testing.T, region string) *cloudwatchevents.CloudWatchEvents {
	client, err := NewCloudWatchEventsClientE(region)
	if err != nil {
		t.Fatal(err)
	}
	return client
}

// NewCloudWatchEventsClientE ... returns an initialized cloudwatchevents client in the given region
func NewCloudWatchEventsClientE(region string) (*cloudwatchevents.CloudWatchEvents, error) {
	sess, err := terratest.NewAuthenticatedSession(region)
	if err != nil {
		return nil, err
	}
	return cloudwatchevents.New(sess), nil
}
