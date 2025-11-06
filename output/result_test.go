package output

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewResult(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc    string
		input   map[string]any
		want    Result
		wantErr bool
	}{
		{
			desc:    "no metadata is an error",
			wantErr: true,
		},
		{
			desc:    "missing msg is an error",
			input:   map[string]any{},
			wantErr: true,
		},
		{
			desc:    "non-string msg is an error",
			input:   map[string]any{"msg": 123},
			wantErr: true,
		},
		{
			desc:  "msg only",
			input: map[string]any{"msg": "message"},
			want: Result{
				Message:  "message",
				Metadata: make(map[string]any),
			},
		},
		{
			desc: "msg with location and metadata",
			input: map[string]any{
				"msg": "message",
				"_loc": map[string]any{
					"file": "some_file.json",
					"line": json.Number("123"),
				},
				"other": "metadata",
			},
			want: Result{
				Message: "message",
				Location: &Location{
					File: "some_file.json",
					Line: json.Number("123"),
				},
				Metadata: map[string]any{"other": "metadata"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			got, err := NewResult(tc.input)
			if gotErr := err != nil; gotErr != tc.wantErr {
				t.Fatalf("NewResult() error = %v, want %v", err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Errorf("NewResult() produced an unexpected diff:\n%s", diff)
			}
		})
	}
}

func TestCheckResultsHelpers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc             string
		results          CheckResults
		wantHasFailure   bool
		wantHasWarning   bool
		wantHasException bool
		wantOnlySuccess  bool
	}{
		{
			desc:            "Empty returns success only",
			wantOnlySuccess: true,
		},
		{
			desc:            "Result with only success returns success only",
			results:         CheckResults{{Successes: 12345}},
			wantOnlySuccess: true,
		},
		{
			desc: "Mix of success and non-success",
			results: CheckResults{
				{Successes: 12345},
				{Failures: []Result{{Message: "Failure"}}},
				{Warnings: []Result{{Message: "Failure"}}},
				{Exceptions: []Result{{Message: "Failure"}}},
			},
			wantHasFailure:   true,
			wantHasWarning:   true,
			wantHasException: true,
		},
		{
			desc: "Failure, warning, exception",
			results: CheckResults{
				{Failures: []Result{{Message: "Failure"}}},
				{Warnings: []Result{{Message: "Failure"}}},
				{Exceptions: []Result{{Message: "Failure"}}},
			},
			wantHasFailure:   true,
			wantHasWarning:   true,
			wantHasException: true,
		},
		{
			desc: "Failure only",
			results: CheckResults{
				{Failures: []Result{{Message: "Failure"}}},
			},
			wantHasFailure: true,
		},
		{
			desc: "Warning only",
			results: CheckResults{
				{Warnings: []Result{{Message: "Failure"}}},
			},
			wantHasWarning: true,
		},
		{
			desc: "Exception only",
			results: CheckResults{
				{Exceptions: []Result{{Message: "Failure"}}},
			},
			wantHasException: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			if gotFailure := tc.results.HasFailure(); gotFailure != tc.wantHasFailure {
				t.Errorf("HasFailure() = %v, want %v", gotFailure, tc.wantHasFailure)
			}
			if gotWarning := tc.results.HasWarning(); gotWarning != tc.wantHasWarning {
				t.Errorf("HasWarning() = %v, want %v", gotWarning, tc.wantHasWarning)
			}
			if gotException := tc.results.HasException(); gotException != tc.wantHasException {
				t.Errorf("HasException() = %v, want %v", gotException, tc.wantHasException)
			}
			if gotSuccess := tc.results.OnlySuccess(); gotSuccess != tc.wantOnlySuccess {
				t.Errorf("OnlySuccess() = %v, want %v", gotSuccess, tc.wantOnlySuccess)
			}
		})
	}
}

func TestExitCode(t *testing.T) {
	t.Parallel()

	warning := CheckResult{
		Warnings: []Result{{}},
	}

	failure := CheckResult{
		Failures: []Result{{}},
	}

	skipped := CheckResult{
		Skipped: []Result{{}},
	}

	testCases := []struct {
		results  CheckResults
		expected int
	}{
		{results: CheckResults{}, expected: 0},
		{results: CheckResults{warning}, expected: 0},
		{results: CheckResults{skipped}, expected: 0},
		{results: CheckResults{failure}, expected: 1},
		{results: CheckResults{warning, failure}, expected: 1},
	}

	for i, testCase := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			actual := testCase.results.ExitCode()
			if actual != testCase.expected {
				t.Errorf("Unexpected error code. expected %v, actual %v", testCase.expected, actual)
			}
		})
	}
}

func TestExitCodeFailOnWarn(t *testing.T) {
	t.Parallel()

	warning := CheckResult{
		Warnings: []Result{{}},
	}

	failure := CheckResult{
		Failures: []Result{{}},
	}

	testCases := []struct {
		results  CheckResults
		expected int
	}{
		{results: CheckResults{}, expected: 0},
		{results: CheckResults{warning}, expected: 1},
		{results: CheckResults{failure}, expected: 2},
		{results: CheckResults{warning, failure}, expected: 2},
	}

	for i, testCase := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			actual := testCase.results.ExitCodeFailOnWarn()
			if actual != testCase.expected {
				t.Errorf("Unexpected error code. expected %v, actual %v", testCase.expected, actual)
			}
		})
	}
}
