package trail

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/cloudtrail"
)

func TestTrail(t *testing.T) {
	trails := []*cloudtrail.TrailInfo{
		{TrailARN: aws.String("a"), Name: aws.String("b"), HomeRegion: aws.String("c")},
		{TrailARN: aws.String("a"), Name: aws.String("b"), HomeRegion: aws.String("d")},
		{TrailARN: aws.String("a"), Name: aws.String("b"), HomeRegion: aws.String("e")},
		{TrailARN: aws.String("a"), Name: aws.String("b"), HomeRegion: aws.String("f")},
	}

	New(nil).Arn("a").Region("c").Name("b").Assert(t, trails...)

	trail := New(nil).Arn("a").Name("b").First(t, trails...).Selected()
	if aws.StringValue(trail.Name) != "c" {
		t.Fatalf("first element invalid, expected c, got: %#s\n", trail.Name)
	}
}
