package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestTable(t *testing.T) {
	tests := []struct {
		name     string
		input    CheckResults
		expected []string
	}{
		{
			name: "No warnings or errors",
			input: CheckResults{
				{
					FileName: "examples/kubernetes/service.yaml",
				},
			},
			expected: []string{},
		},
		{
			name: "A warning, a failure, a skipped",
			input: CheckResults{
				{
					FileName:  "examples/kubernetes/service.yaml",
					Namespace: "namespace",
					Warnings:  []Result{{Message: "first warning"}},
					Failures:  []Result{{Message: "first failure"}},
					Skipped:   []Result{{Message: "first skipped"}},
				},
			},
			expected: []string{
				`┌─────────┬──────────────────────────────────┬───────────┬───────────────┐`,
				`│ RESULT  │               FILE               │ NAMESPACE │    MESSAGE    │`,
				`├─────────┼──────────────────────────────────┼───────────┼───────────────┤`,
				`│ warning │ examples/kubernetes/service.yaml │ namespace │ first warning │`,
				`│ skipped │ examples/kubernetes/service.yaml │ namespace │ first skipped │`,
				`│ failure │ examples/kubernetes/service.yaml │ namespace │ first failure │`,
				`└─────────┴──────────────────────────────────┴───────────┴───────────────┘`,
				``,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected := strings.Join(tt.expected, "\n")

			buf := new(bytes.Buffer)
			if err := NewTable(buf).Output(tt.input); err != nil {
				t.Fatal("output table:", err)
			}
			actual := buf.String()

			if expected != actual {
				t.Errorf("Unexpected output. expected %v actual %v", expected, actual)
			}
		})
	}
}
