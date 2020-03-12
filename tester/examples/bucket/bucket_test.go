package policy

import (
	"fmt"
	"os"
	"testing"

	"github.com/GSA/grace-tftest/aws/s3"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

func TestBucket(t *testing.T) {
	url := "http://localhost:" + os.Getenv("MOTO_PORT")
	fmt.Printf("connecting to: %s\n", url)
	sess, err := session.NewSession(&aws.Config{
		Endpoint:         aws.String(url),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		t.Fatalf("failed to connect to moto: %s -> %v", url, err)
	}

	s3.New(sess).
		Bucket.
		Name("bucket").
		Assert(t)
}
