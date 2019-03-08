package aws

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	terratest "github.com/gruntwork-io/terratest/modules/aws"
)

// S3BucketEncryptionRuleMatcher ... returns a matcher that uses the given properties to match
// a ServerSideEncryptionRule
func S3BucketEncryptionRuleMatcher(keyArn string, algorithm string) func(*s3.ServerSideEncryptionRule) bool {
	return func(r *s3.ServerSideEncryptionRule) bool {
		if len(keyArn) > 0 && !strings.EqualFold(keyArn, aws.StringValue(r.ApplyServerSideEncryptionByDefault.KMSMasterKeyID)) {
			debug("S3BucketEncryptionRuleMatcher: failed to match KMSMasterKeyID values: %q != %q\n", keyArn, aws.StringValue(r.ApplyServerSideEncryptionByDefault.KMSMasterKeyID))
			return false
		}
		if len(algorithm) > 0 && !strings.EqualFold(algorithm, aws.StringValue(r.ApplyServerSideEncryptionByDefault.SSEAlgorithm)) {
			debug("S3BucketEncryptionRuleMatcher: failed to match SSEAlgorithm values: %q != %q\n", algorithm, aws.StringValue(r.ApplyServerSideEncryptionByDefault.SSEAlgorithm))
			return false
		}
		return true
	}
}

// FindS3BucketEncryptionRule ... finds ServerSideEncryptionRule with the given matcher
// on the bucket with the given name
func FindS3BucketEncryptionRule(t *testing.T, region string, name string, matcher func(*s3.ServerSideEncryptionRule) bool) *s3.ServerSideEncryptionRule {
	rule, err := FindS3BucketEncryptionRuleE(region, name, matcher)
	if err != nil {
		t.Fatal(err)
	}
	return rule
}

// FindS3BucketEncryptionRuleE ... finds ServerSideEncryptionRule with the given matcher
// on the bucket with the given name
func FindS3BucketEncryptionRuleE(region string, name string, matcher func(*s3.ServerSideEncryptionRule) bool) (*s3.ServerSideEncryptionRule, error) {
	rules, err := GetS3BucketEncryptionRulesE(region, name)
	if err != nil {
		return nil, err
	}
	for _, r := range rules {
		if matcher(r) {
			return r, nil
		}
	}
	return nil, fmt.Errorf("failed to locate a matching encryption rule for bucket with name %q", name)
}

// GetS3BucketEncryptionRules ... gets the encryption rules for the bucket with the given name
func GetS3BucketEncryptionRules(t *testing.T, region string, name string) []*s3.ServerSideEncryptionRule {
	rules, err := GetS3BucketEncryptionRulesE(region, name)
	if err != nil {
		t.Fatal(err)
	}
	return rules
}

// GetS3BucketEncryptionRulesE ... gets the encryption rules for the bucket with the given name
func GetS3BucketEncryptionRulesE(region string, name string) ([]*s3.ServerSideEncryptionRule, error) {
	client, err := terratest.NewS3ClientE(nil, region)
	if err != nil {
		return nil, err
	}
	out, err := client.GetBucketEncryption(&s3.GetBucketEncryptionInput{
		Bucket: &name,
	})
	if err != nil {
		return nil, err
	}
	return out.ServerSideEncryptionConfiguration.Rules, nil
}

// S3BucketExpirationRuleMatcher ... returns a LifecycleRule matcher with the given properties
func S3BucketExpirationRuleMatcher(status string, method string, date *time.Time, days int64) func(*s3.LifecycleRule) bool {
	return func(r *s3.LifecycleRule) bool {
		if !strings.EqualFold(status, aws.StringValue(r.Status)) {
			debug("S3BucketExpirationRuleMatcher: failed to match Status values: %q != %q\n", status, aws.StringValue(r.Status))
			return false
		}
		if !strings.EqualFold(method, aws.StringValue(r.ID)) {
			debug("S3BucketExpirationRuleMatcher: failed to match ID values: %q != %q\n", method, aws.StringValue(r.ID))
			return false
		}
		if date != nil && !date.IsZero() && aws.TimeValue(date) != aws.TimeValue(r.Expiration.Date) {
			debug("S3BucketExpirationRuleMatcher: failed to match Date values: %v != %v\n", aws.TimeValue(date), aws.TimeValue(r.Expiration.Date))
			return false
		}
		if days > 0 && days != aws.Int64Value(r.Expiration.Days) {
			debug("S3BucketExpirationRuleMatcher: failed to match Days values: %d != %d\n", days, aws.Int64Value(r.Expiration.Days))
			return false
		}
		return true
	}
}

// FindS3BucketLifecycleRule ... retrieves the LifecycleRule for the bucket with the given name that matches the given matcher
func FindS3BucketLifecycleRule(t *testing.T, region string, name string, matcher func(*s3.LifecycleRule) bool) *s3.LifecycleRule {
	rule, err := FindS3BucketLifecycleRuleE(region, name, matcher)
	if err != nil {
		t.Fatal(err)
	}
	return rule
}

// FindS3BucketLifecycleRuleE ... retrieves the lifecycle rule for the bucket with the given name that matches the given matcher
func FindS3BucketLifecycleRuleE(region string, name string, matcher func(*s3.LifecycleRule) bool) (*s3.LifecycleRule, error) {
	rules, err := GetS3BucketLifecycleRulesE(region, name)
	if err != nil {
		return nil, err
	}
	for _, r := range rules {
		if matcher(r) {
			return r, nil
		}
	}
	return nil, fmt.Errorf("failed to locate a matching lifecycle rule for bucket with name %q", name)
}

// GetS3BucketLifecycleRules ... retrieves all lifecycle rules for the bucket with the given name
func GetS3BucketLifecycleRules(t *testing.T, region string, name string) []*s3.LifecycleRule {
	rules, err := GetS3BucketLifecycleRulesE(region, name)
	if err != nil {
		t.Fatal(err)
	}
	return rules
}

// GetS3BucketLifecycleRulesE ... retrieves all lifecycle rules for the bucket with the given name
func GetS3BucketLifecycleRulesE(region string, name string) ([]*s3.LifecycleRule, error) {
	client, err := terratest.NewS3ClientE(nil, region)
	if err != nil {
		return nil, err
	}
	out, err := client.GetBucketLifecycleConfiguration(&s3.GetBucketLifecycleConfigurationInput{
		Bucket: &name,
	})
	if err != nil {
		return nil, err
	}
	return out.Rules, nil
}
