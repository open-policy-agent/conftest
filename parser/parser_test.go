package parser

import (
	"reflect"
	"testing"

	"github.com/open-policy-agent/conftest/parser/docker"
	dotenv "github.com/open-policy-agent/conftest/parser/dotenv"
	"github.com/open-policy-agent/conftest/parser/hcl2"
	"github.com/open-policy-agent/conftest/parser/ignore"
	"github.com/open-policy-agent/conftest/parser/json"
	"github.com/open-policy-agent/conftest/parser/jsonc"
	"github.com/open-policy-agent/conftest/parser/textproto"
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
			"foo.Dockerfile",
			&docker.Parser{},
			false,
		},
		{
			"foo.dockerfile",
			&docker.Parser{},
			false,
		},
		{
			"Dockerfile.foo",
			&docker.Parser{},
			false,
		},
		{
			"dockerfile.foo",
			&docker.Parser{},
			false,
		},
		{
			"test.tf",
			&hcl2.Parser{},
			false,
		},
		{
			"test.tfvars",
			&hcl2.Parser{},
			false,
		},
		{
			"terragrunt.hcl",
			&hcl2.Parser{},
			false,
		},
		{
			"terragrunt.hcl.json",
			&json.Parser{},
			false,
		},
		{
			"terragrunt.hcl.jsonc",
			&jsonc.Parser{},
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
		{
			".env",
			&dotenv.Parser{},
			false,
		},
		{
			"prod.env",
			&dotenv.Parser{},
			false,
		},
		{
			".env.prod",
			&dotenv.Parser{},
			false,
		},
		{
			"foo.textproto",
			&textproto.Parser{},
			false,
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
