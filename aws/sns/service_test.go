package sns

import "testing"

func TestService(t *testing.T) {
	svc := New(nil)
	if svc.Topic.Selected() != nil {
		t.Errorf("svc.Topic.Selected() should be nil")
	}
}
