package cloudtrail

import "testing"

func TestService(t *testing.T) {
	svc := New(nil)
	if svc.Trail.Selected() != nil {
		t.Errorf("svc.Trail.Selected() should be nil")
	}
}
