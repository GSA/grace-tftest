package config

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/lambda"
)

// Validator is an interface for validating properties on a single *lambda.FunctionConfiguration
type Validator func(*lambda.FunctionConfiguration) error

// Config contains the necessary properties for testing a *lambda.FunctionConfiguration
type Config struct {
	client       client.ConfigProvider
	functionName string
	validators   []Validator
}

// New returns a new *Config
func New(client client.ConfigProvider, functionName string) *Config {
	return &Config{client: client, functionName: functionName}
}

// Assert executes all Validators provided against the *lambda.FunctionConfiguration
// if cfg is nil, the *lambda.FunctionConfiguration with be queried from AWS
func (c *Config) Assert(t *testing.T, cfg *lambda.FunctionConfiguration) *Config {
	err := c.validate(cfg)
	if err != nil {
		t.Error(err)
	}
	c.validators = []Validator{}
	return c
}

// Validator adds the custom Validator provided to the validation method chain
func (c *Config) Validator(validator Validator) *Config {
	c.validators = append(c.validators, validator)
	return c
}

// Env validates that the config contains an environment variable
// matching the provided key and value
func (c *Config) Env(key, value string) *Config {
	c.validators = append(c.validators, func(cfg *lambda.FunctionConfiguration) error {
		if cfg.Environment == nil {
			return fmt.Errorf("validator Env() failed: Environment was nil")
		}
		env := aws.StringValueMap(cfg.Environment.Variables)
		if v, ok := env[key]; ok {
			if v == value {
				return nil
			}
			return fmt.Errorf("validator Env() failed: Environment.Variables[%q]: %q != %q", key, v, value)
		}
		return fmt.Errorf("validator Env() failed: Environment.Variables[%q] does not exist", key)
	})
	return c
}

// Handler validates that the config's Handler property matches the provided
// handler value
func (c *Config) Handler(handler string) *Config {
	c.validators = append(c.validators, func(cfg *lambda.FunctionConfiguration) error {
		if handler != aws.StringValue(cfg.Handler) {
			return fmt.Errorf("validator Handler() failed: %q != %q", handler, aws.StringValue(cfg.Handler))
		}
		return nil
	})
	return c
}

// KeyArn validates that the config's KMSKeyArn property matches the provided
// arn value
func (c *Config) KeyArn(arn string) *Config {
	c.validators = append(c.validators, func(cfg *lambda.FunctionConfiguration) error {
		if arn != aws.StringValue(cfg.KMSKeyArn) {
			return fmt.Errorf("validator KeyArn() failed: %q != %q", arn, aws.StringValue(cfg.KMSKeyArn))
		}
		return nil
	})
	return c
}

// Role validates that the config's Role property matches the provided
// role value
func (c *Config) Role(role string) *Config {
	c.validators = append(c.validators, func(cfg *lambda.FunctionConfiguration) error {
		if role != aws.StringValue(cfg.Role) {
			return fmt.Errorf("validator Role() failed: %q != %q", role, aws.StringValue(cfg.Role))
		}
		return nil
	})
	return c
}

// Runtime validates that the config's Runtime property matches the provided
// runtime value
func (c *Config) Runtime(runtime string) *Config {
	c.validators = append(c.validators, func(cfg *lambda.FunctionConfiguration) error {
		if runtime != aws.StringValue(cfg.Runtime) {
			return fmt.Errorf("validator Runtime() failed: %q != %q", runtime, aws.StringValue(cfg.Runtime))
		}
		return nil
	})
	return c
}

// Timeout validates that the config's Timeout property matches the provided
// timeout value
func (c *Config) Timeout(timeout int) *Config {
	c.validators = append(c.validators, func(cfg *lambda.FunctionConfiguration) error {
		if int64(timeout) != aws.Int64Value(cfg.Timeout) {
			return fmt.Errorf("validator Runtime() failed: %d != %d", timeout, aws.Int64Value(cfg.Timeout))
		}
		return nil
	})
	return c
}

func (c *Config) validate(cfg *lambda.FunctionConfiguration) error {
	var err error

	if cfg == nil {
		cfg, err = c.getConfig()
		if err != nil {
			return err
		}
	}

	for _, v := range c.validators {
		err = v(cfg)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) getConfig() (*lambda.FunctionConfiguration, error) {
	svc := lambda.New(c.client)
	cfg, err := svc.GetFunctionConfiguration(&lambda.GetFunctionConfigurationInput{
		FunctionName: aws.String(c.functionName),
	})
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
