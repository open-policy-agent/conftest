package output

import (
	"bytes"
	"testing"
)

func TestTable(t *testing.T) {
	tests := []struct {
		name     string
		input    CheckResult
		expected string
	}{
		{
			name: "no warnings or errors",
			input: CheckResult{
				FileName: "examples/kubernetes/service.yaml",
			},
			expected: "",
		},
		{
			name: "records failure and warnings",
			input: CheckResult{
				FileName: "examples/kubernetes/service.yaml",
				Warnings: []Result{{Message: "first warning"}},
				Failures: []Result{{Message: "first failure"}},
			},
			expected: `+---------+----------------------------------+---------------+
| RESULT  |               FILE               |    MESSAGE    |
+---------+----------------------------------+---------------+
| warning | examples/kubernetes/service.yaml | first warning |
| failure | examples/kubernetes/service.yaml | first failure |
+---------+----------------------------------+---------------+
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			s := NewTableOutputManager(buf)

			if err := s.Put(tt.input); err != nil {
				t.Fatalf("put output: %v", err)
			}

			if err := s.Flush(); err != nil {
				t.Fatalf("flush output: %v", err)
			}

			actual := buf.String()

			if tt.expected != actual {
				t.Errorf("unexpected output. expected %v actual %v", tt.expected, actual)
			}
		})
	}
}
