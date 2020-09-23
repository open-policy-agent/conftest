package plugin

import "testing"

func TestIsDirectory(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{input: "/", expected: true},
		{input: "/abs/path", expected: true},
		{input: "some/path", expected: true},
		{input: "file://some/path", expected: true},
		{input: "C:\\some\\path", expected: true},
		{input: "unknown", expected: true},
		{input: "unknown.com", expected: true},

		{input: "github.com/username/repo", expected: false},
	}

	for _, testCase := range testCases {
		actual, err := isDirectory(testCase.input)
		if err != nil {
			t.Fatal("is directory:", err)
		}

		if actual != testCase.expected {
			t.Errorf("Directory check failed. expected %v, actual %v", testCase.expected, actual)
		}
	}
}
