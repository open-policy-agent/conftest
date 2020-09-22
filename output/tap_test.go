package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestTAP(t *testing.T) {
	tests := []struct {
		name     string
		input    []CheckResult
		expected []string
	}{
		{
			name: "no warnings or errors",
			input: []CheckResult{
				{
					FileName: "examples/kubernetes/service.yaml",
				},
			},
			expected: []string{},
		},
		{
			name: "records failure and warnings",
			input: []CheckResult{
				{
					FileName: "examples/kubernetes/service.yaml",
					Warnings: []Result{{Message: "first warning"}},
					Failures: []Result{{Message: "first failure"}},
				},
			},
			expected: []string{
				"1..2",
				"not ok 1 - examples/kubernetes/service.yaml - first failure",
				"# warnings",
				"not ok 2 - examples/kubernetes/service.yaml - first warning",
				"",
			},
		},
		{
			name: "mixed failure and warnings",
			input: []CheckResult{
				{
					FileName: "examples/kubernetes/service.yaml",
					Failures: []Result{{Message: "first failure"}},
				},
			},
			expected: []string{
				"1..1",
				"not ok 1 - examples/kubernetes/service.yaml - first failure",
				"",
			},
		},
		{
			name: "handles stdin input",
			input: []CheckResult{
				{
					FileName: "-",
					Failures: []Result{{Message: "first failure"}},
				},
			},
			expected: []string{
				"1..1",
				"not ok 1 - first failure",
				"",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected := strings.Join(tt.expected, "\n")

			buf := new(bytes.Buffer)
			if err := NewTAP(buf).Output(tt.input); err != nil {
				t.Fatal("output TAP:", err)
			}

			actual := buf.String()

			if expected != actual {
				t.Errorf("unexpected output. expected %v actual %v", expected, actual)
			}
		})
	}
}
