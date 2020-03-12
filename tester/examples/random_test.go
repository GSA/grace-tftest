package examples

import (
	"os"
	"testing"
)

func TestGoTest(t *testing.T) {
	val := os.Getenv("TESTVAR")
	if val != "test" {
		t.Fatal("failed to validate TESTVAR")
	}
}
