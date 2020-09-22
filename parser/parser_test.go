package parser

import (
	"reflect"
	"testing"

	"github.com/open-policy-agent/conftest/parser/docker"
	"github.com/open-policy-agent/conftest/parser/hcl2"
	"github.com/open-policy-agent/conftest/parser/yaml"
)

func TestNewFromPath(t *testing.T) {
	testCases := []struct {
		path     string
		expected Parser
	}{
		{
			"-",
			&yaml.Parser{},
		},
		{
			"test.yaml",
			&yaml.Parser{},
		},
		{
			"test.yml",
			&yaml.Parser{},
		},
		{
			"dockerfile",
			&docker.Parser{},
		},
		{
			"Dockerfile",
			&docker.Parser{},
		},
		{
			"test.tf",
			&hcl2.Parser{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.path, func(t *testing.T) {
			expectedType := reflect.TypeOf(testCase.expected)

			actual, err := NewFromPath(testCase.path)
			if err != nil {
				t.Fatal("from path:", err)
			}
			actualType := reflect.TypeOf(actual)

			if !reflect.DeepEqual(actualType, expectedType) {
				t.Errorf("Unexpected parser. expected %v actual %v", expectedType, actualType)
			}
		})
	}
}
