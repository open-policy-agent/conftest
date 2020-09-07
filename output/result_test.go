package output

import (
	"testing"
)

func TestGetExitCode(t *testing.T) {
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
		{results: []CheckResult{failure}, expected: 1},
		{results: []CheckResult{warning, failure}, expected: 1},
	}

	for _, testCase := range testCases {
		actual := GetExitCode(testCase.results, false)

		if actual != testCase.expected {
			t.Errorf("Unexpected error code. expected %v, actual %v", testCase.expected, actual)
		}
	}
}

func TestGetExitCode_FailOnWarn(t *testing.T) {
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
		{results: []CheckResult{warning}, expected: 1},
		{results: []CheckResult{failure}, expected: 2},
		{results: []CheckResult{warning, failure}, expected: 2},
	}

	for _, testCase := range testCases {
		actual := GetExitCode(testCase.results, true)

		if actual != testCase.expected {
			t.Errorf("Unexpected error code. expected %v, actual %v", testCase.expected, actual)
		}
	}
}
