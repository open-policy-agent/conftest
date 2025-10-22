package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGitHub(t *testing.T) {
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
				"::group::Testing \"examples/kubernetes/service.yaml\" against 0 policies in namespace \"namespace\"",
				"::notice file=examples/kubernetes/service.yaml,line=1::Number of successful checks: 0",
				"::endgroup::",
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
				"::group::Testing \"examples/kubernetes/service.yaml\" against 2 policies in namespace \"namespace\"",
				"::error file=examples/kubernetes/service.yaml,line=1::first failure",
				"::warning file=examples/kubernetes/service.yaml,line=1::first warning",
				"::notice file=examples/kubernetes/service.yaml,line=1::Number of successful checks: 0",
				"::endgroup::",
				"2 tests, 0 passed, 1 warning, 1 failure, 0 exceptions",
				"",
			},
		},
		{
			name: "mixed failure, warnings, successes and skipped",
			input: CheckResults{
				{
					FileName:  "examples/kubernetes/service.yaml",
					Namespace: "namespace",
					Failures:  []Result{{Message: "first failure"}},
					Skipped:   []Result{{Message: "first skipped"}},
					Successes: 10,
				},
			},
			expected: []string{
				"::group::Testing \"examples/kubernetes/service.yaml\" against 12 policies in namespace \"namespace\"",
				"::error file=examples/kubernetes/service.yaml,line=1::first failure",
				"::notice file=examples/kubernetes/service.yaml,line=1::Test was skipped: first skipped",
				"::notice file=examples/kubernetes/service.yaml,line=1::Number of successful checks: 10",
				"::endgroup::",
				"12 tests, 10 passed, 0 warnings, 1 failure, 0 exceptions",
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
				"::group::Testing \"-\" against 1 policies in namespace \"namespace\"",
				"::error file=-,line=1::first failure",
				"::notice file=-,line=1::Number of successful checks: 0",
				"::endgroup::",
				"1 test, 0 passed, 0 warnings, 1 failure, 0 exceptions",
				"",
			},
		},
		{
			name: "handles location in the same file",
			input: CheckResults{
				{
					FileName:  "foo.json",
					Namespace: "main",
					Failures: []Result{{
						Message: "first failure",
						Location: &Location{
							File: "foo.json",
							Line: json.Number("10"),
						},
					}},
				},
			},
			expected: []string{
				"::group::Testing \"foo.json\" against 1 policies in namespace \"main\"",
				"::error file=foo.json,line=10::first failure",
				"::notice file=foo.json,line=1::Number of successful checks: 0",
				"::endgroup::",
				"1 test, 0 passed, 0 warnings, 1 failure, 0 exceptions",
				"",
			},
		},
		{
			name: "handles location outside of the file",
			input: CheckResults{
				{
					FileName:  "foo.json",
					Namespace: "main",
					Failures: []Result{{
						Message: "first failure",
						Location: &Location{
							File: "../other_file.json",
							Line: json.Number("10"),
						},
					}},
				},
			},
			expected: []string{
				"::group::Testing \"foo.json\" against 1 policies in namespace \"main\"",
				"::error file=../other_file.json,line=10::first failure",
				"::error file=foo.json,line=1::(ORIGINATING FROM ../other_file.json L10) first failure",
				"::notice file=foo.json,line=1::Number of successful checks: 0",
				"::endgroup::",
				"1 test, 0 passed, 0 warnings, 1 failure, 0 exceptions",
				"",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected := strings.Join(tt.expected, "\n")

			buf := new(bytes.Buffer)
			if err := NewGitHub(buf).Output(tt.input); err != nil {
				t.Fatal("output GitHub:", err)
			}
			if diff := cmp.Diff(buf.String(), expected); diff != "" {
				t.Errorf("GitHub.Output() produced an unexpected diff:\n%s", diff)
			}
		})
	}
}
