package s3

import (
	"testing"
)

// var bucketName = ""
// var keyID = ""
// var keyAlg = ""
// var lambdaArn = ""

func TestAll(t *testing.T) {
	/*
		sess := session.Must(
			session.NewSession(&aws.Config{
				Region: aws.String("us-east-1"),
			}),
		)

		// using the provided session, find a bucket with bucketName
		bucket := New(sess).Bucket.Name(bucketName).Assert(t)

		// using the bucket, filter the encryption rules to only rules
		// that use SSE, and verify it is configured with the KeyID
		// and Algorithm
		bucket.Encryption().IsSSE().ID(keyID).Alg(keyAlg).Assert(t)

		// using the bucket, verify it has a notification rule
		// that is targeting lambdaArn, has a prefix filter matching
		// 'logs/', and a suffix filter matching '.gz'
		bucket.Notification().Arn(lambdaArn).Prefix("logs/").Suffix(".gz").Assert(t)

		// using the bucket, filter the lifecycle rules to only rules
		// that have expiration rules attached, and verify the status is
		// enabled, the method is delete, and the expiration days equals 7
		bucket.Lifecycle().IsExp().Status("enabled").Method("delete").ExpDays(7).Assert(t)
	*/
}
