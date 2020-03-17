package rule

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
)

// nolint: funlen
func TestRule(t *testing.T) {
	rules := []*configservice.ConfigRule{
		{
			ConfigRuleArn:             aws.String("a"),
			ConfigRuleId:              aws.String("b"),
			ConfigRuleName:            aws.String("c"),
			ConfigRuleState:           aws.String("d"),
			CreatedBy:                 aws.String("e"),
			Description:               aws.String("f"),
			MaximumExecutionFrequency: aws.String("g"),
			Scope: &configservice.Scope{
				ComplianceResourceId: aws.String("h"),
				ComplianceResourceTypes: aws.StringSlice([]string{
					"i", "j", "k",
				}),
				TagKey:   aws.String("l"),
				TagValue: aws.String("m"),
			},
			Source: &configservice.Source{
				Owner:            aws.String("n"),
				SourceIdentifier: aws.String("o"),
				SourceDetails: []*configservice.SourceDetail{
					{
						EventSource:               aws.String("p"),
						MaximumExecutionFrequency: aws.String("q"),
						MessageType:               aws.String("r"),
					},
				},
			},
		},
		{
			ConfigRuleArn:             aws.String("c"),
			ConfigRuleId:              aws.String("d"),
			ConfigRuleName:            aws.String("e"),
			ConfigRuleState:           aws.String("f"),
			CreatedBy:                 aws.String("g"),
			Description:               aws.String("h"),
			MaximumExecutionFrequency: aws.String("i"),
			Scope: &configservice.Scope{
				ComplianceResourceId: aws.String("j"),
				ComplianceResourceTypes: aws.StringSlice([]string{
					"k", "l", "m",
				}),
				TagKey:   aws.String("n"),
				TagValue: aws.String("o"),
			},
			Source: nil,
		},
		{
			ConfigRuleArn:             aws.String("d"),
			ConfigRuleId:              aws.String("e"),
			ConfigRuleName:            aws.String("f"),
			ConfigRuleState:           aws.String("g"),
			CreatedBy:                 aws.String("h"),
			Description:               aws.String("i"),
			MaximumExecutionFrequency: aws.String("j"),
			Scope:                     nil,
			Source:                    nil,
		},
	}

	svc := New(nil)

	svc.
		Arn("a").
		ID("b").
		Name("c").
		State("d").
		CreatedBy("e").
		Description("f").
		Frequency("g").
		ScopeID("h").
		ScopeTypes("i", "j", "k").
		ScopeTagKey("l").
		ScopeTagValue("m").
		SourceOwner("n").
		SourceID("o").
		SourceDetailEventSource("p").
		SourceDetailFrequency("q").
		SourceDetailMessageType("r").
		Assert(t, rules...)

	rule := svc.Name("e").Assert(t, rules...).Selected()
	if aws.StringValue(rule.ConfigRuleArn) != "c" {
		t.Fatalf("failed to match rule, expected: %s, got: %s", "c", aws.StringValue(rule.ConfigRuleArn))
	}
}
