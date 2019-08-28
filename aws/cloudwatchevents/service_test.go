package cloudwatchevents

import "testing"

func TestService(t *testing.T) {
	svc := New(nil)
	if svc.Rule.Selected() != nil {
		t.Errorf("svc.Rule.Selected() should be nil")
	}
}
