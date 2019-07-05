package util

import "testing"

func TestRepositoryNameContainsTag(t *testing.T) {
	tests := []struct {
		note           string
		name           string
		expectedResult bool
	}{
		{
			note:           "no port, no tag",
			name:           "instrumenta.azurecr.io/test",
			expectedResult: false,
		},
		{
			note:           "no port, contains tag",
			name:           "instrumenta.azurecr.io/test:master",
			expectedResult: true,
		},
		{
			note:           "contains port, no tag",
			name:           "localhost:5000/test",
			expectedResult: false,
		},
		{
			note:           "contains port, contains tag",
			name:           "localhost:5000/test:master",
			expectedResult: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.note, func(t *testing.T) {
			result := RepositoryNameContainsTag(tc.name)
			if result != tc.expectedResult {
				t.Errorf("Expected %v, got %v", tc.expectedResult, result)
			}
		})
	}

}
