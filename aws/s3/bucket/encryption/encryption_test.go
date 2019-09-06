package encryption

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

func TestEncryption(t *testing.T) {
	rules := []*s3.ServerSideEncryptionRule{
		{ApplyServerSideEncryptionByDefault: &s3.ServerSideEncryptionByDefault{
			KMSMasterKeyID: aws.String("a"),
			SSEAlgorithm:   aws.String("b"),
		}},
	}

	New(nil, "").IsSSE().ID("a").Alg("b").Assert(t, rules...)
}
