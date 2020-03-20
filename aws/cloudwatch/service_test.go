package cloudwatch

import "testing"

func TestService(t *testing.T) {
	svc := New(nil)
	if svc.Metric.Selected() != nil {
		t.Errorf("svc.Metric.Selected() should be nil")
	}
}
