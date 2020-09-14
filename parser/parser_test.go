package parser

import (
	"reflect"
	"testing"

	"github.com/open-policy-agent/conftest/parser/cue"
	"github.com/open-policy-agent/conftest/parser/docker"
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
		fileType string
		expected Parser
	}{
		{
			fileType: "hcl1",
			expected: new(hcl1.Parser),
		},
		{
			fileType: "tf",
			expected: new(hcl2.Parser),
		},
		{
			fileType: "hcl",
			expected: new(hcl2.Parser),
		},
		{
			fileType: "toml",
			expected: new(toml.Parser),
		},
		{
			fileType: "cue",
			expected: new(cue.Parser),
		},
		{
			fileType: "ini",
			expected: new(ini.Parser),
		},
		{
			fileType: "json",
			expected: new(json.Parser),
		},
		{
			fileType: "yaml",
			expected: new(yaml.Parser),
		},
		{
			fileType: "yml",
			expected: new(yaml.Parser),
		},
		{
			fileType: "edn",
			expected: new(edn.Parser),
		},
		{
			fileType: "Dockerfile",
			expected: new(docker.Parser),
		},
	}

	for _, testUnit := range testTable {
		t.Run(testUnit.fileType, func(t *testing.T) {
			actual, err := GetParser(testUnit.fileType)
			if err != nil {
				t.Fatal("get parser failed:", err)
			}

			if !reflect.DeepEqual(actual, testUnit.expected) {
				t.Errorf("Unexpected parser. expected %v actual %v", testUnit.expected, actual)
			}
		})
	}
}

func TestGetFileType(t *testing.T) {
	testTable := []struct {
		path     string
		expected string
	}{
		{"-", "yaml"},
		{"test.yaml", "yaml"},
		{"test.yml", "yml"},
		{"some/path/test.toml", "toml"},
		{"dockerfile", "dockerfile"},
	}

	for _, testUnit := range testTable {
		t.Run(testUnit.path, func(t *testing.T) {
			actual := getFileType(testUnit.path)

			if actual != testUnit.expected {
				t.Errorf("Unexpected filetype. expected %v actual %v", testUnit.expected, actual)
			}
		})
	}
}
