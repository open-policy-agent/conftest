package policy

import (
	"testing"
)

func TestRepositoryToPull(t *testing.T) {
	tests := []struct {
		policy   Policy
		expected string
	}{
		{
			policy:   Policy{Repository: "my.url.com/repository:v1", Tag: ""},
			expected: "my.url.com/repository:v1",
		},
		{
			policy:   Policy{Repository: "my.url.com/repository", Tag: ""},
			expected: "my.url.com/repository:latest",
		},
		{
			policy:   Policy{Repository: "my.url.com/repository", Tag: "v1"},
			expected: "my.url.com/repository:v1",
		},
	}

	for _, test := range tests {
		actual := getRepositoryFromPolicy(test.policy)
		if actual != test.expected {
			t.Errorf("Expected %v, got %v", test.expected, actual)
		}
	}
}
