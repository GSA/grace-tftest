package cloudwatchlogs

import "testing"

func TestService(t *testing.T) {
	svc := New(nil)
	if svc.Group.Selected() != nil {
		t.Errorf("svc.Group.Selected() should be nil")
	}
}
