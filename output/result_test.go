package output

import (
	"testing"
)

func TestCheckResultsHelpers(t *testing.T) {
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

	for _, testCase := range testCases {
		actual := testCase.results.ExitCode()
		if actual != testCase.expected {
			t.Errorf("Unexpected error code. expected %v, actual %v", testCase.expected, actual)
		}
	}
}

func TestExitCodeFailOnWarn(t *testing.T) {
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

	for _, testCase := range testCases {
		actual := testCase.results.ExitCodeFailOnWarn()
		if actual != testCase.expected {
			t.Errorf("Unexpected error code. expected %v, actual %v", testCase.expected, actual)
		}
	}
}
