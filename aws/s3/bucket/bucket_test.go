package bucket

import "testing"

func TestBucket(t *testing.T) {
	b := New(nil)

	// use custom checker for offline mode
	//testing commit
	b.checker = func() error { return nil }

	b.Name("test").Assert(t)
}
