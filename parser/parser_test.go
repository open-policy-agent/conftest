package parser

import (
	"reflect"
	"testing"

	"github.com/open-policy-agent/conftest/parser/cue"
	"github.com/open-policy-agent/conftest/parser/edn"
	"github.com/open-policy-agent/conftest/parser/hcl1"
	"github.com/open-policy-agent/conftest/parser/hcl2"
	"github.com/open-policy-agent/conftest/parser/ini"
	"github.com/open-policy-agent/conftest/parser/json"
	"github.com/open-policy-agent/conftest/parser/toml"
	"github.com/open-policy-agent/conftest/parser/yaml"
)

func TestGetParser(t *testing.T) {
	testTable := []struct {
		name        string
		fileType    string
		expected    Parser
		expectError bool
	}{
		{
			name:        "Test getting Terraform parser from HCL1 input",
			fileType:    "hcl1",
			expected:    new(hcl1.Parser),
			expectError: false,
		},
		{
			name:        "Test getting HCL2 parser from .tf input",
			fileType:    "tf",
			expected:    new(hcl2.Parser),
			expectError: false,
		},
		{
			name:        "Test getting TOML parser",
			fileType:    "toml",
			expected:    new(toml.Parser),
			expectError: false,
		},
		{
			name:        "Test getting Cue parser",
			fileType:    "cue",
			expected:    new(cue.Parser),
			expectError: false,
		},
		{
			name:        "Test getting INI parser",
			fileType:    "ini",
			expected:    new(ini.Parser),
			expectError: false,
		},
		{
			name:        "Test getting JSON parser from JSON input",
			fileType:    "json",
			expected:    new(json.Parser),
			expectError: false,
		},
		{
			name:        "Test getting YAML parser from YAML input",
			fileType:    "yaml",
			expected:    new(yaml.Parser),
			expectError: false,
		},
		{
			name:        "Test getting EDN parser",
			fileType:    "edn",
			expected:    new(edn.Parser),
			expectError: false,
		},
		{
			name:        "Test getting invalid filetype",
			fileType:    "epicfailure",
			expected:    nil,
			expectError: true,
		},
	}

	for _, testUnit := range testTable {
		t.Run(testUnit.name, func(t *testing.T) {
			received, err := GetParser(testUnit.fileType)

			if !reflect.DeepEqual(received, testUnit.expected) {
				t.Errorf("expected: %T \n got this: %T", testUnit.expected, received)
			}
			if !testUnit.expectError && err != nil {
				t.Errorf("error here: %v", err)
			}
			if testUnit.expectError && err == nil {
				t.Error("error expected but not received")
			}
		})
	}
}
