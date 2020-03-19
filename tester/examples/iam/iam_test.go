package config

import (
	"os"
	"testing"

	"github.com/GSA/grace-tftest/aws/iam"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

func TestIam(t *testing.T) {
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

	svc := iam.New(sess)

	svc.
		Policy.
		Name("admin").
		Assert(t, nil).
		Statement(t, nil).
		Effect("Allow").
		Action("*").
		Resource("*").
		Assert(t)

	svc.
		Role.
		Name("admin").
		Assert(t, nil).
		Attached().
		Name("admin").
		Assert(t, nil)
}
