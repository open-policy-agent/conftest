package output

import (
	"bytes"
	"log"
	"reflect"
	"strings"
	"testing"
)

func TestStandard(t *testing.T) {
	tests := []struct {
		name     string
		input    CheckResult
		expected []string
	}{
		{
			name: "records failure and Warnings",
			input: CheckResult{
				FileName: "foo.yaml",
				Warnings: []Result{{Message: "first warning"}},
				Failures: []Result{{Message: "first failure"}},
			},
			expected: []string{
				"WARN - foo.yaml - first warning",
				"FAIL - foo.yaml - first failure",
				"",
				"2 tests, 0 passed, 1 warning, 1 failure, 0 exceptions",
			},
		},
		{
			name: "skips filenames for stdin",
			input: CheckResult{
				FileName: "-",
				Warnings: []Result{{Message: "first warning"}},
				Failures: []Result{{Message: "first failure"}},
			},
			expected: []string{
				"WARN - first warning",
				"FAIL - first failure",
				"",
				"2 tests, 0 passed, 1 warning, 1 failure, 0 exceptions",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			s := NewStandardOutputManager(log.New(buf, "", 0), false)

			if err := s.Put(tt.input); err != nil {
				t.Fatalf("put output: %v", err)
			}

			if err := s.Flush(); err != nil {
				t.Fatalf("flush output: %v", err)
			}

			actual := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
			if !reflect.DeepEqual(tt.expected, actual) {
				t.Errorf("unexpected output. expected %v actual %v", tt.expected, actual)
			}
		})
	}
}
