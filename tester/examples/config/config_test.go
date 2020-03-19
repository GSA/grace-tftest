package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/GSA/grace-tftest/aws/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

func TestConfig(t *testing.T) {
	port := os.Getenv("MOTO_PORT")
	if len(port) == 0 {
		t.Skipf("skipping testing, MOTO_PORT not set in environment variables")
	}

	url := "http://localhost:" + os.Getenv("MOTO_PORT")
	t.Logf("connecting to: %s\n", url)
	sess, err := session.NewSession(&aws.Config{
		Endpoint:         aws.String(url),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		t.Fatalf("failed to connect to moto: %s -> %v", url, err)
	}

	svc := config.New(sess)

	svc.
		Recorder.
		Name("config").
		RoleArn("arn:aws:iam::123456789012:role").
		AllSupported(true).
		IncludeGlobalResourceTypes(true).
		Assert(t, nil)

	assert.True(t, svc.Recorder.Recording(t, nil))

	svc.
		DeliveryChannel.
		Name("config").
		BucketName("config").
		Frequency("Three_Hours").
		Assert(t, nil)

	/* Not yet supported by moto
	svc.
		Rule.
		Name("config").
		SourceOwner("AWS").
		SourceID("CLOUDWATCH_ALARM_ACTION_CHECK").
		Assert(t, nil)*/
}
