package config

import "testing"

func TestService(t *testing.T) {
	svc := New(nil)
	if svc.Recorder == nil {
		t.Fatalf("failed create service property 'Recorder' is nil")
	}
}
