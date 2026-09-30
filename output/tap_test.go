package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestTAP(t *testing.T) {
	tests := []struct {
		name     string
		input    CheckResults
		expected []string
	}{
		{
			name: "no warnings or errors",
			input: CheckResults{
				{
					FileName:  "examples/kubernetes/service.yaml",
					Namespace: "namespace",
				},
			},
			expected: []string{},
		},
		{
			name: "records failure and warnings",
			input: CheckResults{
				{
					FileName:  "examples/kubernetes/service.yaml",
					Namespace: "namespace",
					Warnings:  []Result{{Message: "first warning"}},
					Failures:  []Result{{Message: "first failure"}},
				},
			},
			expected: []string{
				"1..2",
				"not ok 1 - examples/kubernetes/service.yaml - namespace - first failure",
				"# warnings",
				"not ok 2 - examples/kubernetes/service.yaml - namespace - first warning",
				"",
			},
		},
		{
			name: "multiple files use a single plan with continuous numbering",
			input: CheckResults{
				{
					FileName:  "examples/kubernetes/service.yaml",
					Namespace: "namespace",
					Failures:  []Result{{Message: "first failure"}},
				},
				{
					FileName:  "examples/kubernetes/deployment.yaml",
					Namespace: "namespace",
					Warnings:  []Result{{Message: "second file warning"}},
					Successes: 1,
				},
			},
			expected: []string{
				"1..3",
				"not ok 1 - examples/kubernetes/service.yaml - namespace - first failure",
				"# warnings",
				"not ok 2 - examples/kubernetes/deployment.yaml - namespace - second file warning",
				"# successes",
				"ok 3 - examples/kubernetes/deployment.yaml - namespace - SUCCESS",
				"",
			},
		},
		{
			name: "files without tests do not truncate later results",
			input: CheckResults{
				{
					FileName:  "examples/kubernetes/service.yaml",
					Namespace: "namespace",
				},
				{
					FileName:  "examples/kubernetes/deployment.yaml",
					Namespace: "namespace",
					Failures:  []Result{{Message: "first failure"}},
				},
			},
			expected: []string{
				"1..1",
				"not ok 1 - examples/kubernetes/deployment.yaml - namespace - first failure",
				"",
			},
		},
		{
			name: "mixed failure, warnings and skipped",
			input: CheckResults{
				{
					FileName:  "examples/kubernetes/service.yaml",
					Namespace: "namespace",
					Failures:  []Result{{Message: "first failure"}},
					Skipped:   []Result{{Message: "first skipped"}},
				},
			},
			expected: []string{
				"1..2",
				"not ok 1 - examples/kubernetes/service.yaml - namespace - first failure",
				"# skip",
				"ok 2 - examples/kubernetes/service.yaml - namespace - first skipped",
				"",
			},
		},
		{
			name: "handles stdin input",
			input: CheckResults{
				{
					FileName:  "-",
					Namespace: "namespace",
					Failures:  []Result{{Message: "first failure"}},
				},
			},
			expected: []string{
				"1..1",
				"not ok 1 - - namespace - first failure",
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
