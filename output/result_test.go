package output

import (
	"testing"
)

func TestExitCode(t *testing.T) {
	warning := CheckResult{
		Warnings: []Result{{}},
	}

	failure := CheckResult{
		Failures: []Result{{}},
	}

	testCases := []struct {
		results  []CheckResult
		expected int
	}{
		{results: []CheckResult{}, expected: 0},
		{results: []CheckResult{warning}, expected: 0},
		{results: []CheckResult{failure}, expected: failureExitCode},
		{results: []CheckResult{warning, failure}, expected: failureExitCode},
	}

	for _, testCase := range testCases {
		actual := ExitCode(testCase.results)

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
		results  []CheckResult
		expected int
	}{
		{results: []CheckResult{}, expected: 0},
		{results: []CheckResult{warning}, expected: warningExitCode},
		{results: []CheckResult{failure}, expected: failureExitCode},
		{results: []CheckResult{warning, failure}, expected: failureExitCode},
	}

	for _, testCase := range testCases {
		actual := ExitCodeFailOnWarn(testCase.results)

		if actual != testCase.expected {
			t.Errorf("Unexpected error code. expected %v, actual %v", testCase.expected, actual)
		}
	}
}
