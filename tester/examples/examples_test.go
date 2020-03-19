package examples

import (
	"testing"

	"github.com/GSA/grace-tftest/tester"
)

func TestRun(t *testing.T) {
	err := tester.Run(&tester.Config{
		Dir:        ".",
		Env:        map[string]string{"TFTEST_DEBUG": "true"},
		JobsPerCPU: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
}
