package output

import (
	"bytes"
	"log"
	"testing"
)

func TestTAP(t *testing.T) {
	tests := []struct {
		name     string
		input    CheckResult
		expected string
	}{
		{
			name:     "no warnings or errors",
			input:    CheckResult{FileName: "testdata/kubernetes/service.yaml"},
			expected: "",
		},
		{
			name: "records failure and warnings",
			input: CheckResult{
				FileName: "testdata/kubernetes/service.yaml",
				Warnings: []Result{{Message: "first warning"}},
				Failures: []Result{{Message: "first failure"}},
			},
			expected: `1..2
not ok 1 - testdata/kubernetes/service.yaml - first failure
# warnings
not ok 2 - testdata/kubernetes/service.yaml - first warning
`,
		},
		{
			name: "mixed failure and warnings",
			input: CheckResult{
				FileName: "testdata/kubernetes/service.yaml",
				Failures: []Result{{Message: "first failure"}},
			},
			expected: `1..1
not ok 1 - testdata/kubernetes/service.yaml - first failure
`,
		},
		{
			name: "handles stdin input",
			input: CheckResult{
				FileName: "-",
				Failures: []Result{{Message: "first failure"}},
			},
			expected: `1..1
not ok 1 - first failure
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			s := NewTAPOutputManager(log.New(buf, "", 0))

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
