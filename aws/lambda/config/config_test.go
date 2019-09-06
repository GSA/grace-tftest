package config

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
)

func TestConfig(t *testing.T) {
	cfg := &lambda.FunctionConfiguration{
		Handler:   aws.String("a"),
		KMSKeyArn: aws.String("b"),
		Role:      aws.String("c"),
		Runtime:   aws.String("d"),
		Timeout:   aws.Int64(1),
		Environment: &lambda.EnvironmentResponse{
			Variables: aws.StringMap(map[string]string{
				"a": "a",
				"b": "b",
				"c": "c",
			}),
		},
	}
	New(nil, "").Handler("a").KeyArn("b").Role("c").Runtime("d").Timeout(1).
		Env("a", "a").Env("b", "b").Env("c", "c").Assert(t, cfg)
}
