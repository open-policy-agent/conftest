package parser

import (
	"reflect"
	"testing"

	"github.com/open-policy-agent/conftest/parser/docker"
	"github.com/open-policy-agent/conftest/parser/hcl2"
	"github.com/open-policy-agent/conftest/parser/ignore"
	"github.com/open-policy-agent/conftest/parser/yaml"
)

func TestNewFromPath(t *testing.T) {
	testCases := []struct {
		path     string
		expected Parser
		wantErr  bool
	}{
		{
			"-",
			&yaml.Parser{},
			false,
		},
		{
			"test.yaml",
			&yaml.Parser{},
			false,
		},
		{
			"test.yml",
			&yaml.Parser{},
			false,
		},
		{
			"dockerfile",
			&docker.Parser{},
			false,
		},
		{
			"Dockerfile",
			&docker.Parser{},
			false,
		},
		{
			"test.tf",
			&hcl2.Parser{},
			false,
		},
		{
			"noextension",
			&yaml.Parser{},
			false,
		},
		{
			".gitignore",
			&ignore.Parser{},
			false,
		},
		{
			"file.unknown",
			nil,
			true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.path, func(t *testing.T) {
			expectedType := reflect.TypeOf(testCase.expected)

			actual, err := NewFromPath(testCase.path)
			if err != nil && !testCase.wantErr {
				t.Fatal("from path:", err)
			}

			actualType := reflect.TypeOf(actual)

			if !reflect.DeepEqual(actualType, expectedType) {
				t.Errorf("Unexpected parser. expected %v actual %v", expectedType, actualType)
			}
		})
	}
}
