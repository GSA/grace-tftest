package bucket

import (
	"errors"
	"testing"

	"github.com/GSA/grace-tftest/aws/s3/bucket/encryption"
	"github.com/GSA/grace-tftest/aws/s3/bucket/lifecycle"
	"github.com/GSA/grace-tftest/aws/s3/bucket/notification"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/s3"
)

type checkFunc func() error

// Bucket contains properties for testing S3 Bucket objects
type Bucket struct {
	client  client.ConfigProvider
	checker checkFunc
	name    string
}

// New returns a new *Bucket
func New(client client.ConfigProvider) *Bucket {
	b := &Bucket{client: client}
	b.checker = b.head
	return b
}

// Notification returns a new *notification.Notification
// instantiated with the current bucket name set by calling Name()
func (b *Bucket) Notification() *notification.Notification {
	return notification.New(b.client, b.name)
}

// Encryption returns a new *encryption.Encryption
// instantiated with the current bucket name set by calling Name()
func (b *Bucket) Encryption() *encryption.Encryption {
	return encryption.New(b.client, b.name)
}

// Lifecycle returns a new *lifecycle.Lifecycle
// instantiated with the current bucket name set by calling Name()
func (b *Bucket) Lifecycle() *lifecycle.Lifecycle {
	return lifecycle.New(b.client, b.name)
}

// Assert executes the checker method (normally s3.Head)
// to verify the bucket with the name give to Name exists
// fails if bucket doesn't exist
func (b *Bucket) Assert(t *testing.T) *Bucket {
	err := b.checker()
	if err != nil {
		t.Error(err)
		return b
	}
	return b
}

// Name sets the bucket name to use when calling
// Assert
func (b *Bucket) Name(name string) *Bucket {
	b.name = name
	return b
}

func (b *Bucket) head() (err error) {
	if len(b.name) == 0 {
		return errors.New("a bucket name must be provided")
	}
	svc := s3.New(b.client)
	_, err = svc.HeadBucket(&s3.HeadBucketInput{Bucket: &b.name})
	return
}
