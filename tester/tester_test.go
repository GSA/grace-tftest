package tester

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"testing"
	"time"
)

func TestMapToKeyValueSlice(t *testing.T) {
	tt := map[string]struct {
		in       map[string]string
		expected []string
	}{
		"standard_map": {
			in:       map[string]string{"a": "a", "b": "b", "c": "c"},
			expected: []string{"a=a", "b=b", "c=c"},
		},
		"empty_map": {
			in:       map[string]string{},
			expected: []string{},
		},
		"nil_map": {
			in:       nil,
			expected: []string{},
		},
	}

	for name, tc := range tt {
		tc := tc
		t.Run(name, func(t *testing.T) {
			actual := mapToKeyValueSlice(tc.in)
			for _, e := range tc.expected {
				var found bool
				for _, a := range actual {
					if a == e {
						found = true
					}
				}
				if !found {
					t.Errorf("actual is missing element: %s, actual: %v", e, actual)
				}
			}
		})
	}
}

func TestMapMerge(t *testing.T) {
	tt := map[string]struct {
		a        map[string]string
		b        map[string]string
		expected map[string]string
	}{
		"standard_maps": {
			a:        map[string]string{"a": "a", "b": "b", "c": "c"},
			b:        map[string]string{"d": "d", "e": "e"},
			expected: map[string]string{"a": "a", "b": "b", "c": "c", "d": "d", "e": "e"},
		},
		"nil_map_a": {
			a:        nil,
			b:        map[string]string{"d": "d", "e": "e"},
			expected: map[string]string{"d": "d", "e": "e"},
		},
		"nil_map_b": {
			a:        map[string]string{"a": "a", "b": "b", "c": "c"},
			b:        nil,
			expected: map[string]string{"a": "a", "b": "b", "c": "c"},
		},
	}

	for name, tc := range tt {
		tc := tc
		t.Run(name, func(t *testing.T) {
			actual := mapMerge(tc.a, tc.b)
			for k, v := range tc.expected {
				a, ok := actual[k]
				if !ok {
					t.Errorf("actual is missing key: %s, actual: %v", k, actual)
					continue
				}
				if v != a {
					t.Errorf("actual value for key %s is invalid, expected: %s, got: %s", k, v, a)
				}
			}
		})
	}
}

func TestStartProcess(t *testing.T) {
	tt := map[string]struct {
		j                     *job
		path                  string
		args                  []string
		env                   []string
		expectedExitCode      int
		expectedOutputMatcher *regexp.Regexp
	}{
		"go_version": {
			j:                     &job{Name: "go_version", Env: []string{"GOPRIVATE=test1"}},
			path:                  "go",
			args:                  []string{"env"},
			expectedExitCode:      0,
			expectedOutputMatcher: regexp.MustCompile("GOPRIVATE=\"?test1\"?"),
		},
		"go_invalid": {
			j:                &job{Name: "go_invalid"},
			path:             "go",
			args:             []string{"invalid"},
			expectedExitCode: 2,
		},
	}

	for name, tc := range tt {
		tc := tc
		t.Run(name, func(t *testing.T) {
			var (
				stdout = &bytes.Buffer{}
				stderr = &bytes.Buffer{}
			)
			tc.j.Stdout = stdout
			tc.j.Stderr = stderr
			p, err := tc.j.startProcess(tc.path, tc.args...)
			if err != nil {
				t.Fatalf("failed to start process: %s -> %v", tc.path, err)
				return
			}
			err = p.Wait()
			if err != nil {
				if xerr, ok := err.(*exec.ExitError); ok {
					if xerr.ExitCode() != tc.expectedExitCode {
						t.Fatalf("exit code invalid, expected: %d, got: %d", tc.expectedExitCode, xerr.ExitCode())
					}
					return
				}
				t.Fatalf("failed to wait on process: %s -> %v", tc.path, err)
			}
			// windows: set GOPRIVATE=value
			// linux: GOPRIVATE="value"
			if tc.expectedOutputMatcher == nil {
				return
			}
			match := tc.expectedOutputMatcher.Match(stdout.Bytes())
			if !match {
				t.Fatalf("failed to match regex with output: %s", tc.expectedOutputMatcher.String())
			}
		})
	}
}

func TestStartProcessGoTest(t *testing.T) {
	tt := map[string]struct {
		j                *job
		path             string
		content          []byte
		expectedExitCode int
	}{
		"go_test": {
			j:                &job{Name: "go_test", Env: []string{"TESTVAR=test"}},
			path:             "go_test.go",
			content:          gotest,
			expectedExitCode: 0,
		},
	}

	for name, tc := range tt {
		tc := tc
		t.Run(name, func(t *testing.T) {
			err := ioutil.WriteFile(tc.path, tc.content, 0600)
			if err != nil {
				t.Fatalf("failed to write file: %s -> %v", tc.path, err)
				return
			}

			defer func() {
				err := retrier(100*time.Millisecond, 5, func() error {
					return os.Remove(tc.path)
				})
				if err != nil {
					fmt.Printf("failed to cleanup: %s -> %v", tc.path, err)
				}
			}()

			tc.j.Stdout = os.Stdout
			tc.j.Stderr = os.Stderr
			p, err := tc.j.startProcess("go", "test", "-v", tc.path)
			if err != nil {
				t.Fatalf("failed to run go test for file: %s -> %v", tc.path, err)
				return
			}
			err = p.Wait()
			if err != nil {
				if xerr, ok := err.(*exec.ExitError); ok {
					if xerr.ExitCode() != tc.expectedExitCode {
						t.Fatalf("exit code invalid, expected: %d, got: %d", tc.expectedExitCode, xerr.ExitCode())
					}
					return
				}
				t.Fatalf("failed to wait on go test for file: %s -> %v", tc.path, err)
			}
		})
	}
}

var gotest = []byte(`package tester

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
`)
