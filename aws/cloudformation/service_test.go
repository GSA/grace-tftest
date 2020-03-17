package cloudformation

import "testing"

func TestService(t *testing.T) {
	svc := New(nil)
	if svc.Stack.Selected() != nil {
		t.Errorf("svc.Stack.Selected() should be nil")
	}
}
