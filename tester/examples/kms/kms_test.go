package kms

import (
	"os"
	"testing"

	"github.com/GSA/grace-tftest/aws/kms"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

func TestKm(t *testing.T) {
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

	svc := kms.New(sess)

	alias := svc.
		Alias.
		Name("alias/key").
		Assert(t)

	alias.
		Policy(t).
		Statement(t, nil).
		Principal("AWS", "arn:aws:iam::*:root").
		Effect("Allow").
		Resource("*").
		Action("kms:*").
		Assert(t)
}
