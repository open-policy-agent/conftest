package output

import (
	"os"
	"reflect"
	"testing"
)

func TestGetOutputter(t *testing.T) {
	testCases := []struct {
		input    string
		expected Outputter
	}{
		{
			input:    OutputStandard,
			expected: NewStandard(os.Stdout),
		},
		{
			input:    OutputJSON,
			expected: NewJSON(os.Stdout),
		},
		{
			input:    OutputTAP,
			expected: NewTAP(os.Stdout),
		},
		{
			input:    OutputTable,
			expected: NewTable(os.Stdout),
		},
		{
			input:    OutputJUnit,
			expected: NewJUnit(os.Stdout, false),
		},
		{
			input:    OutputGitHub,
			expected: NewGitHub(os.Stdout),
		},
		{
			input:    OutputAzureDevOps,
			expected: NewAzureDevOps(os.Stdout),
		},
		{
			input:    OutputSARIF,
			expected: NewSARIF(os.Stdout),
		},
		{
			input:    "unknown_format",
			expected: NewStandard(os.Stdout),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.input, func(t *testing.T) {
			actual := Get(testCase.input, Options{NoColor: true})

			actualType := reflect.TypeOf(actual)

			expectedType := reflect.TypeOf(testCase.expected)
			if expectedType != actualType {
				t.Errorf("Unexpected outputter. expected %v actual %v", expectedType, actualType)
			}
		})
	}
}
