package output

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestStandard(t *testing.T) {
	tests := []struct {
		name     string
		input    []CheckResult
		expected []string
	}{
		{
			name: "records failures and warnings",
			input: []CheckResult{
				{
					FileName: "foo.yaml",
					Warnings: []Result{{Message: "first warning"}},
					Failures: []Result{{Message: "first failure"}},
				},
			},
			expected: []string{
				"WARN - foo.yaml - first warning",
				"FAIL - foo.yaml - first failure",
				"",
				"2 tests, 0 passed, 1 warning, 1 failure, 0 exceptions",
				"",
			},
		},
		{
			name: "skips filenames for stdin",
			input: []CheckResult{
				{
					FileName: "-",
					Warnings: []Result{{Message: "first warning"}},
					Failures: []Result{{Message: "first failure"}},
				},
			},
			expected: []string{
				"WARN - first warning",
				"FAIL - first failure",
				"",
				"2 tests, 0 passed, 1 warning, 1 failure, 0 exceptions",
				"",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected := strings.Join(tt.expected, "\n")

			buf := new(bytes.Buffer)
			standard := Standard{Writer: buf, NoColor: true}
			if err := standard.Output(tt.input); err != nil {
				t.Fatal("output standard:", err)
			}

			actual := buf.String()

			if !reflect.DeepEqual(expected, actual) {
				t.Errorf("Unexpected output. expected %v actual %v", expected, actual)
			}
		})
	}
}
