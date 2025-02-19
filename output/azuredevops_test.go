package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestAzureDevOps(t *testing.T) {
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
			expected: []string{
				"##[section]Testing 'examples/kubernetes/service.yaml' against 0 policies in namespace 'namespace'",
				"##[group]See conftest results",
				"##[endgroup]",
				"0 tests, 0 passed, 0 warnings, 0 failures, 0 exceptions",
				"",
			},
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
				"##[section]Testing 'examples/kubernetes/service.yaml' against 2 policies in namespace 'namespace'",
				"##[group]See conftest results",
				"##vso[task.logissue type=error] file=examples/kubernetes/service.yaml --> first failure",
				"##vso[task.logissue type=warning] file=examples/kubernetes/service.yaml --> first warning",
				"##[endgroup]",
				"2 tests, 0 passed, 1 warning, 1 failure, 0 exceptions",
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
				"##[section]Testing 'examples/kubernetes/service.yaml' against 2 policies in namespace 'namespace'",
				"##[group]See conftest results",
				"##vso[task.logissue type=error] file=examples/kubernetes/service.yaml --> first failure",
				"skipped file=examples/kubernetes/service.yaml first skipped",
				"##[endgroup]",
				"2 tests, 0 passed, 0 warnings, 1 failure, 0 exceptions",
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
				"##[section]Testing '-' against 1 policies in namespace 'namespace'",
				"##[group]See conftest results",
				"##vso[task.logissue type=error] file=- --> first failure",
				"##[endgroup]",
				"1 test, 0 passed, 0 warnings, 1 failure, 0 exceptions",
				"",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected := strings.Join(tt.expected, "\n")

			buf := new(bytes.Buffer)
			if err := NewAzureDevOps(buf).Output(tt.input); err != nil {
				t.Fatal("output Azure DevOps:", err)
			}

			actual := buf.String()

			if expected != actual {
				t.Errorf("unexpected output. expected %v actual %v", expected, actual)
			}
		})
	}
}
